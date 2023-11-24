package database

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestMergeUpdateOptions(t *testing.T) {
	opt := MergeUpdateOptions([]*UpdateOptions{
		nil,
		lo.ToPtr(UpdateOptions{
			Upsert: nil,
		}),
		lo.ToPtr(UpdateOptions{
			Upsert: lo.ToPtr(true),
		}),
	})

	assert.Equal(t, lo.ToPtr(UpdateOptions{
		Upsert: lo.ToPtr(true),
	}), opt)
}

func TestMergeFindOptions(t *testing.T) {
	opt := MergeFindOptions([]*FindOptions{
		nil,
		lo.ToPtr(FindOptions{
			Limit: lo.ToPtr(1),
		}),
		lo.ToPtr(FindOptions{
			Skip: lo.ToPtr(1),
		}),
		lo.ToPtr(FindOptions{
			Sorts: []Sort{{Key: "", Order: OrderASC}},
		}),
	})

	assert.Equal(t, lo.ToPtr(FindOptions{
		Limit: lo.ToPtr(1),
		Skip:  lo.ToPtr(1),
		Sorts: []Sort{{Key: "", Order: OrderASC}},
	}), opt)
}
