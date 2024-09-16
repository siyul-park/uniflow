# 🔧 User Extensions

While most functionalities can already createable through combinations of uniflow nodes, there may occasionally be situations where you need to implement special or additional features directly. This guide explains how to create and apply such features.

Before reading this guide, it's recommended to read the [Key Concepts](https://github.com/siyul-park/uniflow/blob/main/docs/key_concepts.md) and [System Architecture](https://github.com/siyul-park/uniflow/blob/main/docs/architecture.md).

## Setting Up the Development Environment

First, initialize the [Go](https://go.dev) module and download the necessary dependencies:

```shell
go get github.com/siyul-park/uniflow
```

## Writing a Workflow

To connect a new node to a workflow, the node should be implemented first.

but for now, (for ease of explanation) we'll show the final completed form to give an overall picture. Here's a simple workflow that implements a proxy feature to load balance HTTP requests:

```yaml
- kind: listener
  name: listener
  protocol: http
  port: 8000
  ports:
    out:
      - name: proxy
        port: in

# This is the proxy node we'll create
- kind: proxy
  name: proxy
  urls:
    - https://backend1.com/
    - https://backend2.com/
```

When an HTTP request is made to port 8000, the proxy will automatically select one of the given URLs to process the data. Now, we need to create this proxy node.

## Creating a New Node

To create a node, we generally go through three steps: **Define structure and type -> Define action function -> Define creation function**. After defining the node's specifications and deciding on a name for the node type (called 'kind'), implement the action function that defines what the node will do, and create a function to generate this node.

we can call these steps as 'creating a node specification'. After this, create a codec that converts this specification into an actual functioning node, register it with the schema, and finally connect it to the runtime environment.

### Defining Node Specifications

The node specification should conform to the `spec.Spec` interface. The following items are required:

```go
ID uuid.UUID // Unique identifier in UUID format
Kind string // Specifies the type of node
Namespace string // Specifies the namespace the node belongs to
Name string // Specifies the name of the node, which must be unique within the same namespace
Annotations map[string]string // Additional metadata for the node
Ports map[string][]Port // Defines how ports are connected
Env map[string][]Secret // Specifies environment variables needed for the node
```

You can use `spec.Meta` to write this simply:

```go
type ProxyNodeSpec struct {
    spec.Meta `map:",inline"`
}
```

#### Adding Additional Fields

If you need to receive additional values, you can add extra fields. Assuming we need URL information to create the proxy functionality, we can declare a URLS field like this:

```go
type ProxyNodeSpec struct {
    spec.Meta `map:",inline"`
    URLS []string `map:"urls"`
}
```

This field can be used as an additional field when using the node in a workflow later, and can be accepted as initial configuration values such as environment variables.

```yaml
- kind: proxy
  name: proxy
  urls:
    - https://backend1.com/
    - https://backend2.com/
```

### Defining Node Type

Now we define the node type. This type must be accurately written for the runtime to correctly recognize the node:

```go
const KindProxy = "proxy"
```

### Defining Node Structure

Based on the node specification, now we need to define the elements actually needed for the node to function. In simple terms, it should contain information about how the node will communicate and what data it will hold:

```go
type ProxyNode struct {
    *node.OneToOneNode
    proxy *httputil.ReverseProxy
}
```

For nodes to communicate, a communication standard must be defined. uniflow supports `ZeroToOne`, `OneToOne`, `OneToMany`, `ManyToOne`, and `Other` standards. The `OneToOneNode` template used here supports a 1:1 structure and helps easily implement nodes that receive packets from the input port, process them, and transfer them to the output port.

### Implementing Node Action

Now we implement the process of the node handling input packets and generating output packets as a result. Packets contain payloads, and these payloads are represented by one of several common data types that implement the `types.Value` interface.

```go
// Value is an interface representing atomic data types.
type Value interface {
    Kind() Kind // Kind returns the type of the Value.
    Hash() uint64 // Hash returns the hash code of the Value.
    Interface() any // Interface returns the Value as a general interface.
    Equal(other Value) bool // Equal checks if this Value is the same as another Value.
    Compare(other Value) int // Compare compares this Value with another Value.
}
```

Let's continue implementing the example we used earlier. To create a proxy function, we need a structure that can change the URL according to a predetermined order and request to the server with the received packet. Since the incoming packet is a structure that directly requests resources from the server, we need to create a format that requests with the packet data and receives the response value.

### Testing the Node
Write a test to ensure the node functions correctly. Send an input packet to the `in` port and verify that the output packet contains the expected `contents`:
    req := HTTPPayload{}
    if err := types.Unmarshal(inPck.Payload(), &req); err != nil {
        return nil, packet.New(types.NewError(err))
    }

    // Get the body data
    buf := bytes.NewBuffer(nil)
    if err := mime.Encode(buf, req.Body, textproto.MIMEHeader(req.Header)); err != nil {
        return nil, packet.New(types.NewError(err))
    }

    // Now create Request data to use in the proxy environment based on this value
    r := &http.Request{
        Method: req.Method,
        URL: &url.URL{
            Scheme:   req.Scheme,
            Host:     req.Host,
            Path:     req.Path,
            RawQuery: req.Query.Encode(),
        },
        Proto:  req.Protocol,
        Header: req.Header,
        Body:   io.NopCloser(buf),
    }

    // Perform HTTP request through proxy
    w := httptest.NewRecorder()
    n.proxy.ServeHTTP(w, r)

    // Get the result body
    body, err := mime.Decode(w.Body, textproto.MIMEHeader(w.Header()))
    if err != nil {
        return nil, packet.New(types.NewError(err))
    }

    // Create result payload
    res := &HTTPPayload{
        Header: w.Header(),
        Body:   body,
        Status: w.Code,
    }

    // Now create and return the packet to send this result
    outPayload, err := types.Encoder.Encode(res)
    if err != nil {
        return nil, packet.New(types.NewError(err))
    }
    return packet.New(outPayload), nil
}
```

> You might think of other way, such as modifying header values without reconstructing the incoming request. In fact, the `proc` object from the http listener already contains `http.ResponseWriter` and `*http.Request` values that can be retrieved, but tampering with these values carelessly could make it impossible to pre- and post-process requests and responses. Unless necessary, it's best not to tamper with the process structure.

Create a codec that converts the node specification into a node and register it with the schema:

```go
func NewProxyNode(urls []*url.URL) *ProxyNode {
    var index int
    var mu sync.Mutex

    proxy := &httputil.ReverseProxy{
        Rewrite: func(pr *httputil.ProxyRequest) {
            mu.Lock()
            defer mu.Unlock()
            index = (index + 1) % len(urls)
            pr.SetURL(urls[index])
            pr.SetXForwarded()
        },
    }

    n := &ProxyNode{proxy: proxy}
    n.OneToOneNode = node.NewOneToOneNode(n.action)
    return n
}
```

In this example, any additional features are not included such as health-check. (for explanation purposes) but you can add various features that can improve completeness.

### Writing Tests

Write tests to verify that the node works as intended. Send an input packet to the `in` port and verify that the output packet comes out as expected from the `out` port:

```go
func TestProxyNode_SendAndReceive(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
    defer cancel()

    s1 := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
        writer.WriteHeader(http.StatusOK)
        writer.Write([]byte("Backend 1"))
    }))
    defer s1.Close()

    s2 := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
        writer.WriteHeader(http.StatusOK)
        writer.Write([]byte("Backend 2"))
    }))
    defer s2.Close()

    u1, _ := url.Parse(s1.URL)
    u2, _ := url.Parse(s2.URL)

    n := NewProxyNode([]*url.URL{u1, u2})
    defer n.Close()

    in := port.NewOut()
    in.Link(n.In(node.PortIn))

    proc := process.New()
    defer proc.Exit(nil)

    inWriter := in.Open(proc)

    inPayload := types.NewMap(
        types.NewString("method"), types.NewString(http.MethodGet),
        types.NewString("scheme"), types.NewString("http"),
        types.NewString("host"), types.NewString("test"),
        types.NewString("path"), types.NewString("/"),
        types.NewString("protocol"), types.NewString("HTTP/1.1"),
        types.NewString("status"), types.NewInt(0),
    )
    inPck := packet.New(inPayload)
    inWriter.Write(inPck)

    select {
    case outPck := <-inWriter.Receive():
        payload := &HTTPPayload{}
        err := types.Unmarshal(outPck.Payload(), payload)
        assert.NoError(t, err)
        assert.Contains(t, payload.Body.Interface(), "Backend")
    case <-ctx.Done():
        assert.Fail(t, ctx.Err().Error())
    }
}
```

If the assert succeeds, you can confirm that it can perform a complete function as a single node.

## Connecting to Runtime

The content explained so far was about how to create a single node. Now, to connect the node to the system, we need to create a codec, connect it to the schema, and then the node can perform the desired task when executed.

### Creating a Codec

Write a codec that converts the node specification into a node:

```go
func NewProxyNodeCodec() scheme.Codec {
    return scheme.CodecWithType(func(spec *ProxyNodeSpec) (node.Node, error) {
        urls := make([]*url.URL, 0, len(spec.URLS))
        if len(urls) == 0 {
            return nil, errors.WithStack(encoding.ErrUnsupportedValue)
        }
        for _, u := range spec.URLS {
            parsed, err := url.Parse(u)
            if err != nil {
                return nil, err
            }
            urls = append(urls, parsed)
        }
        return NewProxyNode(urls), nil
    })
}
```

### Creating and Adding Schema

Create a schema generation function so that the specification and type used when creating the node can be recognized and used externally:

```go
func AddToScheme() scheme.Register {
    return scheme.RegisterFunc(func(s *scheme.Scheme) error {
        ...
        s.AddKnownType(KindProxy, &ProxyNodeSpec{})
        s.AddCodec(KindProxy, NewProxyNodeCodec())
        ...
        return nil
    })
}
```

Pass this defined `scheme.Register` to `scheme.Builder` to create a schema:

```go
builder := scheme.NewBuilder()
builder.Register(AddToScheme())
scheme, _ := builder.Build()
```

### Running the Runtime Environment

Now, if you pass this schema to the runtime environment, you can run a workflow that includes the node you created directly:

```go
r := runtime.New(runtime.Config{
    Namespace:   namespace,
    Schema:      scheme,
    Hook:        hook,
    SpecStore:   specStore,
    SecretStore: secretStore,
})
defer r.Close()
```

All data about the workflow is now in `r`, so all preparations are complete if you run this variable according to your purpose.

## Integrating with Existing Services

Now you need to add this runtime environment to your existing service, rebuild it, and create an executable file. There are two ways to add services: continuous execution where the runtime keeps running and operating, and simple execution that runs once and ends.

### Continuous Execution

Continuously maintaining the runtime environment allows for immediate response to external requests. Each runtime environment runs in an independent container and is suitable for scenarios requiring continuous workflow execution.

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

### Simple Execution

While there are advantages to continuously maintaining the runtime environment, you might want it to operate only when needed or to operate simply. In this case, you can configure it for simple execution.

```go
r := runtime.New(runtime.Config{
    Namespace:   "default",
    Schema:      scheme,
    Hook:        hook,
    SpecStore:   specStore,
    SecretStore: secretStore,
})
defer r.Close()

r.Load(ctx) // Load all resources

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
