package util

import (
	"encoding/json"
	"log"
	"net/http"
)

func WriteErrorResponse(w http.ResponseWriter, statusCode int, errorMessage string) {
	log.Println("Function WriteErrorResponse called")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	errorResponse := struct {
		Error string `json:"error"`
	}{Error: errorMessage}
	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("WriteErrorResponse: sent status code= ", statusCode, "with error message= ", errorResponse)
}

func WriteSuccessResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	log.Println("Function WriteSuccessResponse called")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("WriteSuccessResponse: sent status code= ", statusCode, " with date= ", data)
}
