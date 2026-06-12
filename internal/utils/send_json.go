package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func SendJson(w http.ResponseWriter, v any, statusCode int) {

	fmt.Println("I AM HERE")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	encoder := json.NewEncoder(w)
	if err := encoder.Encode(v); err != nil {
		http.Error(w, "Internal Server Error-Encode.", http.StatusInternalServerError)
		return
	}
}
