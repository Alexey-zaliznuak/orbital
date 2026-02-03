package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Alexey-zaliznuak/orbital/internal/coordinator/storage/etcd"
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/coordinator"
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/gateway"
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/node"
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/pusher"
	routingrule "github.com/Alexey-zaliznuak/orbital/pkg/entities/routing_rule"
	"github.com/Alexey-zaliznuak/orbital/pkg/entities/storage"
)

// === Health ===

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

// === Nodes ===

// createNode godoc
// @Summary		Создать ноду координатора
// @Description	Регистрирует новую ноду координатора в кластере
// @Tags		Nodes
// @Accept		json
// @Produce		json
// @Param		request	body		CreateNodeRequest	true	"Данные ноды"
// @Success		201		{object}	NodeResponse
// @Failure		400		{object}	ErrorResponse
// @Failure		409		{object}	ErrorResponse	"Нода уже существует"
// @Failure		500		{object}	ErrorResponse
// @Router		/nodes [post]
func (s *Server) createNode(w http.ResponseWriter, r *http.Request) {
	var req CreateNodeRequest
	if err := s.decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Address == "" {
		s.writeError(w, http.StatusBadRequest, "address is required")
		return
	}

	node := coordinator.NewNode(uuid.New(), req.Address)

	if err := s.coordinator.GetStorage().CreateNode(r.Context(), node); err != nil {
		if errors.Is(err, etcd.ErrAlreadyExists) {
			s.writeError(w, http.StatusConflict, "node already exists")
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusCreated, nodeToResponse(node))
}

// getNode godoc
// @Summary		Получить ноду по ID
// @Description	Возвращает информацию о ноде координатора
// @Tags		Nodes
// @Produce		json
// @Param		nodeID	path		string	true	"ID ноды (UUID)"
// @Success		200		{object}	NodeResponse
// @Failure		400		{object}	ErrorResponse	"Невалидный ID"
// @Failure		404		{object}	ErrorResponse	"Нода не найдена"
// @Failure		500		{object}	ErrorResponse
// @Router		/nodes/{nodeID} [get]
func (s *Server) getNode(w http.ResponseWriter, r *http.Request) {
	nodeIDStr := chi.URLParam(r, "nodeID")
	nodeID, err := uuid.Parse(nodeIDStr)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid node ID")
		return
	}

	node, err := s.coordinator.GetStorage().GetNode(r.Context(), nodeID)
	if err != nil {
		if errors.Is(err, etcd.ErrNotFound) {
			s.writeError(w, http.StatusNotFound, "node not found")
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, nodeToResponse(node))
}

// listNodes godoc
// @Summary		Список нод
// @Description	Возвращает список всех нод координатора
// @Tags		Nodes
// @Produce		json
// @Success		200	{array}		NodeResponse
// @Failure		500	{object}	ErrorResponse
// @Router		/nodes [get]
func (s *Server) listNodes(w http.ResponseWriter, r *http.Request) {
	nodes, err := s.coordinator.GetStorage().ListNodes(r.Context())
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := make([]NodeResponse, len(nodes))
	for i, n := range nodes {
		resp[i] = nodeToResponse(n)
	}

	s.writeJSON(w, http.StatusOK, resp)
}

// updateNodeHeartbeat godoc
// @Summary		Обновить heartbeat ноды
// @Description	Обновляет время последнего heartbeat ноды
// @Tags		Nodes
// @Param		nodeID	path	string	true	"ID ноды (UUID)"
// @Success		204		"No Content"
// @Failure		400		{object}	ErrorResponse	"Невалидный ID"
// @Failure		404		{object}	ErrorResponse	"Нода не найдена"
// @Failure		500		{object}	ErrorResponse
// @Router		/nodes/{nodeID}/heartbeat [put]
func (s *Server) updateNodeHeartbeat(w http.ResponseWriter, r *http.Request) {
	nodeIDStr := chi.URLParam(r, "nodeID")
	nodeID, err := uuid.Parse(nodeIDStr)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid node ID")
		return
	}

	if err := s.coordinator.GetStorage().UpdateNodeHeartbeat(r.Context(), nodeID); err != nil {
		if errors.Is(err, etcd.ErrNotFound) {
			s.writeError(w, http.StatusNotFound, "node not found")
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// deleteNode godoc
// @Summary		Удалить ноду
// @Description	Удаляет ноду координатора из кластера
// @Tags		Nodes
// @Param		nodeID	path	string	true	"ID ноды (UUID)"
// @Success		204		"No Content"
// @Failure		400		{object}	ErrorResponse	"Невалидный ID"
// @Failure		404		{object}	ErrorResponse	"Нода не найдена"
// @Failure		500		{object}	ErrorResponse
// @Router		/nodes/{nodeID} [delete]
func (s *Server) deleteNode(w http.ResponseWriter, r *http.Request) {
	nodeIDStr := chi.URLParam(r, "nodeID")
	nodeID, err := uuid.Parse(nodeIDStr)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid node ID")
		return
	}

	if err := s.coordinator.GetStorage().DeleteNode(r.Context(), nodeID); err != nil {
		if errors.Is(err, etcd.ErrNotFound) {
			s.writeError(w, http.StatusNotFound, "node not found")
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// === Gateways ===

// registerGateway godoc
// @Summary		Зарегистрировать Gateway
// @Description	Регистрирует новый Gateway инстанс
// @Tags		Gateways
// @Accept		json
// @Produce		json
// @Param		request	body		RegisterGatewayRequest	true	"Данные Gateway"
// @Success		201		{object}	GatewayResponse
// @Failure		400		{object}	ErrorResponse
// @Failure		409		{object}	ErrorResponse	"Gateway уже существует"
// @Failure		500		{object}	ErrorResponse
// @Router		/gateways [post]
func (s *Server) registerGateway(w http.ResponseWriter, r *http.Request) {
	var req RegisterGatewayRequest
	if err := s.decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ID == "" || req.Address == "" {
		s.writeError(w, http.StatusBadRequest, "id and address are required")
		return
	}

	now := time.Now()
	gw := &gateway.Info{
		ID:            req.ID,
		Address:       req.Address,
		Status:        node.NodeStatusActive,
		RegisteredAt:  now,
		LastHeartbeat: now,
	}

	if err := s.coordinator.GetStorage().RegisterGateway(r.Context(), gw); err != nil {
		if errors.Is(err, etcd.ErrAlreadyExists) {
			s.writeError(w, http.StatusConflict, "gateway already exists")
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusCreated, gatewayToResponse(gw))
}

// getGateway godoc
// @Summary		Получить Gateway по ID
// @Description	Возвращает информацию о Gateway
// @Tags		Gateways
// @Produce		json
// @Param		gatewayID	path		string	true	"ID Gateway"
// @Success		200			{object}	GatewayResponse
// @Failure		404			{object}	ErrorResponse	"Gateway не найден"
// @Failure		500			{object}	ErrorResponse
// @Router		/gateways/{gatewayID} [get]
func (s *Server) getGateway(w http.ResponseWriter, r *http.Request) {
	gatewayID := chi.URLParam(r, "gatewayID")

	gateway, err := s.coordinator.GetStorage().GetGateway(r.Context(), gatewayID)
	if err != nil {
		if errors.Is(err, etcd.ErrNotFound) {
			s.writeError(w, http.StatusNotFound, "gateway not found")
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, gatewayToResponse(gateway))
}

// listGateways godoc
// @Summary		Список Gateways
// @Description	Возвращает список всех зарегистрированных Gateways
// @Tags		Gateways
// @Produce		json
// @Success		200	{array}		GatewayResponse
// @Failure		500	{object}	ErrorResponse
// @Router		/gateways [get]
func (s *Server) listGateways(w http.ResponseWriter, r *http.Request) {
	gateways, err := s.coordinator.GetStorage().ListGateways(r.Context())
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := make([]GatewayResponse, len(gateways))
	for i, g := range gateways {
		resp[i] = gatewayToResponse(g)
	}

	s.writeJSON(w, http.StatusOK, resp)
}

// updateGatewayHeartbeat godoc
// @Summary		Обновить heartbeat Gateway
// @Description	Обновляет время последнего heartbeat Gateway
// @Tags		Gateways
// @Param		gatewayID	path	string	true	"ID Gateway"
// @Success		204			"No Content"
// @Failure		404			{object}	ErrorResponse	"Gateway не найден"
// @Failure		500			{object}	ErrorResponse
// @Router		/gateways/{gatewayID}/heartbeat [put]
func (s *Server) updateGatewayHeartbeat(w http.ResponseWriter, r *http.Request) {
	gatewayID := chi.URLParam(r, "gatewayID")

	if err := s.coordinator.GetStorage().UpdateGatewayHeartbeat(r.Context(), gatewayID); err != nil {
		if errors.Is(err, etcd.ErrNotFound) {
			s.writeError(w, http.StatusNotFound, "gateway not found")
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// unregisterGateway godoc
// @Summary		Удалить Gateway
// @Description	Удаляет Gateway из кластера
// @Tags		Gateways
// @Param		gatewayID	path	string	true	"ID Gateway"
// @Success		204			"No Content"
// @Failure		404			{object}	ErrorResponse	"Gateway не найден"
// @Failure		500			{object}	ErrorResponse
// @Router		/gateways/{gatewayID} [delete]
func (s *Server) unregisterGateway(w http.ResponseWriter, r *http.Request) {
	gatewayID := chi.URLParam(r, "gatewayID")

	if err := s.coordinator.GetStorage().UnregisterGateway(r.Context(), gatewayID); err != nil {
		if errors.Is(err, etcd.ErrNotFound) {
			s.writeError(w, http.StatusNotFound, "gateway not found")
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// === Storages ===

// registerStorage godoc
// @Summary		Зарегистрировать Storage
// @Description	Регистрирует новый Storage инстанс с диапазоном задержек
// @Tags		Storages
// @Accept		json
// @Produce		json
// @Param		request	body		RegisterStorageRequest	true	"Данные Storage"
// @Success		201		{object}	StorageResponse
// @Failure		400		{object}	ErrorResponse
// @Failure		409		{object}	ErrorResponse	"Storage уже существует"
// @Failure		500		{object}	ErrorResponse
// @Router		/storages [post]
func (s *Server) registerStorage(w http.ResponseWriter, r *http.Request) {
	var req RegisterStorageRequest
	if err := s.decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ID == "" || req.Address == "" {
		s.writeError(w, http.StatusBadRequest, "id and address are required")
		return
	}

	minDelay, err := time.ParseDuration(req.MinDelay)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid min_delay format")
		return
	}

	var maxDelay time.Duration
	if req.MaxDelay != "" && req.MaxDelay != "0" {
		maxDelay, err = time.ParseDuration(req.MaxDelay)
		if err != nil {
			s.writeError(w, http.StatusBadRequest, "invalid max_delay format")
			return
		}
	}

	now := time.Now()
	st := &storage.Info{
		ID:            req.ID,
		Address:       req.Address,
		MinDelay:      minDelay,
		MaxDelay:      maxDelay,
		Status:        node.NodeStatusActive,
		RegisteredAt:  now,
		LastHeartbeat: now,
	}

	if err := s.coordinator.GetStorage().RegisterStorage(r.Context(), st); err != nil {
		if errors.Is(err, etcd.ErrAlreadyExists) {
			s.writeError(w, http.StatusConflict, "storage already exists")
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusCreated, storageToResponse(st))
}

// getStorage godoc
// @Summary		Получить Storage по ID
// @Description	Возвращает информацию о Storage
// @Tags		Storages
// @Produce		json
// @Param		storageID	path		string	true	"ID Storage"
// @Success		200			{object}	StorageResponse
// @Failure		404			{object}	ErrorResponse	"Storage не найден"
// @Failure		500			{object}	ErrorResponse
// @Router		/storages/{storageID} [get]
func (s *Server) getStorage(w http.ResponseWriter, r *http.Request) {
	storageID := chi.URLParam(r, "storageID")

	storage, err := s.coordinator.GetStorage().GetStorage(r.Context(), storageID)
	if err != nil {
		if errors.Is(err, etcd.ErrNotFound) {
			s.writeError(w, http.StatusNotFound, "storage not found")
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, storageToResponse(storage))
}

// listStorages godoc
// @Summary		Список Storages
// @Description	Возвращает список всех зарегистрированных Storages
// @Tags		Storages
// @Produce		json
// @Success		200	{array}		StorageResponse
// @Failure		500	{object}	ErrorResponse
// @Router		/storages [get]
func (s *Server) listStorages(w http.ResponseWriter, r *http.Request) {
	storages, err := s.coordinator.GetStorage().ListStorages(r.Context())
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := make([]StorageResponse, len(storages))
	for i, st := range storages {
		resp[i] = storageToResponse(st)
	}

	s.writeJSON(w, http.StatusOK, resp)
}

// updateStorageHeartbeat godoc
// @Summary		Обновить heartbeat Storage
// @Description	Обновляет время последнего heartbeat Storage
// @Tags		Storages
// @Param		storageID	path	string	true	"ID Storage"
// @Success		204			"No Content"
// @Failure		404			{object}	ErrorResponse	"Storage не найден"
// @Failure		500			{object}	ErrorResponse
// @Router		/storages/{storageID}/heartbeat [put]
func (s *Server) updateStorageHeartbeat(w http.ResponseWriter, r *http.Request) {
	storageID := chi.URLParam(r, "storageID")

	if err := s.coordinator.GetStorage().UpdateStorageHeartbeat(r.Context(), storageID); err != nil {
		if errors.Is(err, etcd.ErrNotFound) {
			s.writeError(w, http.StatusNotFound, "storage not found")
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// unregisterStorage godoc
// @Summary		Удалить Storage
// @Description	Удаляет Storage из кластера
// @Tags		Storages
// @Param		storageID	path	string	true	"ID Storage"
// @Success		204			"No Content"
// @Failure		404			{object}	ErrorResponse	"Storage не найден"
// @Failure		500			{object}	ErrorResponse
// @Router		/storages/{storageID} [delete]
func (s *Server) unregisterStorage(w http.ResponseWriter, r *http.Request) {
	storageID := chi.URLParam(r, "storageID")

	if err := s.coordinator.GetStorage().UnregisterStorage(r.Context(), storageID); err != nil {
		if errors.Is(err, etcd.ErrNotFound) {
			s.writeError(w, http.StatusNotFound, "storage not found")
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// === Pushers ===

// registerPusher godoc
// @Summary		Зарегистрировать Pusher
// @Description	Регистрирует новый Pusher инстанс
// @Tags		Pushers
// @Accept		json
// @Produce		json
// @Param		request	body		RegisterPusherRequest	true	"Данные Pusher"
// @Success		201		{object}	PusherResponse
// @Failure		400		{object}	ErrorResponse
// @Failure		409		{object}	ErrorResponse	"Pusher уже существует"
// @Failure		500		{object}	ErrorResponse
// @Router		/pushers [post]
func (s *Server) registerPusher(w http.ResponseWriter, r *http.Request) {
	var req RegisterPusherRequest
	if err := s.decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ID == "" || req.Type == "" || req.Address == "" {
		s.writeError(w, http.StatusBadRequest, "id, type and address are required")
		return
	}

	now := time.Now()
	p := &pusher.Info{
		ID:            req.ID,
		Type:          req.Type,
		Address:       req.Address,
		Status:        node.NodeStatusActive,
		RegisteredAt:  now,
		LastHeartbeat: now,
	}

	if err := s.coordinator.GetStorage().RegisterPusher(r.Context(), p); err != nil {
		if errors.Is(err, etcd.ErrAlreadyExists) {
			s.writeError(w, http.StatusConflict, "pusher already exists")
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusCreated, pusherToResponse(p))
}

// getPusher godoc
// @Summary		Получить Pusher по ID
// @Description	Возвращает информацию о Pusher
// @Tags		Pushers
// @Produce		json
// @Param		pusherID	path		string	true	"ID Pusher"
// @Success		200			{object}	PusherResponse
// @Failure		404			{object}	ErrorResponse	"Pusher не найден"
// @Failure		500			{object}	ErrorResponse
// @Router		/pushers/{pusherID} [get]
func (s *Server) getPusher(w http.ResponseWriter, r *http.Request) {
	pusherID := chi.URLParam(r, "pusherID")

	pusher, err := s.coordinator.GetStorage().GetPusher(r.Context(), pusherID)
	if err != nil {
		if errors.Is(err, etcd.ErrNotFound) {
			s.writeError(w, http.StatusNotFound, "pusher not found")
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, pusherToResponse(pusher))
}

// listPushers godoc
// @Summary		Список Pushers
// @Description	Возвращает список всех зарегистрированных Pushers
// @Tags		Pushers
// @Produce		json
// @Success		200	{array}		PusherResponse
// @Failure		500	{object}	ErrorResponse
// @Router		/pushers [get]
func (s *Server) listPushers(w http.ResponseWriter, r *http.Request) {
	pushers, err := s.coordinator.GetStorage().ListPushers(r.Context())
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := make([]PusherResponse, len(pushers))
	for i, p := range pushers {
		resp[i] = pusherToResponse(p)
	}

	s.writeJSON(w, http.StatusOK, resp)
}

// updatePusherHeartbeat godoc
// @Summary		Обновить heartbeat Pusher
// @Description	Обновляет время последнего heartbeat Pusher
// @Tags		Pushers
// @Param		pusherID	path	string	true	"ID Pusher"
// @Success		204			"No Content"
// @Failure		404			{object}	ErrorResponse	"Pusher не найден"
// @Failure		500			{object}	ErrorResponse
// @Router		/pushers/{pusherID}/heartbeat [put]
func (s *Server) updatePusherHeartbeat(w http.ResponseWriter, r *http.Request) {
	pusherID := chi.URLParam(r, "pusherID")

	if err := s.coordinator.GetStorage().UpdatePusherHeartbeat(r.Context(), pusherID); err != nil {
		if errors.Is(err, etcd.ErrNotFound) {
			s.writeError(w, http.StatusNotFound, "pusher not found")
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// unregisterPusher godoc
// @Summary		Удалить Pusher
// @Description	Удаляет Pusher из кластера
// @Tags		Pushers
// @Param		pusherID	path	string	true	"ID Pusher"
// @Success		204			"No Content"
// @Failure		404			{object}	ErrorResponse	"Pusher не найден"
// @Failure		500			{object}	ErrorResponse
// @Router		/pushers/{pusherID} [delete]
func (s *Server) unregisterPusher(w http.ResponseWriter, r *http.Request) {
	pusherID := chi.URLParam(r, "pusherID")

	if err := s.coordinator.GetStorage().UnregisterPusher(r.Context(), pusherID); err != nil {
		if errors.Is(err, etcd.ErrNotFound) {
			s.writeError(w, http.StatusNotFound, "pusher not found")
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// === Routing Rules ===

// createRoutingRule godoc
// @Summary		Создать правило маршрутизации
// @Description	Создаёт новое правило маршрутизации сообщений к Pusher
// @Tags		RoutingRules
// @Accept		json
// @Produce		json
// @Param		request	body		CreateRoutingRuleRequest	true	"Данные правила"
// @Success		201		{object}	RoutingRuleResponse
// @Failure		400		{object}	ErrorResponse
// @Failure		409		{object}	ErrorResponse	"Правило уже существует"
// @Failure		500		{object}	ErrorResponse
// @Router		/routing-rules [post]
func (s *Server) createRoutingRule(w http.ResponseWriter, r *http.Request) {
	var req CreateRoutingRuleRequest
	if err := s.decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ID == "" || req.Pattern == "" || req.PusherID == "" {
		s.writeError(w, http.StatusBadRequest, "id, pattern and pusher_id are required")
		return
	}

	rule := &routingrule.RoutingRule{
		ID:        req.ID,
		Pattern:   req.Pattern,
		MatchType: routingrule.MatchType(req.MatchType),
		PusherID:  req.PusherID,
		Enabled:   req.Enabled,
	}

	if err := s.coordinator.GetStorage().CreateRoutingRule(r.Context(), rule); err != nil {
		if errors.Is(err, etcd.ErrAlreadyExists) {
			s.writeError(w, http.StatusConflict, "routing rule already exists")
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusCreated, routingRuleToResponse(rule))
}

// getRoutingRule godoc
// @Summary		Получить правило по ID
// @Description	Возвращает информацию о правиле маршрутизации
// @Tags		RoutingRules
// @Produce		json
// @Param		ruleID	path		string	true	"ID правила"
// @Success		200		{object}	RoutingRuleResponse
// @Failure		404		{object}	ErrorResponse	"Правило не найдено"
// @Failure		500		{object}	ErrorResponse
// @Router		/routing-rules/{ruleID} [get]
func (s *Server) getRoutingRule(w http.ResponseWriter, r *http.Request) {
	ruleID := chi.URLParam(r, "ruleID")

	rule, err := s.coordinator.GetStorage().GetRoutingRule(r.Context(), ruleID)
	if err != nil {
		if errors.Is(err, etcd.ErrNotFound) {
			s.writeError(w, http.StatusNotFound, "routing rule not found")
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, routingRuleToResponse(rule))
}

// listRoutingRules godoc
// @Summary		Список правил маршрутизации
// @Description	Возвращает список всех правил маршрутизации
// @Tags		RoutingRules
// @Produce		json
// @Success		200	{array}		RoutingRuleResponse
// @Failure		500	{object}	ErrorResponse
// @Router		/routing-rules [get]
func (s *Server) listRoutingRules(w http.ResponseWriter, r *http.Request) {
	rules, err := s.coordinator.GetStorage().ListRoutingRules(r.Context())
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := make([]RoutingRuleResponse, len(rules))
	for i, rule := range rules {
		resp[i] = routingRuleToResponse(rule)
	}

	s.writeJSON(w, http.StatusOK, resp)
}

// updateRoutingRule godoc
// @Summary		Обновить правило маршрутизации
// @Description	Обновляет существующее правило маршрутизации
// @Tags		RoutingRules
// @Accept		json
// @Produce		json
// @Param		ruleID	path		string						true	"ID правила"
// @Param		request	body		CreateRoutingRuleRequest	true	"Новые данные правила"
// @Success		200		{object}	RoutingRuleResponse
// @Failure		400		{object}	ErrorResponse
// @Failure		404		{object}	ErrorResponse	"Правило не найдено"
// @Failure		500		{object}	ErrorResponse
// @Router		/routing-rules/{ruleID} [put]
func (s *Server) updateRoutingRule(w http.ResponseWriter, r *http.Request) {
	ruleID := chi.URLParam(r, "ruleID")

	var req CreateRoutingRuleRequest
	if err := s.decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	rule := &routingrule.RoutingRule{
		ID:        ruleID,
		Pattern:   req.Pattern,
		MatchType: routingrule.MatchType(req.MatchType),
		PusherID:  req.PusherID,
		Enabled:   req.Enabled,
	}

	if err := s.coordinator.GetStorage().UpdateRoutingRule(r.Context(), rule); err != nil {
		if errors.Is(err, etcd.ErrNotFound) {
			s.writeError(w, http.StatusNotFound, "routing rule not found")
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, routingRuleToResponse(rule))
}

// deleteRoutingRule godoc
// @Summary		Удалить правило маршрутизации
// @Description	Удаляет правило маршрутизации
// @Tags		RoutingRules
// @Param		ruleID	path	string	true	"ID правила"
// @Success		204		"No Content"
// @Failure		404		{object}	ErrorResponse	"Правило не найдено"
// @Failure		500		{object}	ErrorResponse
// @Router		/routing-rules/{ruleID} [delete]
func (s *Server) deleteRoutingRule(w http.ResponseWriter, r *http.Request) {
	ruleID := chi.URLParam(r, "ruleID")

	if err := s.coordinator.GetStorage().DeleteRoutingRule(r.Context(), ruleID); err != nil {
		if errors.Is(err, etcd.ErrNotFound) {
			s.writeError(w, http.StatusNotFound, "routing rule not found")
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// getClusterConfig godoc
// @Summary		Получить конфигурацию
// @Description	Возвращает текущую конфигурацию кластера
// @Tags		coordinator.ClusterConfig
// @Produce		json
// @Success		200	{object}	coordinator.ClusterConfig
// @Failure		500	{object}	ErrorResponse
// @Router		/cluster-config [get]
func (s *Server) getClusterConfig(w http.ResponseWriter, r *http.Request) {
	config := s.coordinator.GetClusterConfig()

	s.writeJSON(w, http.StatusOK, config)
}

// getCoordinatorConfig godoc
// @Summary		Получить конфигурацию
// @Description	Возвращает текущую конфигурацию координатора
// @Tags		coordinator.CoordinatorConfig
// @Produce		json
// @Success		200	{object}	coordinator.CoordinatorConfig
// @Failure		500	{object}	ErrorResponse
// @Router		/coordinator-config [get]
func (s *Server) getCoordinatorConfig(w http.ResponseWriter, r *http.Request) {
	config := s.coordinator.GetCoordinatorConfig()

	s.writeJSON(w, http.StatusOK, config)
}
