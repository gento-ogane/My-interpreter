package lexer

import "monkey/token"

type Lexer struct {
	input        string
	position     int  // 入力における現在の文字位置
	readPosition int  // これから読み込む次の文字位置
	ch           byte //現在検査中の文字, 慣習的に数値量ではなく生データであることを示す
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

//lはレシーバー,*はポインタ. tokenを読み終わって、positionをずらすため
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 //ASCIIで"Null"の意味。ファイルの終わり
	} else {
		l.ch = l.input[l.readPosition] //次の文字をセット
	}
	l.position = l.readPosition
	l.readPosition += 1
}

//現在検査中のchを見て、その文字が何であるかに応じてトークンを返す。
func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace() //スペースを読み飛ばす。

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch //現在の文字
			l.readChar()
			literal := string(ch) + string(l.ch) //現在の文字+次の文字なので"=="
			tok = token.Token{Type: token.EQ, Literal: literal}
		} else {
			tok = newToken(token.ASSIGN, l.ch)
		}
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '-':
		tok = newToken(token.MINUS, l.ch)
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch) //現在の文字+次の文字なので"!="
			tok = token.Token{Type: token.NOT_EQ, Literal: literal}
		} else {
			tok = newToken(token.BANG, l.ch)
		}
	case '/':
		tok = newToken(token.SLASH, l.ch)
	case '*':
		tok = newToken(token.ASTERISK, l.ch)
	case '<':
		tok = newToken(token.LT, l.ch)
	case '>':
		tok = newToken(token.GT, l.ch)
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case '[':
		tok = newToken(token.LBRACKET, l.ch)
	case ']':
		tok = newToken(token.RBRACKET, l.ch)
	case ':':
		tok = newToken(token.COLON, l.ch)
	case '.':
		tok = newToken(token.DOT, l.ch)

	case '"': //string型の追加
		tok.Type = token.STRING
		tok.Literal = l.readString()

	case 0: //ASCIIの"Null"。何もない、ファイルの終わりを表す。
		tok.Literal = ""
		tok.Type = token.EOF
	default: //記号ではなく、文字列がでてきた場合、予約後か識別子かを区別する。
		if isLetter(l.ch) { //文字列だった場合
			tok.Literal = l.readIdentifier()          //文字列のまとまりを読む
			tok.Type = token.LookupIdent(tok.Literal) //識別子typeか予約後typeを判別して代入
			return tok                                //readChar()を呼ぶ必要がないため
		} else if isDigit(l.ch) { //数字だった場合
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			return tok
		} else { //その他搭載されていないILLEGALなToken
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}
	l.readChar()
	return tok //現在検査中のchを見て、その文字が何であるかに応じてトークンを返す。
}

//tokenTypeとchからtokenを生成する。
func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

//識別子を読んで、非英字に到達するまで字句解析器の位置を進めていく。
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

//judge that is a alphabet?
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' //_も英字として認識する。
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\n' || l.ch == '\r' || l.ch == '\t' {
		l.readChar()
	}
}

//識別子を読んで、非数字に到達するまで字句解析器の位置を進めていく。
func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		return l.input[l.readPosition]
	}
}

func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	return l.input[position:l.position]
}
