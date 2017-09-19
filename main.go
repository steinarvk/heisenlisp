package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/fatih/color"

	"github.com/steinarvk/heisenlisp/builtin"
	"github.com/steinarvk/heisenlisp/code"
	"github.com/steinarvk/heisenlisp/types"

	"github.com/steinarvk/heisenlisp/gen/parser"
)

var (
	script  = flag.String("script", "", "execute script in filename instead of stdin")
	verbose = flag.Bool("verbose", false, "increase verbosity level")
)

func mainCoreExecuteScript(filename string) error {
	value, err := code.RunFile(builtin.NewRootEnv(), filename)
	if err != nil {
		return err
	}

	fmt.Println(value)

	return nil
}

func mainCoreREPL() error {
	wr := bufio.NewWriter(os.Stdout)

	reader, err := readline.NewEx(&readline.Config{
		Prompt:      color.GreenString("..? "),
		HistoryFile: ".heisenlisp_history",
	})
	if err != nil {
		return err
	}
	defer reader.Close()

	root := builtin.NewRootEnv()

	for {
		text, err := reader.Readline()
		if err == readline.ErrInterrupt {
			if len(text) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if strings.TrimSpace(text) == "" {
			continue
		}

		verboseString := func(s string) string {
			if *verbose {
				return s
			}
			return ""
		}

		expressionsIntf, err := parser.Parse("<stdin>", []byte(text))
		if err != nil {
			wr.Write([]byte(fmt.Sprintf("==! parsing error: %v\n", err)))
		} else {
			expressions := expressionsIntf.([]interface{})

			for _, expression := range expressions {
				if *verbose {
					wr.Write([]byte(fmt.Sprintf("%s%s %v\n", color.MagentaString(verboseString("(read) ")), color.MagentaString("==>"), expression)))
				}

				evaled, err := expression.(types.Value).Eval(root)
				if err != nil {
					wr.Write([]byte(fmt.Sprintf("==! eval error: %v\n", err)))
				} else {
					wr.Write([]byte(fmt.Sprintf("%s%s %v\n", color.YellowString(verboseString("(eval) ")), color.YellowString("==>"), evaled)))
				}
			}
		}
		wr.Flush()
	}

	return nil
}

func main() {
	flag.Parse()

	if *verbose {
		builtin.Verbose = true
	}

	if *script != "" {
		if err := mainCoreExecuteScript(*script); err != nil {
			log.Printf("fatal: %v", err)
			os.Exit(1)
		}
	} else if err := mainCoreREPL(); err != nil {
		log.Printf("fatal: %v", err)
		os.Exit(1)
	}
}
