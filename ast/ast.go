package ast

import (
	"bytes"
	"monkey/token"
	"strings"
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

//前置演算子式
type PrefixExpression struct {
	Token    token.Token //前置トークン.ex「!」など
	Operator string      //前置オペレーターの文字列
	Right    Expression  //右にかかる式(値)
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")
	//わざと()でくくることで、どのオペランドがどの演算子に属するのかがわかる

	return out.String()
}

type InfixExpression struct {
	Token    token.Token //演算子トークン、例えば「+」
	Left     Expression
	Operator string
	Right    Expression
}

func (oe *InfixExpression) expressionNode()      {}
func (oe *InfixExpression) TokenLiteral() string { return oe.Token.Literal }
func (oe *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(oe.Left.String())
	out.WriteString(" " + oe.Operator + " ")
	out.WriteString(oe.Right.String())
	out.WriteString(")")
	//わざと()でくくることで、どのオペランドがどの演算子に属するのかがわかる

	return out.String()
}

type PostfixExpression struct {
	Token    token.Token //演算子トークン、例えば「+」
	Left     Expression
	Operator string
}

func (pe *PostfixExpression) expressionNode()      {}
func (pe *PostfixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PostfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Left.String())
	out.WriteString(pe.Operator)
	out.WriteString(")")
	//わざと()でくくることで、どのオペランドがどの演算子に属するのかがわかる

	return out.String()
}

type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }
func (b *Boolean) String() string       { return b.Token.Literal }

type IfExpression struct {
	Token       token.Token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	var out bytes.Buffer

	out.WriteString("if")
	out.WriteString(ie.Condition.String())
	out.WriteString(" ")
	out.WriteString(ie.Condition.String())

	if ie.Alternative != nil {
		out.WriteString("else ")
		out.WriteString(ie.Alternative.String())
	}

	return out.String()
}

type BlockStatement struct {
	Token      token.Token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer

	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

type FunctionLiteral struct {
	Token      token.Token   //'fn'トークン
	Parameters []*Identifier //識別子のスライス
	Body       *BlockStatement
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())

	}
	out.WriteString(fl.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ""))
	out.WriteString(")")
	out.WriteString(fl.Body.String())

	return out.String()
}

type FunctionStatement struct {
	Token           token.Token
	Name            *Identifier
	FunctionLiteral *FunctionLiteral
}

func (f *FunctionStatement) statementNode()       {}
func (f *FunctionStatement) TokenLiteral() string { return f.Token.Literal }

func (f *FunctionStatement) String() string {
	var out bytes.Buffer

	out.WriteString("fn ")
	out.WriteString(f.Name.String())
	params := []string{}
	for _, p := range f.FunctionLiteral.Parameters {
		params = append(params, p.String())
	}
	out.WriteString(" (")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString("{ ")
	out.WriteString(f.FunctionLiteral.Body.String())
	out.WriteString(" }")
	//
	return out.String()
}

type CallExpression struct {
	Token     token.Token
	Function  Expression   //Identifier or FunctionLiteral
	Arguments []Expression //引数たち
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer

	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}

type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return sl.Token.Literal }

