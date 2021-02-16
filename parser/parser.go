package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
)

type Parser struct {
	l         *lexer.Lexer //字句解析インスタンスへのポインタ
	errors    []string
	curToken  token.Token //現在のToken
	peekToken token.Token //次のToken
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}} //?これはなにをしているのか？なぜ参照代入？
	//２つのトークンを読み込む
	p.nextToken()
	p.nextToken()

	return p
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
		stmt := p.parseStatement() //どんな種類の文かを判断し、そのstatementを返却する。ない場合は、nil
		if stmt != nil {
			program.Statements = append(program.Statements, stmt) //Statementsに追加する.
			//これはルートノードにあるスライスだった。
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return nil
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
