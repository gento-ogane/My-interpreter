package evaluator

import (
	"fmt"
	"monkey/ast"
	"monkey/object"
)

//真偽値用のインスタンスを予め作成しておく
var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	//Nodeのタイプによってどのeval関数を呼び出すのか場合分け
	switch node := node.(type) {

	//文の配列を受け取った時(初回)
	case *ast.Program:
		return evalProgram(node, env) //文のスライスを分解(一つずつ)して、Evalを呼び出している

	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.BlockStatement:
		return evalBlockStatement(node, env)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.WhileExpression:
		return evalWhileExpression(node, env)
	case *ast.ForLoop:
		return evalForLoopExpression(node, env)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}

	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val) //環境に新しく変数を追加する。
	//識別子の場合
	case *ast.Identifier:
		return evalIdentifier(node, env)

	//関数を認識する
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Env: env, Body: body}

	case *ast.FunctionStatement:
		funcObj := Eval(node.FunctionLiteral, env)
		env.Set(node.Name.String(), funcObj)
		return funcObj

	//関数を呼び出す
	case *ast.CallExpression:
		if node.Function.TokenLiteral() == "quote" {
			return quote(node.Arguments[0])
		}
		function := Eval(node.Function, env) //関数を認識し、関数objectを得る。
		if isError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, env) //引数と環境を渡し、引数の値を計算したobjectスライスを得る。ex) 5+5 => 10
		if len(args) == 1 && isError(args[0]) {      //エラーがある場合,args[0]に格納されている。(objectインスタンスを新しく作成するため)
			return args[0]
		}

		return applyFunction(function, args) //関数Objectと引数Objectを用い、拡張環境を作成してそこで実行する。

	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env) //引数と環境を渡し、配列の値を計算したobjectスライスを得る。ex) 5+5 => 10
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}

	case *ast.IndexExpression: //添字演算子の構文木
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(left, index)

	case *ast.HashLiteral:
		return evalHashLiteral(node, env)

	//式
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value} //オブジェクトシステムの整数型を返す。Valueは受け取ったNodeのValueを入れている。

	case *ast.StringLiteral:
		return &object.String{Value: node.Value} //オブジェクトシステムの文字型を返す。Valueは受け取ったNodeのValueを入れている。

	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value) //オブジェクトシステムの真偽値型を返す。Valueは受け取ったNodeのValueを入れている。

	case *ast.ClassStatement:
		return evalClassStatement(node, env)
	case *ast.ClassLiteral:
		return evalClassLiteral(node, env)
	case *ast.NewExpression:
		return evalNewExpression(node, env)

	case *ast.MethodCallExpression:
		return evalMethodCallExpression(node, env)

	case *ast.PrefixExpression: //前置演算式。Token(type),Operator(string),right(Expression)から成る
		right := Eval(node.Right, env) //まず右の式を評価してObjectを得る。
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right) //オペレータと右の値(上で評価したObject)からObjectを返却する。

	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)

	case *ast.PostfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		return evalPostfixExpression(left, node.Operator)
	case *ast.AssignExpression:
		return evalAssignExpression(node, env)
	}
	return nil
}

func nativeBoolToBooleanObject(input bool) *object.Boolean { //TrueとFalseのオブジェクトを参照する。
	if input {
		return TRUE
	}
	return FALSE
}

//前置演算子の評価関数
func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right) //right(右オペランド)の反転した値をValueに入れたobject.Objectを返却する
	case "-":
		return evalMinusPrefixOperatorExpression(right) //right(右オペランド)の値に-1をかけた値をValueに入れたobject.Objectを返却する
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalPostfixExpression(left object.Object, operator string) object.Object {
	switch operator {
	case "++":
		return evalIncrementPostfixOperatorExpression(left)
	case "--":
		return evalDecrementPostfixOperatorExpression(left)
	default:
		return newError("unknown operator: %s%s", operator, left.Type())
	}
}

func evalIncrementPostfixOperatorExpression(left object.Object) object.Object {
	switch left.Type() {
	case object.INTEGER_OBJ: //Integerのみ後置演算のみ
		leftObj := left.(*object.Integer)
		returnVal := object.NewInteger(leftObj.Value)
		leftObj.Value = leftObj.Value + 1
		return returnVal
	default:
		return NULL
	}
}

