package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/steinarvk/heisenlisp/gen/parser"
)

func mainCore() error {
	wr := bufio.NewWriter(os.Stdout)
	scanner := bufio.NewScanner(os.Stdin)

	prompt := "..? "

	wr.Write([]byte(prompt))
	wr.Flush()

	for scanner.Scan() {
		text := scanner.Text()

		if strings.TrimSpace(text) == "" {
			wr.Write([]byte(prompt))
			wr.Flush()
			continue
		}

		rv, err := parser.Parse("<stdin>", []byte(text))
		if err != nil {
			wr.Write([]byte(fmt.Sprintf("==! parsing error: %v\n", err)))
		} else {
			wr.Write([]byte(fmt.Sprintf("==> %v\n", rv)))
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
	if err := mainCore(); err != nil {
		log.Printf("fatal: %v", err)
		os.Exit(1)
	}
}
