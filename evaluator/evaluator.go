package evaluator

import (
	"monkey/ast"
	"monkey/object"
)

//真偽値用のインスタンスを予め作成しておく
var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node) object.Object {
	switch node := node.(type) {
	//文
	case *ast.Program:
		return evalStatements(node.Statements) //文のスライスを分解(一つずつ)して、Evalを呼び出している
	case *ast.ExpressionStatement:
		return Eval(node.Expression)
	//式
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value} //オブジェクトシステムの整数型を返す
	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value) //オブジェクトシステムの真偽値型を返す

	}
	return nil
}

func evalStatements(stmts []ast.Statement) object.Object {
	var result object.Object

	for _, statement := range stmts {
		result = Eval(statement)
	}

	return result
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}
