package cel

import (
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

type adapter struct{}

var _ types.Adapter = (*adapter)(nil)

func (*adapter) NativeToValue(value interface{}) ref.Val {
	switch v := value.(type) {
	case error:
		return &Error{error: v}
	}
	return types.DefaultTypeAdapter.NativeToValue(value)
}
