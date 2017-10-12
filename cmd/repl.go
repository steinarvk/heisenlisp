package cmd

import (
	"bufio"
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
	"github.com/spf13/cobra"
	"github.com/steinarvk/heisenlisp/builtin"
	"github.com/steinarvk/heisenlisp/code"
	"github.com/steinarvk/heisenlisp/expr"
	"github.com/steinarvk/heisenlisp/gen/parser"
	"github.com/steinarvk/heisenlisp/types"
)

var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "Enters an interactive REPL",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runREPL(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(replCmd)
}

var (
	replScript        *string
	replListenAddress *string
	replMetrics       *bool
)

func init() {
	replScript = replCmd.Flags().String("script", "", "execute script in filename instead of stdin")
	replListenAddress = replCmd.Flags().String("listen_address", "127.0.0.1:6860", "http address on which to serve metrics")
	replMetrics = replCmd.Flags().Bool("metrics", true, "whether to serve metrics")
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

	if *replScript != "" {
		_, err = code.RunFile(root, *replScript)
		if err != nil {
			return err
		}
	}

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

func serveMetrics(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	http.Handle("/metrics", promhttp.Handler())
	log.Printf("listening on: http://%s/metrics", listener.Addr())
	go func() {
		// Listen forever, unless something goes wrong.
		log.Fatal(http.Serve(listener, nil))
	}()

	return nil
}

func runREPL() error {
	if *replMetrics {
		if err := serveMetrics(*replListenAddress); err != nil {
			log.Printf("unable to serve metrics: %v", err)
		}
	}

	return mainCoreREPL()
}
