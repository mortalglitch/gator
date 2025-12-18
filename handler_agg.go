package main

import (
	"context"
	"fmt"
	"time"

	"github.com/mortalglitch/gator/internal/database"
	"github.com/google/uuid"
)

func handlerAgg(s *state, cmd command) error {
	feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return fmt.Errorf("couldn't fetch feed: %w", err)
	}
	fmt.Printf("Feed: %+v\n", feed)
	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	
	if len(cmd.Args) != 2 {
		return fmt.Errorf("usage: %v <title> <url>", cmd.Name)
	}

	name := cmd.Args[0]
	url := cmd.Args[1]

	// Grab current user ID
	//current_user := s.cfg.CurrentUserName
	//user, err := s.db.GetUser(context.Background(), current_user)
	//if err != nil {
	//	return fmt.Errorf("Unable to find user %s", current_user)
	//}

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
