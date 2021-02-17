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

	for p.curToken.Type != token.EOF { //EOFトークンに達するまで入力のトークンを繰り返して読む。
		stmt := p.parseStatement() //どんな種類の文かを判断し、そのstatementを返却する。
		if stmt != nil {
			program.Statements = append(program.Statements, stmt) //Statementsに追加する.
			//これはルートノードにあるスライスだった。
		}
		p.nextToken()
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
	default:
		return p.parseExpressionStatement()
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

	//TODO:セミコロンに遭遇するまで式を読み飛ばしてしまっている
	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt

}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken} //ASTNodeの構築

	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) { //セミコロンは省略可能REPLで楽になる。
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

const (
	_int = iota
	LOWEST
	EQUELS      //==
	LESSGREATER //> OR <
	SUM         //+
	PRODUCT     //*
	PREFIX      //-X OR !X
	CALL        //myFunction(X)
)

//式の構文解析関数
func (p *Parser) parseExpression(produce int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type] //p.curToken.Typeの前置に関連つけられた構文解析関数があるかを確認
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()
	return leftExp
}

//整数リテラルの構文解析関数
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

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}
