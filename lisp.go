package main

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
  symbols map[string](func(_ []int) int)
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
    env.symbols = make(map[string](func(_ []int) int))
  }

  env.symbols["+"] = func(nums []int) int {
    sum := 0
    for _, num := range nums {
      sum += num
    }
    return sum
  }

  env.symbols["-"] = func(nums []int) int {
    diff := nums[0]
    for _, num := range nums[1:] {
      diff -= num
    }
    return diff
  }

  env.symbols["*"] = func(nums []int) int {
    prod := nums[0]
    for _, num := range nums[1:] {
      prod = prod * num
    }
    return prod
  }

  env.symbols["/"] = func(nums []int) int {
    quot := nums[0]
    for _, num := range nums[1:] {
      quot /= num
    }
    return quot
  }
}

func tokenize(s string) []string {
  // Convert a string into a list of tokens
  lParenStr := strings.Replace(s, "("," ( ", -1)
  rParenStr := strings.Replace(lParenStr, ")"," ) ", -1)
  return strings.Fields(rParenStr)
}

func pop(s []string) (string, []string) {
  return s[0], s[1:]
}

func atomize(token string) (int, error) {
  // Numbers become numbers.
  // Later, everything else will become values
  return strconv.Atoi(token)
}

func read(tokens []string, env *Env) (string, error) {
  if len(tokens) == 0 {
    return "", errors.New("unexpected EOF while reading")
  }

  for i, token := range tokens {
    switch {
    case "(" == token:
      return strconv.Itoa(eval(tokens[i+1:], env)), nil
    case ")" == token:
      return "",  errors.New("unexpected ')'")
    default:
      s, e := atomize(token)
      return string(s), e
    }
  }

  return "", nil
}

func eval(exp []string, env *Env) int {
  // fmt.Println("env has: ",env)
  // fmt.Println("We're looking for",exp[0])
  envValue := env.find(exp[0])
  args := make([]int, len(exp[1:])-1)
  for i, exp := range exp[1:] {
    if ")" == exp {
      break
    }
    args[i], _ = atomize(exp)
  }
  // fmt.Println("Eval func: <", exp[0], "> with args: ",arg1,arg2)
  // fmt.Println("We return: ", envValue.symbols[exp[0]](arg1, arg2))
  return envValue.symbols[exp[0]](args)
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

    result, err := read(tokenize(line), &globalEnv)
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
