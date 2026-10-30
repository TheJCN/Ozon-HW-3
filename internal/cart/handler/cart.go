package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"OzonHW2/internal/cart/model"
	"OzonHW2/internal/cart/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	service *service.Cart
}

func New(logic *service.Cart) *Server {
	return &Server{service: logic}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /user/{user_id}/cart/{sku_id}", s.add)
	mux.HandleFunc("GET /user/{user_id}/cart", s.get)
	return mux
}

func (s *Server) add(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(r.PathValue("user_id"), 10, 64)
	if err != nil || userID <= 0 {
		http.Error(w, "user_id must be positive", http.StatusBadRequest)
		return
	}
	sku, err := strconv.ParseInt(r.PathValue("sku_id"), 10, 64)
	if err != nil || sku <= 0 {
		http.Error(w, "sku_id must be positive", http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1024))
	if err != nil {
		http.Error(w, "cannot read request body", http.StatusBadRequest)
		return
	}
	var request model.AddRequest
	if err := json.Unmarshal(body, &request); err != nil {
		http.Error(w, "invalid JSON or count", http.StatusBadRequest)
		return
	}
	if request.Count == 0 {
		http.Error(w, "count must be positive", http.StatusBadRequest)
		return
	}
	if err := s.service.Add(r.Context(), userID, sku, request.Count); err != nil {
		switch status.Code(err) {
		case codes.NotFound:
			http.Error(w, "product not found", http.StatusPreconditionFailed)
		case codes.OutOfRange:
			http.Error(w, "count overflow", http.StatusBadRequest)
		default:
			http.Error(w, "product service unavailable", http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) get(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(r.PathValue("user_id"), 10, 64)
	if err != nil || userID <= 0 {
		http.Error(w, "user_id must be positive", http.StatusBadRequest)
		return
	}
	result, err := s.service.Get(r.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrEmptyCart) {
			http.Error(w, "cart is empty", http.StatusNotFound)
			return
		}
		switch status.Code(err) {
		case codes.NotFound:
			http.Error(w, "product not found", http.StatusPreconditionFailed)
		default:
			http.Error(w, "product service unavailable", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}
