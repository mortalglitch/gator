package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"

	"github.com/mortalglitch/gator/internal/database"
	"github.com/google/uuid"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	httpClient := http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "gator")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	dat, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rssFeed RSSFeed
	err = xml.Unmarshal(dat, &rssFeed)
	if err != nil {
		return nil, err
	}

	rssFeed.Channel.Title = html.UnescapeString(rssFeed.Channel.Title)
	rssFeed.Channel.Description = html.UnescapeString(rssFeed.Channel.Description)
	for i, item := range rssFeed.Channel.Item {
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)
		rssFeed.Channel.Item[i] = item
	}

	return &rssFeed, nil
}

func scrapeFeeds(s *state) error{
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}
	
	fmt.Printf("Checking: %s\n", feed.ID)
	// Mark Fetched
	markComplete := s.db.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams{
		LastFetchedAt: sql.NullTime{time.Now().UTC(), true},
		ID:      feed.ID,
	})
	if markComplete != nil {
		return fmt.Errorf("couldn't mark feed as fetched: %w", err)
	}

	

	rssFeed, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		return err
	}

	fmt.Printf("Channel Result: %v\n", rssFeed.Channel.Title)
	for _, item := range rssFeed.Channel.Item {
		fmt.Printf("* %v\n", item.Title)
		// Add post to DB
		publishTime, err := ParseFlexibleTime(item.PubDate)
		if err != nil {
			return err
		}

		newPost, err := s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Title:       item.Title,
			Url:         item.Link,
			Description: item.Description,
			PublishedAt: publishTime,
			FeedID:      feed.ID,
		})	
		if err != nil {
			fmt.Errorf("Error occured when creating post: %s%v\n", newPost.Title, err)
		}
	}
	
	return nil
}

var commonLayouts = []string{
	time.RFC1123Z,
	time.RFC1123,
	time.RFC3339,
	time.RFC3339Nano,
	"2006-01-02 15:04:05",
	"2006/01/02 15:04:05",
	"02 Jan 2006 15:04:05",
	"2006-01-02",
}

func ParseFlexibleTime(dateString string) (time.Time, error) {
	trimmedDateString := dateString 

	for _, layout := range commonLayouts {
		t, err := time.Parse(layout, trimmedDateString)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("failed to parse date string '%s' using %d layouts", dateString, len(commonLayouts))
}
