package repl

import (
	"fmt"
	"io"
	"monkey/evaluator"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"strings"

	"github.com/chzyer/readline"
)

const PROMPT = "genmaru >> "

func Start(in io.Reader, out io.Writer) {
	env := object.NewEnvironment()

	l, err := readline.NewEx(&readline.Config{
		Prompt:              "\033[34m»»»»\033[0m ",
		HistoryFile:         "/readline.tmp",
		InterruptPrompt:     "^C",
		EOFPrompt:           "お疲れさまでした。",
		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()

	for {
		fmt.Printf(PROMPT)
		line, err := l.Readline() //ここをカスタマイズする必要あり
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		l := lexer.New(line)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		evaluated := evaluator.Eval(program, env)

		switch {
		case line == "":
		default:
			if evaluated != nil {
				io.WriteString(out, evaluated.Inspect()) //結果を出力する(Inspectはstring)
				io.WriteString(out, "\n")
			}
		}
	}
}

func printParserErrors(out io.Writer, errors []string) {
	io.WriteString(out, "Woops! We ran into some monkey business here!\n")
	io.WriteString(out, " parser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}
