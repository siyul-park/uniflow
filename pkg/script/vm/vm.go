package vm

import (
	"errors"
	"fmt"

	"github.com/siyul-park/uniflow/pkg/script/code"
	"github.com/siyul-park/uniflow/pkg/script/compiler"
	"github.com/siyul-park/uniflow/pkg/script/object"
)

const (
	// StackSize is an initial stack size.
	StackSize = 2048

	// GlobalSize is an upper limit of the number of global bindings the VM can support.
	GlobalSize = 1 << 16 // 16 bits

	// MaxFrames is the maximum number of stack frames.
	MaxFrames = 1024
)

var (
	// True is the boolean `true` value.
	True = &object.Boolean{Value: true}
	// False is the boolean `false` value.
	False = &object.Boolean{Value: false}
	// Nil represents the zero value.
	Nil = &object.Nil{}
)

// VM is a virtual machine which interprets and executes bytecode instructions.
type VM struct {
	consts []object.Object

	stack []object.Object
	// Stack pointer always points to the *next* slot on the stack. Top of stack is stack[sp-1].
	sp int

	// globals store
	globals []object.Object

	frames    []*Frame
	framesIdx int
}

// New creates a new VM instance which executes the given bytecode.
func New(bytecode *compiler.Bytecode) *VM {
	return NewWithGlobalStore(bytecode, make([]object.Object, GlobalSize))
}

// NewWithGlobalStore creates a new VM instance which executes the given bytecode with the
// given globals store.
func NewWithGlobalStore(bytecode *compiler.Bytecode, globals []object.Object) *VM {
	mainFn := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainClosure := &object.Closure{Fn: mainFn}
	mainFrame := NewFrame(mainClosure, 0) // Base pointer points to zero

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		consts: bytecode.Constants,

		stack: make([]object.Object, StackSize),
		sp:    0,

		globals: globals,

		frames:    frames,
		framesIdx: 1,
	}
}

// StackTop returns an object on top of the stack.
func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

// LastPoppedStackElem returns an object which was popped off the stack most recently.
func (vm *VM) LastPoppedStackElem() object.Object {
	// vm.sp always points to the *next free* slot in vm.stack
	return vm.stack[vm.sp]
}

