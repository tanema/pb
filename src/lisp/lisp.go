package lisp

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type (
	symbol string
)

var (
	numberPattern = regexp.MustCompile(`^-?[0-9]+\.?[0-9]*$|^-?\.[0-9]+$`)
	tokensPattern = regexp.MustCompile(
		`[\s,]*(~@|[\[\]{}()'` + "`" + `~^@!$#]|"(?:\\.|[^\\"])*"?|;.*|[^\s\[\]{}('"` + "`" + `,;)]*)`,
	)
	eof                = symbol("end of form")
	ErrorUnderflow     = errors.New("underflow, expected end of list was not found")
	ErrMalformedNumber = errors.New("malformed number")
	stdenv             = map[string]any{
		"exit":  exit,
		"+":     add,
		"-":     sub,
		"*":     mul,
		"/":     div,
		"str":   str,
		"print": prin,
		"defun": defun,
		"list":  list,
		"first": first,
		"rest":  rest,
		"nth":   nth,
	}
)

// Eval will interpret a string and return the value
func Eval(src string) (any, error) {
	tokenStream := tokenize(src)
	var final any
	for {
		if res, err := eval(stdenv, read(tokenStream)); err == io.EOF {
			return final, nil
		} else if err != nil {
			return nil, err
		} else {
			final = res
		}
	}
}

func tokenize(src string) chan string {
	matches := tokensPattern.FindAllStringSubmatch(src, -1)
	tokens := make(chan string, len(matches))
	for _, group := range matches {
		tokens <- group[1]
	}
	close(tokens)
	return tokens
}

func read(tokens chan string) any {
	if token, hasNext := <-tokens; !hasNext {
		return io.EOF
	} else if token == ")" {
		return eof
	} else if token == "(" {
		forms := []any{}
		form := read(tokens)
		for ; form != eof && form != io.EOF; form = read(tokens) {
			forms = append(forms, form)
		}
		if form == io.EOF {
			return ErrorUnderflow
		}
		return forms
	} else if token == "nil" {
		return nil
	} else if token == "true" || token == "false" {
		return token == "true"
	} else if match := numberPattern.MatchString(token); match {
		n, err := strconv.ParseFloat(token, 64)
		if err != nil {
			return ErrMalformedNumber
		}
		return n
	} else if (strings.HasPrefix(token, `"`) && strings.HasSuffix(token, `"`)) || (strings.HasPrefix(token, `'`) && strings.HasSuffix(token, `'`)) {
		return strings.Trim(token, `"'`)
	} else {
		return symbol(token)
	}
}

func eval(env map[string]any, object any) (any, error) {
	switch tobj := object.(type) {
	case error:
		return nil, tobj
	case []any:
		if len(tobj) == 0 {
			return nil, nil
		} else if act, err := eval(env, tobj[0]); err != nil {
			return nil, err
		} else if fn, ok := act.(func(env map[string]any, args []any) (any, error)); !ok {
			return nil, fmt.Errorf("%v is not callable", act)
		} else {
			return fn(env, tobj[1:])
		}
	case symbol:
		val, ok := env[string(tobj)]
		if !ok {
			return nil, fmt.Errorf("undefined symbol %v", tobj)
		}
		return val, nil
	default:
		return object, nil
	}
}

func evalAST(env map[string]any, ast []any) ([]any, error) {
	return mapn[any, any](func(i any) (any, error) { return eval(env, i) }, ast)
}

func exit(map[string]any, []any) (any, error) {
	os.Exit(0)
	return nil, nil
}

