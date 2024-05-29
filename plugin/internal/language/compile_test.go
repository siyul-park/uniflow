package language

import (
	"fmt"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/object"
	"github.com/stretchr/testify/assert"
)

func TestCompileTransformWithPrimitive(t *testing.T) {
	fun, err := CompileTransformWithPrimitive("$", "")
	assert.NoError(t, err)

	in := object.NewString(faker.Word())
	out, err := fun(in)
	assert.NoError(t, err)
	assert.Equal(t, in, out)
}

func TestCompileTransform(t *testing.T) {
	t.Run("Detect", func(t *testing.T) {
		fun, err := CompileTransform("$", nil)
		assert.NoError(t, err)

		in := faker.Word()
		out, err := fun(in)
		assert.NoError(t, err)
		assert.Equal(t, in, out)
	})

	t.Run(Text, func(t *testing.T) {
		in := faker.Word()

		fun, err := CompileTransform(in, lo.ToPtr(Text))
		assert.NoError(t, err)

		out, err := fun(in)
		assert.NoError(t, err)
		assert.Equal(t, in, out)
	})

	t.Run(Typescript, func(t *testing.T) {
		t.Run("inline", func(t *testing.T) {
			fun, err := CompileTransform("$", lo.ToPtr(Typescript))
			assert.NoError(t, err)

			in := faker.Word()
			out, err := fun(in)
			assert.NoError(t, err)
			assert.Equal(t, in, out)
		})

		t.Run("multiline", func(t *testing.T) {
			fun, err := CompileTransform("{\n foo: \"bar\" \n}", lo.ToPtr(Typescript))
			assert.NoError(t, err)

			in := faker.Word()
			out, err := fun(in)
			assert.NoError(t, err)
			assert.Equal(t, map[string]any{"foo": "bar"}, out)
		})

		t.Run("details", func(t *testing.T) {
			fun, err := CompileTransform(`export default function (input: any): any {
				return input;
			}`, lo.ToPtr(Typescript))
			assert.NoError(t, err)

			in := faker.Word()
			out, err := fun(in)
			assert.NoError(t, err)
			assert.Equal(t, in, out)
		})
	})

	t.Run(Javascript, func(t *testing.T) {
		t.Run("inline", func(t *testing.T) {
			fun, err := CompileTransform("$", lo.ToPtr(Javascript))
			assert.NoError(t, err)

			in := faker.Word()
			out, err := fun(in)
			assert.NoError(t, err)
			assert.Equal(t, in, out)
		})

		t.Run("multiline", func(t *testing.T) {
			fun, err := CompileTransform("{\n foo: \"bar\" \n}", lo.ToPtr(Javascript))
			assert.NoError(t, err)

			in := faker.Word()
			out, err := fun(in)
			assert.NoError(t, err)
			assert.Equal(t, map[string]any{"foo": "bar"}, out)
		})

		t.Run("details", func(t *testing.T) {
			fun, err := CompileTransform(`export default function (input) {
				return input;
			}`, lo.ToPtr(Javascript))
			assert.NoError(t, err)

			in := faker.Word()
			out, err := fun(in)
			assert.NoError(t, err)
			assert.Equal(t, in, out)
		})
	})

	t.Run(JSON, func(t *testing.T) {
		in := faker.Word()

		fun, err := CompileTransform(fmt.Sprintf("\"%s\"", in), lo.ToPtr(JSON))
		assert.NoError(t, err)

		out, err := fun(in)
		assert.NoError(t, err)
		assert.Equal(t, in, out)
	})

	t.Run(JSONata, func(t *testing.T) {
		fun, err := CompileTransform("$", lo.ToPtr(JSONata))
		assert.NoError(t, err)

		in := faker.Word()
		out, err := fun(in)
		assert.NoError(t, err)
		assert.Equal(t, in, out)
	})

	t.Run(YAML, func(t *testing.T) {
		in := faker.Word()

		fun, err := CompileTransform(fmt.Sprintf("\"%s\"", in), lo.ToPtr(YAML))
		assert.NoError(t, err)

		out, err := fun(in)
		assert.NoError(t, err)
		assert.Equal(t, in, out)
	})
}
