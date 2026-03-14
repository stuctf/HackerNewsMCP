package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var validUsername = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// AllTools returns the MCP tool definitions.
func AllTools() []ToolDef {
	return []ToolDef{
		{
			Name:        "hn_front_page",
			Description: "Fetch stories from a Hacker News feed with full metadata.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"feed": map[string]interface{}{
						"type":        "string",
						"description": "Feed type: top, new, best, ask, show, job",
						"default":     "top",
					},
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Number of stories (1-500)",
						"default":     30,
					},
					"include_urls": map[string]interface{}{
						"type":        "boolean",
						"description": "Include URLs in output (default false)",
						"default":     false,
					},
				},
			},
		},
		{
			Name:        "hn_search",
			Description: "Search Hacker News history via Algolia with filters.",
			InputSchema: map[string]interface{}{
				"type":     "object",
				"required": []string{},
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search terms",
					},
					"type": map[string]interface{}{
						"type":        "string",
						"description": "Filter: story, comment, or all",
						"default":     "story",
					},
					"min_points": map[string]interface{}{
						"type":        "number",
						"description": "Minimum score filter",
						"default":     0,
					},
					"date_from": map[string]interface{}{
						"type":        "string",
						"description": "Start date (ISO 8601, e.g. 2025-01-01)",
					},
					"date_to": map[string]interface{}{
						"type":        "string",
						"description": "End date (ISO 8601)",
					},
					"sort": map[string]interface{}{
						"type":        "string",
						"description": "Sort by: points or date",
						"default":     "points",
					},
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Max results (1-50)",
						"default":     20,
					},
					"include_urls": map[string]interface{}{
						"type":        "boolean",
						"description": "Include URLs in output (default false)",
						"default":     false,
					},
				},
			},
		},
		{
			Name:        "hn_comments",
			Description: "Fetch a story's comment tree with depth and limit control.",
			InputSchema: map[string]interface{}{
				"type":     "object",
				"required": []string{"story_id"},
				"properties": map[string]interface{}{
					"story_id": map[string]interface{}{
						"type":        "number",
						"description": "HN item ID",
					},
					"max_depth": map[string]interface{}{
						"type":        "number",
						"description": "Recursion depth (1 = top-level only)",
						"default":     3,
					},
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Max total comments",
						"default":     50,
					},
				},
			},
		},
		{
			Name:        "hn_user",
			Description: "Fetch a Hacker News user profile and optionally recent submissions.",
			InputSchema: map[string]interface{}{
				"type":     "object",
				"required": []string{"username"},
				"properties": map[string]interface{}{
					"username": map[string]interface{}{
						"type":        "string",
						"description": "Case-sensitive HN username",
					},
					"include_submissions": map[string]interface{}{
						"type":        "boolean",
						"description": "Include recent submissions",
						"default":     false,
					},
					"submission_limit": map[string]interface{}{
						"type":        "number",
						"description": "Number of submissions to fetch",
						"default":     10,
					},
				},
			},
		},
	}
}

// DispatchTool routes a tool call to the appropriate handler.
func DispatchTool(name string, args json.RawMessage) ToolResult {
	switch name {
	case "hn_front_page":
		return handleFrontPage(args)
	case "hn_search":
		return handleSearch(args)
	case "hn_comments":
		return handleComments(args)
	case "hn_user":
		return handleUser(args)
	default:
		return toolError(fmt.Sprintf("unknown tool: %s", name))
	}
}

