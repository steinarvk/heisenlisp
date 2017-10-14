package cmd

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/steinarvk/heisenlisp/builtin"
	"github.com/steinarvk/heisenlisp/tracing"
)

var (
	verbose             *bool
	calltracingFilename *string
	calltracingDetailed *bool
	calltracingBuffered *bool
	listenAddress       *string
	activateMetrics     *bool
	keepAliveAfter      *bool
)

func init() {
	verbose = RootCmd.PersistentFlags().Bool("verbose", false, "increase verbosity")
	calltracingFilename = RootCmd.PersistentFlags().String("output_calltrace", "", "write a JSON call trace to file")
	calltracingDetailed = RootCmd.PersistentFlags().Bool("detailed_calltrace", false, "enable detailed call tracing, even at severe performance cost")
	calltracingBuffered = RootCmd.PersistentFlags().Bool("buffered_calltrace", true, "use an output buffer for writing call trace")
	listenAddress = RootCmd.PersistentFlags().String("listen_address", "127.0.0.1:6860", "http address on which to serve metrics")
	activateMetrics = RootCmd.PersistentFlags().Bool("metrics", true, "serve Prometheus metrics")
	keepAliveAfter = RootCmd.PersistentFlags().Bool("keep_alive", false, "keep process alive after main command terminates (to serve metrics)")
}

const (
	calltracingBufsize = 1048576
)

var globalListener net.Listener

func getListener() (net.Listener, error) {
	if globalListener != nil {
		return globalListener, nil
	}

	listener, err := net.Listen("tcp", *listenAddress)
	if err != nil {
		return nil, err
	}
	globalListener = listener

	go func() {
		// Listen forever, unless something goes wrong.
		log.Fatal(http.Serve(globalListener, nil))
	}()

	return globalListener, nil
}

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
			} else {
				w = os.Stdout
			}

			if *calltracingBuffered {
				callTracingBufWriter = bufio.NewWriterSize(w, calltracingBufsize)
				tracing.Target = callTracingBufWriter
			} else {
				tracing.Target = w
			}

			tracing.Target.Write([]byte("[\n"))

			tracing.Enabled = true
		}

		if *activateMetrics {
			lst, err := getListener()
			if err != nil {
				log.Printf("error: unable to get listener for metrics: %v", err)
			} else {
				http.Handle("/metrics", promhttp.Handler())
				log.Printf("serving metrics on: http://%s/metrics", lst.Addr())
			}
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if callTracingBufWriter != nil {
			callTracingBufWriter.Flush()
		}
		if tracing.Enabled {
			tracing.Target.Write([]byte("]\n"))
		}
		if callTracingFile != nil {
			if err := callTracingFile.Close(); err != nil {
				log.Fatal(err)
			}
		}

		if *keepAliveAfter {
			log.Printf("command completed, keeping process alive")
			for {
				time.Sleep(time.Hour)
			}
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		if err := runREPL(); err != nil {
			log.Fatal(err)
		}
	},
}
