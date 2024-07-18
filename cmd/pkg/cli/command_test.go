package cli

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestCommend_Execute(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	s := scheme.New()
	h := hook.New()
	st := spec.NewStore(memdb.NewCollection(""))
	fsys := afero.NewMemMapFs()

	output := new(bytes.Buffer)

	cmd := NewCommand(Config{
		Scheme: s,
		Hook:   h,
		Store:  st,
		FS:     fsys,
	})
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetContext(ctx)

	err := cmd.Execute()
	assert.NoError(t, err)
}
