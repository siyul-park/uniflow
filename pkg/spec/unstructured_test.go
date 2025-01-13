package spec

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/node"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUnstructured_GetAndSet(t *testing.T) {
	t.Run("KeyID", func(t *testing.T) {
		unstructured := &Unstructured{}
		id := uuid.Must(uuid.NewV4())
		unstructured.Set(KeyID, id)
		val, ok := unstructured.Get(KeyID)
		assert.True(t, ok)
		assert.Equal(t, id, val)
	})

	t.Run("KeyKind", func(t *testing.T) {
		unstructured := &Unstructured{}
		kind := faker.UUIDHyphenated()
		unstructured.Set(KeyKind, kind)
		val, ok := unstructured.Get(KeyKind)
		assert.True(t, ok)
		assert.Equal(t, kind, val)
	})

	t.Run("KeyNamespace", func(t *testing.T) {
		unstructured := &Unstructured{}
		unstructured.Set(KeyNamespace, "default")
		val, ok := unstructured.Get(KeyNamespace)
		assert.True(t, ok)
		assert.Equal(t, "default", val)
	})

	t.Run("KeyName", func(t *testing.T) {
		unstructured := &Unstructured{}
		name := faker.UUIDHyphenated()
		unstructured.Set(KeyName, name)
		val, ok := unstructured.Get(KeyName)
		assert.True(t, ok)
		assert.Equal(t, name, val)
	})

	t.Run("KeyAnnotations", func(t *testing.T) {
		unstructured := &Unstructured{}
		annotations := map[string]string{"key": "value"}
		unstructured.Set(KeyAnnotations, annotations)
		val, ok := unstructured.Get(KeyAnnotations)
		assert.True(t, ok)
		assert.Equal(t, annotations, val)
	})

	t.Run("KeyPorts", func(t *testing.T) {
		unstructured := &Unstructured{}
		ports := map[string][]Port{
			node.PortOut: {
				{
					Name: faker.UUIDHyphenated(),
					Port: node.PortIn,
				},
			},
		}
		unstructured.Set(KeyPorts, ports)
		val, ok := unstructured.Get(KeyPorts)
		assert.True(t, ok)
		assert.Equal(t, ports, val)
	})

	t.Run("KeyEnv", func(t *testing.T) {
		unstructured := &Unstructured{}
		env := map[string]Value{"env1": {Name: "value1", Data: "value1"}}
		unstructured.Set(KeyEnv, env)
		val, ok := unstructured.Get(KeyEnv)
		assert.True(t, ok)
		assert.Equal(t, env, val)
	})

	t.Run("CustomField", func(t *testing.T) {
		unstructured := &Unstructured{}
		customKey := "customField"
		customValue := "customValue"
		unstructured.Set(customKey, customValue)
		val, ok := unstructured.Get(customKey)
		assert.True(t, ok)
		assert.Equal(t, customValue, val)
	})
}

func TestUnstructured_Build(t *testing.T) {
	unstructured := &Unstructured{
		Meta: Meta{
			ID:   uuid.Must(uuid.NewV7()),
			Kind: faker.UUIDHyphenated(),
			Env: map[string]Value{
				"FOO": {
					Data: "foo",
				},
			},
		},
		Fields: map[string]any{
			"foo": "{{ .FOO }}",
		},
	}

	err := unstructured.Build()
	assert.NoError(t, err)
	assert.Equal(t, "foo", unstructured.Fields["foo"])
}
