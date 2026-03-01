package handler

import "net/http"

// Health returns 200 OK. No auth required, no DB access.
func Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
