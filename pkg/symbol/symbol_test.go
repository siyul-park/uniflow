package symbol

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/stretchr/testify/assert"
)

func TestSymbol_Getter(t *testing.T) {
	n := node.NewOneToOneNode(nil)
	defer n.Close()

	meta := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      faker.UUIDHyphenated(),
		Namespace: resource.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
		Annotations: map[string]string{
			faker.UUIDHyphenated(): faker.UUIDHyphenated(),
		},
		Ports: map[string][]spec.Port{
			node.PortOut: {
				{
					ID:   uuid.Must(uuid.NewV7()),
					Port: node.PortIn,
				},
			},
		},
	}

	sb := &Symbol{
		Spec: meta,
		Node: n,
	}

	assert.Equal(t, meta.GetID(), sb.ID())
	assert.Equal(t, meta.GetKind(), sb.Kind())
	assert.Equal(t, meta.GetNamespace(), sb.Namespace())
	assert.Equal(t, meta.GetName(), sb.Name())
	assert.Equal(t, meta.GetAnnotations(), sb.Annotations())
	assert.Equal(t, meta.GetPorts(), sb.Ports())
	assert.Equal(t, meta.GetEnv(), sb.Env())
	assert.Equal(t, n.In(node.PortIn), sb.In(node.PortIn))
	assert.Equal(t, n.Out(node.PortOut), sb.Out(node.PortOut))
	assert.Contains(t, sb.Ins(), node.PortIn)
	assert.Contains(t, sb.Outs(), node.PortOut)
}

func TestSymbol_Setter(t *testing.T) {
	n := node.NewOneToOneNode(nil)
	defer n.Close()

	meta := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      faker.UUIDHyphenated(),
		Namespace: resource.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
		Annotations: map[string]string{
			faker.UUIDHyphenated(): faker.UUIDHyphenated(),
		},
		Ports: map[string][]spec.Port{
			node.PortOut: {
				{
					ID:   uuid.Must(uuid.NewV7()),
					Port: node.PortIn,
				},
			},
		},
	}

	sb := &Symbol{
		Spec: meta,
		Node: n,
	}

	id := uuid.Must(uuid.NewV7())
	sb.SetID(id)
	assert.Equal(t, id, sb.ID())

	kind := faker.UUIDHyphenated()
	sb.SetKind(kind)
	assert.Equal(t, kind, sb.Kind())

	namespace := faker.UUIDHyphenated()
	sb.SetNamespace(namespace)
	assert.Equal(t, namespace, sb.Namespace())

	name := faker.UUIDHyphenated()
	sb.SetName(name)
	assert.Equal(t, name, sb.Name())

	annotations := map[string]string{
		faker.UUIDHyphenated(): faker.UUIDHyphenated(),
	}
	sb.SetAnnotations(annotations)
	assert.Equal(t, annotations, sb.Annotations())

	ports := map[string][]spec.Port{
		node.PortIn: {
			{
				ID:   uuid.Must(uuid.NewV7()),
				Port: node.PortOut,
			},
		},
	}
	sb.SetPorts(ports)
	assert.Equal(t, ports, sb.Ports())

	env := map[string]spec.Value{
		faker.UUIDHyphenated(): {
			Data: faker.UUIDHyphenated(),
		},
	}
	sb.SetEnv(env)
	assert.Equal(t, env, sb.Env())
}
