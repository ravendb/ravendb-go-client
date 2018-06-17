package ravendb

var (
	OpenSubclauseToken_INSTANCE = NewOpenSubclauseToken()
)

type OpenSubclauseToken struct {
	*QueryToken
}

func NewOpenSubclauseToken() *OpenSubclauseToken {
	return &OpenSubclauseToken{
		QueryToken: NewQueryToken(),
	}
}

func (t *OpenSubclauseToken) writeTo(writer *StringBuilder) {
	writer.append("(")
}