func handleFrontPage(args json.RawMessage) ToolResult {
	var p struct {
		Feed        string `json:"feed"`
		Limit       int    `json:"limit"`
		IncludeURLs bool   `json:"include_urls"`
	}
	p.Feed = "top"
	p.Limit = 30
	if len(args) > 0 {
		if err := json.Unmarshal(args, &p); err != nil {
			return toolError(fmt.Sprintf("invalid arguments: %v", err))
		}
	}

	// Validate feed.
	validFeeds := map[string]bool{"top": true, "new": true, "best": true, "ask": true, "show": true, "job": true}
	if !validFeeds[p.Feed] {
		return toolError(fmt.Sprintf("invalid feed: %s", p.Feed))
	}
	if p.Limit < 1 {
		p.Limit = 1
	} else if p.Limit > 500 {
		p.Limit = 500
	}

	ids, err := FetchFeed(p.Feed)
	if err != nil {
		return toolError(err.Error())
	}
	if len(ids) > p.Limit {
		ids = ids[:p.Limit]
	}

	items := FetchItems(ids, 20)

	var b strings.Builder
	fmt.Fprintf(&b, "Feed: %s | %d stories\n\n", p.Feed, len(items))

	i := 0
	for _, item := range items {
		if item == nil {
			continue
		}
		i++
		fmt.Fprintf(&b, "%d. %s\n", i, item.Title)
		fmt.Fprintf(&b, "   %d points | %d comments | by %s | %s\n",
			item.Score, item.Descendants, item.By, FormatTime(item.Time))
		if p.IncludeURLs {
			if item.URL != "" {
				fmt.Fprintf(&b, "   Link: %s\n", item.URL)
			}
			fmt.Fprintf(&b, "   HN: https://news.ycombinator.com/item?id=%d\n", item.ID)
		}
		b.WriteString("\n")
	}

	return toolText(b.String())
}

func handleSearch(args json.RawMessage) ToolResult {
	var p struct {
		Query       string `json:"query"`
		Type        string `json:"type"`
		MinPoints   int    `json:"min_points"`
		DateFrom    string `json:"date_from"`
		DateTo      string `json:"date_to"`
		Sort        string `json:"sort"`
		Limit       int    `json:"limit"`
		IncludeURLs bool   `json:"include_urls"`
	}
	p.Type = "story"
	p.Sort = "points"
	p.Limit = 20
	if len(args) > 0 {
		if err := json.Unmarshal(args, &p); err != nil {
			return toolError(fmt.Sprintf("invalid arguments: %v", err))
		}
	}

	if p.Limit < 1 {
		p.Limit = 1
	} else if p.Limit > 50 {
		p.Limit = 50
	}

	resp, err := SearchHN(SearchParams{
		Query:     p.Query,
		Type:      p.Type,
		MinPoints: p.MinPoints,
		DateFrom:  p.DateFrom,
		DateTo:    p.DateTo,
		Sort:      p.Sort,
		Limit:     p.Limit,
	})
	if err != nil {
		return toolError(err.Error())
	}

	var b strings.Builder

	// Header with search metadata.
	if p.Query != "" {
		fmt.Fprintf(&b, "Search: %q", p.Query)
	} else {
		fmt.Fprintf(&b, "Search: (all)")
	}
	fmt.Fprintf(&b, " | %d results of %d total | sort=%s\n", len(resp.Hits), resp.NbHits, p.Sort)

	// Filters summary.
	var filters []string
	if p.Type != "" {
		filters = append(filters, "type="+p.Type)
	}
	if p.MinPoints > 0 {
		filters = append(filters, fmt.Sprintf("min_points=%d", p.MinPoints))
	}
	if p.DateFrom != "" {
		filters = append(filters, "from="+p.DateFrom)
	}
	if p.DateTo != "" {
		filters = append(filters, "to="+p.DateTo)
	}
	if len(filters) > 0 {
		fmt.Fprintf(&b, "Filters: %s\n", strings.Join(filters, ", "))
	}
	b.WriteString("\n")

	// Results.
	for i, hit := range resp.Hits {
		fmt.Fprintf(&b, "%d. %s\n", i+1, hit.Title)
		fmt.Fprintf(&b, "   %d points | %d comments | by %s | %s\n",
			hit.Points, hit.NumComments, hit.Author, hit.CreatedAt)
		if p.IncludeURLs {
			if hit.URL != "" {
				fmt.Fprintf(&b, "   Link: %s\n", hit.URL)
			}
			fmt.Fprintf(&b, "   HN: https://news.ycombinator.com/item?id=%s\n", hit.ObjectID)
		}
		b.WriteString("\n")
	}

	return toolText(b.String())
}

