package lambdalist

import (
	"errors"
	"fmt"
	"strings"

	"github.com/steinarvk/heisenlisp/env"
	"github.com/steinarvk/heisenlisp/expr"
	"github.com/steinarvk/heisenlisp/types"
	"github.com/steinarvk/heisenlisp/value/symbol"
)

type namedValue struct {
	name uint32
	val  types.Value
}

type LambdaList struct {
	rawValue     types.Value
	requiredArgs []uint32
	optionalArgs []namedValue
	restArgName  uint32
}

func (l *LambdaList) minArgs() int {
	return len(l.requiredArgs)
}

func (l *LambdaList) maxArgs() (int, bool) {
	if l.restArgName != 0 {
		return 0, false
	}
	return len(l.requiredArgs) + len(l.optionalArgs), true
}

func (l *LambdaList) BindArgs(e types.Env, params []types.Value, pure bool) (types.Env, error) {
	if min := l.minArgs(); len(params) < min {
		return nil, fmt.Errorf("too few params (want %d got %d)", min, len(params))
	}

	if max, limited := l.maxArgs(); limited && len(params) > max {
		return nil, fmt.Errorf("too many params (want at most %d got %d)", max, len(params))
	}

	e = env.New(e)
	if pure {
		e.MarkPure()
	}

	for _, reqArgName := range l.requiredArgs {
		e.Bind(reqArgName, params[0])
		params = params[1:]
	}

	var err error

	for _, optArg := range l.optionalArgs {
		var val types.Value
		if len(params) == 0 {
			val, err = optArg.val.Eval(e)
			if err != nil {
				return nil, err
			}
		} else {
			val = params[0]
			params = params[1:]
		}
		e.Bind(optArg.name, val)
	}

	if l.restArgName != 0 {
		e.Bind(l.restArgName, expr.WrapList(params))
	}

	return e, nil
}

func Parse(val types.Value) (*LambdaList, error) {
	rv := &LambdaList{
		rawValue: val,
	}

	xs, err := expr.UnwrapList(val)
	if err != nil {
		return nil, err
	}

	addRequiredArgument := func(name string) {
		rv.requiredArgs = append(rv.requiredArgs, symbol.StringToIdOrPanic(name))
	}

	addOptionalArgument := func(name string, val types.Value) {
		rv.optionalArgs = append(rv.optionalArgs, namedValue{
			name: symbol.StringToIdOrPanic(name),
			val:  val,
		})
	}

	setRestArgument := func(name string) {
		rv.restArgName = symbol.StringToIdOrPanic(name)
	}

	inOptionalMode := false

	for i := 0; i < len(xs); i++ {
		name, err := symbol.Name(xs[i])
		if err != nil {
			if inOptionalMode {
				nameSym, defaultValue, err := expr.UnwrapProperListPair(xs[i])
				if err != nil {
					return nil, err
				}
				name, err = symbol.Name(nameSym)
				if err != nil {
					return nil, err
				}

				addOptionalArgument(name, defaultValue)

				continue
			}
			return nil, err
		}

		if strings.HasPrefix(name, "&") {
			switch name {
			case "&optional":
				inOptionalMode = true

			case "&rest":
				// must be the second-to-last, and next must be a symbol
				if len(xs) != (i + 2) {
					return nil, errors.New("&rest must be penultimate element in lambda list")
				}
				restName, err := symbol.Name(xs[i+1])
				if err != nil {
					return nil, fmt.Errorf("final element after &rest must be symbol (was %v): %v", xs[i+1], err)
				}
				setRestArgument(restName)

				i++ // skip over last

			default:
				return nil, fmt.Errorf("unknown & parameter: %v", xs[i])
			}

			continue
		}

		addRequiredArgument(name)
	}

	return rv, nil
}
