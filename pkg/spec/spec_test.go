package spec

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/value"
	"github.com/stretchr/testify/assert"
)

func TestAs(t *testing.T) {
	meta := &Meta{
		ID:          uuid.Must(uuid.NewV7()),
		Kind:        faker.UUIDHyphenated(),
		Namespace:   "default",
		Name:        faker.UUIDHyphenated(),
		Annotations: map[string]string{"key": "value"},
		Ports:       map[string][]Port{"out": {{Name: faker.UUIDHyphenated(), Port: "in"}}},
		Env:         map[string]Value{"env1": {Name: "value1", Data: "value1"}},
	}

	unstructured := &Unstructured{}
	err := As(meta, unstructured)
	assert.NoError(t, err)
}

func TestMeta_ID(t *testing.T) {
	meta := &Meta{}
	id := uuid.Must(uuid.NewV7())
	meta.SetID(id)
	assert.Equal(t, id, meta.GetID())
}

func TestMeta_Kind(t *testing.T) {
	meta := &Meta{}
	kind := faker.UUIDHyphenated()
	meta.SetKind(kind)
	assert.Equal(t, kind, meta.GetKind())
}

func TestMeta_Namespace(t *testing.T) {
	meta := &Meta{}
	namespace := faker.UUIDHyphenated()
	meta.SetNamespace(namespace)
	assert.Equal(t, namespace, meta.GetNamespace())
}

func TestMeta_Name(t *testing.T) {
	meta := &Meta{}
	name := faker.UUIDHyphenated()
	meta.SetName(name)
	assert.Equal(t, name, meta.GetName())
}

func TestMeta_Annotations(t *testing.T) {
	meta := &Meta{}
	annotations := map[string]string{"key": "value"}
	meta.SetAnnotations(annotations)
	assert.Equal(t, annotations, meta.GetAnnotations())
}

func TestMeta_Env(t *testing.T) {
	meta := &Meta{}
	env := map[string]Value{
		"FOO": {

			ID:   uuid.Must(uuid.NewV7()),
			Data: "baz",
		},
	}
	meta.SetEnv(env)
	assert.Equal(t, env, meta.GetEnv())
}

func TestMeta_Ports(t *testing.T) {
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

func TestMeta_NamespacedName(t *testing.T) {
	t.Run("ID", func(t *testing.T) {
		meta := &Meta{
			ID:          uuid.Must(uuid.NewV7()),
			Kind:        faker.UUIDHyphenated(),
			Namespace:   "default",
			Annotations: map[string]string{"key": "value"},
			Ports:       map[string][]Port{"out": {{Name: faker.UUIDHyphenated(), Port: "in"}}},
			Env:         map[string]Value{"env1": {Name: "value1", Data: "value1"}},
		}
		assert.Equal(t, meta.GetNamespace()+"/"+meta.GetID().String(), meta.GetNamespacedName())
	})
	t.Run("Name", func(t *testing.T) {
		meta := &Meta{
			ID:          uuid.Must(uuid.NewV7()),
			Kind:        faker.UUIDHyphenated(),
			Namespace:   "default",
			Name:        faker.UUIDHyphenated(),
			Annotations: map[string]string{"key": "value"},
			Ports:       map[string][]Port{"out": {{Name: faker.UUIDHyphenated(), Port: "in"}}},
			Env:         map[string]Value{"env1": {Name: "value1", Data: "value1"}},
		}
		assert.Equal(t, meta.GetNamespace()+"/"+meta.GetName(), meta.GetNamespacedName())
	})
}

func TestMeta_IsBound(t *testing.T) {
	sec1 := &value.Value{
		ID: uuid.Must(uuid.NewV7()),
	}
	sec2 := &value.Value{
		ID: uuid.Must(uuid.NewV7()),
	}

	meta := &Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: faker.UUIDHyphenated(),
		Env: map[string]Value{
			"FOO": {
				ID:   sec1.ID,
				Data: "foo",
			},
		},
	}

	assert.True(t, meta.IsBound(sec1))
	assert.False(t, meta.IsBound(sec2))
}

func TestMeta_Bind(t *testing.T) {
	sec := &value.Value{
		ID:   uuid.Must(uuid.NewV7()),
		Data: faker.UUIDHyphenated(),
	}
	meta := &Meta{
		ID:   uuid.Must(uuid.NewV7()),
		Kind: faker.UUIDHyphenated(),
		Env: map[string]Value{
			"FOO": {
				ID:   sec.ID,
				Data: "{{ . }}",
			},
		},
	}

	err := meta.Bind(sec)
	assert.NoError(t, err)
	assert.Equal(t, sec.Data, meta.GetEnv()["FOO"].Data)
}
