package cel

import (
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

type adapter struct{}

var _ types.Adapter = (*adapter)(nil)

var TypeAdapter = &adapter{}

func (*adapter) NativeToValue(value any) ref.Val {
	switch v := value.(type) {
	case error:
		return &Error{error: v}
	}
	return types.DefaultTypeAdapter.NativeToValue(value)
}
