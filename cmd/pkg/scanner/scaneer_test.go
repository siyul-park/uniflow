package scanner

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestScanner_Scan(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	st := spec.NewMemStore()
	fsys := afero.NewMemMapFs()

	kind := faker.UUIDHyphenated()

	filename := "spec.json"

	meta := &spec.Meta{
		Kind:      kind,
		Namespace: spec.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}

	data, _ := json.Marshal(meta)

	_, _ = st.Store(ctx, meta)

	f, _ := fsys.Create(filename)
	f.Write(data)

	scanner := New().
		Store(st).
		Namespace(spec.DefaultNamespace).
		FS(fsys).
		Filename(filename)

	specs, err := scanner.Scan(ctx)
	assert.NoError(t, err)
	assert.Len(t, specs, 1)
}
