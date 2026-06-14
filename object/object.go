package object

import (
	"fmt"
	"webscript/ast"
)

type ObjectType string

const (
	STRING_OBJ   = "STRING"
	INTEGER_OBJ  = "INTEGER"
	FUNCTION_OBJ = "FUNCTION"
	BUILTIN_OBJ  = "BUILTIN"
	NULL_OBJ     = "NULL"
	ERROR_OBJ    = "ERROR"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type String struct {
	Value string
}
func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }

type Integer struct {
	Value int64
}
func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }

type Null struct{}
func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "null" }

type Error struct {
	Message string
}
func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }

type Environment struct {
	store map[string]Object
	outer *Environment
}

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, outer: nil}
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}
func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string  { return "func(...)" }

type BuiltinFunction func(args ...Object) Object
type Builtin struct {
	Fn BuiltinFunction
}
func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return "builtin function" }
