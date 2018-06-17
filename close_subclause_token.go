package ravendb

var (
	CloseSubclauseToken_INSTANCE = NewCloseSubclauseToken()
)

type CloseSubclauseToken struct {
	*QueryToken
}

func NewCloseSubclauseToken() *CloseSubclauseToken {
	return &CloseSubclauseToken{
		QueryToken: NewQueryToken(),
	}
}

func (t *CloseSubclauseToken) writeTo(writer *StringBuilder) {
	writer.append(")")
}
