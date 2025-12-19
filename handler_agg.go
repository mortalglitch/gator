package main

import (
	"context"
	"fmt"
	"time"
	"strconv"

	"github.com/mortalglitch/gator/internal/database"
	"github.com/google/uuid"
)

func handlerAgg(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %v <time_between_reqs (1s, 1m, 1h)>", cmd.Name)
	}

	timeBetweenReqs, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("Error parsing time duration: %v", err)
	}

	ticker := time.NewTicker(timeBetweenReqs)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	
	if len(cmd.Args) != 2 {
		return fmt.Errorf("usage: %v <title> <url>", cmd.Name)
	}

	name := cmd.Args[0]
	url := cmd.Args[1]
	
	feed, err := s.db.AddFeed(context.Background(), database.AddFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	})
	if err != nil {
		return fmt.Errorf("couldn't add feed: %w", err)
	}
	
	fmt.Println("Feed added successfully:")
	printFeed(feed, s)

	// Register as following for current user
	follow , err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:    uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return fmt.Errorf("couldn't add follow: %w", err)
	}
	fmt.Println("Added to follow list: ", follow.ID)

	return nil	
}

func printFeed(feed database.Feed, s *state) {
	fmt.Printf(" * ID:      %v\n", feed.ID)
	fmt.Printf(" * Name:    %v\n", feed.Name)
	fmt.Printf(" * URL:     %v\n", feed.Url)
	user, err := s.db.GetUserByID(context.Background(), feed.UserID)
	if err != nil {
		fmt.Printf("Unable to find user from feed list: %s", feed.UserID)
	}
	fmt.Printf(" * User:    %v\n", user.Name)
}

func handlerListFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("couldn't list feeds: %w", err)
	}
	for _, feed := range feeds {
		printFeed(feed, s)
	}

	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %v <url>", cmd.Name)
	}

	url := cmd.Args[0]
	
	// Grab URL from feeds
	feed, err := s.db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("Unable to find existing feed %s", url)
	}

	follow , err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:    uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return fmt.Errorf("couldn't add follow: %w", err)
	}
	

	fmt.Println("Feed follow added successfully:")
	fmt.Printf("Follow ID: %s\n", follow.ID)
	fmt.Printf("User: %s\n", user.Name)
	fmt.Printf("Followed: %s\n", feed.Name)
	return nil	

}

func handlerFollowing(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("usage: %v", cmd.Name)
	}

	feeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)	
	if err != nil {
		return fmt.Errorf("Unable to find feeds for user: %s", user.Name)
	}

	for _, feed := range feeds {
		fmt.Println("* ", feed.FeedName)
	}
	
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %v <url>", cmd.Name)
	}

	url := cmd.Args[0]

	// Grab URL from feeds
	feed, err := s.db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("Unable to find existing feed %s", url)
	}

  unfollowResult := s.db.DeleteUserFeed(context.Background(), database.DeleteUserFeedParams{
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return fmt.Errorf("couldn't unfollow: %w", unfollowResult)
	}

	return nil
}

func handlerBrowse(s *state, cmd command) error {
	
	amount, err := strconv.Atoi(cmd.Args[0])
	if err != nil {
		return err
	}

	if amount < 2 {
		amount = 2
	}

	posts, err := s.db.GetPostForUser(context.Background(), int32(amount))
	if err != nil {
		return err
	}

	for _, post := range posts {
		fmt.Printf("* %v\n", post.Title)
		fmt.Printf("* %v\n", post.Url)
		fmt.Printf("* %v\n", post.Description)
		fmt.Printf("* %v\n", post.PublishedAt)
	}
	
	return nil
}
