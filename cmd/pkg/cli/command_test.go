package cli

import (
	"bytes"
	"context"
	"testing"
	"testing/fstest"
	"time"

	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/store"
	"github.com/stretchr/testify/assert"
)

func TestCommend_Execute(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	s := scheme.New()
	h := hook.New()
	st, _ := store.New(ctx, memdb.NewCollection(""))
	fsys := make(fstest.MapFS)

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
