# 🔧 사용자 확장

사용자가 uniflow 시스템을 이해하고 사용하여 자신의 서비스를 확장하고 이를 런타임 환경에 통합하는 방법을 설명합니다.

해당 가이드를 읽기 전에 [핵심 키워드](https://github.com/siyul-park/uniflow/blob/main/docs/key_concepts.md)와 [시스템 구조](https://github.com/siyul-park/uniflow/blob/main/docs/architecture.md)를 읽는 것을 권장합니다.

## 개발 환경 설정

우선, [Go](https://go.dev) 모듈을 초기화하고 필요한 의존성을 설치합니다.

```shell
go get github.com/siyul-park/uniflow
```

## 명세서 작성

원하는 서비스를 만들기 위해선 가장 먼저 큰 틀을 잡을 명세서가 구성되어야 합니다. 명세서는 개발 도중 계속해서 바뀔 수 있고 최종적으로 만들어진 형태에서 다시 한번 수정이 이루어지는 경우가 많으므로, 구조에 익숙하다면 코드 구성이 모두 끝난 다음 작성해도 무방합니다.

examples 폴더 안에 있는 .yaml 명세서를 응용해도 되고, 처음부터 새로 명세서를 만들어도 됩니다. 이번 예시에서는 간단하게 ping 명세서를 응용하여 http 서버에서 POST 요청을 하면 날린 메시지를 그대로 서버에서 받아서 출력하는 서비스를 만들어보도록 하겠습니다.

```yaml
- kind: listener
  name: listener
  protocol: http
  port: 8000
  ports:
    out:
      - name: router
        port: in

- kind: router
  name: router
  routes:
    - method: POST
      path: /ping
      port: out[0]
  ports:
    out[0]:
      - name: pong
        port: in

- kind: my-snippet
  name: pong
```

http 서버를 8000번으로 잡고, router를 in으로 받도록 선언하고, router는 pong을 in으로 받도록 선언합니다. (노드를 연결할 때는 name 필드를 기준으로 합니다.)

이제 /ping 으로 POST 요청을 날리면 보낸 메시지를 받아 처리한 후 pong 요청을 서버에 돌려주는 my-snippet 노드를 직접 만들면 기본적인 구조가 완성됩니다.

## 새로운 노드 작성

노드가 만들어지려면 크게 **구조 및 유형 정의 -> 동작 함수 정의 -> 생성 함수 정의** 의 3가지 과정을 거치게 됩니다.

명세서에 사용할 수 있는 노드의 스펙을 먼저 선언한 다음, 노드 유형(kind)에 들어갈 이름을 정한 후, 노드가 할 일을 정의하는 동작 함수를 구현하고 이 노드를 생성하는 함수를 만들면 기본적인 노드 구성이 만들어집니다. 여기까지의 과정을 '노드 명세를 만든다' 라고 하며, 이후 이 명세를 실제 동작하는 노드로 변환해주는 코덱을 만들고 스키마에 등록시키는 과정을 거쳐, 최종적으로 런타임 환경과 연결하게 됩니다.

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

이 때 `spec.Meta`를 사용하면 간단하게 선언할 수 있습니다:

```go
type MySnippetNodeSpec struct {
	spec.Meta `map:",inline"`
  Contents string
}
```

만약 추가로 받을 값이 필요하다면, 필수 항목을 제외하고 추가적인 항목을 작성하면 됩니다. 이후 명세서에서 노드를 사용할 때 추가 필드로 사용할 수 있게 되며, 환경 변수 등 명세서에 선언할 때 초기 설정 값으로 받아들일 수 있습니다. 예시로 다양한 언어를 지원하기 위해 코드가 있는 code 필드와 어떤 코드인지에 대한 정보인 language 필드를 받고 싶다면, 이런 식으로 노드 명세에 추가가 가능합니다:

```go
type MySnippetNodeSpec struct {
	spec.Meta `map:",inline"`
	Language  string `map:"language,omitempty"` // <- 언어 종류
	Code      string `map:"code"` // <- 소스 코드
}
```

이후 이 두개의 값은 명세서에서 추가 필드로 인식되며, 넣은 값을 직접적으로 사용하거나 다른 노드에게 처리 과정을 지시할 수 있게 됩니다.

```yaml
- kind: my-snippet
  name: pong
  language: text # <- 언어 종류 필드
  code: pong # <- 소스 코드 필드
```

### 노드 유형 정의

이제 노드 유형을 선언합니다. 해당 유형이 정확하게 선언되어 있어야 런타임이 노드를 올바르게 인식할 수 있습니다:

```go
const KindMySnippet = "my-snippet"
```

### 노드 타입 정의

이제 노드 명세를 기반으로 노드가 동작하기 위해 실제로 필요한 요소들을 정의해야 합니다. 쉽게 말해서 노드가 어떤 방식으로 통신할 것이고 어떠한 데이터를 받아들일 것인지에 대한 정보가 담겨 있어야 합니다:

```go
type MySnippetNode struct {
  *node.OneToOneNode
}
```

노드끼리 통신을 하려면 통신 규격이 정의되어야 하는데, uniflow에서는 `ZeroToOne`, `OneToOne`, `OneToMany`, `ManyToOne`, `Other` 규격을 지원합니다.

여기서 사용할 `OneToOneNode` 템플릿은 1:1 구조를 지원하며, 입력 포트에서 패킷을 받아 처리한 후, 이를 출력 포트로 바로 전달하는 노드를 쉽게 구현할 수 있도록 도와줍니다.

### 노드 동작 구현

이제 노드가 입력 패킷을 처리하고, 그 결과를 출력 패킷으로 생성하는 방법을 구현합니다. 패킷은 데이터인 페이로드를 담고 있으며, 이 페이로드는 `types.Value` 인터페이스를 구현하는 여러 공용 데이터 타입 중 하나로 표현됩니다.

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

패킷 처리 함수는 정상적으로 처리된 결과를 첫 번째 반환값으로, 오류가 발생한 경우에는 두 번째 반환값으로 반환합니다:

```go
func (n *MySnippetNode) action(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	inPayload := inPck.Payload()
	input := types.InterfaceOf(inPayload)

  if outPayload, err := types.Encoder.Encode(input); err == nil {
    return packet.New(outPayload), nil
  } else {
    return nil, packet.New(types.NewErorr(err))
  }
}
```

### 노드 생성

이제 노드를 실제로 구현해 보겠습니다. 노드를 생성하는 함수를 정의하고, 패킷 처리 방식을 `OneToOneNode` 생성자에 전달합니다:

```go
func NewMySnippetNode() *MySnippetNode {
	n := &MySnippetNode{}
	n.OneToOneNode = node.NewOneToOneNode(n.action)
	return n
}
```

### 테스트 작성

노드가 의도대로 작동하는지 확인하기 위해 테스트를 작성합니다. 입력 패킷을 `in` 포트로 전송하고, `out` 포트에서 출력 패킷이 예상대로 나오는지에 대해 검증합니다:

```go
func TestMySnippetNode_SendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	n := NewMySnippetNode()
	defer n.Close()

	out := port.NewOut()
	out.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload := types.NewString(faker.Word())
	inPck := packet.New(inPayload)

	inWriter.Write(inPck)

	select {
	case outPck := <-inWriter.Receive():
		assert.Equal(t, inPayload, outPck.Payload())
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
func NewMySnippetNodeCodec(module *language.Module) scheme.Codec {
	return scheme.CodecWithType(func(spec *MySnippetNodeSpec) (node.Node, error) {
		return NewMySnippetNode(), nil
	})
}
```

### 스키마 생성 및 추가

노드를 만들 때 사용했던 명세와 유형을 외부에서 인식하고 사용할 수 있도록 스키마 생성 함수를 만듭니다:

```go
func AddToScheme() scheme.Register {
	return scheme.RegisterFunc(func(s *scheme.Scheme) error {
		s.AddKnownType(KindMySnippet, &MySnippetNodeSpec{})
		s.AddCodec(KindMySnippet, NewMySnippetNodeCodec())
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

## 서비스와의 통합

이렇게 만들어진 런타임 환경을 기존 서비스에 통합하는 방법에는 두 가지가 있습니다. 런타임이 계속해서 돌아가면서 운영되는 지속 실행 방법과, 한 번 실행하고 끝나는 단순 실행 방법이 있습니다.

### 지속 실행

런타임 환경을 지속적으로 유지하면 외부 요청에 신속하게 대응할 수 있습니다. 각 런타임 환경은 독립적인 컨테이너에서 실행되며, 지속적인 워크플로우 실행이 필요한 시나리오에 적합합니다.

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

런타임 환경을 지속적으로 유지함으로써 얻는 장점도 있지만, 필요할 때만 동작하기를 원하거나 간단하게 동작하기를 원할 수 있습니다. 이럴 때는 단순 실행을 목적으로 실행할 수 있습니다.

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
