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

type Env struct {
  symbols  map[string](interface{})
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

func reduceInts(init int, f (func(_, _ int) int), nums []int) int {
  result := init
  for _, num := range nums {
    result = f(result, num)
  }
  return result
}

func add_globals(env *Env) {
  if env.symbols == nil {
    env.symbols = make(map[string](interface{}))
  }

  env.symbols["+"] = func(args []interface{}) int {
    sum := args[0].(int)
    for _, num := range args[1:] {
      sum += num.(int)
    }
    return sum
  }

  env.symbols["-"] = func(args []interface{}) int {
    diff := args[0].(int)
    for _, num := range args[1:] {
      diff -= num.(int)
    }
    return diff
  }

  env.symbols["*"] = func(args []interface{}) int {
    prod := args[0].(int)
    for _, num := range args[1:] {
      prod = prod * num.(int)
    }
    return prod
  }

  env.symbols["/"] = func(args []interface{}) int {
    quot := args[0].(int)
    for _, num := range args[1:] {
      quot /= num.(int)
    }
    return quot
  }

  env.symbols["define"] = func(args []interface{}) int {
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

func atomize(token string, env *Env) (interface{}, error) {
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
      return strconv.Itoa(retVal), err
    } else {
      return "", err
    }
  case ")":
    return "", errors.New("Unexpected ')'")
  }

  a, e := atomize(firstToken, env)
  return strconv.Itoa(a.(int)), e
}

func Eval(exp []string, env *Env) (int, error) {
  if len(exp) <= 0 {
    return 0, errors.New("No arguments in expression")
  }

  funcName := exp[0]
  envValue := env.find(funcName)
  expLen := len(exp[1:])
  args := make([]interface{}, expLen)

  ignoreNum := 0
  for i, val := range exp[1:] {
    if ignoreNum == 0 {
      if "(" == val {
        mParen, _ := findMatchingParen(exp[i+2:])
        retVal, _ := Eval(exp[i+2:i+2+mParen], env)
        args[i], _ = atomize(strconv.Itoa(retVal), env)
        ignoreNum = mParen
      } else if val != ")" {
        args[i], _ = atomize(val, env)
      }
    } else {
      ignoreNum -= 1
    }
  }

  fargs := Filter(args, (func(x interface{}) bool { return x != nil }))
  f, _ := envValue.symbols[funcName].(func([]interface{}) int)
  return f(fargs), nil
}

func Filter(s []interface{}, fn func(interface{}) bool) []interface{} {
  var p []interface{} // == nil
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
