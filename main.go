package main

import _ "github.com/lib/pq"

import (
	"context"
	"fmt"
	"os"
	"database/sql"
	"time"
	
	"github.com/mortalglitch/gator/internal/config"
	"github.com/mortalglitch/gator/internal/database"

	"github.com/google/uuid"
)

type state struct {
	cfg *config.Config
	db *database.Queries
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
	currentCommands.register("register", handlerRegister)
	currentCommands.register("reset", handlerReset)
	currentCommands.register("users", handlerUsers)

	db, err := sql.Open("postgres", currentState.cfg.DBURL)
	if err != nil {
		fmt.Println("Error openning database")
		os.Exit(1)
	}
	dbQueries := database.New(db)
	currentState.db = dbQueries

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
	exists, _ := s.db.GetUser(context.Background(), cmd.arguments[0])
	if exists == (database.User{}) {
		fmt.Println("user doesn't exist")
		os.Exit(1)
	}

	s.cfg.CurrentUserName = cmd.arguments[0]
	config.SetUser(s.cfg, s.cfg.CurrentUserName)
	fmt.Printf("user has been set to %s\n", cmd.arguments[0])
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.arguments) < 1 {
		return fmt.Errorf("invalid command length")
	}

	exists, _ := s.db.GetUser(context.Background(), cmd.arguments[0])
	if exists != (database.User{}) {
		fmt.Println("user already exist")
		os.Exit(1)
	}

	userParams := database.CreateUserParams{}
	userParams.ID = uuid.New()
	userParams.CreatedAt = time.Now()
	userParams.UpdatedAt = time.Now()
	userParams.Name = cmd.arguments[0]

	user, err := s.db.CreateUser(context.Background(), userParams)
	if err != nil {
		fmt.Println("error creating user in database")
		os.Exit(1)
	}
  handlerLogin(s, cmd)
	fmt.Println("user registered successfully")
	fmt.Println(user)

	return nil
}

func handlerReset(s *state, cmd command) error {
	ok := s.db.Reset(context.Background())
	if ok != nil {
		fmt.Println("Error clearing database")
		os.Exit(1)
	}
	fmt.Println("Database reset complete.")
	return nil
}

func handlerUsers(s *state, cmd command) error {
	ok, err := s.db.GetUsers(context.Background())
	if err != nil {
		fmt.Println("Error fetching users.")
		os.Exit(1)
	}
	for i := 0; i < len(ok); i++ {
		if ok[i] == s.cfg.CurrentUserName {
			ok[i] = s.cfg.CurrentUserName + " (current)"
		}
		fmt.Println("* " + ok[i])
	}
	
	return nil
}
