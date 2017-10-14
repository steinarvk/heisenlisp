package cmd

import (
	"bufio"
	"io"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/steinarvk/heisenlisp/builtin"
	"github.com/steinarvk/heisenlisp/tracing"
)

var (
	verbose             *bool
	calltracingFilename *string
	calltracingDetailed *bool
)

func init() {
	verbose = RootCmd.PersistentFlags().Bool("verbose", false, "increase verbosity")
	calltracingFilename = RootCmd.PersistentFlags().String("output_calltrace", "", "write a JSON call trace to file")
	calltracingDetailed = RootCmd.PersistentFlags().Bool("detailed_calltrace", false, "enable detailed call tracing, even at severe performance cost")
}

const (
	calltracingBufsize = 1048576
)

var callTracingFile *os.File
var callTracingBufWriter *bufio.Writer

var RootCmd = &cobra.Command{
	Use:   "heisenlisp",
	Short: "Heisenlisp is a Lisp with features for uncertainty",
	Long:  "Heisenlisp is a Lisp with features for uncertainty",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		builtin.Verbose = *verbose
		tracing.Detailed = *calltracingDetailed

		if *calltracingFilename != "" {
			var w io.Writer

			if *calltracingFilename != "-" {
				f, err := os.OpenFile(*calltracingFilename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0640)
				if err != nil {
					log.Fatalf("error opening %q (to write trace): %v", *calltracingFilename, err)
				}
				callTracingFile = f
				w = f
			}

			callTracingBufWriter = bufio.NewWriterSize(w, calltracingBufsize)
			callTracingBufWriter.Write([]byte("[\n"))

			tracing.Target = callTracingBufWriter

			tracing.Enabled = true
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if callTracingBufWriter != nil {
			callTracingBufWriter.Write([]byte("]\n"))
			callTracingBufWriter.Flush()
		}
		if callTracingFile != nil {
			if err := callTracingFile.Close(); err != nil {
				log.Fatal(err)
			}
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		if err := runREPL(); err != nil {
			log.Fatal(err)
		}
	},
}