// Run executes bytecode instructions.
func (vm *VM) Run() error {
	frame := vm.currentFrame()
	insns := frame.Instructions()

	for frame.ip < len(insns)-1 {
		frame.ip++

		ip := frame.ip
		op := code.Opcode(insns[ip])

		switch op {
		case code.OpConstant:
			// Read a 2-byte operand from the next position
			constIdx := code.ReadUint16(insns[ip+1:])
			// Because the operand is 2-byte width and we already read it,
			// increment the pointer by 2 (bytes)
			frame.ip += 2

			if err := vm.push(vm.consts[constIdx]); err != nil {
				return err
			}

		case code.OpTrue:
			if err := vm.push(True); err != nil {
				return err
			}

		case code.OpFalse:
			if err := vm.push(False); err != nil {
				return err
			}

		case code.OpNil:
			if err := vm.push(Nil); err != nil {
				return err
			}

		case code.OpArray:
			numElems := int(code.ReadUint16(insns[ip+1:]))
			frame.ip += 2

			startIdx := vm.sp - numElems
			arr := vm.buildArray(startIdx, vm.sp)
			vm.sp = startIdx

			if err := vm.push(arr); err != nil {
				return err
			}

		case code.OpHash:
			numElems := int(code.ReadUint16(insns[ip+1:]))
			frame.ip += 2

			startIdx := vm.sp - numElems
			hash, err := vm.buildHash(startIdx, vm.sp)
			if err != nil {
				return err
			}
			vm.sp = startIdx

			if err := vm.push(hash); err != nil {
				return err
			}

		case code.OpSetIndex:
			val := vm.pop()
			idx := vm.pop()
			left := vm.pop()

			if err := vm.execSetIndexExpr(left, idx, val); err != nil {
				return err
			}

		case code.OpGetIndex:
			idx := vm.pop()
			left := vm.pop()

			if err := vm.execGetIndexExpr(left, idx); err != nil {
				return err
			}

		case code.OpPop:
			vm.pop()

		case code.OpBang:
			if err := vm.execBangOp(); err != nil {
				return err
			}

		case code.OpMinus:
			if err := vm.execMinusOp(); err != nil {
				return err
			}

		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			if err := vm.execBinaryOp(op); err != nil {
				return err
			}

		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan, code.OpGreaterThanOrEqual:
			if err := vm.execComparison(op); err != nil {
				return err
			}

		case code.OpAnd, code.OpOr:
			if err := vm.execLogicalOp(op); err != nil {
				return err
			}

		case code.OpJump:
			pos := int(code.ReadUint16(insns[ip+1:]))
			// Since we're in a loop that increments `ip` with each iteration, we need to set `ip`
			// to the offset *right before the one* we want.
			frame.ip = pos - 1

		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(insns[ip+1:]))
			frame.ip += 2

			condition := vm.pop()
			if !isTruthy(condition) {
				frame.ip = pos - 1
			}

		case code.OpSetGlobal:
			globalIdx := code.ReadUint16(insns[ip+1:])
			frame.ip += 2

			vm.globals[globalIdx] = vm.pop()

		case code.OpGetGlobal:
			globalIdx := code.ReadUint16(insns[ip+1:])
			frame.ip += 2

			if err := vm.push(vm.globals[globalIdx]); err != nil {
				return err
			}

		case code.OpCall:
			numArgs := int(code.ReadUint8(insns[ip+1:]))
			frame.ip++

			if err := vm.execCall(numArgs); err != nil {
				return err
			}

		case code.OpReturnValue:
			// Pop the return value off the stack before clearing the stack frame
			retVal := vm.pop()

			// Clear the called function's stack frame
			frame := vm.popFrame()
			vm.sp = frame.bp - 1 // -1 for the called function object itself on the stack

			// Push the return value on to the stack again
			if err := vm.push(retVal); err != nil {
				return err
			}

		case code.OpReturn:
			// Clear the called function's stack frame
			frame := vm.popFrame()
			vm.sp = frame.bp - 1 // -1 for the called function object itself on the stack

			// Push the Nil value on to the stack because we have no return value
			if err := vm.push(Nil); err != nil {
				return err
			}

		case code.OpSetLocal:
			localIdx := int(code.ReadUint8(insns[ip+1:]))
			frame.ip++

			vm.stack[frame.bp+localIdx] = vm.pop()

		case code.OpGetLocal:
			localIdx := int(code.ReadUint8(insns[ip+1:]))
			frame.ip++

			if err := vm.push(vm.stack[frame.bp+localIdx]); err != nil {
				return err
			}

		case code.OpGetBuiltin:
			builtinIdx := code.ReadUint8(insns[ip+1:])
			frame.ip++

			def := object.Builtins[builtinIdx]

			if err := vm.push(def.Builtin); err != nil {
				return err
			}

		case code.OpClosure:
			constIdx := int(code.ReadUint16(insns[ip+1:]))
			numFree := int(code.ReadUint8(insns[ip+3:]))
			frame.ip += 3

			if err := vm.pushClosure(constIdx, numFree); err != nil {
				return err
			}

		case code.OpGetFree:
			freeIdx := code.ReadUint8(insns[ip+1:])
			frame.ip++

			currentClosure := frame.cl
			if err := vm.push(currentClosure.Free[freeIdx]); err != nil {
				return err
			}

		case code.OpCurrentClosure:
			currentClosure := frame.cl
			if err := vm.push(currentClosure); err != nil {
				return err
			}
		}

		// Update current frame and instructions for the next interation
		frame = vm.currentFrame()
		insns = frame.Instructions()
	}

	return nil
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIdx-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.framesIdx] = f
	vm.framesIdx++
}

func (vm *VM) popFrame() *Frame {
	vm.framesIdx--
	return vm.frames[vm.framesIdx]
}

func (vm *VM) push(obj object.Object) error {
	if vm.sp >= StackSize {
		return errors.New("stack overflow")
	}

	// Push the object on to the stack
	vm.stack[vm.sp] = obj
	// Increment the stack pointer
	vm.sp++

	return nil
}

func (vm *VM) pop() object.Object {
	if vm.sp == 0 {
		return nil
	}

	// Pop an object off the stack
	obj := vm.stack[vm.sp-1]
	// Decrement the stack pointer
	vm.sp--

	return obj
}

func (vm *VM) buildArray(startIdx, endIdx int) object.Object {
	elems := make([]object.Object, endIdx-startIdx)

	for i := startIdx; i < endIdx; i++ {
		elems[i-startIdx] = vm.stack[i]
	}

	return &object.Array{Elements: elems}
}

func (vm *VM) buildHash(startIdx, endIdx int) (object.Object, error) {
	capacity := (endIdx - startIdx) / 2
	m := make(map[object.HashKey]object.HashPair, capacity)

	for i := startIdx; i < endIdx; i += 2 {
		key := vm.stack[i]
		val := vm.stack[i+1]

		pair := object.HashPair{Key: key, Value: val}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("unusable as hash key: %s", key.Type())
		}

		m[hashKey.HashKey()] = pair
	}

	return &object.Hash{Pairs: m}, nil
}

