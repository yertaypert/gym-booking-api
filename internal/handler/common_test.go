package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParsePathID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/gyms/12", nil)
	req.SetPathValue("id", "12")

	id, err := parsePathID(req, "id")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != 12 {
		t.Fatalf("expected id 12, got %d", id)
	}
}

func TestParsePathIDInvalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/gyms/abc", nil)
	req.SetPathValue("id", "abc")

	_, err := parsePathID(req, "id")
	if err == nil {
		t.Fatal("expected error for invalid id")
	}
}
