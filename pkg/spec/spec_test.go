package spec

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/secret"
	"github.com/stretchr/testify/assert"
)

func TestIsBound(t *testing.T) {
	sec1 := &secret.Secret{
		ID: uuid.Must(uuid.NewV7()),
	}
	sec2 := &secret.Secret{
		ID: uuid.Must(uuid.NewV7()),
	}

	meta := &Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: faker.UUIDHyphenated(),
		Env: map[string][]Value{
			"FOO": {
				{
					ID:   sec1.ID,
					Data: "foo",
				},
			},
		},
	}

	assert.True(t, IsBound(meta, sec1))
	assert.False(t, IsBound(meta, sec2))
}

func TestBind(t *testing.T) {
	scrt := &secret.Secret{
		ID:   uuid.Must(uuid.NewV7()),
		Data: "foo",
	}

	meta := &Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: faker.UUIDHyphenated(),
		Env: map[string][]Value{
			"FOO": {
				{
					ID:   scrt.ID,
					Data: "{{ . }}",
				},
			},
		},
	}

	bind, err := Bind(meta, scrt)
	assert.NoError(t, err)
	assert.Equal(t, "foo", bind.GetEnv()["FOO"][0].Data)
	assert.True(t, IsBound(bind, scrt))
}

func TestMeta_Get(t *testing.T) {
	meta := &Meta{
		ID:          uuid.Must(uuid.NewV7()),
		Kind:        faker.Word(),
		Namespace:   "default",
		Name:        faker.Word(),
		Annotations: map[string]string{"key": "value"},
		Ports:       map[string][]Port{"out": {{Name: faker.Word(), Port: "in"}}},
		Env:         map[string][]Value{"env1": {{Name: "secret1", Data: "value1"}}},
	}

	assert.Equal(t, meta.ID, meta.GetID())
	assert.Equal(t, meta.Kind, meta.GetKind())
	assert.Equal(t, meta.Namespace, meta.GetNamespace())
	assert.Equal(t, meta.Name, meta.GetName())
	assert.Equal(t, meta.Annotations, meta.GetAnnotations())
	assert.Equal(t, meta.Ports, meta.GetPorts())
	assert.Equal(t, meta.Env, meta.GetEnv())
}

func TestMeta_Set(t *testing.T) {
	meta := &Meta{}

	id := uuid.Must(uuid.NewV7())
	meta.SetID(id)
	assert.Equal(t, id, meta.GetID())

	kind := "testKind"
	meta.SetKind(kind)
	assert.Equal(t, kind, meta.GetKind())

	namespace := "testNamespace"
	meta.SetNamespace(namespace)
	assert.Equal(t, namespace, meta.GetNamespace())

	name := "testName"
	meta.SetName(name)
	assert.Equal(t, name, meta.GetName())

	annotations := map[string]string{"key": "value"}
	meta.SetAnnotations(annotations)
	assert.Equal(t, annotations, meta.GetAnnotations())

	ports := map[string][]Port{
		"http": {
			{ID: uuid.Must(uuid.NewV7()), Name: "port1", Port: "8080"},
		},
	}
	meta.SetPorts(ports)
	assert.Equal(t, ports, meta.GetPorts())

	env := map[string][]Value{
		"FOO": {
			{ID: uuid.Must(uuid.NewV7()), Name: "bar", Data: "baz"},
		},
	}
	meta.SetEnv(env)
	assert.Equal(t, env, meta.GetEnv())
}
