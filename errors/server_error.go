package errors

import (
	"net/http"
	"encoding/json"
	"io/ioutil"
)

type ServerError struct{
	Url, Type, Message, Error string
}

func NewServerError(response http.Response) (*ServerError, error){
	var servErr ServerError
	body, err := ioutil.ReadAll(response.Body)
	if err != nil{
		return nil, err
	}
	err = json.Unmarshal(body, servErr)
	if err != nil{
		return nil, err
	}
	return &servErr, nil
}