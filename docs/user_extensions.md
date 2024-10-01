# 🔧 User Extensions

While most functionalities can be implemented by combining existing nodes, sometimes new features may be required. In such cases, you can extend functionality by adding new nodes at runtime.

Before reading this guide, it is recommended to review the [Key Concepts](./key_concepts.md) and [Architecture](./architecture.md).

## Setting Up the Development Environment

Initialize the Go module and install necessary dependencies.

```shell
go get github.com/siyul-park/uniflow
```

## Creating a Workflow

Here is a simple example of a workflow that provides proxy functionality. This workflow receives HTTP requests and performs load balancing across multiple backend servers.

```yaml
- kind: listener
  name: listener
  protocol: http
  port: 8000
  ports:
    out:
      - name: proxy
        port: in

- kind: proxy
  name: proxy
  urls:
    - https://backend1.com/
    - https://backend2.com/
```

In this YAML configuration, HTTP requests coming into port 8000 are handled by the proxy node, which selects one of the backend servers specified in the `urls` for processing the request.

## Adding a Node

To support a new node, you need to define the node specification, implement the node, and then connect the node at runtime.

Define the node's specification and type (kind), implement the node's behavior function, and create a function to instantiate the node. After that, create a codec to convert the specification into an operational node and register it with the schema for runtime integration.

### Define Node Specification

Node specifications should conform to the `spec.Spec` interface. The following fields are required:

```go
ID uuid.UUID // Unique identifier in UUID format.
Kind string // Specifies the type of the node.
Namespace string // Specifies the namespace the node belongs to.
Name string // Specifies the name of the node, which must be unique within the same namespace.
Annotations map[string]string // Additional metadata about the node.
Ports map[string][]Port // Defines the port connections.
Env map[string][]Secret // Specifies environment variables required by the node.
```

You can simplify this with `spec.Meta`:

```go
type ProxyNodeSpec struct {
	spec.Meta `map:",inline"`
	URLs      []string `map:"urls"`
}
```

The specification includes fields like UUID, node kind, namespace, and additional settings such as `URLs`.

```yaml
- kind: proxy
  name: proxy
  urls:
    - https://backend1.com/
    - https://backend2.com/
```

### Define Node Type

Define the node type so that it can be recognized at runtime. Here is the definition for the proxy node type.

```go
const KindProxy = "proxy"
```

### Define Node Implementation

Based on the node specification, define the actual behavior of the node. This should include details on how the node communicates and processes data:

```go
type ProxyNode struct {
	*node.OneToOneNode
	proxy *httputil.ReverseProxy
}
```

Next, select the communication specification for the node. Supported specifications include `ZeroToOne`, `OneToOne`, `OneToMany`, `ManyToOne`, and `Other`.

The `OneToOneNode` template supports 1:1 structure, which simplifies the implementation of nodes that receive packets from an input port and directly pass them to an output port.

Implement the process for handling input packets and generating output packets. Packets contain payloads, which are represented by one of the public data types implementing the `types.Value` interface.

```go
// Value is an interface representing atomic data types.
type Value interface {
	Kind() Kind              // Kind returns the type of the Value.
	Hash() uint64            // Hash returns the hash code of the Value.
	Interface() any          // Interface returns the Value as a general interface.
	Equal(other Value) bool  // Equal checks if this Value is equal to another Value.
	Compare(other Value) int // Compare compares this Value with another Value.
}
```

To implement the proxy functionality, the node should be able to modify URLs according to a predetermined order and send requests to the server. The implementation should handle direct requests using packet data and process the responses.

```go
func (n *ProxyNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	req := &HTTPPayload{}
	if err := types.Unmarshal(inPck.Payload(), req); err != nil {
		return nil, packet.New(types.NewError(err))
	}

	buf := bytes.NewBuffer(nil)
	if err := mime.Encode(buf, req.Body, textproto.MIMEHeader(req.Header)); err != nil {
		return nil, packet.New(types.NewError(err))
	}

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
	w := httptest.NewRecorder()

	n.proxy.ServeHTTP(w, r)

	body, err := mime.Decode(w.Body, textproto.MIMEHeader(w.Header()))
	if err != nil {
		return nil, packet.New(types.NewError(err))
	}

	res := &HTTPPayload{
		Method:   req.Method,
		Scheme:   req.Scheme,
		Host:     req.Host,
		Path:     req.Path,
		Query:    req.Query,
		Protocol: req.Protocol,
		Header:   w.Header(),
		Body:     body,
		Status:   w.Code,
	}

	outPayload, err := types.Marshal(res)
	if err != nil {
		return nil, packet.New(types.NewError(err))
	}
	return packet.New(outPayload), nil
}
```