func evalDecrementPostfixOperatorExpression(left object.Object) object.Object {
	switch left.Type() {
	case object.INTEGER_OBJ: //Integerのみ後置演算のみ
		leftObj := left.(*object.Integer)
		returnVal := object.NewInteger(leftObj.Value)
		leftObj.Value = leftObj.Value - 1
		return returnVal
	default:
		return NULL
	}
}

//right(右オペランド)の反転した値を返却する。
func evalBangOperatorExpression(right object.Object) object.Object {
	//ここでのTRUEやFALSEはobject型。参照している。
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

//前置演算子-の処理
func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ { //オペランドが整数かどうかのcheck
		return newError("unknown operator: -%s", right.Type())
	}
	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value} //-1がかかった値を返却する
}

func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	//オペランド(演算対象)として、整数値が入れられた場合
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)

	//オペランドとして、真偽値が入れられた場合
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)

	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)

	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s",
			left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

//中値演算子式。オペランドとして整数が入れられた場合。
func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence, env) //真文
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env) //else文
	} else {
		return NULL
	}
}

//null,false以外はtrue
func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range program.Statements {
		result = Eval(statement, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error: //errorの時、評価を中断する
			return result
		}
	}
	return result
}

func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = Eval(statement, env)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ { //returnとerrorの時の両方で評価を中断する
				return result
			}
		}
	}
	return result
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}

}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

//識別子から、その値を参照する関数
func evalIdentifier(
	node *ast.Identifier,
	env *object.Environment,
) object.Object {
	if val, ok := env.Get(node.Value); ok { //環境から識別子をキーとしてGetする、ない場合はエラーが出る(そんな変数定義れてないよ！)
		return val
	}
	if builtin, ok := builtins[node.Value]; ok { //与えられた識別子が現在の環境で値に束縛されていない時、フォールバックして組み込み関数を探す
		return builtin
	}
	return newError("identifier not found: " + node.Value)
}

//引数と環境を渡し、引数の値を計算したスライスを得る。ex) 5+5 => 10
func evalExpressions(
	exps []ast.Expression,
	env *object.Environment,
) []object.Object {
	var result []object.Object

	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) { //評価の中止
			return []object.Object{evaluated} //evaluatedはエラーobjectなので、それを返却する。
		}
		result = append(result, evaluated)
	}
	return result
}

//関数Objectと引数Objectを用い、拡張環境を作成してそこで実行する。
func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		extendEnv := extendFunctionEnv(fn, args) //関数が保持する環境に包まれた新環境で変数を束縛し、その環境を返す。
		evaluated := Eval(fn.Body, extendEnv)    //その関数のBodyと環境を入れ、Evalする！
		return unwrapReturnValue(evaluated)      //returnの場合、アンラップしないとBlockの外まできて評価を中止してしまう。
	case *object.Builtin:
		return fn.Fn(args...)
	default: //objectが手に入っていない場合はエラーを発生
		return newError("not a function: %s", fn.Type())
	}
}

//拡張された環境の作成
func extendFunctionEnv(
	fn *object.Function,
	args []object.Object,
) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env) //関数独自に持つ環境を外側にもつ環境を作成(環境を拡張する。)

	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx]) //拡張した環境にparams(引数)変数を保存する。
	}

	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

func evalStringInfixExpression(
	operator string,
	left, right object.Object) object.Object {

	//演算子は+以外受け付けない
	if operator != "+" {
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}

	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value
	return &object.String{Value: leftVal + rightVal}
}

func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	case left.Type() == object.HASH_OBJ:
		return evalHashIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	if idx < 0 || idx > max { //配列の長さが0未満、maxより大きい場合
		return NULL
	}
	return arrayObject.Elements[idx]
}

func evalHashLiteral(
	node *ast.HashLiteral,
	env *object.Environment,
) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for keyNode, valueNode := range node.Pairs {
		key := Eval(keyNode, env)
		if isError(key) {
			return key
		}

		hashKey, ok := key.(object.Hashable) //Hashableインターフェースのアサーション
		if !ok {
			return newError("unusable as hash key: %s", key.Type())
		}

		value := Eval(valueNode, env)
		if isError(value) {
			return value
		}
		hashed := hashKey.HashKey() //hashkeyからhashを取り出す！
		pairs[hashed] = object.HashPair{Key: key, Value: value}
	}
	return &object.Hash{Pairs: pairs}
}

