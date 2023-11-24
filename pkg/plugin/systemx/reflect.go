package systemx

import (
	"context"

	"github.com/oklog/ulid/v2"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
)

type (
	ReflectNodeConfig struct {
		ID      ulid.ULID
		OP      string
		Storage *storage.Storage
	}

	ReflectNode struct {
		*node.OneToOneNode
		op      string
		storage *storage.Storage
	}

	ReflectSpec struct {
		scheme.SpecMeta `map:",inline"`
		OP              string `map:"op"`
	}
)

const (
	KindReflect = "reflect"
)

const (
	OPDelete = "delete"
	OPInsert = "insert"
	OPSelect = "select"
	OPUpdate = "update"
)

func NewReflectNode(config ReflectNodeConfig) *ReflectNode {
	id := config.ID
	op := config.OP
	storage := config.Storage

	n := &ReflectNode{
		op:      op,
		storage: storage,
	}
	n.OneToOneNode = node.NewOneToOneNode(node.OneToOneNodeConfig{
		ID:     id,
		Action: n.action,
	})

	return n
}

func (n *ReflectNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-proc.Done()
		cancel()
	}()

	inPayload := inPck.Payload()

	batch := true
	var examples []*primitive.Map
	if v, ok := inPayload.(*primitive.Map); ok {
		examples = append(examples, v)
		batch = false
	} else if v, ok := inPayload.(*primitive.Slice); ok {
		for i := 0; i < v.Len(); i++ {
			if e, ok := v.Get(i).(*primitive.Map); ok {
				examples = append(examples, e)
			}
		}
	}

	switch n.op {
	case OPDelete:
		filter, err := examplesToFilter(examples)
		if err != nil {
			return nil, packet.NewError(err, inPck)
		}

		specs, err := n.storage.FindMany(ctx, filter)
		if err != nil {
			return nil, packet.NewError(err, inPck)
		}

		var ids []ulid.ULID
		for _, spec := range specs {
			ids = append(ids, spec.GetID())
		}

		if _, err := n.storage.DeleteMany(ctx, storage.Where[ulid.ULID](scheme.KeyID).IN(ids...)); err != nil {
			return nil, packet.NewError(err, inPck)
		}

		if len(specs) == 0 {
			return nil, inPck
		}
		if outPayload, err := specsToExamples(specs, batch); err != nil {
			return nil, packet.NewError(err, inPck)
		} else {
			return packet.New(outPayload), nil
		}
	case OPInsert:
		specs := examplesToSpecs(examples)

		ids, err := n.storage.InsertMany(ctx, specs)
		if err != nil {
			return nil, packet.NewError(err, inPck)
		}

		specs, err = n.storage.FindMany(ctx, storage.Where[ulid.ULID](scheme.KeyID).IN(ids...), &database.FindOptions{
			Limit: lo.ToPtr(len(ids)),
		})
		if err != nil {
			return nil, packet.NewError(err, inPck)
		}

		if len(specs) == 0 {
			return nil, inPck
		}
		if outPayload, err := specsToExamples(specs, batch); err != nil {
			return nil, packet.NewError(err, inPck)
		} else {
			return packet.New(outPayload), nil
		}
	case OPSelect:
		filter, err := examplesToFilter(examples)
		if err != nil {
			return nil, packet.NewError(err, inPck)
		}

		specs, err := n.storage.FindMany(ctx, filter)
		if err != nil {
			return nil, packet.NewError(err, inPck)
		}

		if len(specs) == 0 {
			return nil, inPck
		}
		if outPayload, err := specsToExamples(specs, batch); err != nil {
			return nil, packet.NewError(err, inPck)
		} else {
			return packet.New(outPayload), nil
		}
	case OPUpdate:
		specs := examplesToSpecs(examples)

		var ids []ulid.ULID
		patches := map[ulid.ULID]*primitive.Map{}
		for i, spec := range specs {
			id := spec.GetID()

			if id != (ulid.ULID{}) {
				ids = append(ids, id)
				patches[id] = examples[i]
			}
		}

		specs, err := n.storage.FindMany(ctx, storage.Where[ulid.ULID](scheme.KeyID).IN(ids...), &database.FindOptions{
			Limit: lo.ToPtr(len(ids)),
		})
		if err != nil {
			return nil, packet.NewError(err, inPck)
		}

		var merges []scheme.Spec
		for _, spec := range specs {
			unstructured := scheme.NewUnstructured(nil)
			if err := unstructured.Marshal(spec); err != nil {
				return nil, packet.NewError(err, inPck)
			}

			patch := patches[spec.GetID()]

			doc := unstructured.Doc()
			for _, k := range patch.Keys() {
				doc = doc.Set(k, patch.GetOr(k, nil))
			}

			merges = append(merges, scheme.NewUnstructured(doc))
		}

		if _, err := n.storage.UpdateMany(ctx, merges); err != nil {
			return nil, packet.NewError(err, inPck)
		}

		specs, err = n.storage.FindMany(ctx, storage.Where[ulid.ULID](scheme.KeyID).IN(ids...), &database.FindOptions{
			Limit: lo.ToPtr(len(ids)),
		})
		if err != nil {
			return nil, packet.NewError(err, inPck)
		}

		if len(specs) == 0 {
			return nil, inPck
		}
		if outPayload, err := specsToExamples(specs, batch); err != nil {
			return nil, packet.NewError(err, inPck)
		} else {
			return packet.New(outPayload), nil
		}
	}

	return inPck, nil
}

func examplesToFilter(examples []*primitive.Map) (*storage.Filter, error) {
	var filter *storage.Filter
	for _, example := range examples {
		var sub *storage.Filter

		spec := scheme.SpecMeta{}
		unstructured := scheme.NewUnstructured(example)
		if err := unstructured.Unmarshal(&spec); err != nil {
			return nil, err
		}

		if spec.ID != (ulid.ULID{}) {
			sub = sub.And(storage.Where[ulid.ULID](scheme.KeyID).EQ(spec.ID))
		}
		if spec.Kind != "" {
			sub = sub.And(storage.Where[string](scheme.KeyKind).EQ(spec.Kind))
		}
		if spec.Name != "" {
			sub = sub.And(storage.Where[string](scheme.KeyName).EQ(spec.Name))
		}
		if spec.Namespace != "" {
			sub = sub.And(storage.Where[string](scheme.KeyName).EQ(spec.Namespace))
		}

		filter = filter.And(sub)
	}

	return filter, nil
}

func examplesToSpecs(examples []*primitive.Map) []scheme.Spec {
	var specs []scheme.Spec
	for _, example := range examples {
		unstructured := scheme.NewUnstructured(example)
		specs = append(specs, unstructured)
	}
	return specs
}

func specsToExamples(specs []scheme.Spec, batch bool) (primitive.Object, error) {
	if batch || len(specs) > 1 {
		return primitive.MarshalText(specs)
	} else {
		return primitive.MarshalText(specs[0])
	}
}
