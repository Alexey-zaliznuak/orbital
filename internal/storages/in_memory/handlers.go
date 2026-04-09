package inmemory

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	storageapi "github.com/Alexey-zaliznuak/orbital/pkg/sdk/storage/api"
)

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	health, err := s.storage.HealthCheck(r.Context())
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": string(health)})
}

func (s *Server) store(w http.ResponseWriter, r *http.Request) {
	var req storageapi.StoreMessageRequest
	if err := s.decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	msg := req.ToMessage()
	if err := s.storage.Store(r.Context(), msg); err != nil {
		if errors.Is(err, ErrAlreadyExists) {
			s.writeError(w, http.StatusConflict, err.Error())
			return
		}

		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusCreated, storageapi.MessageResponseFromMessage(msg))
}

func (s *Server) getByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	msg, err := s.storage.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			s.writeError(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, ErrNotInitialized) {
			s.writeError(w, http.StatusServiceUnavailable, err.Error())
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, storageapi.MessageResponseFromMessage(msg))
}

func (s *Server) count(w http.ResponseWriter, r *http.Request) {
	n, err := s.storage.Count(r.Context())
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, storageapi.CountResponse{Count: n})
}

// === Helpers ===

func (s *Server) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) writeError(w http.ResponseWriter, status int, msg string) {
	s.writeJSON(w, status, storageapi.ErrorResponse{Error: msg})
}

func (s *Server) decodeJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}
