package chart

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/value"
	"github.com/stretchr/testify/assert"
)

func TestChart_IsBound(t *testing.T) {
	t.Run("NoValues", func(t *testing.T) {
		chrt := &Chart{
			ID: uuid.Must(uuid.NewV7()),

			Env: map[string][]spec.Value{
				"FOO": {
					{
						ID:   uuid.Must(uuid.NewV7()),
						Data: "foo",
					},
				},
			},
		}
		assert.False(t, chrt.IsBound())
	})

	t.Run("WithValues", func(t *testing.T) {
		sec1 := &value.Value{
			ID: uuid.Must(uuid.NewV7()),
		}
		sec2 := &value.Value{
			ID: uuid.Must(uuid.NewV7()),
		}
		chrt := &Chart{
			ID: uuid.Must(uuid.NewV7()),

			Env: map[string][]spec.Value{
				"FOO": {
					{
						ID:   sec1.ID,
						Data: "foo",
					},
				},
			},
		}
		assert.True(t, chrt.IsBound(sec1))
		assert.False(t, chrt.IsBound(sec2))
	})
}

func TestChart_Bind(t *testing.T) {
	t.Run("NoMatchingValue", func(t *testing.T) {
		scrt := &value.Value{
			ID:   uuid.Must(uuid.NewV7()),
			Data: "foo",
		}
		chrt := &Chart{
			ID: uuid.Must(uuid.NewV7()),
			Env: map[string][]spec.Value{
				"FOO": {
					{
						ID:   uuid.Must(uuid.NewV7()),
						Data: "{{ . }}",
					},
				},
			},
		}
		err := chrt.Bind(scrt)
		assert.Error(t, err)
	})

	t.Run("MatchingValue", func(t *testing.T) {
		scrt := &value.Value{
			ID:   uuid.Must(uuid.NewV7()),
			Data: "foo",
		}
		chrt := &Chart{
			ID: uuid.Must(uuid.NewV7()),
			Env: map[string][]spec.Value{
				"FOO": {
					{
						ID:   scrt.ID,
						Data: "{{ . }}",
					},
				},
			},
		}
		err := chrt.Bind(scrt)
		assert.NoError(t, err)
		assert.Equal(t, "foo", chrt.GetEnv()["FOO"][0].Data)
	})
}

func TestChart_Build(t *testing.T) {
	t.Run("NoEnv", func(t *testing.T) {
		chrt := &Chart{
			ID:   uuid.Must(uuid.NewV7()),
			Name: faker.UUIDHyphenated(),
			Specs: []*spec.Unstructured{
				{
					Meta: spec.Meta{
						ID:   uuid.Must(uuid.NewV7()),
						Kind: faker.UUIDHyphenated(),
					},
					Fields: map[string]any{
						"foo": "{{ .FOO }}",
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
	})

	t.Run("WithEnv", func(t *testing.T) {
		chrt := &Chart{
			ID:   uuid.Must(uuid.NewV7()),
			Name: faker.UUIDHyphenated(),
			Specs: []*spec.Unstructured{
				{
					Meta: spec.Meta{
						ID:   uuid.Must(uuid.NewV7()),
						Kind: faker.UUIDHyphenated(),
					},
					Fields: map[string]any{
						"foo": "{{ .FOO }}",
					},
				},
			},
			Env: map[string][]spec.Value{
				"FOO": {
					{
						Data: "foo",
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
	})
}

func TestChart_ID(t *testing.T) {
	chrt := New()
	id := uuid.Must(uuid.NewV7())
	chrt.SetID(id)
	assert.Equal(t, id, chrt.GetID())
}

func TestChart_Namespace(t *testing.T) {
	chrt := New()
	namespace := "test-namespace"
	chrt.SetNamespace(namespace)
	assert.Equal(t, namespace, chrt.GetNamespace())
}

func TestChart_Name(t *testing.T) {
	chrt := New()
	name := "test-chart"
	chrt.SetName(name)
	assert.Equal(t, name, chrt.GetName())
}

func TestChart_Annotations(t *testing.T) {
	chrt := New()
	annotations := map[string]string{"key": "value"}
	chrt.SetAnnotations(annotations)
	assert.Equal(t, annotations, chrt.GetAnnotations())
}

func TestChart_Specs(t *testing.T) {
	chrt := New()
	specs := []*spec.Unstructured{
		{
			Meta: spec.Meta{
				ID:   uuid.Must(uuid.NewV7()),
				Kind: "test",
			},
		},
	}
	chrt.SetSpecs(specs)
	assert.Equal(t, specs, chrt.GetSpecs())
}

func TestChart_Inbounds(t *testing.T) {
	chrt := New()
	ports := map[string][]spec.Port{
		"http": {
			{
				ID:   uuid.Must(uuid.NewV7()),
				Name: "http",
				Port: "80",
			},
		},
	}
	chrt.SetInbounds(ports)
	assert.Equal(t, ports, chrt.GetInbounds())
}

func TestChart_Outbounds(t *testing.T) {
	chrt := New()
	ports := map[string][]spec.Port{
		"http": {
			{
				ID:   uuid.Must(uuid.NewV7()),
				Name: "http",
				Port: "80",
			},
		},
	}
	chrt.SetOutbounds(ports)
	assert.Equal(t, ports, chrt.GetOutbounds())
}

func TestChart_Env(t *testing.T) {
	chrt := New()
	env := map[string][]spec.Value{
		"FOO": {
			{
				ID:   uuid.Must(uuid.NewV7()),
				Data: "bar",
			},
		},
	}
	chrt.SetEnv(env)
	assert.Equal(t, env, chrt.GetEnv())
}