func (vm *VM) execBangOp() error {
	operand := vm.pop()

	switch operand {
	case True:
		return vm.push(False)
	case False, Nil:
		return vm.push(True)
	default:
		return vm.push(False)
	}
}

func (vm *VM) execMinusOp() error {
	switch operand := vm.pop().(type) {
	case *object.Integer:
		return vm.push(&object.Integer{Value: -operand.Value})
	case *object.Float:
		return vm.push(&object.Float{Value: -operand.Value})
	default:
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}
}

func (vm *VM) execBinaryOp(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	switch {
	case isFloatArithmeticRequired(op, left, right):
		return vm.execBinaryFloatOp(op, left, right)
	case isBothType(object.IntegerType, left, right):
		return vm.execBinaryIntOp(op, left, right)
	case isBothType(object.StringType, left, right):
		return vm.execBinaryStrOp(op, left, right)
	default:
		return fmt.Errorf(
			"unsupported types for binary operation %d: %s and %s", op, left.Type(), right.Type(),
		)
	}
}

func (vm *VM) execBinaryIntOp(op code.Opcode, left, right object.Object) error {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	var result int64

	switch op {
	case code.OpAdd:
		result = leftVal + rightVal
	case code.OpSub:
		result = leftVal - rightVal
	case code.OpMul:
		result = leftVal * rightVal
	case code.OpDiv:
		result = leftVal / rightVal
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}

	return vm.push(&object.Integer{Value: result})
}

func (vm *VM) execBinaryFloatOp(op code.Opcode, left, right object.Object) error {
	leftVal, err := castToFloat(left)
	if err != nil {
		return err
	}

	rightVal, err := castToFloat(right)
	if err != nil {
		return err
	}

	var result float64

	switch op {
	case code.OpAdd:
		result = leftVal + rightVal
	case code.OpSub:
		result = leftVal - rightVal
	case code.OpMul:
		result = leftVal * rightVal
	case code.OpDiv:
		result = leftVal / rightVal
	default:
		return fmt.Errorf("unknown float operator: %d", op)
	}

	return vm.push(&object.Float{Value: result})
}

func (vm *VM) execBinaryStrOp(op code.Opcode, left, right object.Object) error {
	if op != code.OpAdd {
		return fmt.Errorf("unknown string operator: %d", op)
	}

	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value

	return vm.push(&object.String{Value: leftVal + rightVal})
}

func (vm *VM) execSetIndexExpr(left, idx, val object.Object) error {
	leftType := left.Type()
	switch {
	case leftType == object.ArrayType && idx.Type() == object.IntegerType:
		return vm.execArraySetIndex(left, idx, val)
	case leftType == object.HashType:
		return vm.execHashSetIndex(left, idx, val)
	default:
		return fmt.Errorf("index operator not supported: %s", leftType)
	}
}

func (vm *VM) execArraySetIndex(array, idx, val object.Object) error {
	arr := array.(*object.Array)
	i := idx.(*object.Integer).Value
	max := int64(len(arr.Elements) - 1)

	if i < 0 || i > max {
		return fmt.Errorf("array index %d out of range", i)
	}

	arr.Elements[i] = val

	return nil
}

func (vm *VM) execHashSetIndex(hash, idx, val object.Object) error {
	h := hash.(*object.Hash)

	key, ok := idx.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", idx.Type())
	}

	h.Pairs[key.HashKey()] = object.HashPair{Key: idx, Value: val}

	return nil
}

func (vm *VM) execGetIndexExpr(left, idx object.Object) error {
	leftType := left.Type()
	switch {
	case leftType == object.ArrayType && idx.Type() == object.IntegerType:
		return vm.execArrayGetIndex(left, idx)
	case leftType == object.HashType:
		return vm.execHashGetIndex(left, idx)
	default:
		return fmt.Errorf("index operator not supported: %s", leftType)
	}
}

func (vm *VM) execArrayGetIndex(array, idx object.Object) error {
	arr := array.(*object.Array)
	i := idx.(*object.Integer).Value
	max := int64(len(arr.Elements) - 1)

	if i < 0 || i > max {
		return vm.push(Nil)
	}

	return vm.push(arr.Elements[i])
}

func (vm *VM) execHashGetIndex(hash, idx object.Object) error {
	h := hash.(*object.Hash)

	key, ok := idx.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", idx.Type())
	}

	pair, ok := h.Pairs[key.HashKey()]
	if !ok {
		return vm.push(Nil)
	}

	return vm.push(pair.Value)
}

func (vm *VM) execComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if isEitherType(object.FloatType, left, right) {
		return vm.execFloatComparison(op, left, right)
	} else if isBothType(object.IntegerType, left, right) {
		return vm.execIntComparison(op, left, right)
	}

	var result bool

	switch op {
	case code.OpEqual:
		result = left == right
	case code.OpNotEqual:
		result = left != right
	default:
		return fmt.Errorf("unknown operator %d: %s and %s", op, left.Type(), right.Type())
	}

	return vm.push(nativeBoolToBooleanObject(result))
}

