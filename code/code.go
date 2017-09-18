package code

import (
	"fmt"
	"io/ioutil"

	"github.com/steinarvk/heisenlisp/gen/parser"
	"github.com/steinarvk/heisenlisp/types"
)

func Run(env types.Env, name string, code []byte) (types.Value, error) {
	expressionsIntf, err := parser.Parse(name, code)
	if err != nil {
		return nil, err
	}

	var lastResult types.Value

	expressions := expressionsIntf.([]interface{})

	for _, expression := range expressions {
		val := expression.(types.Value)

		lastResult, err = val.Eval(env)
		if err != nil {
			return nil, fmt.Errorf("error evaluating %v: %v", val, err)
		}
	}

	return lastResult, nil
}

func RunFile(env types.Env, filename string) (types.Value, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return Run(env, filename, data)
}
