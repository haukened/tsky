package main

import (
	"fmt"

	"github.com/haukened/tsky/internal/auth"
	"github.com/haukened/tsky/internal/config"
)

func main() {
	c, err := config.New("~/.config/tsky/config.yaml")
	if err != nil {
		panic(err)
	}
	err = c.Load()
	if err != nil {
		panic(err)
	}
	err = auth.AuthUser(c)
	if err != nil {
		panic(err)
	}
	fmt.Println("Done")
}
