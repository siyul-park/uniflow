package symbol

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"

	"github.com/siyul-park/uniflow/pkg/meta"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/spec"
)

func TestSymbol_Getter(t *testing.T) {
	n := node.NewOneToOneNode(nil)
	defer n.Close()

	meta := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      faker.UUIDHyphenated(),
		Namespace: meta.DefaultNamespace,
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

	require.Equal(t, meta.GetID(), sb.ID())
	require.Equal(t, meta.GetKind(), sb.Kind())
	require.Equal(t, meta.GetNamespace(), sb.Namespace())
	require.Equal(t, meta.GetName(), sb.Name())
	require.Equal(t, meta.GetNamespacedName(), sb.NamespacedName())
	require.Equal(t, meta.GetAnnotations(), sb.Annotations())
	require.Equal(t, meta.GetPorts(), sb.Ports())
	require.Equal(t, meta.GetEnv(), sb.Env())
	require.Equal(t, n.In(node.PortIn), sb.In(node.PortIn))
	require.Equal(t, n.Out(node.PortOut), sb.Out(node.PortOut))
	require.Contains(t, sb.Ins(), node.PortIn)
	require.Contains(t, sb.Outs(), node.PortOut)
}

func TestSymbol_Setter(t *testing.T) {
	n := node.NewOneToOneNode(nil)
	defer n.Close()

	meta := &spec.Meta{
		ID:        uuid.Must(uuid.NewV7()),
		Kind:      faker.UUIDHyphenated(),
		Namespace: meta.DefaultNamespace,
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
	require.Equal(t, id, sb.ID())

	kind := faker.UUIDHyphenated()
	sb.SetKind(kind)
	require.Equal(t, kind, sb.Kind())

	namespace := faker.UUIDHyphenated()
	sb.SetNamespace(namespace)
	require.Equal(t, namespace, sb.Namespace())

	name := faker.UUIDHyphenated()
	sb.SetName(name)
	require.Equal(t, name, sb.Name())

	annotations := map[string]string{
		faker.UUIDHyphenated(): faker.UUIDHyphenated(),
	}
	sb.SetAnnotations(annotations)
	require.Equal(t, annotations, sb.Annotations())

	ports := map[string][]spec.Port{
		node.PortIn: {
			{
				ID:   uuid.Must(uuid.NewV7()),
				Port: node.PortOut,
			},
		},
	}
	sb.SetPorts(ports)
	require.Equal(t, ports, sb.Ports())

	env := map[string]spec.Value{
		faker.UUIDHyphenated(): {
			Data: faker.UUIDHyphenated(),
		},
	}
	sb.SetEnv(env)
	require.Equal(t, env, sb.Env())
}
