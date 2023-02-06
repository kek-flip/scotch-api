package main

import "github.com/kek-flip/scotch-api/internal/apiserver"

func main() {
	if err := apiserver.StartServer(); err != nil {
		panic(err)
	}
}
