package ravendb

type OperationExceptionResult struct {
	Type       string `json:"Type"`
	Message    string `json:"Message"`
	Error      string `json:"Error"`
	StatusCode int    `json:"StatusCode"`
}

/*
    public string getType() {
        return type;
    }

    public void setType(string type) {
        this.type = type;
    }

    public string getMessage() {
        return message;
    }

    public void setMessage(string message) {
        this.message = message;
    }

    public string getError() {
        return error;
    }

    public void setError(string error) {
        this.error = error;
    }

    public int getStatusCode() {
        return statusCode;
    }

    public void setStatusCode(int statusCode) {
        this.statusCode = statusCode;
	}
*/
