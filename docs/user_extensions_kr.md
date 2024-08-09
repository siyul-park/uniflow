# 🔧 사용자 확장

이 가이드는 사용자가 자신의 서비스를 확장하고 이를 런타임 환경에 통합하는 방법을 설명합니다.

## 개발 환경 설정

먼저, [Go](https://go.dev) 모듈을 초기화하고 필요한 의존성을 설치합니다.

```shell
go get github.com/siyul-park/uniflow
```

## 새로운 노드 추가

새로운 노드를 지원하려면, 먼저 노드의 작성 및 저장 방식을 선언적으로 정의하는 명세를 작성하고, 이를 바탕으로 실제로 패킷을 처리하는 노드를 구현해야 합니다. 그런 다음, 노드 명세를 노드로 변환하는 코덱을 스키마에 등록하여 런타임 환경과 연결해야 합니다.

### 명세 정의

노드 명세는 `spec.Spec` 인터페이스를 구현해야 하며, 이를 구현한 `spec.Meta`를 사용하면 쉽게 정의할 수 있습니다:

```go
type TextNodeSpec struct {
	spec.Meta `map:",inline"`
	Contents  string `map:"contents"`
}
```

새로운 노드 유형을 정의하는 방법은 다음과 같습니다:

```go
const KindText = "text"
```

### 노드 생성

이제 노드를 실제로 구현해 보겠습니다. `OneToOneNode` 템플릿은 입력 포트에서 패킷을 받아 처리한 후, 이를 출력 포트로 전달하는 노드를 쉽게 구현할 수 있도록 도와줍니다.

```go
type TextNode struct {
	*node.OneToOneNode
	contents string
}
```

노드를 생성하는 함수를 정의하고, 패킷 처리 방식을 `OneToOneNode` 생성자에 전달합니다:

```go
func NewTextNode(contents string) *TextNode {
	n := &TextNode{contents: contents}
	n.OneToOneNode = node.NewOneToOneNode(n.action)

	return n
}
```

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

패킷 처리 함수는 정상적으로 처리된 결과를 첫 번째 반환값으로, 오류가 발생한 경우에는 두 번째 반환값으로 반환합니다. 이 예제에서는 노드의 `contents` 값을 그대로 출력합니다.

```go
func (n *TextNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	outPck := packet.New(types.NewString(n.contents))
	return outPck, nil
}
```

### 테스트 작성

노드가 의도대로 작동하는지 확인하기 위해 테스트를 작성합니다. 입력 패킷을 `in` 포트로 전송하고, 출력 패킷이 예상대로 `contents` 값을 포함하는지 검증합니다:

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

### 스키마 및 코덱 등록

이제 노드 명세와 노드를 연결해야 합니다. 노드 명세를 노드로 변환하는 코덱을 작성하고 이를 스키마에 등록합니다:

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

이렇게 정의된 `scheme.Register`를 `scheme.Builder`에 전달하여 스키마를 생성합니다:

```go
builder := scheme.NewBuilder()
builder.Register(AddToScheme())

scheme, _ := builder.Build()
```

### 런타임 환경 실행

이제 이 스키마를 런타임 환경에 전달하여 확장된 노드를 포함한 워크플로우를 실행할 수 있습니다:

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

## 서비스와의 통합

런타임 환경을 서비스에 통합하는 방법에는 두 가지가 있습니다.

### 지속적 통합

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

### 임시 실행

워크플로우를 일시적으로 실행하고, 실행이 끝난 후 환경을 정리할 수 있습니다. 이 방법은 단기적인 환경에서 워크플로우를 실행하는 데 적합합니다:

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