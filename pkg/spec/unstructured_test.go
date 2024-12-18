package spec

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUnstructured_GetSet(t *testing.T) {
	unstructured := &Unstructured{}

	id := uuid.Must(uuid.NewV7())
	unstructured.Set(KeyID, id)
	val, ok := unstructured.Get(KeyID)
	assert.True(t, ok)
	assert.Equal(t, id, val)

	kind := faker.Word()
	unstructured.Set(KeyKind, kind)
	val, ok = unstructured.Get(KeyKind)
	assert.True(t, ok)
	assert.Equal(t, kind, val)

	unstructured.Set(KeyNamespace, "default")
	val, ok = unstructured.Get(KeyNamespace)
	assert.True(t, ok)
	assert.Equal(t, "default", val)

	name := faker.Word()
	unstructured.Set(KeyName, name)
	val, ok = unstructured.Get(KeyName)
	assert.True(t, ok)
	assert.Equal(t, name, val)

	annotations := map[string]string{"key": "value"}
	unstructured.Set(KeyAnnotations, annotations)
	val, ok = unstructured.Get(KeyAnnotations)
	assert.True(t, ok)
	assert.Equal(t, annotations, val)

	ports := map[string][]Port{"port1": {{Name: faker.Word(), Port: "8080"}}}
	unstructured.Set(KeyPorts, ports)
	val, ok = unstructured.Get(KeyPorts)
	assert.True(t, ok)
	assert.Equal(t, ports, val)

	env := map[string][]Value{"env1": {{Name: "secret1", Data: "value1"}}}
	unstructured.Set(KeyEnv, env)
	val, ok = unstructured.Get(KeyEnv)
	assert.True(t, ok)
	assert.Equal(t, env, val)

	customKey := "customField"
	customValue := "customValue"
	unstructured.Set(customKey, customValue)
	val, ok = unstructured.Get(customKey)
	assert.True(t, ok)
	assert.Equal(t, customValue, val)
}

func TestUnstructured_Build(t *testing.T) {
	unstructured := &Unstructured{
		Meta: Meta{
			ID:   uuid.Must(uuid.NewV7()),
			Kind: faker.UUIDHyphenated(),
			Env: map[string][]Value{
				"FOO": {
					{
						Data: "foo",
					},
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
