package ravendb

var _ QueryToken = &CloseSubclauseToken{}

var (
	CloseSubclauseToken_INSTANCE = NewCloseSubclauseToken()
)

type CloseSubclauseToken struct {
}

func NewCloseSubclauseToken() *CloseSubclauseToken {
	return &CloseSubclauseToken{}
}

func (t *CloseSubclauseToken) WriteTo(writer *StringBuilder) {
	writer.append(")")
}
