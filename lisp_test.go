package main

import "fmt"

func ExampleAdd() {
	env := Env{}
	add_globals(&env)

	result, err := Read(Tokenize("(+ 2 2)"), &env)

	if err == nil {
		fmt.Println(result)
	} else {
		fmt.Println(err)
	}

	// Output:
	// 4
}
