package ravendb

var (
	CloseSubclauseToken_INSTANCE = NewCloseSubclauseToken()
)

type CloseSubclauseToken struct {
}

func NewCloseSubclauseToken() *CloseSubclauseToken {
	return &CloseSubclauseToken{}
}

func (t *CloseSubclauseToken) writeTo(writer *StringBuilder) {
	writer.append(")")
}
