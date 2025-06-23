package main

import (
	"github.com/Edofo/bitly-clone/cmd"
	_ "github.com/Edofo/bitly-clone/cmd/cli"    // Importe le package 'cli' pour que ses init() soient exécutés
	_ "github.com/Edofo/bitly-clone/cmd/server" // Importe le package 'server' pour que ses init() soient exécutés
)

func main() {
	// Exécute la commande racine de Cobra.
	cmd.Execute()
}
