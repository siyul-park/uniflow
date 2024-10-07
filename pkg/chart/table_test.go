package chart

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/stretchr/testify/assert"
)

func TestTable_Insert(t *testing.T) {
	tb := NewTable()
	defer tb.Close()

	chrt1 := &Chart{
		ID:        uuid.Must(uuid.NewV7()),
		Namespace: resource.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
		Specs:     []spec.Spec{},
	}
	chrt2 := &Chart{
		ID:        uuid.Must(uuid.NewV7()),
		Namespace: resource.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
		Specs: []spec.Spec{
			&spec.Meta{
				Kind:      chrt1.GetName(),
				Namespace: resource.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			},
		},
	}

	err := tb.Insert(chrt1)
	assert.NoError(t, err)
	assert.NotNil(t, tb.Lookup(chrt1.GetID()))

	err = tb.Insert(chrt2)
	assert.NoError(t, err)
	assert.NotNil(t, tb.Lookup(chrt2.GetID()))
}

func TestTable_Free(t *testing.T) {
	tb := NewTable()
	defer tb.Close()

	chrt := &Chart{
		ID:        uuid.Must(uuid.NewV7()),
		Namespace: resource.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
		Specs: []spec.Spec{
			&spec.Meta{
				Kind:      faker.UUIDHyphenated(),
				Namespace: resource.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			},
		},
	}

	err := tb.Insert(chrt)
	assert.NoError(t, err)

	ok, err := tb.Free(chrt.GetID())
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestTable_Lookup(t *testing.T) {
	tb := NewTable()
	defer tb.Close()

	chrt := &Chart{
		ID:        uuid.Must(uuid.NewV7()),
		Namespace: resource.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
		Specs: []spec.Spec{
			&spec.Meta{
				Kind:      faker.UUIDHyphenated(),
				Namespace: resource.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			},
		},
	}

	err := tb.Insert(chrt)
	assert.NoError(t, err)
	assert.Equal(t, chrt, tb.Lookup(chrt.GetID()))
}

func TestTable_Hook(t *testing.T) {
	loaded := 0
	unloaded := 0

	tb := NewTable(TableOption{
		LoadHooks: []LoadHook{
			LoadFunc(func(_ *Chart) error {
				loaded += 1
				return nil
			}),
		},
		UnloadHooks: []UnloadHook{
			UnloadFunc(func(_ *Chart) error {
				unloaded += 1
				return nil
			}),
		},
	})
	defer tb.Close()

	chrt1 := &Chart{
		ID:        uuid.Must(uuid.NewV7()),
		Namespace: resource.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
		Specs:     []spec.Spec{},
	}
	chrt2 := &Chart{
		ID:        uuid.Must(uuid.NewV7()),
		Namespace: resource.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
		Specs: []spec.Spec{
			&spec.Meta{
				Kind:      chrt1.GetName(),
				Namespace: resource.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			},
		},
	}

	err := tb.Insert(chrt2)
	assert.NoError(t, err)
	assert.Equal(t, 0, loaded)
	assert.Equal(t, 0, unloaded)

	err = tb.Insert(chrt1)
	assert.NoError(t, err)
	assert.Equal(t, 2, loaded)
	assert.Equal(t, 0, unloaded)
}
