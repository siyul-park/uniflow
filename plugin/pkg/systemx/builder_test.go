package systemx

import (
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
	"github.com/stretchr/testify/assert"
)

func TestAddToScheme(t *testing.T) {
	s := scheme.New()
	st, _ := storage.New(context.Background(), storage.Config{
		Scheme:   s,
		Database: memdb.New(faker.UUIDHyphenated()),
	})

	err := AddToScheme(st)(s)
	assert.NoError(t, err)

	_, ok := s.Codec(KindReflect)
	assert.True(t, ok)
}
