package tests

import "time"

type Post struct {
	ID            string
	Title         string    `json:"title,omitempty"`
	Desc          string    `json:"desc,omitempty"`
	Comments      []*Post   `json:"comments"`
	AttachmentIds string    `json:"attachmentIds,omitempty"`
	CreatedAt     time.Time `json:"createdAt,omitempty"`
}
