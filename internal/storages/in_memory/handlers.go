package inmemory

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	storageapi "github.com/Alexey-zaliznuak/orbital/pkg/storage/api"
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
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusCreated, storageapi.MessageResponseFromMessage(msg))
}

func (s *Server) fetchReady(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			s.writeError(w, http.StatusBadRequest, "invalid limit parameter")
			return
		}
		limit = parsed
	}

	msgs, err := s.storage.FetchReady(r.Context(), limit)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := make([]storageapi.MessageResponse, len(msgs))
	for i, msg := range msgs {
		response[i] = storageapi.MessageResponseFromMessage(msg)
	}
	s.writeJSON(w, http.StatusOK, storageapi.FetchReadyResponse{Messages: response})
}

func (s *Server) acknowledge(w http.ResponseWriter, r *http.Request) {
	var req storageapi.AcknowledgeRequest
	if err := s.decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := s.storage.Acknowledge(r.Context(), req.IDs); err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) getByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	msg, err := s.storage.GetByID(r.Context(), id)
	if err != nil {
		s.writeError(w, http.StatusNotFound, "message not found")
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
