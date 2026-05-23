package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// parsePathID extracts an integer ID from the request path.
func parsePathID(r *http.Request, name string) (int, error) {
	return strconv.Atoi(r.PathValue(name))
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes a plain text error response.
func writeError(w http.ResponseWriter, message string, status int) {
	http.Error(w, message, status)
}
