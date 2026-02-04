package http

import (
	"encoding/json"
	"net/http"
)

// healthCheck godoc
// @Summary		Проверка здоровья сервиса
// @Description	Возвращает статус работоспособности gateway
// @Tags		Health
// @Produce		json
// @Success		200	{object}	map[string]string	"Сервис работает"
// @Router		/api/v1/health [get]
func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// === Helpers ===

func (s *Server) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) writeError(w http.ResponseWriter, status int, msg string) {
	s.writeJSON(w, status, ErrorResponse{Error: msg})
}

func (s *Server) decodeJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// getGatewayConfig godoc
// @Summary		Получить конфигурацию gateway
// @Description	Возвращает текущую конфигурацию gateway сервиса
// @Tags		Config
// @Produce		json
// @Success		200	{object}	gateway.GatewayConfig	"Конфигурация gateway"
// @Failure		500	{object}	ErrorResponse			"Внутренняя ошибка сервера"
// @Router		/api/v1/config [get]
func (s *Server) getGatewayConfig(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, s.gateway.GetConfig())
}

// consumeMessage godoc
// @Summary		Отправить сообщение
// @Description	Принимает сообщение для последующей доставки через storage ноды
// @Tags		Messages
// @Accept		json
// @Produce		json
// @Param		request	body		NewMessageRequest	true	"Данные сообщения"
// @Success		201		{object}	NewMessageResponse	"Сообщение успешно создано"
// @Failure		400		{object}	ErrorResponse		"Некорректный запрос"
// @Failure		500		{object}	ErrorResponse		"Внутренняя ошибка сервера"
// @Router		/api/v1/message [post]
func (s *Server) consumeMessage(w http.ResponseWriter, r *http.Request) {
	var req NewMessageRequest
	var err error

	if err = s.decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	msg := req.ToMessage()

	if err = s.gateway.Consume(msg); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	s.writeJSON(w, http.StatusCreated, NewMessageResponseFromMessage(msg))
}
