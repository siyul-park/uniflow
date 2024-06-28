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
	"github.com/stretchr/testify/assert"
)

func TestCommend_Execute(t *testing.T) {
	s := scheme.New()
	h := hook.New()
	db := memdb.New("")
	fsys := make(fstest.MapFS)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	output := new(bytes.Buffer)

	cmd := NewCommand(Config{
		Scheme:   s,
		Hook:     h,
		FS:       fsys,
		Database: db,
	})
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetContext(ctx)

	err := cmd.Execute()
	assert.NoError(t, err)
}
