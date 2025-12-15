package main

import _ "github.com/lib/pq"

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"net/http"
	"io"
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

type RSSFeed struct {
	Channel struct {
		Title        string    `xml:"title"`
		Link         string    `xml:"link"`
		Description  string    `xml:"description"`
		Item         []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
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
	currentCommands.register("agg", handlerAgg)

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

func handlerAgg(s *state, cmd command) error {
	//feed, err := fetchFeed(context.Background(), "https://mashable.com/feeds/rss/all")
	feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		fmt.Println("Error fetching feed")
		os.Exit(1)
	}

	items := feed.Channel.Item
	fmt.Println(feed.Channel.Title)	
	fmt.Println(feed.Channel.Link)	
	fmt.Println(feed.Channel.Description)	
	for _, item := range items {
		fmt.Printf("%s\n", item.Title)
		fmt.Printf("%s\n", item.Link)
		fmt.Printf("%s\n", item.Description)
		fmt.Printf("%s\n", item.PubDate)
	}

	return nil
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	newClient := NewClient(5 * time.Second)
	fmt.Println("Attempting to read feed from: ", feedURL)

	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		fmt.Println("An error occurred building a request.")
		os.Exit(1)
	}
	req.Header.Set("User-Agent", "gator")

	resp, err := newClient.httpClient.Do(req)
	if err != nil {
		fmt.Println("An error occurred fetching information.")
		os.Exit(1)
	}
	defer resp.Body.Close()
	
	dat, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("An error occurred reading data")
		os.Exit(1)
	}
	
	feedResult := RSSFeed{}
	err = xml.Unmarshal(dat, &feedResult)
	if err != nil {
		fmt.Println("An error occurred moving the data into the struct.", err)
		os.Exit(1)
	}

	fmt.Println("HTTP status:", resp.StatusCode)
	fmt.Println("First 300 bytes:\n", string(dat[:300]))

	fmt.Println("Channel title after unmarshal:", feedResult.Channel.Title)
	fmt.Println("Items count:", len(feedResult.Channel.Item))

	// Need to unescape the sequence here
	feedResult.Channel.Title = html.UnescapeString(feedResult.Channel.Title)
	feedResult.Channel.Description = html.UnescapeString(feedResult.Channel.Description)
	for i, item := range feedResult.Channel.Item {
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)
		feedResult.Channel.Item[i] = item
	}

	return &feedResult, nil
}

// Client -
type Client struct {
	httpClient http.Client
}

// NewClient -
func NewClient(timeout time.Duration) Client {
	return Client{
		httpClient: http.Client{
			Timeout: timeout,
		},
	}
}
