# 🔧 Extending User Functionality

This guide explains how users can extend their services and integrate them into the runtime environment.

## Setting Up the Development Environment

First, initialize the [Go](https://go.dev) module and download the required dependencies:

```shell
go get github.com/siyul-park/uniflow
```

## Adding a New Node

To add new functionality, define the node specification and register a codec to convert it into a node.

Node specifications implement the `spec.Spec` interface and can be defined using `spec.Meta`:

```go
type TextNodeSpec struct {
	spec.Meta `map:",inline"`
	Contents  string `map:"contents"`
}
```

Define a new node type:

```go
const KindText = "text"
```

Implement the actual node functionality. Use the `OneToOneNode` template provided to receive input packets, process them, and send output packets:

```go
type TextNode struct {
	*node.OneToOneNode
	contents string
}
```

Define a function to create the node, passing the packet processing method to the `OneToOneNode` constructor:

```go
func NewTextNode(contents string) *TextNode {
	n := &TextNode{contents: contents}
	n.OneToOneNode = node.NewOneToOneNode(n.action)

	return n
}
```

Implement the packet processing function. This function uses the first return value for successful processing and the second return value for errors:

```go
func (n *TextNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	outPck := packet.New(types.NewString(n.contents))
	return outPck, nil
}
```

Write a test to verify that the node functions as intended. Send an input packet to the `in` port and check that the output packet contains the `contents`:

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

### Registering Schema and Codec

Create a codec to convert the node specification into a node and register it with the schema:

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

Use `scheme.Builder` to build the schema:

```go
builder := scheme.NewBuilder()
builder.Register(AddToScheme())

scheme, _ := builder.Build()
```

### Running the Runtime Environment

Now pass the schema to the runtime environment to execute workflows containing the extended node:

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

## Integrating with Services

There are two ways to integrate the runtime environment with a service:

### Continuous Integration

Maintain the runtime environment continuously for rapid responses to external requests. Each runtime environment runs in an independent container and is suitable for scenarios requiring ongoing workflow execution.

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

	sbuilder.Register(control.AddToScheme(langs, cel.Language))
	sbuilder.Register(io.AddToScheme())
	sbuilder.Register(network.AddToScheme())
	sbuilder.Register(system.AddToScheme(stable))

	hbuilder.Register(control.AddToHook())
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
		Namespace:   "default",
		Schema:      scheme,
		Hook:        hook,
		SpecStore:   specStore,
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

Run workflows temporarily and remove the execution environment. This method is suitable for executing workflows in transient environments:

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
