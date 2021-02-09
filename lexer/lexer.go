package lexer

import "monkey/token"

type Lexer struct {
	input string
	position int// 入力における現在の文字
	readPosition int // これから読み込む位置(現在の文字の次)
	ch byte //現在検査中の文字, 慣習的に数値量ではなく生データであることを示す
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return  l
}

//lはレシーバー,*はポインタ. tokenを読み終わって、positionをずらすため
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

func (l *Lexer) NextToken() token.Token  {
	var tok token.Token

	switch l.ch {
	case '=':
		tok = newToken(token.ASSIGN, l.ch)
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch){
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}
	l.readChar()
	return tok
}

//tokenTypeとchからtokenを生成する。
func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{ Type: tokenType, Literal: string(ch) }
}

//識別子を読んで、非英字に到達するまで字句解析器の位置を進めていく。
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch){
		l.readChar()
	}
	return l.input[ position:l.position ]
}

//judge that is alphabet
func isLetter(ch byte) bool {
	return  'a' <= ch && ch <='z' || 'A' <= ch && ch <='Z' || ch == '_' //_も英字として認識する。
}




