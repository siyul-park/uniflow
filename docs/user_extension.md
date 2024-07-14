# 🔧 User Extension

This guide explains how to extend your services and integrate them into the runtime environment.

## Setting Up the Development Environment

First, initialize the [Go](https://go.dev) module and download the necessary dependencies:

```shell
go get github.com/siyul-park/uniflow
```

## Adding a New Node

To introduce new functionality, define a node specification and register a codec to convert this specification into a node.

Define the node specification using `spec.Meta`:

```go
type TextNodeSpec struct {
	spec.Meta `map:",inline"`
	Contents  string `map:"contents"`
}
```

Specify the new node type:

```go
const KindText = "text"
```

Implement a node that performs the desired functionality. Use the `OneToOneNode` template, which receives input packets, processes them, and emits output packets:

```go
type TextNode struct {
	*node.OneToOneNode
	contents string
}
```

Create a function to instantiate the node, configuring packet processing by passing a handler function to the `OneToOneNode` constructor:

```go
func NewTextNode(contents string) *TextNode {
	n := &TextNode{contents: contents}
	n.OneToOneNode = node.NewOneToOneNode(n.action)

	return n
}
```

Implement the function with the following signature to handle normal operation and error processing:

```go
func (proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet)
```

Convert the input packet into a new packet containing the `contents` and send it out:

```go
func (n *TextNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	outPck := packet.New(types.NewString(n.contents))
	return outPck, nil
}
```

To ensure the node specification functions correctly, write tests. Send an input packet to the `in` port and verify that the output packet contains the expected `contents`:

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

Create a codec to convert the node specification to a node and register it with the scheme:

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

Then, use `scheme.Builder` to construct the scheme:

```go
builder := scheme.NewBuilder()
builder.Register(AddToScheme())

scheme, _ := builder.Build()
```

### Running the Runtime Environment

Pass this scheme to the runtime environment to execute workflows with the extended node:

```go
r := runtime.New(runtime.Config{
	Namespace: namespace,
	Schema:    scheme,
	Hook:      hook,
	Store:     store,
})
defer r.Close()

r.Listen(ctx)
```

## Integrating with Your Service

Integrate the runtime environment into your service in two ways:

### Continuous Integration

Maintain the runtime environment to respond quickly to external requests. Each runtime environment runs in an independent container, suitable for scenarios requiring continuous workflow execution. Configure the runtime environment with built-in extensions:

```go
func main() {
	col := memdb.NewCollection("")
	store := store.New(col)

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
	stable.Store(system.CodeCreateNodes, system.CreateNodes(store))
	stable.Store(system.CodeReadNodes, system.ReadNodes(store))
	stable.Store(system.CodeUpdateNodes, system.UpdateNodes(store))
	stable.Store(system.CodeDeleteNodes, system.DeleteNodes(store))

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
		Namespace: "default",
		Scheme:    scheme,
		Hook:      hook,
		Store:     store,
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

Execute a workflow temporarily and then remove the execution environment. This approach is suitable for running workflows in temporary environments:

```go
r := runtime.New(runtime.Config{
	Namespace: "default",
	Schema:    scheme,
	Hook:      hook,
	Store:     store,
})
defer r.Close()

sym, _ := r.LookupByName(ctx, "main")

in := port.NewOut()
defer in.Close()

in.Link(sym.In(node.PortIn))

proc := process.New()
defer proc.Exit(nil)

inWriter := in.Open(proc)

outPayload := types.NewString(faker.Word())
outPck := packet.New(outPayload)

backPck := packet.Call(inWriter, outPck)
```