func evalHashIndexExpression(hash, index object.Object) object.Object {
	hashObject := hash.(*object.Hash)

	key, ok := index.(object.Hashable) //添字として使うものがHashableである必要がある
	if !ok {
		return newError("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()] //keyからPairを取得
	if !ok {
		return NULL
	}
	return pair.Value //PairのValueを返す
}

func evalWhileExpression(
	we *ast.WhileExpression,
	env *object.Environment,
) object.Object {
	condition := Eval(we.Condition, env)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		Eval(we.Consequence, env)
		return Eval(we, env) //真文
	}
	return NULL
}

func evalClassLiteral(c *ast.ClassLiteral, env *object.Environment) object.Object {
	clsObj := &object.Class{
		Name:    c.Name,
		Members: c.Members,
		Methods: make(map[string]*object.Function, len(c.Methods)),
	}

	newScope := object.NewEnclosedEnvironment(env)
	for _, member := range c.Members {
		Eval(member, newScope) //拡張環境先で変数を入れる。
	}
	for k, f := range c.Methods {
		clsObj.Methods[k] = Eval(f, newScope).(*object.Function)
	}

	return clsObj
}

func evalClassStatement(c *ast.ClassStatement, env *object.Environment) object.Object {

	clsObj := evalClassLiteral(c.ClassLiteral, env)

	env.Set(c.Name.Value, clsObj) //環境にセットする

	return NULL
}

func evalNewExpression(n *ast.NewExpression, env *object.Environment) object.Object {
	class := Eval(n.Class, env) //そもそもクラス文ってExpressionなのか？

	clsObj, ok := class.(*object.Class)
	if !ok {
		fmt.Println("error2")
	}
	newScope := object.NewEnclosedEnvironment(env)
	for _, member := range clsObj.Members {
		Eval(member, newScope)
	}
	for k, f := range clsObj.Methods {
		newScope.Set(k, f)
	}

	instance := &object.Instance{Class: clsObj, Env: newScope} //閉じた環境にclsObjを入れる。

	return instance
}

func evalMethodCallExpression(call *ast.MethodCallExpression, env *object.Environment) object.Object {
	obj := Eval(call.Object, env)
	if obj.Type() == object.ERROR_OBJ {
		return obj
	}
	switch m := obj.(type) {
	case *object.Instance: //.の左側のtype
		instanceObj := m
		switch o := call.Call.(type) { //.の右側のtype
		case *ast.Identifier:
			val, ok := instanceObj.Env.Get(o.Value)
			if ok {
				return val
			}
		case *ast.CallExpression:
			return Eval(o, instanceObj.Env)
		}
	}
	return NULL
}

func evalAssignExpression(a *ast.AssignExpression, env *object.Environment) object.Object {
	val := Eval(a.Value, env)
	if val.Type() == object.ERROR_OBJ {
		return val
	}

	var name string
	switch nodeType := a.Name.(type) {
	case *ast.Identifier:
		name = nodeType.Value
	case *ast.IndexExpression:
		name = nodeType.Left.(*ast.Identifier).Value
	}
	v, ok := env.Reset(name, val)
	if ok {
		return v
	}
	return NULL
}

func evalForLoopExpression(fl *ast.ForLoop, env *object.Environment) object.Object { //fl:For Loop
	innerScope := object.NewEnclosedEnvironment(env)

	if fl.Init != nil {
		init := Eval(fl.Init, innerScope)
		if init.Type() == object.ERROR_OBJ {
			return init
		}
	}

	condition := Eval(fl.Cond, innerScope)
	if condition.Type() == object.ERROR_OBJ {
		return condition
	}

	var result object.Object
	for isTruthy(condition) {
		newSubScope := object.NewEnclosedEnvironment(innerScope)
		result = Eval(fl.Block, newSubScope)
		if result.Type() == object.ERROR_OBJ {
			return result
		}

		if fl.Update != nil {
			newVal := Eval(fl.Update, newSubScope)
			if newVal.Type() == object.ERROR_OBJ {
				return newVal
			}
		}

		condition = Eval(fl.Cond, newSubScope)
		if condition.Type() == object.ERROR_OBJ {
			return condition
		}
	}

	if result == nil {
		return NULL
	}
	return result
}
