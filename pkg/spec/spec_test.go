package spec

import (
	"github.com/siyul-park/uniflow/pkg/secret"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestConvert(t *testing.T) {
	meta := &Meta{
		ID:          uuid.Must(uuid.NewV7()),
		Kind:        faker.Word(),
		Namespace:   "default",
		Name:        faker.Word(),
		Annotations: map[string]string{"key": "value"},
		Ports:       map[string][]Port{"out": {{Name: faker.Word(), Port: "in"}}},
		Env:         map[string][]Value{"env1": {{Name: "secret1", Data: "value1"}}},
	}

	unstructured := &Unstructured{}
	err := Convert(meta, unstructured)
	assert.NoError(t, err)
}

func TestMeta_SetID(t *testing.T) {
	meta := &Meta{}
	id := uuid.Must(uuid.NewV7())
	meta.SetID(id)
	assert.Equal(t, id, meta.GetID())
}

func TestMeta_SetKind(t *testing.T) {
	meta := &Meta{}
	kind := faker.Word()
	meta.SetKind(kind)
	assert.Equal(t, kind, meta.GetKind())
}

func TestMeta_SetNamespace(t *testing.T) {
	meta := &Meta{}
	namespace := faker.Word()
	meta.SetNamespace(namespace)
	assert.Equal(t, namespace, meta.GetNamespace())
}

func TestMeta_SetName(t *testing.T) {
	meta := &Meta{}
	name := faker.Word()
	meta.SetName(name)
	assert.Equal(t, name, meta.GetName())
}

func TestMeta_SetAnnotations(t *testing.T) {
	meta := &Meta{}
	annotations := map[string]string{"key": "value"}
	meta.SetAnnotations(annotations)
	assert.Equal(t, annotations, meta.GetAnnotations())
}

func TestMeta_SetPorts(t *testing.T) {
	meta := &Meta{}
	ports := map[string][]Port{
		"out": {
			{
				ID:   uuid.Must(uuid.NewV7()),
				Port: "in",
			},
		},
	}
	meta.SetPorts(ports)
	assert.Equal(t, ports, meta.GetPorts())
}

func TestMeta_SetEnv(t *testing.T) {
	meta := &Meta{}
	env := map[string][]Value{
		"FOO": {
			{
				ID:   uuid.Must(uuid.NewV7()),
				Data: "baz",
			},
		},
	}
	meta.SetEnv(env)
	assert.Equal(t, env, meta.GetEnv())
}

func TestMeta_IsBound(t *testing.T) {
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

	assert.True(t, meta.IsBound(sec1))
	assert.False(t, meta.IsBound(sec2))
}

func TestMeta_Bind(t *testing.T) {
	sec := &secret.Secret{
		ID:   uuid.Must(uuid.NewV7()),
		Data: faker.UUIDHyphenated(),
	}
	meta := &Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: faker.UUIDHyphenated(),
		Env: map[string][]Value{
			"FOO": {
				{
					ID:   sec.ID,
					Data: "{{ . }}",
				},
			},
		},
	}

	err := meta.Bind(sec)
	assert.NoError(t, err)
	assert.Equal(t, sec.Data, meta.GetEnv()["FOO"][0].Data)
}
