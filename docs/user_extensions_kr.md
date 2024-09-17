# 🔧 사용자 확장

기존 노드들을 조합하여 대부분의 기능을 구현할 수 있지만, 때때로 새로운 기능이 필요할 수 있습니다. 이럴 때는 런타임에 새로운 노드를 추가하여 기능을 확장할 수 있습니다.

이 가이드를 읽기 전에 [핵심 개념](./key_concepts_kr.md)과 [아키텍처](./architecture_kr.md)를 참고하는 것이 좋습니다.

## 개발 환경 설정

Go 모듈을 초기화하고 필요한 의존성을 설치합니다.

```shell
go get github.com/siyul-park/uniflow
```

## 워크플로우 작성

다음은 프록시 기능을 제공하는 간단한 워크플로우 예시입니다. 이 워크플로우는 HTTP 요청을 받아 여러 백엔드 서버로 로드 밸런싱을 수행합니다.

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

이 YAML 구성에서는 HTTP 요청이 8000번 포트로 들어오면, 프록시 노드가 `urls`에 명시된 백엔드 서버 중 하나를 선택하여 요청을 처리합니다.

## 노드 추가

새로운 노드를 지원하려면 노드 명세를 정의하고, 노드를 구현한 후, 런타임에 노드를 연결해야 합니다.

노드의 스펙을 정의하고, 노드 유형(kind)을 지정한 후, 노드의 동작 함수를 구현하고 노드를 생성하는 함수를 작성하면 기본적인 노드 구성이 완료됩니다. 이후, 이 명세를 실제 동작하는 노드로 변환해주는 코덱을 만들고 스키마에 등록하여 런타임 환경과 연결합니다.

### 노드 명세 정의

노드 명세는 `spec.Spec` 인터페이스에 맞춰 구성되어야 합니다. 다음 항목이 필요합니다:

```go
ID uuid.UUID // UUID 형식의 고유 식별자입니다.
Kind string // 노드의 종류를 지정합니다.
Namespace string // 노드가 속한 네임스페이스를 지정합니다.
Name string // 노드의 이름을 지정하며, 동일한 네임스페이스 내에서 고유해야 합니다.
Annotations map[string]string // 노드에 대한 추가 메타데이터입니다.
Ports map[string][]Port // 포트의 연결 방식을 정의합니다.
Env map[string][]Secret // 노드에 필요한 환경 변수를 지정합니다.
```

`spec.Meta`를 사용하면 간단하게 작성할 수 있습니다:

```go
type ProxyNodeSpec struct {
	spec.Meta `map:",inline"`
	URLs      []string `map:"urls"`
}
```

명세는 `spec.Meta` 필드를 포함하여 UUID, 노드 종류(kind), 네임스페이스 등을 정의하며, `URLs`와 같은 추가 설정 값을 포함할 수 있습니다.

```yaml
- kind: proxy
  name: proxy
  urls:
    - https://backend1.com/
    - https://backend2.com/
```

### 노드 유형 정의

노드의 유형을 정의하여 런타임에서 인식할 수 있도록 합니다. 아래는 프록시 노드의 유형 정의입니다.

```go
const KindProxy = "proxy"
```

### 노드 정의

노드 명세를 기반으로 실제 동작을 정의합니다. 노드가 어떻게 통신하고 어떤 데이터를 처리할 것인지에 대한 정보를 담고 있어야 합니다:

```go
type ProxyNode struct {
  *node.OneToOneNode
  proxy *httputil.ReverseProxy
}
```

그 후, 노드 간 통신 규격을 선택해야 합니다. `ZeroToOne`, `OneToOne`, `OneToMany`, `ManyToOne`, `Other` 규격을 지원합니다.

`OneToOneNode` 템플릿은 1:1 구조를 지원하며, 입력 포트에서 패킷을 받아 처리한 후 출력 포트로 바로 전달하는 노드를 쉽게 구현할 수 있도록 돕습니다.

이제 노드가 입력 패킷을 처리하고 결과를 출력 패킷으로 생성하는 과정을 구현합니다. 패킷은 페이로드를 담고 있으며, 페이로드는 `types.Value` 인터페이스를 구현하는 공용 데이터 타입 중 하나로 표현됩니다.

```go
// Value는 원자적 데이터 타입을 표현하는 인터페이스입니다.
type Value interface {
  Kind() Kind              // Kind는 Value의 타입을 반환합니다.
  Hash() uint64            // Hash는 Value의 해시 코드를 반환합니다.
  Interface() any          // Interface는 Value를 일반 인터페이스로 반환합니다.
  Equal(other Value) bool  // Equal은 이 Value와 다른 Value가 같은지를 확인합니다.
  Compare(other Value) int // Compare는 이 Value와 다른 Value를 비교합니다.
}
```

