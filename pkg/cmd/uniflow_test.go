package cmd

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

func TestExecute(t *testing.T) {
	s := scheme.New()
	h := hook.New()
	db := memdb.New("")
	fsys := make(fstest.MapFS)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	output := new(bytes.Buffer)

	cmd := NewUniflowCommand(UniflowConfig{
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
