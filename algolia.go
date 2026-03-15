package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const algoliaBase = "https://hn.algolia.com/api/v1"

// AlgoliaResponse is the raw response from the Algolia API.
type AlgoliaResponse struct {
	Hits        []AlgoliaHit `json:"hits"`
	NbHits      int          `json:"nbHits"`
	Page        int          `json:"page"`
	NbPages     int          `json:"nbPages"`
	HitsPerPage int          `json:"hitsPerPage"`
}

// AlgoliaHit is a single search result.
type AlgoliaHit struct {
	ObjectID    string `json:"objectID"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Author      string `json:"author"`
	Points      int    `json:"points"`
	NumComments int    `json:"num_comments"`
	CreatedAt   string `json:"created_at"`
	StoryText   string `json:"story_text"`
}

// SearchParams holds the parameters for a search query.
type SearchParams struct {
	Query     string
	Type      string
	MinPoints int
	DateFrom  string
	DateTo    string
	Sort      string
	Limit     int
}

// SearchHN queries the Algolia HN Search API.
func SearchHN(p SearchParams) (*AlgoliaResponse, error) {
	// Pick endpoint based on sort.
	endpoint := "/search"
	if p.Sort == "date" {
		endpoint = "/search_by_date"
	}

	params := url.Values{}
	params.Set("query", p.Query)
	params.Set("hitsPerPage", strconv.Itoa(p.Limit))

	// Type filter.
	if p.Type == "story" || p.Type == "comment" {
		params.Set("tags", p.Type)
	}

	// Numeric filters.
	var filters []string
	if p.MinPoints > 0 {
		filters = append(filters, fmt.Sprintf("points>%d", p.MinPoints))
	}
	if p.DateFrom != "" {
		ts, err := isoToUnix(p.DateFrom)
		if err != nil {
			return nil, fmt.Errorf("invalid date_from: %w", err)
		}
		filters = append(filters, fmt.Sprintf("created_at_i>%d", ts))
	}
	if p.DateTo != "" {
		ts, err := isoToUnix(p.DateTo)
		if err != nil {
			return nil, fmt.Errorf("invalid date_to: %w", err)
		}
		filters = append(filters, fmt.Sprintf("created_at_i<%d", ts))
	}
	if len(filters) > 0 {
		params.Set("numericFilters", strings.Join(filters, ","))
	}

	reqURL := fmt.Sprintf("%s%s?%s", algoliaBase, endpoint, params.Encode())
	resp, err := httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("algolia search: %w", err)
	}
	defer resp.Body.Close()
	if err := checkStatus(resp, "algolia search"); err != nil {
		return nil, err
	}

	var result AlgoliaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode algolia response: %w", err)
	}
	return &result, nil
}

func isoToUnix(iso string) (int64, error) {
	t, err := time.Parse("2006-01-02", iso)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}
