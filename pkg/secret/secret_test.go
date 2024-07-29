package secret

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMatch(t *testing.T) {
	id1 := uuid.Must(uuid.NewV7())
	id2 := uuid.Must(uuid.NewV7())

	spc := &Secret{ID: id1, Namespace: "default", Name: "secret1"}
	examples := []*Secret{
		{ID: id1, Namespace: "default", Name: "secret1"},
		{ID: id1},
		{Namespace: "default", Name: "secret1"},
		{ID: id2, Namespace: "default", Name: "secret2"},
		{ID: id2},
		{Namespace: "default", Name: "secret2"},
	}

	expeced := []*Secret{examples[0], examples[1], examples[2]}

	assert.Equal(t, expeced, Match(spc, examples...))
}
