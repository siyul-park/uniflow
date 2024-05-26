package code

import (
	"fmt"
	"testing"
)

func TestInstructionsString(t *testing.T) {
	insns := []Instructions{
		Make(OpAdd),
		Make(OpGetLocal, 0x1),
		Make(OpConstant, 0x2),
		Make(OpConstant, 0xFFFF),
		Make(OpClosure, 0xFFFF, 0xFF),
	}

	want := `0000 OpAdd
0001 OpGetLocal 0x1
0003 OpConstant 0x2
0006 OpConstant 0xFFFF
0009 OpClosure 0xFFFF 0xFF
`

	concat := make(Instructions, 0, len(insns))
	for _, ins := range insns {
		concat = append(concat, ins...)
	}

	fmt.Println(concat)
	got := concat.String()
	if got != want {
		t.Errorf("instructions wrongly formatted.\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestMake(t *testing.T) {
	tests := []struct {
		op       Opcode
		operands []int
		want     []byte
	}{
		{
			op:       OpConstant,
			operands: []int{65534},
			want:     []byte{byte(OpConstant), 255, 254},
		},
		{
			op:       OpAdd,
			operands: nil,
			want:     []byte{byte(OpAdd)},
		},
		{
			op:       OpGetLocal,
			operands: []int{255},
			want:     []byte{byte(OpGetLocal), 255},
		},
		{
			op:       OpClosure,
			operands: []int{65534, 255},
			want:     []byte{byte(OpClosure), 255, 254, 255},
		},
	}

	for _, tt := range tests {
		insn := Make(tt.op, tt.operands...)

		if len(insn) != len(tt.want) {
			t.Errorf("instruction has wrong length; want=%d, got=%d", len(tt.want), len(insn))
		}

		for i, b := range tt.want {
			if insn[i] != b {
				t.Errorf("wrong byte at pos %d. want=%d, got=%d", i, b, insn[i])
			}
		}
	}
}
func TestReadOperands(t *testing.T) {
	tests := []struct {
		op        Opcode
		operands  []int
		bytesRead int
	}{
		{op: OpConstant, operands: []int{0xFFFF}, bytesRead: 2},
		{op: OpGetLocal, operands: []int{0xFF}, bytesRead: 1},
		{op: OpClosure, operands: []int{0xFFFF, 0xFF}, bytesRead: 3},
	}

	for _, tt := range tests {
		b := byte(tt.op)
		def, err := Lookup(b)
		if err != nil {
			t.Fatalf("definition for byte %x not found: %s", b, err)
		}

		insns := Make(tt.op, tt.operands...)
		operandsRead, n := ReadOperands(def, insns[1:])
		if n != tt.bytesRead {
			t.Fatalf("number of bytes read wrong. want=%d, got=%d", tt.bytesRead, n)
		}

		for i, want := range tt.operands {
			if operandsRead[i] != want {
				t.Errorf("operand wrong. want=%d, got=%d", want, operandsRead[i])
			}
		}
	}
}
