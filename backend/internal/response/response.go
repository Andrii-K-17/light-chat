package response

import (
	"encoding/json"
	"net/http"
)

// JSON sends a JSON response with a given status code and data.
func JSON(res http.ResponseWriter, status int, data any) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(status)
	_ = json.NewEncoder(res).Encode(data)
}

// Error sends a JSON response with a specific error message.
func Error(res http.ResponseWriter, status int, msg string) {
	JSON(res, status, map[string]string{"error": msg})
}
