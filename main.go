package main

import (
	"github.com/Edofo/bitly-clone/cmd"
	_ "github.com/Edofo/bitly-clone/cmd/cli"
	_ "github.com/Edofo/bitly-clone/cmd/server"
)

func main() {
	cmd.Execute()
}
