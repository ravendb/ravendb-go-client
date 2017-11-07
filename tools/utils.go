package tools

import (
	"regexp"
	"errors"
	"io/ioutil"
	"net/http"
	"encoding/json"
	"math/big"
	"strings"
)

const noValidName = "Database name can only contain only A-Z, a-z, \"_\", \".\" or \"-\" but was: "
var dbValidName = regexp.MustCompile(`^([A-Za-z0-9_\-\.]+)$`)
// DatabaseNameValidation return error is name not valid
func DatabaseNameValidation(name string) error{
	if name == "" {
		return errors.New(("None name is not valid"))
	}
	if !dbValidName.MatchString(name) {
		return errors.New( noValidName + name)
	}

	return nil
}

func ResponseToJSON(resp *http.Response) (out []byte, err error) {
	if data, err := ioutil.ReadAll(resp.Body); err != nil{
		return nil, err
	} else {
		out, err = json.Marshal(data)
	}

	return
}
// todo: implemented later accuracy
func Uuid4() uint64 {

	uuid := `25f64dba-634d-4613-9516-9ca61b161454`
	var i big.Int
	i.SetString(strings.Replace(uuid, "-", "", 4), 16)
	return i.Uint64()
}

func GetChangeVectorFromHeader(resp *http.Response) string{
	headers, ok := resp.Header["ETag"]
	if ok && strings.HasPrefix( headers[0], `"`) {
		header := headers[0]
		return header[1: len(header)-2]
	}

	return ""
}