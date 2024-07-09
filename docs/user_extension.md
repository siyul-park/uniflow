# User Extension

This guide provides detailed instructions on how users can extend their services and integrate them into the runtime environment.

## Setting Up Development Environment

First, initialize the [Go](https://go.dev) module and download necessary dependencies.

```shell
go get github.com/siyul-park/uniflow
```

## Adding a New Node

To support new functionalities, define a node specification and register a codec in the scheme to convert it into a node.

Implement the node specification, which implements the `spec.Spec` interface, and define it simply using `spec.Meta`:

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

Now, implement a functional node. Utilize the `OneToOneNode` provided as a base template, which receives input packets, processes them, and sends output packets:

```go
type TextNode struct {
	*node.OneToOneNode
	contents string
}
```

Create a function to instantiate the node. Configure the packet processing method through the constructor of `OneToOneNode`:

```go
func NewTextNode(contents string) *TextNode {
	n := &TextNode{contents: contents}
	n.OneToOneNode = node.NewOneToOneNode(n.action)

	return n
}
```

Implement a function with the following specification. Use the first return value upon successful processing and the second in case of errors:

```go
func (proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet)
```

Here, convert the input packet into a packet containing `contents` and transmit it:

```go
func (n *TextNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	outPck := packet.New(types.NewString(n.contents))
	return outPck, nil
}
```

Next, write tests to verify that the node specification operates as intended. Transmit input packets to the `in` port and ensure output packets contain the `contents`:

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

Write a codec to convert the node specification into a node and register it in the scheme:

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

Use `scheme.Builder` to configure the scheme:

```go
builder := scheme.NewBuilder()
builder.Register(AddToScheme())

scheme, _ := builder.Build()
```

Now, pass this scheme to the runtime environment to execute workflows that include the extended node:

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

## Service Integration

Integrate the runtime environment into your service using two approaches:

**Continuous Integration**: Maintain the runtime environment continuously to promptly respond to external requests. Each runtime environment runs in independent containers, suitable for scenarios requiring continuous workflow execution.

Explore the [built-in extension](../ext/README.md) to configure the runtime environment:

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

**Temporary Execution**: Execute workflows intermittently and remove the runtime afterward. This method uses the embedded runtime environment, suitable for environments where workflows are executed temporarily:

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