func (vm *VM) execIntComparison(op code.Opcode, left, right object.Object) error {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	var result bool

	switch op {
	case code.OpEqual:
		result = leftVal == rightVal
	case code.OpNotEqual:
		result = leftVal != rightVal
	case code.OpGreaterThan:
		result = leftVal > rightVal
	case code.OpGreaterThanOrEqual:
		result = leftVal >= rightVal
	default:
		return fmt.Errorf("unknown operator %d for integers", op)
	}

	return vm.push(nativeBoolToBooleanObject(result))
}

func (vm *VM) execFloatComparison(op code.Opcode, left, right object.Object) error {
	leftVal, err := castToFloat(left)
	if err != nil {
		return err
	}

	rightVal, err := castToFloat(right)
	if err != nil {
		return err
	}

	var result bool

	switch op {
	case code.OpEqual:
		result = leftVal == rightVal
	case code.OpNotEqual:
		result = leftVal != rightVal
	case code.OpGreaterThan:
		result = leftVal > rightVal
	case code.OpGreaterThanOrEqual:
		result = leftVal >= rightVal
	default:
		return fmt.Errorf("unknown operator %d for floats", op)
	}

	return vm.push(nativeBoolToBooleanObject(result))
}

func (vm *VM) execLogicalOp(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	switch op {
	case code.OpAnd:
		if isTruthy(left) {
			return vm.push(right)
		}
		return vm.push(left)
	case code.OpOr:
		if isTruthy(left) {
			return vm.push(left)
		}
		return vm.push(right)
	default:
		return fmt.Errorf("unknow logical operator: %d", op)
	}
}

func (vm *VM) execCall(numArgs int) error {
	callee := vm.stack[vm.sp-1-numArgs]
	switch callee := callee.(type) {
	case *object.Closure:
		return vm.callClosure(callee, numArgs)
	case *object.Builtin:
		return vm.callBuiltin(callee, numArgs)
	default:
		var typ interface{}
		if callee != nil {
			typ = callee.Type()
		}
		return fmt.Errorf("calling non-function and non-built-in: type %v", typ)
	}
}

func (vm *VM) callClosure(cl *object.Closure, numArgs int) error {
	if numArgs != cl.Fn.NumParameters {
		return fmt.Errorf(
			"wrong number of arguments: want=%d, got=%d", cl.Fn.NumParameters, numArgs,
		)
	}

	// Create a new stack frame
	basePtr := vm.sp - numArgs
	frame := NewFrame(cl, basePtr)
	vm.pushFrame(frame)

	vm.sp = frame.bp + cl.Fn.NumLocals // Reserve slots for local bindings on the stack

	return nil
}

func (vm *VM) callBuiltin(builtin *object.Builtin, numArgs int) error {
	args := vm.stack[vm.sp-numArgs : vm.sp]

	// Execute the built-in function itself
	result := builtin.Fn(args...)
	// Take the arguments and the function we just executed off the stack
	vm.sp -= (numArgs + 1)

	if result == nil {
		return vm.push(Nil)
	}
	return vm.push(result)
}

func (vm *VM) pushClosure(constIdx int, numFree int) error {
	// Fetch a closure itself
	c := vm.consts[constIdx]
	fn, ok := c.(*object.CompiledFunction)
	if !ok {
		return fmt.Errorf("not a function: %+v", c)
	}

	// Fetch free variables
	free := make([]object.Object, numFree)
	copy(free, vm.stack[vm.sp-numFree:vm.sp])
	vm.sp -= numFree

	// Create a closure and push it on to the stack
	closure := &object.Closure{Fn: fn, Free: free}
	return vm.push(closure)
}

func castToFloat(obj object.Object) (float64, error) {
	switch obj := obj.(type) {
	case *object.Integer:
		return float64(obj.Value), nil
	case *object.Float:
		return obj.Value, nil
	default:
		return 0.0, fmt.Errorf("could not cast to float: %s", obj.Type())
	}
}

func isFloatArithmeticRequired(op code.Opcode, left, right object.Object) bool {
	// Division always returns a floating-point number
	return op == code.OpDiv || isEitherType(object.FloatType, left, right)
}

func isBothType(typ object.Type, left, right object.Object) bool {
	return left.Type() == typ && right.Type() == typ
}

func isEitherType(typ object.Type, left, right object.Object) bool {
	return left.Type() == typ || right.Type() == typ
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return True
	}
	return False
}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.Nil:
		return false
	default:
		return true
	}
}
