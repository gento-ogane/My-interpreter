package ast

import (
	"monkey/token"
)

type Node interface {
	TokenLiteral() string
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

type Program struct {
	Statements []Statement
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

type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
