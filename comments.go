package main

import "sync/atomic"

// FetchCommentTree fetches a story's comments recursively up to maxDepth and limit.
func FetchCommentTree(storyID, maxDepth, limit int) (*FirebaseItem, []*Comment, int, bool) {
	story, err := FetchItem(storyID)
	if err != nil || story == nil {
		return nil, nil, 0, false
	}

	var count int64
	maxReached := false
	comments := buildTree(story.Kids, 1, maxDepth, limit, &count, &maxReached)
	return story, comments, int(count), maxReached
}

func buildTree(kidIDs []int, depth, maxDepth, limit int, count *int64, maxReached *bool) []*Comment {
	if len(kidIDs) == 0 || depth > maxDepth {
		if depth > maxDepth && len(kidIDs) > 0 {
			*maxReached = true
		}
		return nil
	}

	items := FetchItems(kidIDs, 20)
	var comments []*Comment

	for _, item := range items {
		if item == nil || item.Deleted || item.Dead {
			continue
		}
		if int(atomic.LoadInt64(count)) >= limit {
			break
		}
		atomic.AddInt64(count, 1)

		c := &Comment{
			ID:   item.ID,
			By:   item.By,
			Text: StripHTML(item.Text),
			Time: FormatTime(item.Time),
		}
		c.Children = buildTree(item.Kids, depth+1, maxDepth, limit, count, maxReached)
		comments = append(comments, c)
	}
	return comments
}
