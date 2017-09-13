package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/steinarvk/heisenlisp/builtin"
	"github.com/steinarvk/heisenlisp/code"
	"github.com/steinarvk/heisenlisp/types"

	"github.com/steinarvk/heisenlisp/gen/parser"
)

var (
	script = flag.String("script", "", "execute script in filename instead of stdin")
)

func mainCoreExecuteScript(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	value, err := code.Run(builtin.NewRootEnv(), filename, data)
	if err != nil {
		return err
	}

	fmt.Println(value)

	return nil
}

func mainCoreREPL() error {
	wr := bufio.NewWriter(os.Stdout)
	scanner := bufio.NewScanner(os.Stdin)

	prompt := "..? "

	root := builtin.NewRootEnv()

	wr.Write([]byte(prompt))
	wr.Flush()

	for scanner.Scan() {
		text := scanner.Text()

		if strings.TrimSpace(text) == "" {
			wr.Write([]byte(prompt))
			wr.Flush()
			continue
		}

		expressionsIntf, err := parser.Parse("<stdin>", []byte(text))
		if err != nil {
			wr.Write([]byte(fmt.Sprintf("==! parsing error: %v\n", err)))
		} else {
			expressions := expressionsIntf.([]interface{})

			for _, expression := range expressions {
				wr.Write([]byte(fmt.Sprintf("(read) ==> %v\n", expression)))

				evaled, err := expression.(types.Value).Eval(root)
				if err != nil {
					wr.Write([]byte(fmt.Sprintf("==! eval error: %v\n", err)))
				} else {
					wr.Write([]byte(fmt.Sprintf("(eval) ==> %v\n", evaled)))
				}
			}
		}

		wr.Write([]byte(prompt))
		wr.Flush()
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %v", err)
	}

	return nil
}

func main() {
	flag.Parse()

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
