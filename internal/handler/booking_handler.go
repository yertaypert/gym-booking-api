package handler

import (
	"encoding/json"
	"net/http"

	"github.com/yertaypert/gym-booking-api/internal/usecase"
)

type BookingHandler struct {
	bookingUsecase *usecase.BookingUsecase // ← changed
}

func NewBookingHandler(bookingUsecase *usecase.BookingUsecase) *BookingHandler { // ← changed
	return &BookingHandler{
		bookingUsecase: bookingUsecase,
	}
}

type CreateBookingRequest struct {
	UserID    int `json:"user_id"`
	SessionID int `json:"session_id"`
}

type CreateBookingResponse struct {
	BookingID int    `json:"booking_id"`
	Message   string `json:"message"`
}

func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req CreateBookingRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	bookingID, err := h.bookingUsecase.CreateBooking(r.Context(), req.UserID, req.SessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := CreateBookingResponse{
		BookingID: bookingID,
		Message:   "Booking Created Successfully",
	}
	_ = json.NewEncoder(w).Encode(response)
}

type CancelBookingRequest struct {
	BookingID int `json:"booking_id"`
}

type CancelBookingResponse struct {
	Message string `json:"message"`
}

func (h *BookingHandler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req CancelBookingRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.bookingUsecase.CancelBooking(r.Context(), req.BookingID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := CancelBookingResponse{
		Message: "Booking Cancelled Successfully",
	}

	_ = json.NewEncoder(w).Encode(response)
}
