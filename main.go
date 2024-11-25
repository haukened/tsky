package main

import (
	"fmt"
	"os"

	"github.com/haukened/tsky/internal/auth"
	"github.com/haukened/tsky/internal/config"
	tokensvc "github.com/haukened/tsky/internal/tokenSvc"
	"github.com/haukened/tsky/internal/utils"
)

var Version string = "dev"

func dontPanic(err error) {
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}

func main() {
	utils.SetVersion(Version)
	c, err := config.New("~/.config/tsky/config.yaml")
	dontPanic(err)
	err = c.Load()
	dontPanic(err)
	err = auth.AuthUser(c)
	dontPanic(err)
	tsvc, err := tokensvc.NewRefresher(c.Server, c.RefreshJwt)
	dontPanic(err)
	c.RefreshJwt = tsvc.RefreshToken()
	err = c.Save()
	dontPanic(err)
	fmt.Println(utils.UserAgent())
}
