package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"testing"
)

func TestLetStatements(t *testing.T) {
	input := `
let x=5;
let y =10;
let foobar = 838383;
`
	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)
	// parseが何も返さない場合
	if program == nil {
		t.Fatalf("ParseProgram() return nil")
	}

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d", len(program.Statements))
	}

	tests := []struct {
		expectedIdentifier string
	}{
		{"x"},
		{"y"},
		{"foobar"},
	}

	for i, tt := range tests {
		stmt := program.Statements[i]
		if !testLetStatement(t, stmt, tt.expectedIdentifier) {
			return
		}
	}
}

func testLetStatement(t *testing.T, s ast.Statement, name string) bool {
	if s.TokenLiteral() != "let" {
		t.Errorf("s.TokenLiteral not 'let'. got=%q", s.TokenLiteral())
		return false
	}

	letStmt, ok := s.(*ast.LetStatement)
	if !ok {
		t.Errorf("s not *ast.LetStatement. got=%T", s)
		return false
	}
	if letStmt.Name.Value != name {
		t.Errorf("letStmt.Name.Value not '%s'. got=%s", name, letStmt.Name.Value)
		return false
	}
	if letStmt.Name.TokenLiteral() != name {
		t.Errorf("letStmt.Name.TokenLiteral not '%s'. got=%s", name, letStmt.Name.TokenLiteral())
		return false
	}
	return true
}

//errorを検知し、あれば出力する。
func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parse error: %q", msg)
	}
	t.FailNow()
}

func TestReturnStatements(t *testing.T) {
	input := `
return 5;
return 10;
return 993322;
`
	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d", len(program.Statements))
	}

	for _, stmt := range program.Statements {
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Errorf("stmt not *ast.returnStatement. got=%T", stmt)
			continue
		}
		if returnStmt.TokenLiteral() != "return" {
			t.Errorf("returnStmt.TokenLiteral not 'return'. got=%q", returnStmt.TokenLiteral())
		}
	}

}

//識別子オンリーの式構文解析
func TestIdentifierExpression(t *testing.T) {
	input := "foobar;"

	l := lexer.New(input)
	t.Log(*l)
	p := New(l)
	t.Log(p)
	program := p.ParseProgram()
	t.Log(program)
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has not enough statements. got =%d", len(program.Statements))
	}

	//program.Statementsに含まれる唯一の文が*ast.ExpressionStatementであることの確認
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement) //okはmapに含まれているかどうかの確認
	if !ok {
		t.Fatalf("program.Satatements[0] is not *astExpressionStatements. got =%T", program.Statements[0])
	}
	//同上
	ident, ok := stmt.Expression.(*ast.Identifier) //okはmapに含まれているかどうかの確認
	if !ok {
		t.Fatalf("exp not *ast.Identifier. got =%T", stmt.Expression)
	}
	//識別子が正しいか確認
	if ident.Value != "foobar" {
		t.Errorf("ident.Value not %s. got=%s", "foobar", ident.Value)
	}
	if ident.TokenLiteral() != "foobar" {
		t.Errorf("ident.TokenLiteral not %s. got=%s", "foobar", ident.TokenLiteral())
	}

}

//整数リテラルオンリーの式構文解析
func TestIntegerLiteralExpression(t *testing.T) {
	input := "5;"

	l := lexer.New(input)
	t.Log(*l)
	p := New(l)
	t.Log(p)
	program := p.ParseProgram()
	t.Log(program)
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has not enough statements. got =%d", len(program.Statements))
	}

	//program.Statementsに含まれる唯一の文が*ast.ExpressionStatementであることの確認
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement) //okはmapに含まれているかどうかの確認
	if !ok {
		t.Fatalf("program.Satatements[0] is not *astExpressionStatements. got =%T", program.Statements[0])
	}
	//同上
	literal, ok := stmt.Expression.(*ast.IntegerLiteral) //okはmapに含まれているかどうかの確認
	if !ok {
		t.Fatalf("exp not *ast.IntegerLiteral. got =%T", stmt.Expression)
	}
	//識別子が正しいか確認
	if literal.Value != 5 {
		t.Errorf("literal.Value not %d. got=%d", 5, literal.Value)
	}
	if literal.TokenLiteral() != "5" {
		t.Errorf("literal.TokenLiteral not %s. got=%s", "5", literal.TokenLiteral())
	}

}

func TestParsingPrefixExpressions(t *testing.T) {
	prefixTests := []struct {
		input        string
		operator     string
		integerValue int64
	}{
		{"!5;", "!", 5},
		{"-15", "-", 15},
	}
	//一つ一つをテスト
	for _, tt := range prefixTests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does contain %d staments. got =%d\n", 1, len(program.Statements))
		}
		//program.Statementsに含まれる唯一の文が*ast.ExpressionStatementであることの確認
		stmt, ok := program.Statements[0].(*ast.ExpressionStatement) //okはmapに含まれているかどうかの確認
		if !ok {
			t.Fatalf("program.Satatements[0] is not ast.ExpressionStatements. got =%T",
				program.Statements[0])
		}
		//同上
		exp, ok := stmt.Expression.(*ast.PrefixExpression) //okはmapに含まれているかどうかの確認
		if !ok {
			t.Fatalf("stmt not *ast.PrefixExpression. got =%T", stmt.Expression)
		}
		//識別子が正しいか確認
		if exp.Operator != tt.operator {
			t.Fatalf("exp.Operator is not '%s'. got=%s", tt.operator, exp.Operator)
		}
		//正しい整数リテラルか判定
		if !testIntegerLiteral(t, exp.Right, tt.integerValue) {
			return
		}
	}
}

//正しい整数リテラルか判定
func testIntegerLiteral(t *testing.T, il ast.Expression, value int64) bool {
	integ, ok := il.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("il not *ast.IntegerLiteral. got=%T", il)
		return false
	}
	if integ.Value != value {
		t.Errorf("integ.Value not %d. got=%d", value, integ.Value)
		return false
	}
	if integ.TokenLiteral() != fmt.Sprintf("%d", value) {
		t.Errorf("integ.TokenLiteral not %d. got =%s", value, integ.TokenLiteral())
		return false
	}
	return true //testはpass
}
