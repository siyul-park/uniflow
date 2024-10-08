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

func TestIsBound(t *testing.T) {
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

	assert.True(t, IsBound(chrt, sec1))
	assert.False(t, IsBound(chrt, sec2))
}

func TestBind(t *testing.T) {
	sec := &secret.Secret{
		ID:   uuid.Must(uuid.NewV7()),
		Data: "foo",
	}

	chrt := &Chart{
		ID: uuid.Must(uuid.NewV7()),
		Env: map[string][]Value{
			"FOO": {
				{
					ID:    sec.ID,
					Value: "{{ . }}",
				},
			},
		},
	}

	bind, err := Bind(chrt, sec)
	assert.NoError(t, err)
	assert.Equal(t, "foo", bind.GetEnv()["FOO"][0].Value)
	assert.True(t, IsBound(bind, sec))
}

func TestChart_GetSet(t *testing.T) {
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
		Ports: map[string][]Port{"out": {{Name: faker.Word(), Port: "in"}}},
		Env:   map[string][]Value{"env1": {{Name: "secret1", Value: "value1"}}},
	}

	assert.Equal(t, chrt.ID, chrt.GetID())
	assert.Equal(t, chrt.Namespace, chrt.GetNamespace())
	assert.Equal(t, chrt.Name, chrt.GetName())
	assert.Equal(t, chrt.Annotations, chrt.GetAnnotations())
	assert.Equal(t, chrt.Specs, chrt.GetSpecs())
	assert.Equal(t, chrt.Ports, chrt.GetPorts())
	assert.Equal(t, chrt.Env, chrt.GetEnv())
}