프록시 기능을 구현하기 위해, 받은 패킷을 정해진 순서에 맞춰 URL을 변경하여 서버에 요청할 수 있는 구조를 필요로 합니다. 패킷 데이터만을 사용해 직접 요청하고 응답값을 처리하는 형태로 만들어야 합니다.

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

마지막으로, 노드를 생성하는 함수를 정의하여 실제로 노드를 생성하고 동작을 처리할 수 있도록 설정합니다.

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

### 테스트 작성

노드가 의도대로 작동하는지 확인하기 위해 테스트를 작성합니다. 입력 패킷을 `in` 포트로 전송하고, `out` 포트에서 출력 패킷이 예상대로 나오는지 검증합니다.

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

## 런타임 연결

이제 시스템에 노드를 연결하려면 코덱을 생성하고, 스키마와 연결해야 합니다. 이 과정이 완료되면 노드가 실행 시 올바르게 동작하게 됩니다.

### 코덱 생성

먼저, 노드 명세를 실제 노드 객체로 변환하는 코덱을 작성해야 합니다. 이를 통해 명세를 기반으로 노드를 생성할 수 있습니다.

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

이 코덱 함수는 `ProxyNodeSpec` 명세를 입력으로 받아, URL들을 파싱한 후 `NewProxyNode` 함수를 통해 노드를 생성합니다. 이 과정에서 오류가 발생하면 적절한 오류 메시지를 반환합니다.

### 스키마 생성 및 추가

이제 노드 명세와 유형을 외부에서 인식할 수 있도록 스키마를 생성하고 등록하는 함수를 만듭니다. 이렇게 하면 시스템이 새로운 노드 타입을 인식하고 사용할 수 있게 됩니다.

```go
func AddToScheme() scheme.Register {
	return scheme.RegisterFunc(func(s *scheme.Scheme) error {
		s.AddKnownType(KindProxy, &ProxyNodeSpec{})
		s.AddCodec(KindProxy, NewProxyNodeCodec())
		return nil
	})
}
```

위 함수는 `KindProxy`와 연관된 `ProxyNodeSpec`과 `NewProxyNodeCodec`를 스키마에 추가합니다. 이를 통해 새로운 노드 유형이 시스템에 등록됩니다.

스키마를 실제로 생성하려면, `scheme.Register`를 `scheme.Builder`에 전달하여 빌드합니다.

```go
builder := scheme.NewBuilder()
builder.Register(AddToScheme())

scheme, _ := builder.Build()
```

### 런타임 환경 실행

이제 생성한 스키마를 런타임 환경에 전달하여 만든 노드가 포함된 워크플로우를 실행할 수 있습니다. 이 과정에서는 런타임 환경을 설정하고 초기화합니다.

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

위 코드에서는 `runtime.New`를 사용하여 새로운 런타임 환경을 생성하고, 필요한 모든 구성 요소를 설정합니다. `defer`를 사용하여 종료 시 리소스를 정리합니다.

## 기존 서비스와 통합

이제 만든 런타임 환경을 기존 서비스에 통합하고, 다시 빌드하여 실행 파일을 생성해야 합니다.

### 지속 실행

런타임 환경을 지속적으로 유지하면 외부 요청에 즉시 대응할 수 있습니다. 각 런타임 환경은 독립적인 컨테이너에서 실행되며, 지속적인 워크플로우 실행이 필요한 시나리오에 적합합니다.

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
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		_ = r.Close()
	}()

	r.Listen(context.TODO())
}
```

위 코드에서는 런타임 환경을 지속적으로 실행하여 외부 신호에 반응하도록 설정합니다. `os.Signal`을 통해 종료 신호를 수신하면 런타임 환경을 안전하게 종료합니다.

### 단순 실행

때로는 런타임 환경을 지속적으로 유지하는 대신, 필요할 때만 실행하고 종료하는 간단한 방식이 더 적합할 수 있습니다. 이럴 때는 단순 실행 방식을 사용할 수 있습니다.

```go
r := runtime.New(runtime.Config{
	Namespace:   "default",
	Schema:      scheme,
	Hook:        hook,
	SpecStore:   specStore,
	SecretStore: secretStore,
})
defer r.Close()

r.Load(ctx) // 모든 리소스 로드

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
