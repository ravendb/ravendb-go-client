package ravendb

import "time"

type Post struct {
	ID            String
	Title         String
	Desc          String
	Comments      []*Post
	AttachmentIds String
	CreatedAt     time.Time
}

func NewPost() *Post {
	return &Post{}
}

func (p *Post) getId() String {
	return p.ID
}

func (p *Post) setId(id String) {
	p.ID = id
}

func (p *Post) getTitle() String {
	return p.Title
}

func (p *Post) setTitle(title String) {
	p.Title = title
}

func (p *Post) getDesc() String {
	return p.Desc
}

func (p *Post) setDesc(desc String) {
	p.Desc = desc
}

func (p *Post) getComments() []*Post {
	return p.Comments
}

func (p *Post) setComments(comments []*Post) {
	p.Comments = comments
}

func (p *Post) getAttachmentIds() String {
	return p.AttachmentIds
}

func (p *Post) setAttachmentIds(attachmentIds String) {
	p.AttachmentIds = attachmentIds
}

func (p *Post) getCreatedAt() time.Time {
	return p.CreatedAt
}

func (p *Post) setCreatedAt(createdAt time.Time) {
	p.CreatedAt = createdAt
}
