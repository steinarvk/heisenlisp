package cmd

import (
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/steinarvk/heisenlisp/builtin"
	"github.com/steinarvk/heisenlisp/code"
	"github.com/steinarvk/heisenlisp/env"
	"github.com/steinarvk/secrets"

	_ "github.com/lib/pq"
)

var replServerCmd = &cobra.Command{
	Use:   "repl-server",
	Short: "Web server serving a REPL demo of Heisenlisp",
	Long: `repl-server starts up a web server serving an interface allowing users
to play around with Heisenlisp.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runReplServer(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(replServerCmd)
}

var (
	rsListenAddress        *string
	rsLoggingDatabaseCreds *string
)

func init() {
	rsListenAddress = replServerCmd.Flags().String("listen_address", "127.0.0.1:6861", "http address on which to serve")
	rsLoggingDatabaseCreds = replServerCmd.Flags().String("logging_database_credentials", "", "logging database credentials")
}

var (
	metricRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "hlisp",
			Name:      "repl_server_requests",
			Help:      "Requests to the REPL server",
		},
		[]string{"page"},
	)

	metricEvaluated = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "hlisp",
			Name:      "repl_server_evaluated",
			Help:      "Number of queries evaluated by the REPL server",
		},
	)

	metricEvaluationErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "hlisp",
			Name:      "repl_server_evaluation_errors",
			Help:      "Number of queries evaluated by the REPL server resulting in errors",
		},
	)
)

func init() {
	prometheus.MustRegister(metricRequests)
	prometheus.MustRegister(metricEvaluated)
	prometheus.MustRegister(metricEvaluationErrors)
}

var (
	mainPageData = []byte(`
	<html>
		<head>
			<title>heisenlisp REPL server</title>

			<script type="text/javascript"
						  src="https://cdnjs.cloudflare.com/ajax/libs/underscore.js/1.8.3/underscore-min.js">
			</script>

			<script
			  src="https://code.jquery.com/jquery-3.2.1.min.js"
				integrity="sha256-hwg4gsxgFZhOsEEamdOYGBf13FyQuiTwlAQgxVSNgt4="
				crossorigin="anonymous"></script>


		</head>
		<body>
			<div id="evaluated">
				<p>
					Enter some expressions in the text box below to evaluate them.
				</p>

				<p>
					Things to try:
					<ul>
						<li><code>(+ 2 2)</code></li>
						<li><code>(* 2 (any-of 10 30))</code></li>
						<li><code>(= (any-of 0 1) 1)</code></li>
					</ul>
				</p>

				<p>
					Note that queries may be logged and stored for debugging purposes.
					Don't submit anything that you don't want to be logged.
				</p>
			</div>
			<div id="form">
				<form>
					<input id="textfield" type="text" placeholder="(+ 2 2)">
					<input id="submitbutton" type="submit" value="Evaluate"></input>
				</form>
			</div>
			<style>
				#evaluated {
					position: absolute;
					top: 0;
					left: 0;
					width: 100%;
					height: 80%;
					overflow: scroll;
				}
				#form {
					position: fixed;
					top: 80%;
					left: 0;
					width: 100%;
					height: 20%;
				}

				.input {
					color: #2bb;
					margin: 0;
				}
				.output {
					color: #22b;
					margin: 0;
					margin-bottom: 1em;
				}
				.error {
					color: #b22;
					margin: 0;
					margin-bottom: 1em;
				}
			</style>
			<script type="text/javascript">
				const fmtSuccess = _.template("<p class=\"input\"><%- input %></p><p class=\"output\"><%- output %></p>");
				const fmtError = _.template("<p class=\"input\"><%- input %></p><p class=\"error\"><%- error %></p>");

				function addItem(obj) {
					const newdiv = document.createElement("div");
					if (obj.result) {
						newdiv.innerHTML = fmtSuccess({input: obj.input, output: obj.result});
					} else {
						newdiv.innerHTML = fmtError({input: obj.input, error: obj.error});
					}
				  document.getElementById("evaluated").appendChild(newdiv);
				}

				function tryEvaluate(expr) {
					$.post("/api/eval", expr).done(function(data) {
						addItem(data);
					})
				}

				$("#submitbutton").click(function() {
					tryEvaluate($("#textfield").val());
					$("#textfield").val("");
					return false;
				});
			</script>
		</body>
	</html>
