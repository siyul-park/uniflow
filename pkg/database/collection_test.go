package database

import (
	"testing"

	"github.com/siyul-park/uniflow/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestMergeUpdateOptions(t *testing.T) {
	opt := MergeUpdateOptions([]*UpdateOptions{
		nil,
		util.Ptr(UpdateOptions{
			Upsert: nil,
		}),
		util.Ptr(UpdateOptions{
			Upsert: util.Ptr(true),
		}),
	})

	assert.Equal(t, util.Ptr(UpdateOptions{
		Upsert: util.Ptr(true),
	}), opt)
}

func TestMergeFindOptions(t *testing.T) {
	opt := MergeFindOptions([]*FindOptions{
		nil,
		util.Ptr(FindOptions{
			Limit: util.Ptr(1),
		}),
		util.Ptr(FindOptions{
			Skip: util.Ptr(1),
		}),
		util.Ptr(FindOptions{
			Sorts: []Sort{{Key: "", Order: OrderASC}},
		}),
	})

	assert.Equal(t, util.Ptr(FindOptions{
		Limit: util.Ptr(1),
		Skip:  util.Ptr(1),
		Sorts: []Sort{{Key: "", Order: OrderASC}},
	}), opt)
}
