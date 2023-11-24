package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEqual(t *testing.T) {
	var testCase = []struct {
		when   []any
		expect bool
	}{
		{
			when:   []any{uint8(0), uint8(0)},
			expect: true,
		},
		{
			when:   []any{uint16(0), uint16(0)},
			expect: true,
		},
		{
			when:   []any{uint32(0), uint32(0)},
			expect: true,
		},
		{
			when:   []any{uint64(0), uint64(0)},
			expect: true,
		},
		{
			when:   []any{int8(0), int8(0)},
			expect: true,
		},
		{
			when:   []any{int16(0), int16(0)},
			expect: true,
		},
		{
			when:   []any{int32(0), int32(0)},
			expect: true,
		},
		{
			when:   []any{int64(0), int64(0)},
			expect: true,
		},
		{
			when:   []any{int8(0), uint8(0)},
			expect: true,
		},
		{
			when:   []any{int16(0), uint16(0)},
			expect: true,
		},
		{
			when:   []any{int32(0), uint32(0)},
			expect: true,
		},
		{
			when:   []any{int64(0), uint64(0)},
			expect: true,
		},

		{
			when:   []any{0, 1},
			expect: false,
		},
		{
			when:   []any{false, true},
			expect: false,
		},
		{
			when:   []any{"0", "1"},
			expect: false,
		},
	}

	for _, tc := range testCase {
		r := Equal(tc.when[0], tc.when[1])
		assert.Equal(t, tc.expect, r)
	}
}

