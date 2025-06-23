package main

import (
	_ "github.com/Edofo/bitly-clone/cmd/cli"    // Importe le package 'cli' pour que ses init() soient exécutés
	_ "github.com/Edofo/bitly-clone/cmd/server" // Importe le package 'server' pour que ses init() soient exécutés
)

func main() {
	// TODO Exécute la commande racine de Cobra.
}
