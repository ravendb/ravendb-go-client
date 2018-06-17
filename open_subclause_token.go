package ravendb

var (
	OpenSubclauseToken_INSTANCE = NewOpenSubclauseToken()
)

type OpenSubclauseToken struct {
}

func NewOpenSubclauseToken() *OpenSubclauseToken {
	return &OpenSubclauseToken{}
}

func (t *OpenSubclauseToken) writeTo(writer *StringBuilder) {
	writer.append("(")
}