type ArrayLiteral struct {
	Token    token.Token
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer

	elements := []string{}
	for _, el := range al.Elements {
		elements = append(elements, el.String())
	}
	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

type IndexExpression struct {
	Token token.Token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("])")

	return out.String()
}

type HashLiteral struct {
	Token token.Token //"{"トークン
	Pairs map[Expression]Expression
}

func (hl *HashLiteral) expressionNode()      {}
func (hl *HashLiteral) TokenLiteral() string { return hl.Token.Literal }
func (hl *HashLiteral) String() string {
	var out bytes.Buffer

	pairs := []string{}
	for key, value := range hl.Pairs {
		pairs = append(pairs, key.String()+":"+value.String())
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}

type WhileExpression struct {
	Token       token.Token //WhileToken
	Condition   Expression
	Consequence *BlockStatement
}

func (we *WhileExpression) expressionNode()      {}
func (we *WhileExpression) TokenLiteral() string { return we.Token.Literal }
func (we *WhileExpression) String() string {
	var out bytes.Buffer

	out.WriteString("WHILE")
	out.WriteString(we.Condition.String())
	out.WriteString(" ")
	out.WriteString(we.Consequence.String())

	return out.String()
}

type ClassLiteral struct {
	Token   token.Token //'class'トークン
	Name    string
	Members []*LetStatement //識別子のスライス
	Methods map[string]*FunctionStatement
	Body    *BlockStatement
	Block   *BlockStatement //mainly used for debugging purpose
}

func (c *ClassLiteral) expressionNode()      {}
func (c *ClassLiteral) TokenLiteral() string { return c.Token.Literal }

func (c *ClassLiteral) String() string {
	var out bytes.Buffer

	out.WriteString(c.TokenLiteral() + " ")
	out.WriteString(c.Name)

	out.WriteString("{ ")
	out.WriteString(c.Block.String())
	out.WriteString("} ")

	return out.String()
}

type ClassStatement struct {
	Token        token.Token
	Name         *Identifier //Class name
	ClassLiteral *ClassLiteral
}

func (c *ClassStatement) statementNode()       {}
func (c *ClassStatement) TokenLiteral() string { return c.Token.Literal }
func (c *ClassStatement) String() string {
	var out bytes.Buffer

	out.WriteString(c.Token.Literal + " ")
	out.WriteString(c.Name.Value)
	out.WriteString("{ ")
	out.WriteString(c.ClassLiteral.Block.String())
	out.WriteString(" }")

	return out.String()
}

//newトークンとクラス
type NewExpression struct {
	Token token.Token //newToken
	Class Expression  //class
}

func (n *NewExpression) expressionNode()      {}
func (n *NewExpression) TokenLiteral() string { return n.Token.Literal }
func (n *NewExpression) String() string {
	var out bytes.Buffer

	out.WriteString(n.TokenLiteral() + " ")
	out.WriteString(n.Class.String())
	out.WriteString("(")
	out.WriteString(") ")

	return out.String()
}

//.メソッドでの呼び出し式
type MethodCallExpression struct {
	Token  token.Token
	Object Expression //呼び出し元の値
	Call   Expression //呼び出し先の値
}

func (mc *MethodCallExpression) expressionNode()      {}
func (mc *MethodCallExpression) TokenLiteral() string { return mc.Token.Literal }
func (mc *MethodCallExpression) String() string {
	var out bytes.Buffer
	out.WriteString(mc.Object.String())
	out.WriteString(".")
	out.WriteString(mc.Call.String())

	return out.String()
}

type ForLoop struct {
	Token  token.Token
	Init   Expression
	Cond   Expression
	Update Expression
	Block  *BlockStatement
}

func (fl *ForLoop) expressionNode()      {}
func (fl *ForLoop) TokenLiteral() string { return fl.Token.Literal }

func (fl *ForLoop) String() string {
	var out bytes.Buffer

	out.WriteString("for")
	out.WriteString(" ( ")
	out.WriteString(fl.Init.String())
	out.WriteString(" ; ")
	out.WriteString(fl.Cond.String())
	out.WriteString(" ; ")
	out.WriteString(fl.Update.String())
	out.WriteString(" ) ")
	out.WriteString(" { ")
	out.WriteString(fl.Block.String())
	out.WriteString(" }")

	return out.String()
}

type AssignExpression struct {
	Token token.Token
	Name  Expression
	Value Expression
}

func (ae *AssignExpression) expressionNode()      {}
func (ae *AssignExpression) TokenLiteral() string { return ae.Token.Literal }

func (ae *AssignExpression) String() string {
	var out bytes.Buffer

	out.WriteString(ae.Name.String())
	out.WriteString(" = ")
	out.WriteString(ae.Token.Literal)
	out.WriteString(ae.Value.String())

	return out.String()
}
