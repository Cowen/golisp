package golisp

import (
  "errors"
  "strings"
  "strconv"
  "bufio"
  "os"
  "fmt"
  "io"
)

type Env struct {
  symbols map[string](interface{})
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

func reduceInts(init int, f (func(_,_ int) int), nums []int) int {
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
  lParenStr := strings.Replace(s, "("," ( ", -1)
  rParenStr := strings.Replace(lParenStr, ")"," ) ", -1)
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

func findNextParen(args []string) (int,error) {
  for i, arg := range args {
    if arg == ")" {
      return i, nil
    }
  }
  return 0, errors.New("Unbalanced parens")
}

func Read(tokens []string, env *Env) (string, error) {
  if len(tokens) == 0 {
    return "", errors.New("Unexpected EOF while reading")
  }

  for i, token := range tokens {
    switch {
    case "(" == token:
      nextParen, err := findNextParen(tokens[i+1:])
      if err == nil {
        retVal, err := Eval(tokens[i+1:nextParen+1], env)
        return strconv.Itoa(retVal), err
      } else {
        return "", err
      }
    case ")" == token:
      return "",  errors.New("Unexpected ')'")
    default:
      a, e := atomize(token, env)
      return strconv.Itoa(a.(int)), e
    }
  }

  return "", nil
}

func Eval(exp []string, env *Env) (int, error) {
  if len(exp) <= 0 {
    return 0, errors.New("No arguments in expression")
  }

  envValue := env.find(exp[0])
  expLen := len(exp[1:]);
  args := make([]interface{}, expLen)
  for i, exp := range exp[1:] {
    atom, _ := atomize(exp, env)
    args[i] = atom
  }

  f, _ := envValue.symbols[exp[0]].(func([]interface{}) int)
  return f(args), nil
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
