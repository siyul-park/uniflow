package chart

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/secret"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/stretchr/testify/assert"
)

func TestChart_IsBound(t *testing.T) {
	sec1 := &secret.Secret{
		ID: uuid.Must(uuid.NewV7()),
	}
	sec2 := &secret.Secret{
		ID: uuid.Must(uuid.NewV7()),
	}

	chrt := &Chart{
		ID: uuid.Must(uuid.NewV7()),
		Env: map[string][]Value{
			"FOO": {
				{
					ID:    sec1.ID,
					Value: "foo",
				},
			},
		},
	}

	assert.True(t, chrt.IsBound(sec1))
	assert.False(t, chrt.IsBound(sec2))
}

func TestChart_Bind(t *testing.T) {
	scrt := &secret.Secret{
		ID:   uuid.Must(uuid.NewV7()),
		Data: "foo",
	}

	chrt := &Chart{
		ID: uuid.Must(uuid.NewV7()),
		Env: map[string][]Value{
			"FOO": {
				{
					ID:    scrt.ID,
					Value: "{{ . }}",
				},
			},
		},
	}

	err := chrt.Bind(scrt)
	assert.NoError(t, err)
	assert.Equal(t, "foo", chrt.GetEnv()["FOO"][0].Value)
}

func TestChart_Build(t *testing.T) {
	chrt := &Chart{
		ID:   uuid.Must(uuid.NewV7()),
		Name: faker.UUIDHyphenated(),
		Specs: []spec.Spec{
			&spec.Unstructured{
				Meta: spec.Meta{
					ID:   uuid.Must(uuid.NewV7()),
					Kind: faker.UUIDHyphenated(),
				},
				Fields: map[string]any{
					"foo": "{{ .FOO }}",
				},
			},
		},
		Env: map[string][]Value{
			"FOO": {
				{
					Value: "foo",
				},
			},
		},
	}

	meta := &spec.Meta{
		Kind:      chrt.GetName(),
		Namespace: resource.DefaultNamespace,
	}

	specs, err := chrt.Build(meta)
	assert.NoError(t, err)
	assert.Len(t, specs, 1)
}

func TestChart_Get(t *testing.T) {
	chrt := &Chart{
		ID:          uuid.Must(uuid.NewV7()),
		Namespace:   "default",
		Name:        faker.Word(),
		Annotations: map[string]string{"key": "value"},
		Specs: []spec.Spec{
			&spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      faker.UUIDHyphenated(),
				Namespace: resource.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			},
		},
		Inbound: map[string][]Port{"out": {{Name: faker.Word(), Port: "in"}}},
		Env:     map[string][]Value{"env1": {{Name: "secret1", Value: "value1"}}},
	}

	assert.Equal(t, chrt.ID, chrt.GetID())
	assert.Equal(t, chrt.Namespace, chrt.GetNamespace())
	assert.Equal(t, chrt.Name, chrt.GetName())
	assert.Equal(t, chrt.Annotations, chrt.GetAnnotations())
	assert.Equal(t, chrt.Specs, chrt.GetSpecs())
	assert.Equal(t, chrt.Inbound, chrt.GetInbound())
	assert.Equal(t, chrt.Env, chrt.GetEnv())
}

func TestChart_Set(t *testing.T) {
	chrt := New()

	id := uuid.Must(uuid.NewV7())
	namespace := "test-namespace"
	name := "test-chart"
	annotations := map[string]string{"key": "value"}
	specs := []spec.Spec{
		&spec.Unstructured{
			Meta: spec.Meta{
				ID:   uuid.Must(uuid.NewV7()),
				Kind: "test",
			},
		},
	}
	ports := map[string][]Port{
		"http": {
			{
				ID:   uuid.Must(uuid.NewV7()),
				Name: "http",
				Port: "80",
			},
		},
	}
	env := map[string][]Value{
		"FOO": {
			{
				ID:    uuid.Must(uuid.NewV7()),
				Value: "bar",
			},
		},
	}

	chrt.SetID(id)
	assert.Equal(t, id, chrt.GetID())

	chrt.SetNamespace(namespace)
	assert.Equal(t, namespace, chrt.GetNamespace())

	chrt.SetName(name)
	assert.Equal(t, name, chrt.GetName())

	chrt.SetAnnotations(annotations)
	assert.Equal(t, annotations, chrt.GetAnnotations())

	chrt.SetSpecs(specs)
	assert.Equal(t, specs, chrt.GetSpecs())

	chrt.SetInbound(ports)
	assert.Equal(t, ports, chrt.GetInbound())

	chrt.SetOutbound(ports)
	assert.Equal(t, ports, chrt.GetOutbound())

	chrt.SetEnv(env)
	assert.Equal(t, env, chrt.GetEnv())
}
