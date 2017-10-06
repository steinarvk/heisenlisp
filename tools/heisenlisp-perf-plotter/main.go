package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/steinarvk/heisenlisp/builtin"
	"github.com/steinarvk/heisenlisp/code"
	"github.com/steinarvk/heisenlisp/number"
	"github.com/steinarvk/heisenlisp/types"
	"github.com/steinarvk/heisenlisp/value/unknowns/anyof"
)

var (
	trialTimeLimit    = flag.Duration("trial_time_limit", 5*time.Second, "max time per trial")
	totalTimeLimit    = flag.Duration("total_time_limit", time.Minute, "max time for total measurements")
	exponentialFactor = flag.Float64("exponential_factor", 1.1, "exponential factor when increasing n")
	script            = flag.String("script", "", "script file to load (optional)")
	expression        = flag.String("expression", "timing-testcase", "expression (function dependent on one integer parameter")
	disableAnyofLimit = flag.Bool("disable_anyof_limit", true, "disable the limitation on number of values in an anyof")
)

type trialResult struct {
	n        int64
	duration time.Duration
}

func performTrial(e types.Env, callable types.Callable, n int64) (*trialResult, error) {
	args := []types.Value{number.FromInt64(n)}
	t0 := time.Now()
	_, err := callable.Call(args)
	if err != nil {
		return nil, err
	}
	t1 := time.Now()
	return &trialResult{
		n:        n,
		duration: t1.Sub(t0),
	}, nil
}

func performTrials(e types.Env, callable types.Callable, outCh chan<- *trialResult) error {
	n := int64(1)
	t0 := time.Now()
	for {
		t1 := time.Now()
		result, err := performTrial(e, callable, n)
		if err != nil {
			return err
		}
		t2 := time.Now()

		outCh <- result

		if t2.Sub(t1) > *trialTimeLimit {
			return nil
		}
		if t2.Sub(t0) > *totalTimeLimit {
			return nil
		}

		lastN := n
		n = int64(float64(n) * *exponentialFactor)
		if n == lastN {
			n++
		}
	}
}

func printTrials(results <-chan *trialResult) {
	for result := range results {
		fmt.Println(result.n, int64(result.duration))
	}
}

func mainCore() error {
	if *disableAnyofLimit {
		log.Printf("disabling anyof limit (was %v)", anyof.MaxAnyOfElements)
		anyof.MaxAnyOfElements = math.MaxInt64
	}

	e := builtin.NewRootEnv()
	if *script != "" {
		log.Printf("running script %q", *script)
		_, err := code.RunFile(e, *script)
		if err != nil {
			return err
		}
	}
	value, err := code.Run(e, "<expression>", []byte(*expression))
	if err != nil {
		return err
	}
	callable, ok := value.(types.Callable)
	if !ok {
		return fmt.Errorf("value of %q not callable: %v", *expression, value)
	}

	ch := make(chan *trialResult)
	go printTrials(ch)
	if err := performTrials(e, callable, ch); err != nil {
		return err
	}
	close(ch)

	return nil
}

func main() {
	flag.Parse()

	if err := mainCore(); err != nil {
		log.Fatal(err)
	}
}
