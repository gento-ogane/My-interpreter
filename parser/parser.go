package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
	"strconv"
)

type Parser struct {
	l              *lexer.Lexer //字句解析インスタンスへのポインタ
	errors         []string
	curToken       token.Token                       //現在のToken
	peekToken      token.Token                       //次のToken
	prefixParseFns map[token.TokenType]prefixParseFn //前置のtoken.Typeから対応する関数を呼び出す
	infixParseFns  map[token.TokenType]infixParseFn  //中置のtoken.Typeから対応する関数を呼び出す
}

const (
	_int = iota
	LOWEST
	ASSIGN
	EQUELS      //==
	LESSGREATER //> OR <
	SUM         //+
	PRODUCT     //*
	PREFIX      //-X OR !X
	CALL        //myFunction(X)
	INDEX
	INCREMENT
)

var precedences = map[token.TokenType]int{
	token.ASSIGN:    ASSIGN,
	token.EQ:        EQUELS,
	token.NOT_EQ:    EQUELS,
	token.LT:        LESSGREATER,
	token.GT:        LESSGREATER,
	token.PLUS:      SUM,
	token.MINUS:     SUM,
	token.SLASH:     PRODUCT,
	token.ASTERISK:  PRODUCT,
	token.LPAREN:    CALL,
	token.DOT:       CALL,
	token.LBRACKET:  INDEX,
	token.INCREMENT: INCREMENT,
	token.DECREMENT: INCREMENT,
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn) //mapの初期化(makeは指定された型の、初期化された使用できるようにしたマップを返す)
	p.registerPrefix(token.IDENT, p.parseIdentifier)           //識別子型の構文解析関数の登録
	p.registerPrefix(token.INT, p.parseIntegerLiteral)         //整数リテラル型の構文解析関数の登録
	p.registerPrefix(token.BANG, p.parsePrefixExpression)      //前置!の構文解析関数の登録
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)     //前置-の構文解析関数の登録
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(token.LBRACE, p.parseHashLiteral)
	p.registerPrefix(token.WHILE, p.parseWhileExpression)
	p.registerPrefix(token.NEW, p.parseNewExpression)
	p.registerPrefix(token.FOR, p.parseForLoopExpression)
	p.registerPrefix(token.INCREMENT, p.parsePrefixExpression)
	p.registerPrefix(token.DECREMENT, p.parsePrefixExpression)

	p.infixParseFns = make(map[token.TokenType]infixParseFn) //mapの初期化(makeは指定された型の、初期化された使用できるようにしたマップを返す)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression) //添字演算式の構文解析関数
	p.registerInfix(token.DOT, p.parseMethodCallExpression)
	p.registerInfix(token.INCREMENT, p.parsePostfixExpression)
	p.registerInfix(token.DECREMENT, p.parsePostfixExpression)
	p.registerInfix(token.ASSIGN, p.parseAssignExpression)

	//２つのトークンを読み込む
	p.nextToken()
	p.nextToken()

	return p
}

//*ast.Identifierを返却する
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

//Parserを受け取って、astを返却する。 メインの処理文
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{} //最初にastルートノードの作成。
	program.Statements = []ast.Statement{}

	if p.curTokenIs(token.SEMICOLON) && p.peekTokenIs(token.EOF) {
		return program
	}

	for p.curToken.Type != token.EOF { //EOFトークンに達するまで入力のトークンを繰り返して読む。
		stmt := p.parseStatement() //どんな種類の文かを判断し、そのstatementを返却する。
		if stmt != nil {
			program.Statements = append(program.Statements, stmt) //Statementsに追加する.
			//これはルートノードにあるスライスだった。
		}
		p.nextToken() //token.EOFの次へ...(次のStatementへ)
	}
	return program
}

