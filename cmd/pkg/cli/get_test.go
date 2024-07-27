package cli

import (
	"bytes"
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/stretchr/testify/assert"
)

func TestGetCommand_Execute(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	st := spec.NewStore()

	kind := faker.UUIDHyphenated()

	meta := &spec.Meta{
		Kind:      kind,
		Namespace: spec.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}

	_, _ = st.Store(ctx, meta)

	output := new(bytes.Buffer)

	cmd := NewGetCommand(GetConfig{
		SpecStore: st,
	})
	cmd.SetOut(output)
	cmd.SetErr(output)

	err := cmd.Execute()
	assert.NoError(t, err)

	assert.Contains(t, output.String(), meta.ID.String())
}