func TestCompare(t *testing.T) {
	var testCase1 = []struct {
		when   []any
		expect int
	}{
		{
			when:   []any{uint8(0), uint8(0)},
			expect: 0,
		},
		{
			when:   []any{uint8(1), uint8(0)},
			expect: 1,
		},
		{
			when:   []any{uint8(0), uint8(1)},
			expect: -1,
		},
		{
			when:   []any{uint16(0), uint16(0)},
			expect: 0,
		},
		{
			when:   []any{uint16(1), uint16(0)},
			expect: 1,
		},
		{
			when:   []any{uint16(0), uint16(1)},
			expect: -1,
		},
		{
			when:   []any{uint32(0), uint32(0)},
			expect: 0,
		},
		{
			when:   []any{uint32(1), uint32(0)},
			expect: 1,
		},
		{
			when:   []any{uint32(0), uint32(1)},
			expect: -1,
		},
		{
			when:   []any{uint64(0), uint64(0)},
			expect: 0,
		},
		{
			when:   []any{uint64(1), uint64(0)},
			expect: 1,
		},
		{
			when:   []any{uint64(0), uint64(1)},
			expect: -1,
		},
		{
			when:   []any{int8(0), int8(0)},
			expect: 0,
		},
		{
			when:   []any{int8(1), int8(0)},
			expect: 1,
		},
		{
			when:   []any{int8(0), int8(1)},
			expect: -1,
		},
		{
			when:   []any{int16(0), int16(0)},
			expect: 0,
		},
		{
			when:   []any{int16(1), int16(0)},
			expect: 1,
		},
		{
			when:   []any{int16(0), int16(1)},
			expect: -1,
		},
		{
			when:   []any{int32(0), int32(0)},
			expect: 0,
		},
		{
			when:   []any{int32(1), int32(0)},
			expect: 1,
		},
		{
			when:   []any{int32(0), int32(1)},
			expect: -1,
		},
		{
			when:   []any{int64(0), int64(0)},
			expect: 0,
		},
		{
			when:   []any{int64(1), int64(0)},
			expect: 1,
		},
		{
			when:   []any{int64(0), int64(1)},
			expect: -1,
		},
		{
			when:   []any{float32(0), float32(0)},
			expect: 0,
		},
		{
			when:   []any{float32(1), float32(0)},
			expect: 1,
		},
		{
			when:   []any{float32(0), float32(1)},
			expect: -1,
		},
		{
			when:   []any{float64(0), float64(0)},
			expect: 0,
		},
		{
			when:   []any{float64(1), float64(0)},
			expect: 1,
		},
		{
			when:   []any{float64(0), float64(1)},
			expect: -1,
		},
		{
			when:   []any{"0", "0"},
			expect: 0,
		},
		{
			when:   []any{"1", "0"},
			expect: 1,
		},
		{
			when:   []any{"0", "1"},
			expect: -1,
		},
		{
			when:   []any{0, 0},
			expect: 0,
		},
		{
			when:   []any{1, 0},
			expect: 1,
		},
		{
			when:   []any{0, 1},
			expect: -1,
		},
		{
			when:   []any{uint(0), uint(0)},
			expect: 0,
		},
		{
			when:   []any{uint(1), uint(0)},
			expect: 1,
		},
		{
			when:   []any{uint(0), uint(1)},
			expect: -1,
		},
		{
			when:   []any{uintptr(0), uintptr(0)},
			expect: 0,
		},
		{
			when:   []any{uintptr(1), uintptr(0)},
			expect: 1,
		},
		{
			when:   []any{uintptr(0), uintptr(1)},
			expect: -1,
		},
		{
			when:   []any{nil, 0},
			expect: -1,
		},
		{
			when:   []any{0, nil},
			expect: 1,
		},
		{
			when:   []any{nil, nil},
			expect: 0,
		},
	}

	for _, tc := range testCase1 {
		r := Compare(tc.when[0], tc.when[1])
		assert.Equal(t, tc.expect, r)
	}

	var testCase2 = []struct {
		whenX  any
		whenY  any
		expect int
		ok     bool
	}{
		{
			whenX:  []uint8{uint8(0), uint8(0)},
			whenY:  []uint8{uint8(0), uint8(0)},
			expect: 0,
		},
		{
			whenX:  []uint8{uint8(0), uint8(1)},
			whenY:  []uint8{uint8(0), uint8(0)},
			expect: 1,
		},
		{
			whenX:  []uint8{uint8(0), uint8(1)},
			whenY:  []uint8{uint8(1), uint8(0)},
			expect: -1,
		},

		{
			whenX:  []uint16{uint16(0), uint16(0)},
			whenY:  []uint16{uint16(0), uint16(0)},
			expect: 0,
		},
		{
			whenX:  []uint16{uint16(0), uint16(1)},
			whenY:  []uint16{uint16(0), uint16(0)},
			expect: 1,
		},
		{
			whenX:  []uint16{uint16(0), uint16(1)},
			whenY:  []uint16{uint16(1), uint16(0)},
			expect: -1,
		},

		{
			whenX:  []uint32{uint32(0), uint32(0)},
			whenY:  []uint32{uint32(0), uint32(0)},
			expect: 0,
		},
		{
			whenX:  []uint32{uint32(0), uint32(1)},
			whenY:  []uint32{uint32(0), uint32(0)},
			expect: 1,
		},
		{
			whenX:  []uint32{uint32(0), uint32(1)},
			whenY:  []uint32{uint32(1), uint32(0)},
			expect: -1,
		},

		{
			whenX:  []uint64{uint64(0), uint64(0)},
			whenY:  []uint64{uint64(0), uint64(0)},
			expect: 0,
		},
		{
			whenX:  []uint64{uint64(0), uint64(1)},
			whenY:  []uint64{uint64(0), uint64(0)},
			expect: 1,
		},
		{
			whenX:  []uint64{uint64(0), uint64(1)},
			whenY:  []uint64{uint64(1), uint64(0)},
			expect: -1,
		},

		{
			whenX:  []int8{int8(0), int8(0)},
			whenY:  []int8{int8(0), int8(0)},
			expect: 0,
		},
		{
			whenX:  []int8{int8(0), int8(1)},
			whenY:  []int8{int8(0), int8(0)},
			expect: 1,
		},
		{
			whenX:  []int8{int8(0), int8(1)},
			whenY:  []int8{int8(1), int8(0)},
			expect: -1,
		},

		{
			whenX:  []int16{int16(0), int16(0)},
			whenY:  []int16{int16(0), int16(0)},
			expect: 0,
		},
		{
			whenX:  []int16{int16(0), int16(1)},
			whenY:  []int16{int16(0), int16(0)},
			expect: 1,
		},
		{
			whenX:  []int16{int16(0), int16(1)},
			whenY:  []int16{int16(1), int16(0)},
			expect: -1,
		},

		{
			whenX:  []int32{int32(0), int32(0)},
			whenY:  []int32{int32(0), int32(0)},
			expect: 0,
		},
		{
			whenX:  []int32{int32(0), int32(1)},
			whenY:  []int32{int32(0), int32(0)},
			expect: 1,
		},
		{
			whenX:  []int32{int32(0), int32(1)},
			whenY:  []int32{int32(1), int32(0)},
			expect: -1,
		},

		{
			whenX:  []int64{int64(0), int64(0)},
			whenY:  []int64{int64(0), int64(0)},
			expect: 0,
		},
		{
			whenX:  []int64{int64(0), int64(1)},
			whenY:  []int64{int64(0), int64(0)},
			expect: 1,
		},
		{
			whenX:  []int64{int64(0), int64(1)},
			whenY:  []int64{int64(1), int64(0)},
			expect: -1,
		},

		{
			whenX:  []float32{float32(0), float32(0)},
			whenY:  []float32{float32(0), float32(0)},
			expect: 0,
		},
		{
			whenX:  []float32{float32(0), float32(1)},
			whenY:  []float32{float32(0), float32(0)},
			expect: 1,
		},
		{
			whenX:  []float32{float32(0), float32(1)},
			whenY:  []float32{float32(1), float32(0)},
			expect: -1,
		},

		{
			whenX:  []float64{float64(0), float64(0)},
			whenY:  []float64{float64(0), float64(0)},
			expect: 0,
		},
		{
			whenX:  []float64{float64(0), float64(1)},
			whenY:  []float64{float64(0), float64(0)},
			expect: 1,
		},
		{
			whenX:  []float64{float64(0), float64(1)},
			whenY:  []float64{float64(1), float64(0)},
			expect: -1,
		},

		{
			whenX:  []string{"0", "0"},
			whenY:  []string{"0", "0"},
			expect: 0,
		},
		{
			whenX:  []string{"0", "1"},
			whenY:  []string{"0", "0"},
			expect: 1,
		},
		{
			whenX:  []string{"0", "1"},
			whenY:  []string{"1", "0"},
			expect: -1,
		},
	}

	for _, tc := range testCase2 {
		r := Compare(tc.whenX, tc.whenY)
		assert.Equal(t, tc.expect, r)
	}
}
