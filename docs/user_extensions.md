# 🔧 User Extensions

This guide explains how to extend your service and integrate it into the runtime environment.

## Development Environment Setup

First, initialize the [Go](https://go.dev) module and download the necessary dependencies:

```shell
go get github.com/siyul-park/uniflow
```

## Adding a New Node

To add new functionality, define the node specification and register the codec that converts this specification into a node.

### Define Node Specification

A node specification implements the `spec.Spec` interface and can be defined using `spec.Meta`:

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

### Implement the Node

Create a node that performs the actual work. Use the `OneToOneNode` template to handle input packets, process them, and send output packets:

```go
type TextNode struct {
	*node.OneToOneNode
	contents string
}
```

Define a function to create the node and set up packet processing using the `OneToOneNode` constructor:

```go
func NewTextNode(contents string) *TextNode {
	n := &TextNode{contents: contents}
	n.OneToOneNode = node.NewOneToOneNode(n.action)

	return n
}
```

Implement the packet processing function. This function returns the first value on success and the second value on error:

```go
func (n *TextNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	outPck := packet.New(types.NewString(n.contents))
	return outPck, nil
}
```

### Testing the Node

Write a test to ensure the node functions correctly. Send an input packet to the `in` port and verify that the output packet contains the expected `contents`:

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

Create a codec that converts the node specification into a node and register it with the schema:

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

Use the `scheme.Builder` to construct the schema:

```go
builder := scheme.NewBuilder()
builder.Register(AddToScheme())

scheme, _ := builder.Build()
```

### Running the Runtime Environment

Pass the schema to the runtime environment to execute workflows that include the extended node:

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

There are two ways to integrate the runtime environment into your service:

### Continuous Integration

Maintain the runtime environment continuously to respond quickly to external requests. Each runtime environment runs in an independent container, suitable for scenarios requiring continuous workflow execution:

```go
func main() {
	specStore := spec.NewStore()
	secretStore := secret.NewStore()

	schemeBuilder := scheme.NewBuilder()
	hookBuilder := hook.NewBuilder()

	scheme, err := schemeBuilder.Build()
	if err != nil {
		log.Fatal(err)
	}
	hook, err := hookBuilder.Build()
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
	signal.Notify(sigs, native.SIGINT, native.SIGTERM)

	go func() {
		<-sigs
		_ = r.Close()
	}()

	r.Listen(context.TODO())
}
```

### Temporary Execution

Run the workflow temporarily and then dispose of the execution environment. This method is suitable for running workflows in temporary environments:

```go
r := runtime.New(runtime.Config{
	Namespace:   "default",
	Schema:      scheme,
	Hook:        hook,
	SpecStore:   specStore,
	SecretStore: secretStore,
})
defer r.Close()

r.Load(ctx) // Load all

symbols, _ := r.Load(ctx, &spec.Meta{
	Name: "main",
})

sb := symbols[0]

in := port.NewOut()
defer in.Close()

in.Link(sb.In(node.PortIn))

payload := types.NewString(faker.Word())
payload, err := port.Call(in, payload)
```
