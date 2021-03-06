package object

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"monkey/ast"
	"strings"
)

type ObjectType string

const (
	INTEGER_OBJ      = "INTEGER"
	BOOLEAN_OBJ      = "BOOLEAN"
	NULL_OBJ         = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	ERROR_OBJ        = "ERROR"
	FUNCTION_OBJ     = "FUNCTION"
	STRING_OBJ       = "STRING"
	BUILTIN_OBJ      = "BUILTIN"
	ARRAY_OBJ        = "ARRAY"
	HASH_OBJ         = "HASH"
	QUOTE_OBJ        = "QUOTE"
	CLASS_OBJ        = "CLASS"
	INSTANCE_OBJ     = "INSTANCE_OBJ"
)

type Object interface {
	Type() ObjectType //オブジェクトのタイプ
	Inspect() string  //文字を返す。ぱっと見で中身がわかるため
}

//整数
type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }

func NewInteger(i int64) *Integer {
	return &Integer{Value: i}
}

//真偽値
type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }

//NULL
type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "null" }

//RETURN
type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

//ERROR
type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}

	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(")")

	return out.String()
}

//文字列オブジェクト
type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }

type BuiltinFunction func(args ...Object) Object

type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return "builtin function" }

type Array struct {
	Elements []Object
}

func (ao *Array) Type() ObjectType { return ARRAY_OBJ }
func (ao *Array) Inspect() string {
	var out bytes.Buffer

	elements := []string{}
	for _, e := range ao.Elements {
		elements = append(elements, e.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", ")) //配列の間に,を差し込む
	out.WriteString("]")

	return out.String()
}

type HashPair struct { //こうするとInspect()で割り易く取り出せる
	Key   Object
	Value Object
}

type HashKey struct {
	Type  ObjectType
	Value uint64
}

type Hash struct {
	Pairs map[HashKey]HashPair
}

type Hashable interface {
	HashKey() HashKey
}

//HashKey経由で正しいkeyとりだせるようになる！

func (b *Boolean) HashKey() HashKey {
	var value uint64

	if b.Value {
		value = 1
	} else {
		value = 0
	}
	return HashKey{Type: b.Type(), Value: value}
}
func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))

	return HashKey{Type: s.Type(), Value: h.Sum64()}
}

func (h *Hash) Type() ObjectType { return HASH_OBJ }

func (h *Hash) Inspect() string {
	var out bytes.Buffer

	pairs := []string{}
	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s",
			pair.Key.Inspect(), pair.Value.Inspect())) //keyとvalueを使用してきれいに出力する
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}

type Quote struct {
	Node ast.Node
}

func (q *Quote) Type() ObjectType { return QUOTE_OBJ }
func (q *Quote) Inspect() string {
	return "QUOTE(" + q.Node.String() + ")"
}

type Class struct {
	Name    string
	Parent  *Class
	Members []*ast.LetStatement
	Methods map[string]*Function
	Env     *Environment
}

func (c *Class) Inspect() string {
	var out bytes.Buffer
	out.WriteString("<class:" + c.Name + ">")

	members := []string{}
	for _, mem := range c.Members {
		members = append(members, mem.String())
	}
	out.WriteString("{")
	out.WriteString(strings.Join(members, ", "))
	out.WriteString("}")

	return out.String()
}

func (c *Class) Type() ObjectType { return CLASS_OBJ }

type Instance struct {
	Class *Class
	Env   *Environment
}

func (oi *Instance) Inspect() string  { return "<Instance:" + oi.Class.Name + ">" }
func (oi *Instance) Type() ObjectType { return INSTANCE_OBJ }