//どんな種類の文かを判断し、そのstatementを返却する。
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.FUNC_DEC:
		return p.parseFunctionStatement()
	case token.CLASS:
		return p.parseClassStatement()
	default:
		return p.parseExpressionStatement() //letでもreturnでもなかったら
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)

	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt

}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken} //ASTNodeの構築

	stmt.Expression = p.parseExpression(LOWEST) //最初はLOWESTで始める。

	if p.peekTokenIs(token.SEMICOLON) { //文末セミコロンは省略可能REPLで楽になる。
		p.nextToken()
	}
	return stmt
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

//構文解析器で見られるアサーション関数である。peekTokenの形をcheckし、その型が正しい場合に限ってnextTokenを呼んで、tokenを進める。
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}
func (p *Parser) Errors() []string {
	return p.errors
}

//peekTokenが期待にそぐわない場合、errorsスライスにmsgを追加
func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s,got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

type (
	prefixParseFn func() ast.Expression                          //前置構文解析関数,-1とか
	infixParseFn  func(expression ast.Expression) ast.Expression //中置構文解析関数、引数は演算子の左側,5*8とか
)

//tokenTypeに応じて適切な関数を追加する。
func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}
func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

//式の構文解析関数
func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type] //p.curToken.Typeの前置に関連つけられた構文解析関数があるかを確認
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix() //一回目は現在のトークンに結びついた前置構文解析関数を実行

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() { //次の演算子トークンの左結合力が現在の右結合力(precedenc)よりも高いかを判定する
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp) //二回目以降は前回の結果を用いている。
	}

	return leftExp
}

//整数リテラルの構文解析関数
//トークンのリテラルをintにparseして、&ast.IntegerLiteralのvalueに入れる。
func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value

	return lit
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg) //構文解析器のerrorsに追加する。
}

//rightには、expressionをparseした値を入れる。()
func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,         //前置トークン
		Operator: p.curToken.Literal, //前置文字列
	}
	p.nextToken() //トークンを次へ進める(前置の次の式へ)

	expression.Right = p.parseExpression(PREFIX) //前置の右、つまりトークンを進めたあとのものを式としてparseした値

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,         //前置トークン
		Operator: p.curToken.Literal, //前置文字列
		Left:     left,               //前置文字列
	}
	precedences := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedences)
	return expression
}

func (p *Parser) parsePostfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.PostfixExpression{Token: p.curToken, Left: left, Operator: p.curToken.Literal}
	return expression
}

//tokenを読み込んで、それの優先順位を把握する
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}
func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

//現在の値がBooleanでTRUEガFLASEが判定し、BOOLEANそのままトークンの値に使用
func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	expression.Consequence = p.parseBlockStatement() //curTokenが{に来た時にする

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		expression.Alternative = p.parseBlockStatement()
	}

	return expression

}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}
	return block
}

//function genmaru(x){x+1}
func (p *Parser) parseFunctionStatement() *ast.FunctionStatement {
	stmt := &ast.FunctionStatement{Token: p.curToken}
	p.nextToken()
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	stmt.FunctionLiteral = p.parseFunctionLiteral().(*ast.FunctionLiteral)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	lit.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	lit.Body = p.parseBlockStatement()

	return lit
}

//カンマで区切られたリストから識別子を繰り返し構築し、パラメータのスライスを組み立てる。
func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	//条件式の最後にきたら(リストが空の場合はすぐに終わるようになっている)
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers //からのスライスが返される。
	}
	//パラメータが存在する場合以下
	p.nextToken()

	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() //この時点コンマ
		p.nextToken() //この時点次のパラメータ(引数)
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return identifiers

}

//呼び出し式の構文解析関数
func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	//空の場合
	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	//空じゃない場合
	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list

}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}
	array.Elements = p.parseExpressionList(token.RBRACKET)
	return array
}

//添字演算式の構文解析関数
func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}

	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACKET) { //tokenの次に]がなかったらエラー
		return nil
	}
	return exp
}

