# 🔧 User Extensions

This guide explains how to extend your service and integrate it into the runtime environment.

## Setting Up the Development Environment

First, initialize the [Go](https://go.dev) module and download the necessary dependencies.

```shell
go get github.com/siyul-park/uniflow
```

## Adding a New Node

To add new functionality, define the node specification and register the codec that converts it to a node in the schema.

The node specification implements the `spec.Spec` interface and can be defined using `spec.Meta`:

```go
type TextNodeSpec struct {
	spec.Meta `map:",inline"`
	Contents  string `map:"contents"`
}
```

Define the new node type:

```go
const KindText = "text"
```

Now, implement the node that will perform the actual functionality. Use the provided `OneToOneNode` template to receive an input packet, process it, and send an output packet:

```go
type TextNode struct {
	*node.OneToOneNode
	contents string
}
```

Define a function to create the node and set up the packet processing using the `OneToOneNode` constructor:

```go
func NewTextNode(contents string) *TextNode {
	n := &TextNode{contents: contents}
	n.OneToOneNode = node.NewOneToOneNode(n.action)

	return n
}
```

Implement the processing function, which will return the first value on success and the second value in case of an error:

```go
func (proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet)
```

In this example, the input packet is transformed into a packet containing the `contents` and sent:

```go
func (n *TextNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	outPck := packet.New(types.NewString(n.contents))
	return outPck, nil
}
```

Now, write tests to ensure the node specification works as intended. Send an input packet to the `in` port and verify that the output packet contains the `contents`:

```go
func TestTextNode_SendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	contents := faker.Word()

	n := NewTextNode(contents)
	defer n.Close()

	out := port.NewOut()
	out.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	outWriter := out.Open(proc)

	payload := types.NewString(faker.Word())
	pck := packet.New(payload)

	outWriter.Write(pck)

	select {
	case outPck := <-outWriter.Receive():
		assert.Equal(t, contents, outPck.Payload().Interface())
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}
```

### Registering the Schema and Codec

Create a codec to convert the node specification to a node and register it in the schema:

```go
func NewTextNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *TextNodeSpec) (node.Node, error) {
		return NewTextNode(spec.Contents), nil
	})
}

func AddToScheme() scheme.Register {
	return scheme.RegisterFunc(func(s *scheme.Scheme) error {
		s.AddKnownType(KindText, &TextNodeSpec{})
		s.AddCodec(KindText, NewTextNodeCodec())
		return nil
	})
}
```

Use the `scheme.Builder` to build the schema:

```go
builder := scheme.NewBuilder()
builder.Register(AddToScheme())

scheme, _ := builder.Build()
```

### Running the Runtime Environment

Pass this schema to the runtime environment to run a workflow that includes the extended node:

```go
r := runtime.New(runtime.Config{
	Namespace:   namespace,
	Schema:      scheme,
	Hook:        hook,
	SpecStore:   specStore,
	SecretStore: secretStore,
})
defer r.Close()

r.Listen(ctx)
```

## Service Integration

You can integrate the runtime environment into your service in two ways.

### Persistent Integration

Maintain a persistent runtime environment to quickly respond to external requests. Each runtime environment runs in an independent container, suitable for scenarios requiring continuous workflow execution.

```go
func main() {
	specStore := spec.NewStore()
	secretStore := secret.NewStore()

	sbuilder := scheme.NewBuilder()
	hbuilder := hook.NewBuilder()

	langs := language.NewModule()
	langs.Store(text.Language, text.NewCompiler())
	langs.Store(json.Language, json.NewCompiler())
	langs.Store(yaml.Language, yaml.NewCompiler())
	langs.Store(cel.Language, cel.NewCompiler())
	langs.Store(javascript.Language, javascript.NewCompiler())
	langs.Store(typescript.Language, typescript.NewCompiler())

	stable := system.NewTable()
	stable.Store(system.CodeCreateNodes, system.CreateNodes(specStore))
	stable.Store(system.CodeReadNodes, system.ReadNodes(specStore))
	stable.Store(system.CodeUpdateNodes, system.UpdateNodes(specStore))
	stable.Store(system.CodeDeleteNodes, system.DeleteNodes(specStore))

	broker := event.NewBroker()
	defer broker.Close()

	sbuilder.Register(control.AddToScheme(langs, cel.Language))
	sbuilder.Register(event.AddToScheme(broker, broker))
	sbuilder.Register(io.AddToScheme())
	sbuilder.Register(network.AddToScheme())
	sbuilder.Register(system.AddToScheme(stable))

	hbuilder.Register(control.AddToHook())
	hbuilder.Register(event.AddToHook(broker))
	hbuilder.Register(network.AddToHook())

	scheme, err := sbuilder.Build()
	if err != nil {
		log.Fatal(err)
	}
	hook, err := hbuilder.Build()
	if err != nil {
		log.Fatal(err)
	}

	r := runtime.New(runtime.Config{
		Namespace:	 "default",
		Scheme:		 scheme,
		Hook:		 hook,
		SpecStore:	 specStore,
		SecretStore: secretStore,
	})
	defer r.Close()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		_ = r.Close()
	}()

	r.Listen(context.TODO())
}
```

### Temporary Execution

Execute the workflow temporarily and then remove the execution environment. This method is suitable for running workflows in temporary environments:

```go
r := runtime.New(runtime.Config{
	Namespace: "default",
	Schema:    scheme,
	Hook:      hook,
	Store:     store,
})
defer r.Close()

symbols, _ := r.Load(ctx, &spec.Meta{
	Name: "main",
})

sym := symbols[0]

in := port.NewOut()
defer in.Close()

in.Link(sym.In(node.PortIn))

payload := types.NewString(faker.Word())
payload, err := port.Call(in, payload)
```
