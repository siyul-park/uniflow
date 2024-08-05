# 🔧 사용자 확장

이 가이드는 사용자가 자신의 서비스를 확장하고 런타임 환경에 통합하는 방법을 설명합니다.

## 개발 환경 설정

먼저, [Go](https://go.dev) 모듈을 초기화하고 필요한 의존성을 다운로드합니다.

```shell
go get github.com/siyul-park/uniflow
```

## 새로운 노드 추가

새로운 기능을 추가하려면 노드 명세를 정의하고 이를 노드로 변환하는 코덱을 스키마에 등록합니다.

노드 명세는 `spec.Spec` 인터페이스를 구현하며, `spec.Meta`를 사용하여 정의할 수 있습니다:

```go
type TextNodeSpec struct {
	spec.Meta `map:",inline"`
	Contents  string `map:"contents"`
}
```

새로운 노드 유형을 정의합니다:

```go
const KindText = "text"
```

이제 실제로 동작할 노드를 구현합니다. 기본 템플릿으로 제공되는 `OneToOneNode`를 사용하여 입력 패킷을 수신하고 처리한 후 출력 패킷을 전송합니다:

```go
type TextNode struct {
	*node.OneToOneNode
	contents string
}
```

노드를 생성하는 함수를 정의하고, 패킷 처리 방식을 `OneToOneNode` 생성자에 전달하여 설정합니다:

```go
func NewTextNode(contents string) *TextNode {
	n := &TextNode{contents: contents}
	n.OneToOneNode = node.NewOneToOneNode(n.action)

	return n
}
```

패킷을 처리하는 함수를 구현합니다. 이 함수는 정상 처리 시 첫 번째 반환값을 사용하고, 오류 발생 시 두 번째 반환값을 사용합니다:

```go
func (n *TextNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	outPck := packet.New(types.NewString(n.contents))
	return outPck, nil
}
```

노드가 의도한 대로 작동하는지 확인하기 위해 테스트를 작성합니다. 입력 패킷을 `in` 포트로 전송하고, 출력 패킷이 `contents`를 포함하는지 확인합니다:

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

노드 명세를 노드로 변환하는 코덱을 작성하고 이를 스키마에 등록합니다:

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

그리고 `scheme.Builder`를 사용하여 스키마를 구성합니다:

```go
builder := scheme.NewBuilder()
builder.Register(AddToScheme())

scheme, _ := builder.Build()
```

### 런타임 환경 실행

이제 이 스키마를 런타임 환경에 전달하여 확장된 노드가 포함된 워크플로우를 실행합니다:

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

## 서비스 통합

런타임 환경을 서비스에 통합하는 방법에는 두 가지가 있습니다.

### 지속적 통합

외부 요청에 신속하게 대응하기 위해 런타임 환경을 지속적으로 유지합니다. 각 런타임 환경은 독립적인 컨테이너에서 실행되며, 지속적인 워크플로우 실행이 필요한 시나리오에 적합합니다.

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

워크플로우를 일시적으로 실행하고 실행 환경을 제거합니다. 이 방법은 일시적인 환경에서 워크플로우를 실행할 때 적합합니다:

```go
r := runtime.New(runtime.Config{
	Namespace:   "default",
	Schema:      scheme,
	Hook:        hook,
	SpecStore:   specStore,
	SecretStore: secretStore,
})
defer r.Close()

r.Load(ctx) // Load All

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