package function

import (
	"errors"
	"fmt"
	"strings"

	"github.com/steinarvk/heisenlisp/env"
	"github.com/steinarvk/heisenlisp/expr"
	"github.com/steinarvk/heisenlisp/types"
	"github.com/steinarvk/heisenlisp/value/symbol"
)

func NameIsPure(s string) bool {
	return !strings.HasSuffix(s, "!")
}

type FunctionValue struct {
	name       string
	lexicalEnv types.Env
	lambdaList *lambdaList
	body       []types.Value
	pure       bool
}

type MacroValue struct {
	name       string
	lexicalEnv types.Env
	lambdaList *lambdaList
	body       []types.Value
	pure       bool
}

type namedValue struct {
	name string
	val  types.Value
}

type lambdaList struct {
	rawValue     types.Value
	requiredArgs []string
	optionalArgs []namedValue
	restArgName  string
}

func (l *lambdaList) minArgs() int {
	return len(l.requiredArgs)
}

func (l *lambdaList) maxArgs() (int, bool) {
	if l.restArgName != "" {
		return 0, false
	}
	return len(l.requiredArgs) + len(l.optionalArgs), true
}

func (l *lambdaList) bindArgs(e types.Env, params []types.Value, pure bool) (types.Env, error) {
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

	if l.restArgName != "" {
		e.Bind(l.restArgName, expr.WrapList(params))
	}

	return e, nil
}

func parseLambdaList(val types.Value) (*lambdaList, error) {
	rv := &lambdaList{
		rawValue: val,
	}

	xs, err := expr.UnwrapList(val)
	if err != nil {
		return nil, err
	}

	addRequiredArgument := func(name string) {
		rv.requiredArgs = append(rv.requiredArgs, name)
	}

	addOptionalArgument := func(name string, val types.Value) {
		rv.optionalArgs = append(rv.optionalArgs, namedValue{
			name: name,
			val:  val,
		})
	}

	setRestArgument := func(name string) {
		rv.restArgName = name
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

func New(env types.Env, name string, formalParams types.Value, body []types.Value) (*FunctionValue, error) {
	ll, err := parseLambdaList(formalParams)
	if err != nil {
		return nil, fmt.Errorf("invalid lambda list: %v", err)
	}
	rv := &FunctionValue{
		name:       name,
		lexicalEnv: env,
		lambdaList: ll,
		body:       body,
	}
	if NameIsPure(name) {
		rv.pure = true
	}
	return rv, nil
}

func NewMacro(env types.Env, name string, formalParams types.Value, body []types.Value) (*MacroValue, error) {
	ll, err := parseLambdaList(formalParams)
	if err != nil {
		return nil, fmt.Errorf("invalid lambda list: %v", err)
	}
	rv := &MacroValue{
		name:       name,
		lexicalEnv: env,
		lambdaList: ll,
		body:       body,
	}
	if NameIsPure(name) {
		rv.pure = true
	}
	return rv, nil
}

func (_ *FunctionValue) TypeName() string { return "function" }
func (f *FunctionValue) errorprefix() string {
	if f.name == "" {
		return "(anonymous function): "
	}
	return fmt.Sprintf("%s: ", f.name)
}

func (f *FunctionValue) Call(params []types.Value) (types.Value, error) {
	var rv types.Value
	var err error

	env, err := f.lambdaList.bindArgs(f.lexicalEnv, params, f.pure)
	if err != nil {
		return nil, fmt.Errorf("%s%v", f.errorprefix(), err)
	}

	for _, stmt := range f.body {
		rv, err = stmt.Eval(env)
		if err != nil {
			return nil, err
		}
	}
	return rv, nil
}

func (f *FunctionValue) String() string {
	if f.name == "" {
		return "#<anonymous function>"
	}
	return fmt.Sprintf("#<function %q>", f.name)
}
func (f *FunctionValue) Eval(_ types.Env) (types.Value, error) { return f, nil }

func (f *FunctionValue) IsPure() bool {
	return f.pure
}

func (f *FunctionValue) Falsey() bool { return false }

func (_ *MacroValue) TypeName() string { return "macro" }
func (f *MacroValue) errorprefix() string {
	if f.name == "" {
		return "(anonymous macro): "
	}
	return fmt.Sprintf("%s: ", f.name)
}

func (f *MacroValue) Expand(params []types.Value) (types.Value, error) {
	var rv types.Value
	var err error

	env, err := f.lambdaList.bindArgs(f.lexicalEnv, params, f.pure)
	if err != nil {
		return nil, fmt.Errorf("%s%v", f.errorprefix(), err)
	}

	for _, stmt := range f.body {
		rv, err = stmt.Eval(env)
		if err != nil {
			return nil, err
		}
	}
	return rv, nil
}

func (f *MacroValue) IsPure() bool {
	return f.pure
}

func (f *MacroValue) String() string {
	if f.name == "" {
		return "#<anonymous macro>"
	}
	return fmt.Sprintf("#<macro %q>", f.name)
}
func (f *MacroValue) Eval(_ types.Env) (types.Value, error) { return f, nil }

func (f *MacroValue) Falsey() bool { return false }
