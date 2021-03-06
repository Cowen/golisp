package main

import (
  "bufio"
  "errors"
  "fmt"
  "io"
  "os"
  "strconv"
  "strings"
)

type Value interface{}

type Env struct {
  symbols  map[string](Value)
  outerEnv *Env
}

func (env Env) find(v string) Env {
  if _, ok := env.symbols[v]; ok {
    return env
  }

  if env.outerEnv != nil {
    return env.outerEnv.find(v)
  }

  return Env{}
}

func add_globals(env *Env) {
  if env.symbols == nil {
    env.symbols = make(map[string](Value))
  }

  env.symbols["+"] = func(args []Value) Value {
    sum := args[0].(int)
    for _, num := range args[1:] {
      sum += num.(int)
    }
    return sum
  }

  env.symbols["-"] = func(args []Value) Value {
    diff := args[0].(int)
    for _, num := range args[1:] {
      diff -= num.(int)
    }
    return diff
  }

  env.symbols["*"] = func(args []Value) Value {
    prod := args[0].(int)
    for _, num := range args[1:] {
      prod = prod * num.(int)
    }
    return prod
  }

  env.symbols["/"] = func(args []Value) Value {
    quot := args[0].(int)
    for _, num := range args[1:] {
      quot /= num.(int)
    }
    return quot
  }

  env.symbols["define"] = func(args []Value) Value {
    symbol, val := args[0].(string), args[1]
    env.symbols[symbol] = val
    return 0
  }

}

func Tokenize(s string) []string {
  // Convert a string into a list of tokens
  lParenStr := strings.Replace(s, "(", " ( ", -1)
  rParenStr := strings.Replace(lParenStr, ")", " ) ", -1)
  return strings.Fields(rParenStr)
}

func atomize(token string, env *Env) (Value, error) {
  if v, ok := env.symbols[token]; ok {
    return v, nil
  }

  if num, err := strconv.Atoi(token); err == nil {
    return num, err
  }

  return token, nil
}

func findNextParen(args []string) (int, error) {
  for i, arg := range args {
    if arg == ")" {
      return i, nil
    }
  }
  return 0, errors.New("Unbalanced parens")
}

func findMatchingParen(args []string) (int, error) {
  numLeftP := 0
  for i, arg := range args {
    switch {
    case arg == "(":
      numLeftP += 1
    case arg == ")":
      if numLeftP == 0 {
        return i, nil
      } else {
        numLeftP -= 1
      }
    }
  }

  return 0, errors.New("Unbalanced parens")
}

func Read(tokens []string, env *Env) (string, error) {
  if len(tokens) == 0 {
    return "", errors.New("Unexpected EOF while reading")
  }

  firstToken := tokens[0]

  switch firstToken {
  case "(":
    mParen, err := findMatchingParen(tokens[1:])
    if err == nil {
      retVal, err := Eval(tokens[1:mParen+1], env)
      return fmt.Sprintf("%v", retVal), err
    } else {
      return "", err
    }
  case ")":
    return "", errors.New("Unexpected ')'")
  }

  a, e := atomize(firstToken, env)
  return fmt.Sprintf("%v", a), e
}

func Eval(exp []string, env *Env) (Value, error) {
  if len(exp) <= 0 {
    return 0, errors.New("No arguments in expression")
  }

  funcName := exp[0]
  envValue := env.find(funcName)
  expLen := len(exp[1:])
  args := make([]Value, expLen)

  ignoreNum := 0
  for i, val := range exp[1:] {
    if ignoreNum == 0 {
      if "(" == val {
        mParen, _ := findMatchingParen(exp[i+2:])
        retVal, _ := Eval(exp[i+2:i+2+mParen], env)
        args[i], _ = atomize(fmt.Sprintf("%v",retVal), env)
        ignoreNum = mParen
      } else if val != ")" {
        args[i], _ = atomize(val, env)
      }
    } else {
      ignoreNum -= 1
    }
  }

  fargs := Filter(args, (func(x Value) bool { return x != nil }))
  f, _ := envValue.symbols[funcName].(func([]Value) Value)
  return f(fargs), nil
}

func Filter(s []Value, fn func(Value) bool) []Value {
  var p []Value // == nil
  for _, i := range s {
    if fn(i) {
      p = append(p, i)
    }
  }
  return p
}

func main() {
  globalEnv := Env{}
  add_globals(&globalEnv)

  prompt := "* "
  reader := bufio.NewReader(os.Stdin)

  fmt.Print(prompt)
  buf, isPrefix, err := reader.ReadLine()
  for err == nil && !isPrefix {
    line := string(buf)

    result, err := Read(Tokenize(line), &globalEnv)
    if err == nil {
      fmt.Println(result)
    } else {
      fmt.Println(err)
    }

    fmt.Print(prompt)
    buf, isPrefix, err = reader.ReadLine()
  }

  if isPrefix {
    fmt.Println(errors.New("Buffer size is too small"))
    return
  }

  if err != io.EOF {
    fmt.Println(err)
    return
  }

  return
}
