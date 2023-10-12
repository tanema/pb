package lisp

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
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
		"nil":    nil,
		"true":   true,
		"false":  false,
		"env":    env,
		"doc":    doc,
		"exit":   exit,
		"+":      add,
		"-":      sub,
		"*":      mul,
		"/":      div,
		"str":    str,
		"print":  prin,
		"defun":  defun,
		"list":   list,
		"first":  first,
		"rest":   rest,
		"nth":    nth,
		"length": length,
		"empty?": empty,
		"let":    let,
		"if":     ifelse,
		">":      gt,
		">=":     gte,
		"<":      lt,
		"<=":     lte,
		"not":    not,
		"eq":     eq,
		"and":    and,
		"or":     or,
	}
)

func NewEnv(binds map[string]any) map[string]any {
	return childEnv(stdenv, binds)
}

// Eval will interpret a string and return the value
func EvalSrc(env map[string]any, src string) (any, error) {
	tokenStream := tokenize(src)
	var final any
	for {
		if res, err := EvalForm(env, read(tokenStream)); err == io.EOF {
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
		if group[1] == "" {
			continue
		}
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

func EvalForm(env map[string]any, object any) (any, error) {
	switch tobj := object.(type) {
	case error:
		return nil, tobj
	case []any:
		if len(tobj) == 0 {
			return nil, nil
		} else if act, err := EvalForm(env, tobj[0]); err != nil {
			return nil, err
		} else if fn, ok := act.(func(env map[string]any, args []any) (any, error)); !ok {
			return nil, fmt.Errorf("'%v' is not callable", act)
		} else {
			return fn(env, tobj[1:])
		}
	case symbol:
		if val, ok := env[string(tobj)]; ok {
			return val, nil
		}
		return nil, fmt.Errorf("undefined symbol '%v'", tobj)
	default:
		return object, nil
	}
}

func EvalAST(env map[string]any, ast []any) ([]any, error) {
	return mapn[any, any](func(i any) (any, error) { return EvalForm(env, i) }, ast)
}

func childEnv(env map[string]any, binds map[string]any) map[string]any {
	newEnv := map[string]any{}
	for k, v := range env {
		newEnv[k] = v
	}
	for k, v := range binds {
		newEnv[k] = v
	}
	return newEnv
}

func exit(env map[string]any, args []any) (any, error) {
	if IsDocCall(env, args) {
		return `exit will end the execution of the program

Usage: (exit [exitCode])`, nil
	}
	exitCode := 0
	if len(args) > 0 {
		if code, ok := args[0].(float64); ok {
			exitCode = int(code)
		}
	}
	os.Exit(exitCode)
	return nil, nil
}

func env(env map[string]any, args []any) (any, error) {
	if IsDocCall(env, args) {
		return `env will output all of the defined symbols in the current environment

Usage: (env)`, nil
	}
	output := []string{}
	for k := range env {
		output = append(output, k)
	}
	sort.Strings(output)
	fmt.Println(output)
	return nil, nil
}

func IsDocCall(env map[string]any, args []any) bool {
	return env == nil && args == nil
}

func doc(env map[string]any, args []any) (any, error) {
	if IsDocCall(env, args) {
		return `doc will print out documentation for a defined symbole if it exists

Usage:   (doc funcName)
Example: (doc defun)`, nil
	} else if len(args) == 0 {
		return 0, errors.New("no symbol provided to doc")
	} else if val, err := EvalForm(env, args[0]); err != nil {
		return nil, err
	} else if fn, ok := val.(func(env map[string]any, args []any) (any, error)); !ok {
		return nil, fmt.Errorf("cannot provide documentation for non callable %v", args[0])
	} else if docVal, err := fn(nil, nil); err != nil {
		return nil, fmt.Errorf("no documentation for %v defined", args[0])
	} else if doc, ok := docVal.(string); !ok {
		return nil, fmt.Errorf("no documentation for %v defined", args[0])
	} else {
		return doc, nil
	}
}

func mapn[T any, V any](fn func(V any) (T, error), data []V) ([]T, error) {
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

func reduce[T any, V any](start V, fn func(res V, i T) V, data []T) V {
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

func toBool(val any) bool {
	switch tVal := val.(type) {
	case string:
		return tVal != ""
	case bool:
		return tVal
	case float64:
		return tVal != 0
	default:
		return false
	}
}

func toBools(vals []any) []bool {
	bools, _ := mapn[bool, any](func(i any) (bool, error) { return toBool(i), nil }, vals)
	return bools
}

func arithmetic(env map[string]any, args []any, fn func(r, i float64) float64) (any, error) {
	if len(args) == 0 {
		return 0, nil
	} else if forms, err := EvalAST(env, args); err != nil {
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
	if IsDocCall(env, args) {
		return `str will convert and combine two or more values and return the resulting string.
Any value passed that is not a string will be converted to string.

Usage:   (str n0 [n1 n2 ...])
Example: (str "hello" " " "world")
         => "hello world"`, nil
	}
	forms, err := EvalAST(env, args)
	if err != nil {
		return nil, err
	}
	strs, _ := mapn[string, any](toString, forms)
	return reduce("", func(r, i string) string { return r + i }, strs), nil
}

func prin(env map[string]any, args []any) (any, error) {
	if IsDocCall(env, args) {
		return `print will convert and combine the arguments provided and output the result
to stdout. Any value passed that is not a string will be converted to string.

Usage:   (print n0 [n1 n2 ...])
Example: (print "hello" " " "world")`, nil
	}
	str, err := str(env, args)
	if err == nil {
		fmt.Println(str)
	}
	return nil, err
}

func defun(env map[string]any, args []any) (any, error) {
	if IsDocCall(env, args) {
		return `defun will define a callable func in the current environment.

Usage:   (defun fnName (param1 param2 ...) (body))
Example:

  ;; define fibonacci function
  (defun fibonacci (n)
    (if (<= n 1)
      1
      (+ (fibonacci (- n 1)) (fibonacci (- n 2)))))

  ;; call fibonacci function
  (fibonacci 5)

`, nil
	} else if sym, ok := args[0].(symbol); !ok {
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
		paramVals, err := EvalAST(env, args)
		if err != nil {
			return nil, err
		}
		binds := map[string]any{}
		for i, param := range params {
			binds[param] = paramVals[i]
		}
		return EvalForm(childEnv(env, binds), body)
	}
}

func list(env map[string]any, args []any) (any, error) {
	if IsDocCall(env, args) {
		return `list will create a data list from the provided data

Usage:   (list n0 [n1 n2 ...])
Example: (list 1 22 "hello" "world" false)
         => (1 22 "hello" "world" false)`, nil
	}
	return args, nil
}

func first(env map[string]any, args []any) (any, error) {
	if IsDocCall(env, args) {
		return `first will return the first item of a list.

Usage:   (first list)
Example: (first (list 1 2 3))
         => 1`, nil
	} else if len(args) == 0 {
		return nil, fmt.Errorf("not enough params passed to first")
	} else if param, err := EvalForm(env, args[0]); err != nil {
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
	if IsDocCall(env, args) {
		return `rest will return all of the list provided without the first element.

Usage:   (rest list)
Example: (rest (list 1 2 3))
         => (2 3)`, nil
	} else if len(args) == 0 {
		return nil, fmt.Errorf("not enough params passed to rest")
	} else if param, err := EvalForm(env, args[0]); err != nil {
		return nil, err
	} else if lst, ok := param.([]any); !ok {
		return nil, fmt.Errorf("cannot perform list actions on non list %v", args[0])
	} else {
		return lst[1:], nil
	}
}

func nth(env map[string]any, args []any) (any, error) {
	if IsDocCall(env, args) {
		return `nth will return the element at the index provided. If the index is
negative or beyond the length of the list, nil will be returned.

Usage:   (nth n list)
Example: (nth 1 (list 1 2 3))
         => 2`, nil
	} else if len(args) <= 0 {
		return nil, fmt.Errorf("not enough params passed to rest")
	} else if index, err := EvalForm(env, args[0]); err != nil {
		return nil, err
	} else if i, ok := index.(float64); !ok {
		return nil, fmt.Errorf("cannot index with non-number %v", args[0])
	} else if param, err := EvalForm(env, args[1]); err != nil {
		return nil, err
	} else if lst, ok := param.([]any); !ok {
		return nil, fmt.Errorf("cannot perform list actions on non list %v", args[1])
	} else if i < 0 || int(i) > len(lst)-1 {
		return nil, nil
	} else {
		return lst[int(i)], nil
	}
}

func length(env map[string]any, args []any) (any, error) {
	if IsDocCall(env, args) {
		return `length will count the items in a countable object and return as a number.

Usage:   (length list)
Example: (length (list 1 2 3))
         => 3
         (length "my string")
         => 9`, nil
	} else if len(args) != 1 {
		return nil, fmt.Errorf("not enough params passed to length")
	}

	val, err := EvalForm(env, args[0])
	if err != nil {
		return nil, err
	}

	switch tObj := val.(type) {
	case string:
		return float64(len(tObj)), nil
	case []any:
		return float64(len(tObj)), nil
	default:
		return nil, fmt.Errorf("cannot check length on non countable %v", args[0])
	}
}

func empty(env map[string]any, args []any) (any, error) {
	if IsDocCall(env, args) {
		return `empty? will check if a countables length is zero and return true if so.

Usage:   (empty? countable)
Example: (empty? (list 1 2 3))
         => false
         (empty? "")
         => true`, nil
	} else if len(args) != 1 {
		return nil, fmt.Errorf("not enough params passed to empty?")
	} else if val, err := length(env, args); err != nil {
		return nil, err
	} else {
		return val.(float64) == 0, nil
	}
}

func let(env map[string]any, args []any) (any, error) {
	if IsDocCall(env, args) {
		return `let will define variables in a scope to be used. let will return the
value of the final form evaluation.

Usage:   (let defns evalBody)
Example: (let ((x 22) (y 42)) (+ x y))`, nil
	} else if len(args) != 2 {
		return nil, errors.New("not enough params passed to let")
	}

	varSettings, ok := args[0].([]any)
	if !ok {
		return nil, errors.New("malformed vars in let declaration")

	}

	binds := map[string]any{}
	for _, setting := range varSettings {
		if kv, ok := setting.([]any); !ok || len(kv) != 2 {
			return nil, errors.New("malformed vars in let declaration")
		} else if sym, ok := kv[0].(symbol); !ok {
			return nil, errors.New("cannot bind to non-symbol in let declaration")
		} else if val, err := EvalForm(env, kv[1]); err != nil {
			return nil, err
		} else {
			binds[string(sym)] = val
		}
	}

	return EvalForm(childEnv(env, binds), args[1])
}

func ifelse(env map[string]any, args []any) (any, error) {
	if IsDocCall(env, args) {
		return `if is boolean control flow. When

Usage:   (if (booleanForm) (ifTrueBody) [(elseBody)])
Example: (if (= x "yes") (print "x is yes") (print "x is not yes"))`, nil
	} else if len(args) < 2 {
		return nil, errors.New("not enough params passed to if")
	} else if val, err := EvalForm(env, args[0]); err != nil {
		return nil, err
	} else if !toBool(val) && len(args) > 2 {
		return EvalForm(env, args[2])
	} else {
		return EvalForm(env, args[1])
	}
}

func cmpr(env map[string]any, args []any, fn func(a, b float64) bool) (any, error) {
	if len(args) != 2 {
		return nil, errors.New("not enough params passed to comparison")
	} else if vals, err := EvalAST(env, args); err != nil {
		return nil, err
	} else if a, err := toFloat(vals[0]); err != nil {
		return nil, err
	} else if b, err := toFloat(vals[1]); err != nil {
		return nil, err
	} else {
		return fn(a, b), nil
	}
}

func gt(env map[string]any, args []any) (any, error) {
	return cmpr(env, args, func(a, b float64) bool { return a > b })
}

func gte(env map[string]any, args []any) (any, error) {
	return cmpr(env, args, func(a, b float64) bool { return a >= b })
}

func lt(env map[string]any, args []any) (any, error) {
	return cmpr(env, args, func(a, b float64) bool { return a < b })
}

func lte(env map[string]any, args []any) (any, error) {
	return cmpr(env, args, func(a, b float64) bool { return a <= b })
}

func allEq[T comparable](all []any) bool {
	first, ok := all[0].(T)
	if !ok {
		return false
	}
	for _, val := range all[1:] {
		if tVal, ok := val.(T); !ok {
			return false
		} else if first != tVal {
			return false
		}
	}
	return true
}

func eq(env map[string]any, args []any) (any, error) {
	if IsDocCall(env, args) {
		return `eq will compare two or more data points and return true if they are eq, false otherwise.

Usage:   (eq n1 n2 [n3 n4 ...])
Example: (eq "yes" "yes" "no")
         =>
         false`, nil
	} else if len(args) < 2 {
		return nil, errors.New("not enough params passed to eq")
	}

	vals, err := EvalAST(env, args)
	if err != nil {
		return nil, err
	}

	switch tVal := vals[0].(type) {
	case string:
		return allEq[string](vals), nil
	case float64:
		return allEq[float64](vals), nil
	case bool:
		return allEq[bool](vals), nil
	default:
		return nil, fmt.Errorf("unable to compare %v", tVal)
	}
}

func not(env map[string]any, args []any) (any, error) {
	if IsDocCall(env, args) {
		return `not will the opposite boolean value of what ever value it is provided.

Usage:   (not n1)
Example: (not true)
         =>
         false`, nil
	} else if len(args) != 1 {
		return nil, errors.New("not enough params passed to not")
	} else if val, err := EvalForm(env, args[0]); err != nil {
		return nil, err
	} else {
		return !toBool(val), nil
	}
}

func and(env map[string]any, args []any) (any, error) {
	if IsDocCall(env, args) {
		return `and will true if all of the values passed to it, evaluate to truthy values.
it will return false otherwise.

Usage:   (and n1 n2 [n3 n4 ...])
Example: (and true "truthy string" 42)
         =>
         true`, nil
	} else if len(args) < 2 {
		return false, nil
	} else if vals, err := EvalAST(env, args); err != nil {
		return nil, err
	} else {
		return reduce(true, func(res, i bool) bool { return res && i }, toBools(vals)), nil
	}
}

func or(env map[string]any, args []any) (any, error) {
	if IsDocCall(env, args) {
		return `or will true if any of the values passed to it, evaluate to truthy values.
it will return false otherwise.

Usage:   (or n1 n2 [n3 n4 ...])
Example: (or false "" 0 true)
         =>
         true`, nil
	} else if len(args) < 2 {
		return false, nil
	} else if vals, err := EvalAST(env, args); err != nil {
		return nil, err
	} else {
		return reduce(false, func(res, i bool) bool { return res || i }, toBools(vals)), nil
	}
}
