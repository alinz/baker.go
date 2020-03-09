package response

import (
	"encoding/json"
	"net/http"
)

// AsJSON is converts the given message to json response
// if body is an error type, it creates a custom struct with error message
func AsJSON(w http.ResponseWriter, statusCode int, body interface{}) error {
	if v, ok := body.(error); ok {
		body = struct {
			Error string `json:"error"`
		}{
			Error: v.Error(),
		}
	}

	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(body)
}