`)
)

type databaseLogger struct {
	db             *sql.DB
	serverHostname string
}

func newDatabaseLogger(secretsFilename string) (*databaseLogger, error) {
	if secretsFilename == "" {
		return nil, nil
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	dbSecrets := secrets.Postgres{}
	if err := secrets.FromYAML(secretsFilename, &dbSecrets); err != nil {
		return nil, err
	}

	url, err := dbSecrets.AsURL()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	return &databaseLogger{db, hostname}, nil
}

func (d *databaseLogger) logRequest(t0 time.Time, req *http.Request, data []byte) (int64, error) {
	if d == nil {
		return 0, nil
	}

	jsonHeaders, err := json.Marshal(req.Header)
	if err != nil {
		return 0, err
	}

	var requestId int64

	err = d.db.QueryRow(`
		INSERT INTO request_log (
			timestamp_utcnano,
			server_hostname,
			client_hostname,
			http_headers,
			expr
		) VALUES (
			$1,
			$2,
			$3,
			$4,
			$5
		) RETURNING request_id
	`, t0.UnixNano(), d.serverHostname, req.Host, jsonHeaders, string(data)).Scan(&requestId)
	return requestId, err
}

func (d *databaseLogger) logResponse(t1 time.Time, dur time.Duration, id int64, result string, err error) error {
	if d == nil {
		return nil
	}

	var errorDesc *string
	var resultDesc *string

	if err != nil {
		s := err.Error()
		errorDesc = &s
	}
	if result != "" {
		resultDesc = &result
	}

	_, err = d.db.Exec(`
		INSERT INTO result_log (
			timestamp_utcnano,
			duration_nanos,
			request_id,
			result,
			error
		) VALUES (
			$1,
			$2,
			$3,
			$4,
			$5
		)`, t1.UnixNano(), int64(dur), id, resultDesc, errorDesc)
	return err
}

func runReplServer() error {
	flag.Parse()

	os.Unsetenv("PGPASSFILE")

	listener, err := net.Listen("tcp", *rsListenAddress)
	if err != nil {
		return err
	}

	root := builtin.NewRootEnv()

	var mu sync.Mutex

	evaluate := func(data []byte) (string, error) {
		mu.Lock()
		defer mu.Unlock()

		metricEvaluated.Inc()
		val, err := code.Run(env.New(root), "<request data>", data)
		if err != nil {
			metricEvaluationErrors.Inc()
			return "", err
		}

		return val.String(), nil
	}

	requestLogger, err := newDatabaseLogger(*rsLoggingDatabaseCreds)
	if err != nil {
		return err
	}

	handler := func(w http.ResponseWriter, req *http.Request) error {
		if req.Method != "POST" {
			return errors.New("only POST allowed")
		}
		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return fmt.Errorf("read error: %v", err)
		}

		var responseData []byte

		t0 := time.Now()
		id, err := requestLogger.logRequest(t0, req, data)
		if err != nil {
			return fmt.Errorf("database logging error (request): %v", err)
		}

		s, evaluationErr := evaluate(data)

		t1 := time.Now()
		if err := requestLogger.logResponse(t1, t1.Sub(t0), id, s, evaluationErr); err != nil {
			return fmt.Errorf("database logging error (result): %v", err)
		}

		if evaluationErr != nil {
			responseData, err = json.Marshal(struct {
				Input string `json:"input"`
				Ok    bool   `json:"ok"`
				Error string `json:"error"`
			}{
				Ok:    false,
				Input: strings.TrimSpace(string(data)),
				Error: evaluationErr.Error(),
			})
		} else {
			responseData, err = json.Marshal(struct {
				Ok     bool   `json:"ok"`
				Input  string `json:"input"`
				Result string `json:"result"`
			}{
				Ok:     true,
				Input:  strings.TrimSpace(string(data)),
				Result: s,
			})
		}

		if err != nil {
			return fmt.Errorf("JSON marshalling error: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(responseData)

		return nil
	}

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/api/eval", func(w http.ResponseWriter, req *http.Request) {
		metricRequests.WithLabelValues("api-eval").Inc()
		if err := handler(w, req); err != nil {
			log.Printf("error handling /api/eval: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("successfully handled /api/eval request")
	})
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		metricRequests.WithLabelValues("root").Inc()
		w.Header().Set("Content-Type", "text/html")
		w.Write(mainPageData)
	})

	log.Printf("listening on: http://%s", listener.Addr())

	return http.Serve(listener, nil)
}