//ハッシュリテラルの構文解析。
func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curToken}
	hash.Pairs = make(map[ast.Expression]ast.Expression)

	for !p.peekTokenIs(token.RBRACE) { //空ハッシュじゃなかったら
		p.nextToken()
		key := p.parseExpression(LOWEST) //keyを解析

		if !p.expectPeek(token.COLON) { //keyの次に:がなかったらエラー
			return nil
		}

		p.nextToken()
		value := p.parseExpression(LOWEST) //valueを解析

		hash.Pairs[key] = value

		if !p.peekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) { //valueの後に}がなくて、
			return nil
		}
	}
	if !p.expectPeek(token.RBRACE) { //末尾に}がないならエラー
		return nil
	}
	return hash
}

//While構文解析
func (p *Parser) parseWhileExpression() ast.Expression {
	expression := &ast.WhileExpression{Token: p.curToken}
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST) //条件式のparse

	if !p.expectPeek(token.RPAREN) { //ifがokならすすむ
		return nil
	}

	if !p.expectPeek(token.LBRACE) { //ifがokならすすむ
		return nil
	}

	expression.Consequence = p.parseBlockStatement() //curTokenが{に来た時にする

	return expression
}

//class宣言の{}部分のparse
func (p *Parser) parseClassLiteral() ast.Expression {
	cls := &ast.ClassLiteral{
		Token:   p.curToken,
		Members: make([]*ast.LetStatement, 0),
		Methods: make(map[string]*ast.FunctionStatement),
	}
	p.nextToken()

	if !p.curTokenIs(token.LBRACE) {
		return nil
	}
	cls.Block = p.parseBlockStatement()
	for _, statement := range cls.Block.Statements {
		switch s := statement.(type) {
		case *ast.LetStatement:
			cls.Members = append(cls.Members, s)
		case *ast.FunctionStatement:
			cls.Methods[s.Name.String()] = s
		default:
			return nil
		}
	}
	return cls
}

//class文のparse
func (p *Parser) parseClassStatement() *ast.ClassStatement { //CLASStokenから始まる
	stmt := &ast.ClassStatement{Token: p.curToken}
	p.nextToken()
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	stmt.ClassLiteral = p.parseClassLiteral().(*ast.ClassLiteral)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	stmt.ClassLiteral.Name = stmt.Name.Value
	return stmt
}

//new
func (p *Parser) parseNewExpression() ast.Expression {
	newExp := &ast.NewExpression{Token: p.curToken}

	p.nextToken()
	exp := p.parseExpression(LOWEST)

	call, ok := exp.(*ast.CallExpression)
	if !ok {
		return nil
	}

	newExp.Class = call.Function

	return newExp
}

func (p *Parser) parseMethodCallExpression(obj ast.Expression) ast.Expression {
	methodCall := &ast.MethodCallExpression{Token: p.curToken, Object: obj}
	p.nextToken()

	name := p.parseIdentifier()
	if !p.peekTokenIs(token.LPAREN) {
		methodCall.Call = p.parseExpression(CALL)
	} else {
		p.nextToken()
		methodCall.Call = p.parseCallExpression(name)
	}
	return methodCall
}

func (p *Parser) parseForLoopExpression() ast.Expression {

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	init := p.parseExpression(LOWEST)
	p.nextToken()
	p.nextToken()
	condition := p.parseExpression(LOWEST)
	p.nextToken()
	p.nextToken()
	update := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	loop := &ast.ForLoop{Token: p.curToken, Init: init, Cond: condition, Update: update}
	loop.Block = p.parseBlockStatement()

	fmt.Println(loop)

	return loop
}

func (p *Parser) parseAssignExpression(name ast.Expression) ast.Expression {
	e := &ast.AssignExpression{Token: p.curToken}
	if n, ok := name.(*ast.Identifier); ok {
		e.Name = n
	}
	p.nextToken()
	e.Value = p.parseExpression(LOWEST)
	return e
}
