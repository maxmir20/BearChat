package api

import "time"

type Post struct {
	Content  string    `json:"content"`
	PostID   string    `json:"postID"`
	AuthorID string    `json:"authorID"`
	PostTime time.Time `json:"postTime"`
}
