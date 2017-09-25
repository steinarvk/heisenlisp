package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/steinarvk/heisenlisp/builtin"
	"github.com/steinarvk/heisenlisp/code"
	"github.com/steinarvk/heisenlisp/env"
)

var (
	listenAddress = flag.String("listen_address", "127.0.0.1:6861", "http address on which to serve")
)

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

func main() {
	flag.Parse()

	listener, err := net.Listen("tcp", *listenAddress)
	if err != nil {
		log.Fatal(err)
	}

	root := builtin.NewRootEnv()

	var mu sync.Mutex

	evaluate := func(data []byte) (string, error) {
		mu.Lock()
		defer mu.Unlock()

		val, err := code.Run(env.New(root), "<request data>", data)
		if err != nil {
			return "", err
		}

		return val.String(), nil
	}

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/api/eval", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			http.Error(w, "only POST allowed", http.StatusMethodNotAllowed)
			return
		}
		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(w, "read error", http.StatusBadRequest)
			return
		}

		var responseData []byte

		s, err := evaluate(data)
		if err != nil {
			responseData, err = json.Marshal(struct {
				Input string `json:"input"`
				Ok    bool   `json:"ok"`
				Error string `json:"error"`
			}{
				Ok:    false,
				Input: strings.TrimSpace(string(data)),
				Error: err.Error(),
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
			http.Error(w, "JSON marshalling error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(responseData)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write(mainPageData)
	})

	log.Printf("listening on: http://%s", listener.Addr())

	log.Fatal(http.Serve(listener, nil))
}
