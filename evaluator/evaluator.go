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
	case *ast.PrefixExpression:
		right := Eval(node.Right)
		return evalPrefixExpression(node.Operator, right)
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

//前置演算子の評価関数
func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusOperatorExpression(right)
	default:
		return NULL
	}
}

//right(右オペランド)の反転した値を返却する。
func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evalMinusOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ { //オペランドが整数かどうかのcheck
		return NULL
	}
	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value} //-1がかかった値を返却する
}
