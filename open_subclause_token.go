package ravendb

var _ QueryToken = &OpenSubclauseToken{}

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
