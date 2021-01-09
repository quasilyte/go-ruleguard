package quasigo

import (
	"fmt"
	"reflect"
)

const maxFuncLocals = 8

func eval(env *EvalEnv, fn *Func, args []interface{}) interface{} {
	pc := 0
	code := fn.code
	stack := env.stack
	var locals [maxFuncLocals]interface{}

	for {
		switch op := opcode(code[pc]); op {
		case opPushParam:
			index := code[pc+1]
			stack.Push(args[index])
			pc += 2

		case opPushLocal:
			index := code[pc+1]
			stack.Push(locals[index])
			pc += 2

		case opSetLocal:
			index := code[pc+1]
			locals[index] = stack.Pop()
			pc += 2

		case opIncLocal:
			index := code[pc+1]
			locals[index] = locals[index].(int) + 1
			pc += 2

		case opDecLocal:
			index := code[pc+1]
			locals[index] = locals[index].(int) - 1
			pc += 2

		case opPop:
			stack.Discard()
			pc++
		case opDup:
			stack.Dup()
			pc++

		case opPushConst:
			id := code[pc+1]
			stack.Push(fn.constants[id])
			pc += 2

		case opPushTrue:
			stack.Push(true)
			pc++
		case opPushFalse:
			stack.Push(false)
			pc++

		case opReturnTrue:
			return true
		case opReturnFalse:
			return false
		case opReturnTop:
			return stack.Top()

		case opCallBuiltin:
			id := decode16(code, pc+1)
			fn := env.nativeFuncs[id].mappedFunc
			fn(&stack)
			pc += 3

		case opJump:
			offset := decode16(code, pc+1)
			pc += offset

		case opJumpFalse:
			if !stack.Top().(bool) {
				offset := decode16(code, pc+1)
				pc += offset
			} else {
				pc += 3
			}
		case opJumpTrue:
			if stack.Top().(bool) {
				offset := decode16(code, pc+1)
				pc += offset
			} else {
				pc += 3
			}

		case opNot:
			stack.Push(!stack.Pop().(bool))
			pc++

		case opConcat:
			x, y := stack.Pop2()
			stack.Push(x.(string) + y.(string))
			pc++

		case opAdd:
			x, y := stack.Pop2()
			stack.Push(x.(int) + y.(int))
			pc++

		case opSub:
			x, y := stack.Pop2()
			stack.Push(x.(int) - y.(int))
			pc++

		case opEqInt:
			x, y := stack.Pop2()
			stack.Push(x.(int) == y.(int))
			pc++

		case opNotEqInt:
			x, y := stack.Pop2()
			stack.Push(x.(int) != y.(int))
			pc++

		case opGtInt:
			x, y := stack.Pop2()
			stack.Push(x.(int) > y.(int))
			pc++

		case opGtEqInt:
			x, y := stack.Pop2()
			stack.Push(x.(int) >= y.(int))
			pc++

		case opLtInt:
			x, y := stack.Pop2()
			stack.Push(x.(int) < y.(int))
			pc++

		case opLtEqInt:
			x, y := stack.Pop2()
			stack.Push(x.(int) <= y.(int))
			pc++

		case opEqString:
			x, y := stack.Pop2()
			stack.Push(x.(string) == y.(string))
			pc++

		case opNotEqString:
			x, y := stack.Pop2()
			stack.Push(x.(string) != y.(string))
			pc++

		case opIsNil:
			x := stack.Pop()
			stack.Push(x == nil || reflect.ValueOf(x).IsNil())
			pc++

		case opIsNotNil:
			x := stack.Pop()
			stack.Push(x != nil && !reflect.ValueOf(x).IsNil())
			pc++

		case opStringSlice:
			to := stack.Pop().(int)
			from := stack.Pop().(int)
			s := stack.Pop().(string)
			stack.Push(s[from:to])
			pc++

		case opStringSliceFrom:
			from := stack.Pop().(int)
			s := stack.Pop().(string)
			stack.Push(s[from:])
			pc++

		case opStringSliceTo:
			to := stack.Pop().(int)
			s := stack.Pop().(string)
			stack.Push(s[:to])
			pc++

		case opStringLen:
			stack.Push(len(stack.Pop().(string)))
			pc++

		default:
			panic(fmt.Sprintf("malformed bytecode: unexpected %s found", op))
		}
	}
}
