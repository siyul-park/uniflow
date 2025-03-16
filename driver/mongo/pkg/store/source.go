package store

import (
	"github.com/siyul-park/uniflow/pkg/store"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Source struct {
	database *mongo.Database
}

var _ store.Source = (*Source)(nil)

func NewSource(database *mongo.Database) *Source {
	return &Source{database: database}
}

func (s *Source) Open(name string) (store.Store, error) {
	return New(s.database.Collection(name)), nil
}

func (s *Source) Close() error {
	return nil
}
