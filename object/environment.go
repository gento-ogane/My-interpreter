package object

//拡張する対象の環境へのポインタ
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer //引数で受け取った環境を外側の環境として、envを包む。
	return env
}

//環境の追加(string:Objectのハッシュマップ)
func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s}
}

type Environment struct {
	store map[string]Object
	outer *Environment
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil { //この環境にはなく、この環境を包み込む外側の環境がある場合
		obj, ok = e.outer.Get(name) //外側の環境へ参照を行う。
	}
	return obj, ok
}

func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}

func (s *Environment) Reset(name string, val Object) (Object, bool) {
	var ok bool
	_, ok = s.store[name]
	if ok {
		s.store[name] = val
	}

	if !ok {
		s.store[name] = val
		ok = true
	}
	return val, ok
}
