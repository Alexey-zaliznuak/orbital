package http

import (
	"encoding/json"
	"net/http"
)

// healthCheck godoc
// @Summary		Health check
// @Description	Проверка работоспособности сервиса
// @Tags		Health
// @Produce		json
// @Success		200	{object}	map[string]string
// @Router		/health [get]
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

// getCoordinatorConfig godoc
// @Summary		Получить конфигурацию
// @Description	Возвращает текущую конфигурацию координатора
// @Tags		coordinator.CoordinatorConfig
// @Produce		json
// @Success		200	{object}	coordinator.CoordinatorConfig
// @Failure		500	{object}	ErrorResponse
// @Router		/coordinator-config [get]
func (s *Server) getGatewayConfig(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, s.gateway.GetConfig())
}

func (s *Server) consumeMessage(w http.ResponseWriter, r *http.Request) {
	r.Body.Read()
	// err := s.gateway.Consume(message)
	s.writeJSON(w, http.StatusOK, s.gateway.GetConfig())
}
