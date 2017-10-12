package cmd

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/spf13/cobra"
	"github.com/steinarvk/heisenlisp/builtin"
	"github.com/steinarvk/heisenlisp/code"
	"github.com/steinarvk/heisenlisp/number"
	"github.com/steinarvk/heisenlisp/types"
	"github.com/steinarvk/heisenlisp/value/unknowns/anyof"
)

var (
	ptTrialTimeLimit    *time.Duration
	ptTotalTimeLimit    *time.Duration
	ptExponentialFactor *float64
	ptScript            *string
	ptExpression        *string
	ptTrialIterations   *int
	ptDisableAnyofLimit *bool
)

func init() {
	ptTrialTimeLimit = perfTestCmd.Flags().Duration("trial_time_limit", 5*time.Second, "max time per trial")
	ptTotalTimeLimit = perfTestCmd.Flags().Duration("total_time_limit", time.Minute, "max time for total measurements")
	ptExponentialFactor = perfTestCmd.Flags().Float64("exponential_factor", 1.1, "exponential factor when increasing n")
	ptScript = perfTestCmd.Flags().String("script", "", "script file to load (optional)")
	ptExpression = perfTestCmd.Flags().String("expression", "timing-testcase", "expression (function dependent on one integer parameter")
	ptTrialIterations = perfTestCmd.Flags().Int("iterations", 1, "iterations per trial (mean will be taken)")
	ptDisableAnyofLimit = perfTestCmd.Flags().Bool("disable_anyof_limit", true, "disable the limitation on number of values in an anyof")
}

var perfTestCmd = &cobra.Command{
	Use:   "perf-test",
	Short: "Executes a function with varying n to determine its preformance ",
	Long: `perf-test runs an expression multiple times and records the execution times,
attempting to ascertain the performance characteristics (i.e. complexity)
of the given unary function.

It outputs numbers in two columns: first n (the argument passed to the
function), and then the time taken for one execution in nanoseconds.
(Scripts written to consume this format should tolerate additional columns.)

For instance, if you run perf-test on the expression 
    (lambda (n) (sorted (range n)))
You would expect to get output showing the performance of the "sorted"
function, ideally with y ~= x log x.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runPerfTest(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(perfTestCmd)
}

type trialResult struct {
	n        int64
	duration time.Duration
}

func performTrial(e types.Env, callable types.Callable, n int64) (*trialResult, error) {
	args := []types.Value{number.FromInt64(n)}
	t0 := time.Now()
	for i := 0; i < *ptTrialIterations; i++ {
		_, err := callable.Call(args)
		if err != nil {
			return nil, err
		}
	}
	t1 := time.Now()
	return &trialResult{
		n:        n,
		duration: time.Duration(float64(t1.Sub(t0)) / float64(*ptTrialIterations)),
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

		if t2.Sub(t1) > *ptTrialTimeLimit {
			return nil
		}
		if t2.Sub(t0) > *ptTotalTimeLimit {
			return nil
		}

		lastN := n
		n = int64(float64(n) * *ptExponentialFactor)
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

func runPerfTest() error {
	if *ptDisableAnyofLimit {
		log.Printf("disabling anyof limit (was %v)", anyof.MaxAnyOfElements)
		anyof.MaxAnyOfElements = math.MaxInt64
	}

	e := builtin.NewRootEnv()
	if *ptScript != "" {
		log.Printf("running script %q", *ptScript)
		_, err := code.RunFile(e, *ptScript)
		if err != nil {
			return err
		}
	}
	value, err := code.Run(e, "<expression>", []byte(*ptExpression))
	if err != nil {
		return err
	}
	callable, ok := value.(types.Callable)
	if !ok {
		return fmt.Errorf("value of %q not callable: %v", *ptExpression, value)
	}

	ch := make(chan *trialResult)
	go printTrials(ch)
	if err := performTrials(e, callable, ch); err != nil {
		return err
	}
	close(ch)

	return nil
}
