package main

import "time"

// Story represents a HN story with metadata.
type Story struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	URL         string `json:"url,omitempty"`
	Score       int    `json:"score"`
	By          string `json:"by"`
	Time        string `json:"time"`
	Descendants int    `json:"descendants"`
	HNURL       string `json:"hn_url"`
}

// Comment represents a single comment in a tree.
type Comment struct {
	ID       int        `json:"id"`
	By       string     `json:"by"`
	Text     string     `json:"text"`
	Time     string     `json:"time"`
	Children []*Comment `json:"children"`
}

// User represents a HN user profile.
type User struct {
	Username          string       `json:"username"`
	Karma             int          `json:"karma"`
	Created           string       `json:"created"`
	About             string       `json:"about,omitempty"`
	RecentSubmissions []Submission `json:"recent_submissions,omitempty"`
}

// Submission is a brief item summary for user profiles.
type Submission struct {
	ID    int    `json:"id"`
	Title string `json:"title,omitempty"`
	Type  string `json:"type"`
	Score int    `json:"score"`
	Time  string `json:"time"`
}

// FirebaseItem is the raw item shape from the Firebase API.
type FirebaseItem struct {
	ID          int    `json:"id"`
	Type        string `json:"type"`
	By          string `json:"by"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Text        string `json:"text"`
	Score       int    `json:"score"`
	Time        int64  `json:"time"`
	Descendants int    `json:"descendants"`
	Kids        []int  `json:"kids"`
	Dead        bool   `json:"dead"`
	Deleted     bool   `json:"deleted"`
}

// FirebaseUser is the raw user shape from the Firebase API.
type FirebaseUser struct {
	ID        string `json:"id"`
	Karma     int    `json:"karma"`
	Created   int64  `json:"created"`
	About     string `json:"about"`
	Submitted []int  `json:"submitted"`
}

// FormatTime converts a unix timestamp to an ISO 8601 string.
func FormatTime(unix int64) string {
	return time.Unix(unix, 0).UTC().Format(time.RFC3339)
}

// FormatDate converts a unix timestamp to a date string.
func FormatDate(unix int64) string {
	return time.Unix(unix, 0).UTC().Format("2006-01-02")
}
