package util

import "github.com/mitchellh/hashstructure/v2"

func Hash(val any) (uint64, error) {
	return hashstructure.Hash(val, hashstructure.FormatV2, nil)
}
