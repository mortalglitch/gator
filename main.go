package main

import (
	"fmt"
	"os"

	"github.com/mortalglitch/gator/internal/config"
)

type state struct {
	cfg *config.Config
}

type command struct {
	name      string
	arguments []string
}

type commands struct {
	commandList map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.commandList[cmd.name]
	if !ok {
		return fmt.Errorf("unknown command: %s", cmd.name)
	}

	return handler(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.commandList[name] = f
}

func main() {
	currentState := state{}
	temp, err := config.Read()
	if err != nil {
		fmt.Println("error reading gator config")
		os.Exit(1)
	}
	currentState.cfg = &temp
	currentCommands := &commands{}
	currentCommands.commandList = make(map[string]func(*state, command) error)
	currentCommands.register("login", handlerLogin)

	arguments := os.Args
	if len(arguments) > 1 {
		newCommand := command{}
		newCommand.name = arguments[1]

		if len(arguments) > 2 {
			argumentBundle := []string{arguments[2]}
			newCommand.arguments = argumentBundle
		}
		err := currentCommands.run(&currentState, newCommand)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else {
			os.Exit(0)
		}

	}
	if len(arguments) < 2 {
		fmt.Println("too few arguments")
		os.Exit(1)
	}
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) < 1 {
		return fmt.Errorf("invalid command")
	}
	s.cfg.CurrentUserName = cmd.arguments[0]
	config.SetUser(s.cfg, s.cfg.CurrentUserName)
	fmt.Printf("user has been set to %s\n", cmd.arguments[0])
	return nil
}