func mapn[T any, V any](fn func(V any) (T, error), data []any) ([]T, error) {
	var err error
	result := make([]T, len(data))
	for i, val := range data {
		result[i], err = fn(val)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func reduce[T any](start T, fn func(res, i T) T, data []T) T {
	result := start
	for _, val := range data {
		result = fn(result, val)
	}
	return result
}

func toFloat(val any) (float64, error) {
	n, ok := val.(float64)
	if !ok {
		return -1, fmt.Errorf("arithmetic performed on non-numeric value %v", val)
	}
	return n, nil
}

func toString(val any) (string, error) {
	return fmt.Sprintf("%v", val), nil
}

func arithmetic(env map[string]any, args []any, fn func(r, i float64) float64) (any, error) {
	if len(args) == 0 {
		return 0, nil
	} else if forms, err := evalAST(env, args); err != nil {
		return nil, err
	} else if nums, err := mapn[float64, any](toFloat, forms); err != nil {
		return nil, err
	} else {
		return reduce(nums[0], fn, nums[1:]), nil
	}
}

func add(env map[string]any, args []any) (any, error) {
	return arithmetic(env, args, func(r, i float64) float64 { return r + i })
}

func sub(env map[string]any, args []any) (any, error) {
	return arithmetic(env, args, func(r, i float64) float64 { return r - i })
}

func mul(env map[string]any, args []any) (any, error) {
	return arithmetic(env, args, func(r, i float64) float64 { return r * i })
}

func div(env map[string]any, args []any) (any, error) {
	return arithmetic(env, args, func(r, i float64) float64 { return r / i })
}

func str(env map[string]any, args []any) (any, error) {
	forms, err := evalAST(env, args)
	if err != nil {
		return nil, err
	}
	strs, _ := mapn[string, any](toString, forms)
	return reduce("", func(r, i string) string { return r + i }, strs), nil
}

func prin(env map[string]any, args []any) (any, error) {
	str, err := str(env, args)
	if err == nil {
		fmt.Println(str)
	}
	return nil, err
}

func defun(env map[string]any, args []any) (any, error) {
	if sym, ok := args[0].(symbol); !ok {
		return nil, fmt.Errorf("non-symbol bind value %v", args[0])
	} else if paramDefs, ok := args[1].([]any); !ok {
		return nil, fmt.Errorf("improperly formatted func, expected params, found %v", args[1])
	} else if params, err := mapn[string, any](func(val any) (string, error) {
		sym, ok := val.(symbol)
		if !ok {
			return "", fmt.Errorf("non-symbol function parameter %v", val)
		}
		return string(sym), nil
	}, paramDefs); err != nil {
		return nil, err
	} else {
		env[string(sym)] = callfn(string(sym), params, args[2])
		return nil, nil
	}
}

func callfn(name string, params []string, body any) func(map[string]any, []any) (any, error) {
	return func(env map[string]any, args []any) (any, error) {
		if len(params) > len(args) {
			return nil, fmt.Errorf("not enough arguments provided to fn %v", name)
		}
		paramVals, err := evalAST(env, args)
		if err != nil {
			return nil, err
		}
		env = childEnv(env)
		for i, param := range params {
			env[param] = paramVals[i]
		}
		return eval(env, body)
	}
}

func list(env map[string]any, args []any) (any, error) {
	return args, nil
}

func first(env map[string]any, args []any) (any, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("not enough params passed to first")
	} else if param, err := eval(env, args[0]); err != nil {
		return nil, err
	} else if lst, ok := param.([]any); !ok {
		return nil, fmt.Errorf("cannot perform list actions on non list %v", args[0])
	} else if len(lst) == 0 {
		return nil, nil
	} else {
		return lst[0], nil
	}
}

func rest(env map[string]any, args []any) (any, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("not enough params passed to rest")
	} else if param, err := eval(env, args[0]); err != nil {
		return nil, err
	} else if lst, ok := param.([]any); !ok {
		return nil, fmt.Errorf("cannot perform list actions on non list %v", args[0])
	} else {
		return lst[1:], nil
	}
}

func nth(env map[string]any, args []any) (any, error) {
	if len(args) <= 0 {
		return nil, fmt.Errorf("not enough params passed to rest")
	} else if index, err := eval(env, args[0]); err != nil {
		return nil, err
	} else if i, ok := index.(float64); !ok {
		return nil, fmt.Errorf("cannot index with non-number %v", args[0])
	} else if param, err := eval(env, args[1]); err != nil {
		return nil, err
	} else if lst, ok := param.([]any); !ok {
		return nil, fmt.Errorf("cannot perform list actions on non list %v", args[1])
	} else if i < 0 || int(i) > len(lst)-1 {
		return nil, nil
	} else {
		return lst[int(i)], nil
	}
}

func childEnv(env map[string]any) map[string]any {
	newEnv := map[string]any{}
	for k, v := range env {
		newEnv[k] = v
	}
	return newEnv
}
