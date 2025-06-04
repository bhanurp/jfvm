package main

import (
	"log"
	"os"

	"github.com/bhanurp/jfvm/cmd"
	"github.com/urfave/cli/v2"
)

func main() {
	log.Println("Starting jfvm CLI...")
	app := &cli.App{
		Name:                 "jfvm",
		Usage:                "Manage multiple versions of JFrog CLI",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			// Core functionality
			cmd.Install,
			cmd.Use,
			cmd.List,
			cmd.Remove,
			cmd.Clear,
			cmd.Alias,
			cmd.Link,

			// Smart features
			cmd.Init,
			cmd.Detect,

			// Fun features
			cmd.Fortune,
			cmd.Roulette,
			cmd.Pet,
			cmd.Motivate,
			cmd.Translate,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Error running jfvm CLI: %v", err)
	}
}
