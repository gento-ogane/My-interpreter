package evaluator

import "monkey/object"

var builtins = map[string]*object.Builtin{
	"len": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			//引数が一つではない時
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			switch arg := args[0].(type) {

			//stringを受け取った時(きちんと動作する時)
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))} //goで文字列をlenし、IntegerObjectに渡してreturnしている。

			//stringではない引数を受け取った時
			default:
				return newError("argument to `len` not supported, got=%s",
					args[0].Type())

			}
		},
	},
}
