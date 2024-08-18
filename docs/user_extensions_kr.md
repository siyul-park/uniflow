# 🔧 사용자 확장

현재 대부분의 기능은 이미 uniflow의 노드들의 조합을 통해 구현할 수 있지만, 가끔 특수한 기능이나 추가적인 기능을 직접 구현해야 하는 상황이 생길 수 있습니다. 이런 상황에서 어떻게 기능을 만들고 적용하는지에 대한 가이드입니다.

해당 가이드를 읽기 전에 [핵심 키워드](https://github.com/siyul-park/uniflow/blob/main/docs/key_concepts.md)와 [시스템 구조](https://github.com/siyul-park/uniflow/blob/main/docs/architecture.md)를 읽는 것을 권장합니다.

## 개발 환경 설정

우선, [Go](https://go.dev) 모듈을 초기화하고 필요한 의존성을 설치합니다.

```shell
go get github.com/siyul-park/uniflow
```

## 워크플로우 작성하기

새로운 기능을 워크플로우에 연결하려면 먼저 기능이 구현되어 있어야 하는 것이 정상이지만, 여기서는 설명을 편하게 하기 위해 전체적인 그림을 보여주는 느낌으로 최종적으로 완성되는 형태를 먼저 보여드리겠습니다.

아주 간단하게 프록시 기능을 구현하여 http 요청을 하면 날린 메시지를 로드밸런싱하여 처리하는 워크플로우입니다.

```yaml
- kind: listener
  name: listener
  protocol: http
  port: 8000
  ports:
    out:
      - name: proxy
        port: in

# 직접 만들 proxy 노드입니다.
- kind: proxy
  name: proxy
  urls:
    - https://backend1.com/
    - https://backend2.com/
```

8000번 포트를 통해 http 요청을 넣으면, 프록시가 주어진 urls 중 하나를 선택하여 자동적으로 데이터를 처리하게 됩니다. 이제 이 proxy 노드를 만들면 됩니다.

## 새로운 노드 작성

노드가 만들어지려면 크게 **구조 및 유형 정의 -> 동작 함수 정의 -> 생성 함수 정의** 의 3가지 과정을 거치게 됩니다.

노드의 스펙을 정의하고, 노드 유형(kind)에 들어갈 이름을 정한 후, 노드가 할 일을 정의하는 동작 함수를 구현하고 이 노드를 생성하는 함수를 만들면 기본적인 노드 구성이 만들어집니다. 여기까지의 과정을 '노드 명세를 만든다' 라고 하며, 이후 이 명세를 실제 동작하는 노드로 변환해주는 코덱을 만들고 스키마에 등록시키는 과정을 거쳐, 최종적으로 런타임 환경과 연결하게 됩니다.

### 노드 명세 정의

노드 명세는 `spec.Spec` 인터페이스에 맞춰 구성을 갖춰야 합니다. 아래의 항목이 필요합니다.

```go
ID uuid.UUID // UUID 형식의 고유 식별자입니다.
Kind string // 노드의 종류를 지정합니다.
Namespace string // 노드가 속한 네임스페이스를 지정합니다.
Name string // 노드의 이름을 지정하며, 동일한 네임스페이스 내에서 고유해야 합니다.
Annotations map[string]string // 노드에 대한 추가 메타데이터입니다.
Ports map[string][]Port // 포트의 연결 방식을 정의합니다.
Env map[string][]Secret // 노드에 필요한 환경 변수를 지정합니다.
```

이 때 `spec.Meta`를 사용하면 간단하게 작성할 수 있습니다:

```go
type ProxyNodeSpec struct {
  spec.Meta `map:",inline"`
}
```

#### 추가 필드 받기

만약 추가로 받을 값이 필요하다면, 필수 항목을 제외하고 추가적인 항목을 작성하면 됩니다. 프록시 기능을 만들기 위해 URL 정보가 필요하다고 가정하면, 아래와 같이 URLS 필드를 선언하면 됩니다.

```go
type ProxyNodeSpec struct {
  spec.Meta `map:",inline"`
  URLS      []string `map:"urls"`
}
```

이 필드는 이후 워크플로우에서 노드를 사용할 때 추가 필드로 사용할 수 있게 되며, 환경 변수 등 초기 설정 값으로 받아들일 수 있습니다.

```yaml
- kind: proxy
  name: proxy
  urls:
    - https://backend1.com/
    - https://backend2.com/
```

### 노드 유형 정의

이제 노드 유형을 정의합니다. 해당 유형이 정확하게 작성되어 있어야 런타임이 노드를 올바르게 인식할 수 있습니다:

```go
const KindProxy = "proxy"
```

### 노드 타입 정의

이제 노드 명세를 기반으로 노드가 동작하기 위해 실제로 필요한 요소들을 정의해야 합니다. 쉽게 말해서 노드가 어떤 방식으로 통신할 것이고 어떠한 데이터를 가질 것인지에 대한 정보가 담겨 있어야 합니다:

```go
type ProxyNode struct {
  *node.OneToOneNode
  proxy *httputil.ReverseProxy
}
```

노드끼리 통신을 하려면 통신 규격이 정의되어야 하는데, uniflow에서는 `ZeroToOne`, `OneToOne`, `OneToMany`, `ManyToOne`, `Other` 규격을 지원합니다.

여기서 사용할 `OneToOneNode` 템플릿은 1:1 구조를 지원하며, 입력 포트에서 패킷을 받아 처리한 후, 이를 출력 포트로 바로 전달하는 노드를 쉽게 구현할 수 있도록 도와줍니다.

### 노드 동작 구현

이제 노드가 입력 패킷을 처리하고, 그 결과를 출력 패킷으로 생성하는 과정을 구현합니다. 패킷은 페이로드를 담고 있으며, 이 페이로드는 `types.Value` 인터페이스를 구현하는 여러 공용 데이터 타입 중 하나로 표현됩니다.

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

앞서 사용했던 예시를 계속 구현해봅시다. 프록시 기능을 만드려면 받은 패킷을 정해진 순서에 맞춰 URL을 바꾸어 서버에 요청할 수 있는 구조가 필요합니다. 이 때 들어오는 패킷은 서버에 직접적으로 리소스를 요청하는 구조이므로, 패킷 데이터만 가져다 직접 요청하고 응답값을 받는 형식으로 만들어야 합니다.

```go
func (n *ProxyNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
  // 페이로드를 다룰 수 있는 형태로 변경합니다.
  req := HTTPPayload{}
  if err := types.Unmarshal(inPck.Payload(), &req); err != nil {
    return nil, packet.New(types.NewError(err))
  }

  // body 데이터를 가져옵니다.
  buf := bytes.NewBuffer(nil)
  if err := mime.Encode(buf, req.Body, textproto.MIMEHeader(req.Header)); err != nil {
    return nil, packet.New(types.NewError(err))
  }

  // 이제 이 값을 기반으로 프록시 환경에서 사용할 Request 데이터를 만듭니다.
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

  // 프록시로 http 요청을 수행합니다.
  w := httptest.NewRecorder()
  n.proxy.ServeHTTP(w, r)

  // 결과값 body를 가지고 옵니다.
  body, err := mime.Decode(w.Body, textproto.MIMEHeader(w.Header()))
  if err != nil {
    return nil, packet.New(types.NewError(err))
  }

  // 결과값 페이로드를 만듭니다.
  res := &HTTPPayload{
    Header: w.Header(),
    Body:   body,
    Status: w.Code,
  }

  // 이제 해당 결과를 보내줄 패킷 형태로 만들어 반환하면 됩니다.
  outPayload, err := types.Encoder.Encode(res)
  if err != nil {
    return nil, packet.New(types.NewError(err))
  }

  return packet.New(outPayload), nil
}
```

> 들어오는 요청을 재구성하지 않고, 헤더 값만 살짝 수정하는 식으로 다른 방법을 생각할 수도 있습니다. 실제로 http listener에서 들어온 `proc` 객체에서 이미 `http.ResponseWriter`, `*http.Request` 두 값이 존재하고 이를 얻어올 수 있으나, 이 둘의 값을 함부로 건드리면 요청 응답의 전후처리가 불가능해질 수도 있습니다. 정말 필요한 상황이 아니라면 프로세스 구조를 건드리지 않는 것이 좋습니다.

최종적으로 완성된 outPayload를 반환하면 동작 함수가 모두 완성됩니다. 반환할 때는 정상적으로 처리된 결과를 첫 번째 반환값으로, 오류가 발생한 경우에는 두 번째 반환값으로 반환합니다.

### 노드 생성

이제 노드를 실제로 구현해 보겠습니다. 노드를 생성하는 함수를 정의하고, 패킷 처리 방식을 `OneToOneNode` 생성자에 전달합니다:

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

예제에서는 설명을 위해 어떠한 추가 기능도 구현하지 않았지만, 상태 확인 등 완성도를 높일 수 있는 다양한 기능들을 추가할 수 있습니다.

### 테스트 작성

노드가 의도대로 작동하는지 확인하기 위해 테스트를 작성합니다. 입력 패킷을 `in` 포트로 전송하고, `out` 포트에서 출력 패킷이 예상대로 나오는지에 대해 검증합니다:

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

assert가 성공하면, 하나의 노드로써 온전한 기능을 수행할 수 있음을 확인할 수 있습니다.

## 런타임 연결

지금까지 앞에서 설명했던 내용은 하나의 노드를 어떻게 만드는가에 대한 이야기였습니다. 이제 시스템에 노드를 연결하기 위해 코덱을 만들고, 스키마와 연결해야 실행했을 때 노드가 동작하여 원하는 작업을 수행할 수 있게 됩니다.

### 코덱 생성

노드 명세를 노드로 변환하는 코덱을 작성합니다.

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

### 스키마 생성 및 추가

노드를 만들 때 사용했던 명세와 유형을 외부에서 인식하고 사용할 수 있도록 스키마 생성 함수를 만듭니다:

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

이렇게 정의된 `scheme.Register`를 `scheme.Builder`에 전달하여 스키마를 생성합니다:

```go
builder := scheme.NewBuilder()
builder.Register(AddToScheme())

scheme, _ := builder.Build()
```

### 런타임 환경 실행

이제 이 스키마를 런타임 환경에 전달하면 직접 만든 노드가 포함된 워크플로우를 실행할 수 있습니다:

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

워크플로우에 대한 모든 데이터는 이제 `r`에 있으니, 해당 변수를 목적에 따라 실행하면 모든 준비는 끝납니다.

## 기존 서비스와 통합

이제 이렇게 만들어진 런타임 환경을 기존 서비스에 추가하고, 다시 빌드해서 실행 파일을 만들어야 합니다.

서비스를 추가하는 방법은 두 가지로, 런타임이 계속해서 돌아가면서 운영되는 지속 실행 방법과, 한 번 실행하고 끝나는 단순 실행 방법이 있습니다.

### 지속 실행

런타임 환경을 지속적으로 유지하면 외부 요청에 즉각적으로 대응할 수 있습니다. 각 런타임 환경은 독립적인 컨테이너에서 실행되며, 지속적인 워크플로우 실행이 필요한 시나리오에 적합합니다.

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

### 단순 실행

런타임 환경을 지속적으로 유지함으로써 얻는 장점도 있지만, 필요할 때만 동작하기를 원하거나 간단하게 동작하기를 원할 수 있습니다. 이럴 때는 단순 실행을 목적으로 구성할 수 있습니다.

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

sym := symbols[0]

in := port.NewOut()
defer in.Close()

in.Link(sym.In(node.PortIn))

payload := types.NewString(faker.Word())
payload, err := port.Call(in, payload)
```
