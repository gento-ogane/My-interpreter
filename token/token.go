package token

type TokenType string

type Token struct {
	Type TokenType
	Literal string
}

const (
	ILLEGAL = "ILLEGAL"
	EOF = "EOF"

	IDENT = "IDENT"
	INT = "INT"

	//演算子
	ASSIGN = "="
	PLUS = "+"

	//デリミタ
	COMMA = ","
	SEMICOLON = ";"

	LPAREN = "("
	RPAREN = ")"
	LBRACE = "{"
	RBRACE = "}"

	//キーワード
	FUNCTION = "FUNCTION"
	LET ="LET"
)

var keywords = map[string]TokenType{
	"fn": FUNCTION,
	"let": LET,
}

//渡された識別子がキーワードかどうかを確認、違うのならばTokenType定数を返す。
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok{
		 return tok
	} else {
		return IDENT
	}
}
