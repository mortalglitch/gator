
package main

import (
	"errors"
	"context"
	"fmt"

	"github.com/mortalglitch/gator/internal/database"
)

type command struct {
	Name string
	Args []string
}

type commands struct {
	registeredCommands map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.registeredCommands[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	f, ok := c.registeredCommands[cmd.Name]
	if !ok {
		return errors.New("command not found")
	}
	return f(s, cmd)
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		// Grab current user ID
		current_user := s.cfg.CurrentUserName
		user, err := s.db.GetUser(context.Background(), current_user)
		if err != nil {
			return fmt.Errorf("Unable to find user %s", current_user)
		}

		return handler(s, cmd, user)
	}
}
