package ravendb

type CounterDetail struct {
	DocumentID    string           `json:"DocumentId"`
	CounterName   string           `json:"CounterName"`
	TotalValue    int64            `json:"TotalValue"`
	Etag          int64            `json:"Etag"`
	CounterValues map[string]int64 `json:"CounterValues"`
	ChangeVector  string           `json:"ChangeVector"`
}
