package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/steinarvk/heisenlisp/builtin"
	"github.com/steinarvk/heisenlisp/code"
	"github.com/steinarvk/heisenlisp/expr"
	"github.com/steinarvk/heisenlisp/types"

	"github.com/steinarvk/heisenlisp/gen/parser"
)

var (
	script        = flag.String("script", "", "execute script in filename instead of stdin")
	verbose       = flag.Bool("verbose", false, "increase verbosity level")
	listenAddress = flag.String("listen_address", "127.0.0.1:6860", "http address on which to serve metrics")
	metrics       = flag.Bool("metrics", true, "whether to serve metrics")
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

	builtin.Unary(root, "_load-file!", func(a types.Value) (types.Value, error) {
		s, err := expr.StringValue(a)
		if err != nil {
			return nil, err
		}
		return code.RunFile(root, s)
	})

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

	if *metrics {
		listener, err := net.Listen("tcp", *listenAddress)
		if err != nil {
			log.Fatal(err)
		}

		http.Handle("/metrics", promhttp.Handler())
		log.Printf("listening on: http://%s/metrics", listener.Addr())
		go func() {
			// Listen forever, unless something goes wrong.
			log.Fatal(http.Serve(listener, nil))
		}()
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