Finally, define a function to create and configure the node for actual operation.

```go
func NewProxyNode(urls []*url.URL) *ProxyNode {
	var index int
	var mu sync.Mutex

	transport := &http.Transport{}
	http2.ConfigureTransport(transport)

	proxy := &httputil.ReverseProxy{
		Transport: transport,
		Rewrite: func(r *httputil.ProxyRequest) {
			mu.Lock()
			defer mu.Unlock()

			index = (index + 1) % len(urls)

			r.SetURL(urls[index])
			r.SetXForwarded()
		},
	}

	n := &ProxyNode{proxy: proxy}
	n.OneToOneNode = node.NewOneToOneNode(n.action)
	return n
}
```

### Writing Tests

Write tests to ensure the node operates as expected. Send an input packet through the `in` port and verify that the output packet from the `out` port meets expectations.

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
		types.NewString("query"), types.NewMap(),
		types.NewString("protocol"), types.NewString("HTTP/1.1"),
		types.NewString("header"), types.NewMap(),
		types.NewString("body"), types.NewBytes([]byte("")),
	)
	inWriter.Send(packet.New(inPayload))

	pck, err := inWriter.Receive(ctx)
	if err != nil {
		t.Fatal(err)
	}

	payload := &HTTPPayload{}
	if err := types.Unmarshal(pck.Payload(), payload); err != nil {
		t.Fatal(err)
	}

	if payload.Status != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, payload.Status)
	}
}
```

## Runtime Integration

Now that you have created a node and schema, you need to integrate them into the runtime environment. This process involves creating a codec to convert node specifications into actual node objects, registering the schema, and setting up the runtime environment for execution.

### Creating the Codec

The codec is responsible for converting node specifications into actual node objects. Here's how you can create a codec for the `ProxyNode`:

```go
func NewProxyNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *ProxyNodeSpec) (node.Node, error) {
		urls := make([]*url.URL, 0, len(spec.URLs))
		if len(spec.URLs) == 0 {
			return nil, errors.WithStack(encoding.ErrUnsupportedValue)
		}

		for _, u := range spec.URLs {
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

This codec function takes a `ProxyNodeSpec` specification, parses the URLs, and creates a `ProxyNode` using the `NewProxyNode` function. It returns an error if something goes wrong during this process.

### Creating and Adding the Schema

To make your node type recognizable by the system, you need to create and register a schema. This step allows the system to identify and use your new node type.

```go
func AddToScheme() scheme.Register {
	return scheme.RegisterFunc(func(s *scheme.Scheme) error {
		s.AddKnownType(KindProxy, &ProxyNodeSpec{})
		s.AddCodec(KindProxy, NewProxyNodeCodec())
		return nil
	})
}
```

The `AddToScheme` function registers the `ProxyNodeSpec` and its codec with the schema, allowing the system to recognize and work with the `KindProxy` node type.

To actually build the schema, you need to pass the `scheme.Register` to a `scheme.Builder` and build it:

```go
builder := scheme.NewBuilder()
builder.Register(AddToScheme())

scheme, _ := builder.Build()
```

### Running the Runtime Environment

With the schema created and registered, you can now set up the runtime environment and run workflows that include your new node type. Initialize the runtime environment with the schema and other required components:

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

This code creates a new runtime environment using the provided schema, hook, specification store, and secret store. The `defer` statement ensures that resources are cleaned up when done.

## Integration with Existing Services

To integrate the runtime environment with existing services and build an executable, you need to set up the environment to run continuously or in a simpler execution mode.

```go
func main() {
	ctx := context.TODO()

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
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		_ = r.Close()
	}()

	r.Watch(ctx)
	r.Load(ctx)
	r.Reconcile(ctx)
}
```

This code keeps the runtime environment running and responsive to external signals. It uses `os.Signal` to listen for termination signals and safely shuts down the environment.
