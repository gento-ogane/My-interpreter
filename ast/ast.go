package ast

import (
	"bytes"
	"monkey/token"
)

//部分木
type Node interface {
	TokenLiteral() string
	String() string
}

//文（Statement）は値を生成しない
type Statement interface {
	Node
	statementNode()
}

//式（Expression）は値を生成する
type Expression interface {
	Node
	expressionNode()
}

//programはstatements(文の集合)を保持する。
type Program struct {
	Statements []Statement
}

//デバッグ時にASTノードを表示したり、他のASTと比較したりできる。各Nodeに定義されたString()で実際に仕事している。
func (p *Program) String() string {
	var out bytes.Buffer //バッファの作成

	//バッファにそれぞれのString()の戻り値を格納する。
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String() //バッファを文字列として返却する。
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

//let文
type LetStatement struct {
	Token token.Token
	Name  *Identifier //識別子の名前が入る
	Value Expression  //代入する値が入る。
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }

func (ls *LetStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")

	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}
	out.WriteString(";")

	return out.String()
}

//return文
type ReturnStatement struct {
	Token       token.Token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }

func (rs *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString(rs.TokenLiteral() + " ")

	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}
	out.WriteString(";")

	return out.String()
}

//式文。行にある x + 5;のような単体式
type ExpressionStatement struct {
	Token      token.Token //式の最初のトークン
	Expression Expression  //式を保持する。よって、Statementスライスに入れることができる。
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }

func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

//識別子(値)
type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

//整数リテラル(値)
type IntegerLiteral struct {
	Token token.Token
	Value int64 //"5"という値を5に変換する必要がある。
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }
