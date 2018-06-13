package ravendb

import "time"

type Post struct {
	ID            string
	Title         string
	Desc          string
	Comments      []*Post
	AttachmentIds string
	CreatedAt     time.Time
}

func NewPost() *Post {
	return &Post{}
}

func (p *Post) getId() string {
	return p.ID
}

func (p *Post) setId(id string) {
	p.ID = id
}

func (p *Post) getTitle() string {
	return p.Title
}

func (p *Post) setTitle(title string) {
	p.Title = title
}

func (p *Post) getDesc() string {
	return p.Desc
}

func (p *Post) setDesc(desc string) {
	p.Desc = desc
}

func (p *Post) getComments() []*Post {
	return p.Comments
}

func (p *Post) setComments(comments []*Post) {
	p.Comments = comments
}

func (p *Post) getAttachmentIds() string {
	return p.AttachmentIds
}

func (p *Post) setAttachmentIds(attachmentIds string) {
	p.AttachmentIds = attachmentIds
}

func (p *Post) getCreatedAt() time.Time {
	return p.CreatedAt
}

func (p *Post) setCreatedAt(createdAt time.Time) {
	p.CreatedAt = createdAt
}
