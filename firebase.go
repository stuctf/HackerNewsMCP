package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const firebaseBase = "https://hacker-news.firebaseio.com/v0"

func checkStatus(resp *http.Response, context string) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	return fmt.Errorf("%s: HTTP %d", context, resp.StatusCode)
}

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

// FetchFeed retrieves a list of item IDs for a given feed type.
func FetchFeed(feed string) ([]int, error) {
	url := fmt.Sprintf("%s/%sstories.json", firebaseBase, feed)
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch feed %s: %w", feed, err)
	}
	defer resp.Body.Close()
	if err := checkStatus(resp, "fetch feed "+feed); err != nil {
		return nil, err
	}

	var ids []int
	if err := json.NewDecoder(resp.Body).Decode(&ids); err != nil {
		return nil, fmt.Errorf("decode feed %s: %w", feed, err)
	}
	return ids, nil
}

// FetchItem retrieves a single item by ID.
func FetchItem(id int) (*FirebaseItem, error) {
	url := fmt.Sprintf("%s/item/%d.json", firebaseBase, id)
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch item %d: %w", id, err)
	}
	defer resp.Body.Close()
	if err := checkStatus(resp, fmt.Sprintf("fetch item %d", id)); err != nil {
		return nil, err
	}

	var item FirebaseItem
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, fmt.Errorf("decode item %d: %w", id, err)
	}
	return &item, nil
}

// FetchItems concurrently fetches multiple items by ID.
func FetchItems(ids []int, concurrency int) []*FirebaseItem {
	results := make([]*FirebaseItem, len(ids))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for i, id := range ids {
		wg.Add(1)
		sem <- struct{}{}
		go func(i, id int) {
			defer wg.Done()
			defer func() { <-sem }()
			item, err := FetchItem(id)
			if err == nil {
				results[i] = item
			}
		}(i, id)
	}

	wg.Wait()
	return results
}

// FetchUser retrieves a user profile by username.
func FetchUser(username string) (*FirebaseUser, error) {
	url := fmt.Sprintf("%s/user/%s.json", firebaseBase, username)
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch user %s: %w", username, err)
	}
	defer resp.Body.Close()
	if err := checkStatus(resp, "fetch user "+username); err != nil {
		return nil, err
	}

	var user FirebaseUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("decode user %s: %w", username, err)
	}
	return &user, nil
}
