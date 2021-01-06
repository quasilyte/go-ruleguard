// Package quasigo implements a Go subset compiler and interpreter.
//
// The implementation details are not part of the contract of this package.
package quasigo

import (
	"go/ast"
	"go/token"
	"go/types"
)

// TODO(quasilyte): document what is thread-safe and what not.
// TODO(quasilyte): add a readme.

// Env is used to hold both compilation and evaluation data.
type Env struct {
	// TODO(quasilyte): store both builtin and user func ids in one map?

	nativeFuncs         []nativeFunc
	nameToBuiltinFuncID map[funcKey]uint16

	userFuncs    []*Func
	nameToFuncID map[funcKey]uint16

	// debug contains all information that is only needed
	// for better debugging and compiled code introspection.
	// Right now it's always enabled, but we may allow stripping it later.
	debug *debugInfo
}

// EvalEnv is a goroutine-local handle for Env.
// To get one, use Env.GetEvalEnv() method.
type EvalEnv struct {
	nativeFuncs []nativeFunc
	userFuncs   []*Func

	stack ValueStack
}

// NewEnv creates a new empty environment.
func NewEnv() *Env {
	return newEnv()
}

// GetEvalEnv creates a new goroutine-local handle of env.
func (env *Env) GetEvalEnv() *EvalEnv {
	return &EvalEnv{
		nativeFuncs: env.nativeFuncs,
		userFuncs:   env.userFuncs,
		stack:       make([]interface{}, 0, 32),
	}
}

// AddNativeMethod binds `$typeName.$methodName` symbol with f.
// A typeName should be fully qualified, like `github.com/user/pkgname.TypeName`.
// It method is defined only on pointer type, the typeName should start with `*`.
func (env *Env) AddNativeMethod(typeName, methodName string, f func(*ValueStack)) {
	env.addNativeFunc(funcKey{qualifier: typeName, name: methodName}, f)
}

// AddNativeFunc binds `$pkgPath.$funcName` symbol with f.
// A pkgPath should be a full package path in which funcName is defined.
func (env *Env) AddNativeFunc(pkgPath, funcName string, f func(*ValueStack)) {
	env.addNativeFunc(funcKey{qualifier: pkgPath, name: funcName}, f)
}

// AddFunc binds `$pkgPath.$funcName` symbol with f.
func (env *Env) AddFunc(pkgPath, funcName string, f *Func) {
	env.addFunc(funcKey{qualifier: pkgPath, name: funcName}, f)
}

// GetFunc finds previously bound function searching for the `$pkgPath.$funcName` symbol.
func (env *Env) GetFunc(pkgPath, funcName string) *Func {
	id := env.nameToFuncID[funcKey{qualifier: pkgPath, name: funcName}]
	return env.userFuncs[id]
}

// CompileContext is used to provide necessary data to the compiler.
type CompileContext struct {
	// Env is shared environment that should be used for all functions
	// being compiled; then it should be used to execute these functions.
	Env *Env

	Types *types.Info
	Fset  *token.FileSet
}

// Compile prepares an executable version of fn.
func Compile(ctx *CompileContext, fn *ast.FuncDecl) (compiled *Func, err error) {
	return compile(ctx, fn)
}

// Call invokes a given function with provided arguments.
func Call(env *EvalEnv, fn *Func, args ...interface{}) interface{} {
	env.stack = env.stack[:0]
	return eval(env, fn, args)
}

// Disasm returns the compiled function disassembly text.
// This output is not guaranteed to be stable between versions
// and should be used only for debugging purposes.
func Disasm(env *Env, fn *Func) string {
	return disasm(env, fn)
}

// Func is a compiled function that is ready to be executed.
type Func struct {
	code []byte

	constants []interface{}
}

// ValueStack is used to manipulate runtime values during the evaluation.
// Function arguments are pushed to the stack.
// Function results are returned via stack as well.
type ValueStack []interface{}

// Pop removes the top stack element and returns it.
func (s *ValueStack) Pop() interface{} {
	x := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return x
}

// Pop2 removes the two top stack elements and returns them.
//
// Note that it returns the popped elements in the reverse order
// to make it easier to map the order in which they were pushed.
func (s *ValueStack) Pop2() (second interface{}, top interface{}) {
	x := (*s)[len(*s)-2]
	y := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-2]
	return x, y
}

// Push adds x to the stack.
func (s *ValueStack) Push(x interface{}) { *s = append(*s, x) }

// Top returns top of the stack without popping it.
func (s *ValueStack) Top() interface{} { return (*s)[len(*s)-1] }

// Dup copies the top stack element.
// Identical to s.Push(s.Top()), but more concise.
func (s *ValueStack) Dup() { *s = append(*s, (*s)[len(*s)-1]) }

// Discard drops the top stack element.
// Identical to s.Pop() without using the result.
func (s *ValueStack) Discard() { *s = (*s)[:len(*s)-1] }
