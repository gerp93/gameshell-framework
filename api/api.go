package api

import "net/http"

func WriteGoodHeader(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Add("HX-Retarget", "find .htmx-result-good")
	writeHeader(w, statusCode, message)
}

func WriteBadHeader(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Add("HX-Retarget", "find .htmx-result-bad")
	writeHeader(w, statusCode, message)
}

func writeHeader(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	w.Write([]byte(message))
}
