package inmemory

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Alexey-zaliznuak/orbital/pkg/entities/message"
	storageapi "github.com/Alexey-zaliznuak/orbital/pkg/sdk/storage/api"
)

// healthCheck godoc
//
//	@Summary		Health check
//	@Description	Проверка работоспособности хранилища
//	@Tags			Health
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Failure		500	{object}	storageapi.ErrorResponse
//	@Router			/health [get]
func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	health, err := s.storage.HealthCheck(r.Context())
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": string(health)})
}

// store godoc
//
//	@Summary		Сохранить сообщение
//	@Description	Сохраняет новое сообщение в in-memory хранилище
//	@Tags			Messages
//	@Accept			json
//	@Produce		json
//	@Param			request	body		storageapi.StoreMessageRequest	true	"Данные сообщения"
//	@Success		201		{object}	storageapi.MessageResponse
//	@Failure		400		{object}	storageapi.ErrorResponse
//	@Failure		409		{object}	storageapi.ErrorResponse	"Сообщение уже существует"
//	@Failure		500		{object}	storageapi.ErrorResponse
//	@Router			/messages [post]
func (s *Server) store(w http.ResponseWriter, r *http.Request) {
	var req storageapi.StoreMessageRequest
	if err := s.decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	msg := req.ToMessage()
	if err := s.storage.Store(r.Context(), []*message.Message{msg}); err != nil {
		if errors.Is(err, ErrAlreadyExists) {
			s.writeError(w, http.StatusConflict, err.Error())
			return
		}

		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusCreated, storageapi.MessageResponseFromMessage(msg))
}

// getByID godoc
//
//	@Summary		Получить сообщение по ID
//	@Description	Возвращает сообщение из хранилища по его идентификатору
//	@Tags			Messages
//	@Produce		json
//	@Param			id	path		string	true	"Идентификатор сообщения"
//	@Success		200	{object}	storageapi.MessageResponse
//	@Failure		404	{object}	storageapi.ErrorResponse	"Сообщение не найдено"
//	@Failure		503	{object}	storageapi.ErrorResponse	"Хранилище не инициализировано"
//	@Failure		500	{object}	storageapi.ErrorResponse
//	@Router			/messages/{id} [get]
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

// count godoc
//
//	@Summary		Количество сообщений
//	@Description	Возвращает общее количество сообщений в хранилище
//	@Tags			Messages
//	@Produce		json
//	@Success		200	{object}	storageapi.CountResponse
//	@Failure		500	{object}	storageapi.ErrorResponse
//	@Router			/messages/count [get]
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
