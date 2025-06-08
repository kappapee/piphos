package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/kappapee/piphos/internal/config"
)

type Command struct {
	Name        string
	Description string
	Handler     func(cfg config.Config, args []string) (string, error)
}

type CLI struct {
	Cfg      *config.Config
	Commands map[string]*Command
}

func NewCLI() *CLI {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("unable to read configuration file: %v", err)
	}
	if cfg.Token == "" {
		log.Fatal("TOKEN must be set in the configuration file")
	}

	return &CLI{
		Cfg:      &cfg,
		Commands: make(map[string]*Command),
	}
}

func (cli *CLI) AddCommand(name, description string, handler func(cfg config.Config, args []string) (string, error)) {
	cli.Commands[name] = &Command{
		Name:        name,
		Description: description,
		Handler:     handler,
	}
}

func (cli *CLI) Run() (string, error) {
	var showHelp bool
	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.BoolVar(&showHelp, "h", false, "Show help")
	flag.Parse()

	args := flag.Args()

	if showHelp || len(args) == 0 {
		cli.showHelp()
		return "", nil
	}

	cmdName := args[0]
	cmdArgs := args[1:]

	if cmd, exists := cli.Commands[cmdName]; exists {
		return cmd.Handler(*cli.Cfg, cmdArgs)
	}

	cli.showHelp()
	return "", fmt.Errorf("unknown command: %s", cmdName)
}

func (cli *CLI) showHelp() {
	fmt.Println("Usage: piphos [flags] <command> [args...]")
	fmt.Println("\nFlags: ")
	flag.PrintDefaults()
	fmt.Println("\nAvailable commands: ")
	for name, cmd := range cli.Commands {
		fmt.Printf("  %-10s %s\n", name, cmd.Description)
	}
}
