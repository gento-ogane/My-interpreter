package token

type TokenType string

type Token struct {
	Type    TokenType //属性(識別子とか{とか数字とか)
	Literal string    //文字部(実体の部分)
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	IDENT = "IDENT"
	INT   = "INT"

	STRING = "STRING"

	//演算子
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"

	LT = "<"
	GT = ">"

	EQ        = "=="
	NOT_EQ    = "!="
	INCREMENT = "++"
	DECREMENT = "--"

	//デリミタ
	COMMA     = ","
	SEMICOLON = ";"

	COLON = ":"
	DOT   = "."

	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"

	//キーワード
	FUNCTION = "FUNCTION"
	FUNC_DEC = "FUNC_DEC"
	WHILE    = "WHILE"
	LET      = "LET"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
	CLASS    = "CLASS"
	NEW      = "NEW"
	FOR      = "FOR"
)

var keywords = map[string]TokenType{
	"fn":       FUNCTION,
	"function": FUNC_DEC,
	"while":    WHILE,
	"let":      LET,
	"true":     TRUE,
	"false":    FALSE,
	"if":       IF,
	"else":     ELSE,
	"return":   RETURN,
	"class":    CLASS,
	"new":      NEW,
	"for":      FOR,
}

//渡された識別子がキーワードかどうかを確認、違うのならばTokenType定数を返す。
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	} else {
		return IDENT
	}
}