func handleComments(args json.RawMessage) ToolResult {
	var p struct {
		StoryID  int `json:"story_id"`
		MaxDepth int `json:"max_depth"`
		Limit    int `json:"limit"`
	}
	p.MaxDepth = 3
	p.Limit = 50
	if len(args) > 0 {
		if err := json.Unmarshal(args, &p); err != nil {
			return toolError(fmt.Sprintf("invalid arguments: %v", err))
		}
	}

	if p.StoryID == 0 {
		return toolError("story_id is required")
	}
	if p.MaxDepth < 1 {
		p.MaxDepth = 1
	} else if p.MaxDepth > 10 {
		p.MaxDepth = 10
	}
	if p.Limit < 1 {
		p.Limit = 1
	} else if p.Limit > 500 {
		p.Limit = 500
	}

	story, comments, totalFetched, maxReached := FetchCommentTree(p.StoryID, p.MaxDepth, p.Limit)
	if story == nil {
		return toolError(fmt.Sprintf("could not fetch story %d", p.StoryID))
	}

	result := map[string]interface{}{
		"story": map[string]interface{}{
			"id":    story.ID,
			"title": story.Title,
			"url":   story.URL,
			"score": story.Score,
			"by":    story.By,
		},
		"comments":          comments,
		"total_fetched":     totalFetched,
		"max_depth_reached": maxReached,
	}
	return toolJSON(result)
}

func handleUser(args json.RawMessage) ToolResult {
	var p struct {
		Username           string `json:"username"`
		IncludeSubmissions bool   `json:"include_submissions"`
		SubmissionLimit    int    `json:"submission_limit"`
	}
	p.SubmissionLimit = 10
	if len(args) > 0 {
		if err := json.Unmarshal(args, &p); err != nil {
			return toolError(fmt.Sprintf("invalid arguments: %v", err))
		}
	}

	if p.Username == "" {
		return toolError("username is required")
	}
	if !validUsername.MatchString(p.Username) {
		return toolError("invalid username: must contain only alphanumeric characters, hyphens, or underscores")
	}
	if p.SubmissionLimit < 1 {
		p.SubmissionLimit = 1
	} else if p.SubmissionLimit > 100 {
		p.SubmissionLimit = 100
	}

	fbUser, err := FetchUser(p.Username)
	if err != nil {
		return toolError(err.Error())
	}

	user := User{
		Username: fbUser.ID,
		Karma:    fbUser.Karma,
		Created:  FormatDate(fbUser.Created),
		About:    StripHTML(fbUser.About),
	}

	if p.IncludeSubmissions && len(fbUser.Submitted) > 0 {
		subIDs := fbUser.Submitted
		if len(subIDs) > p.SubmissionLimit {
			subIDs = subIDs[:p.SubmissionLimit]
		}
		items := FetchItems(subIDs, 20)
		for _, item := range items {
			if item == nil {
				continue
			}
			user.RecentSubmissions = append(user.RecentSubmissions, Submission{
				ID:    item.ID,
				Title: item.Title,
				Type:  item.Type,
				Score: item.Score,
				Time:  FormatTime(item.Time),
			})
		}
	}

	return toolJSON(user)
}

func toolText(text string) ToolResult {
	return ToolResult{
		Content: []ToolContent{{Type: "text", Text: text}},
	}
}

func toolJSON(v interface{}) ToolResult {
	data, err := json.Marshal(v)
	if err != nil {
		return toolError("failed to marshal result")
	}
	return ToolResult{
		Content: []ToolContent{{Type: "text", Text: string(data)}},
	}
}

func toolError(msg string) ToolResult {
	return ToolResult{
		Content: []ToolContent{{Type: "text", Text: msg}},
		IsError: true,
	}
}
