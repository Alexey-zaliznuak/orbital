package http

import (
	"time"

	"github.com/Alexey-zaliznuak/orbital/pkg/entities/coordinator"
	routingrule "github.com/Alexey-zaliznuak/orbital/pkg/entities/routing_rule"
)

// === Common ===

type ErrorResponse struct {
	Error string `json:"error"`
}

// === Nodes ===

type CreateNodeRequest struct {
	Address string `json:"address"`
}

type NodeResponse struct {
	ID            string `json:"id"`
	Address       string `json:"address"`
	Status        string `json:"status"`
	RegisteredAt  string `json:"registered_at"`
	LastHeartbeat string `json:"last_heartbeat"`
}

func nodeToResponse(n *coordinator.Node) NodeResponse {
	return NodeResponse{
		ID:            n.ID().String(),
		Address:       n.Address(),
		Status:        n.Status().String(),
		RegisteredAt:  n.RegisteredAt().Format(time.RFC3339),
		LastHeartbeat: n.LastHeartbeat().Format(time.RFC3339),
	}
}

// === Gateways ===

type RegisterGatewayRequest struct {
	ID      string `json:"id"`
	Address string `json:"address"`
}

type GatewayResponse struct {
	ID            string `json:"id"`
	Address       string `json:"address"`
	Status        string `json:"status"`
	RegisteredAt  string `json:"registered_at"`
	LastHeartbeat string `json:"last_heartbeat"`
}

func gatewayToResponse(g *coordinator.GatewayInfo) GatewayResponse {
	return GatewayResponse{
		ID:            g.ID,
		Address:       g.Address,
		Status:        g.Status.String(),
		RegisteredAt:  g.RegisteredAt.Format(time.RFC3339),
		LastHeartbeat: g.LastHeartbeat.Format(time.RFC3339),
	}
}

// === Storages ===

type RegisterStorageRequest struct {
	ID       string `json:"id"`
	Address  string `json:"address"`
	MinDelay string `json:"min_delay"` // e.g. "0s", "1m", "1h"
	MaxDelay string `json:"max_delay"` // e.g. "1m", "1h", "0" (unlimited)
}

type StorageResponse struct {
	ID            string `json:"id"`
	Address       string `json:"address"`
	MinDelay      string `json:"min_delay"`
	MaxDelay      string `json:"max_delay"`
	Status        string `json:"status"`
	RegisteredAt  string `json:"registered_at"`
	LastHeartbeat string `json:"last_heartbeat"`
}

func storageToResponse(s *coordinator.StorageInfo) StorageResponse {
	maxDelay := s.MaxDelay.String()
	if s.MaxDelay == 0 {
		maxDelay = "unlimited"
	}
	return StorageResponse{
		ID:            s.ID,
		Address:       s.Address,
		MinDelay:      s.MinDelay.String(),
		MaxDelay:      maxDelay,
		Status:        s.Status.String(),
		RegisteredAt:  s.RegisteredAt.Format(time.RFC3339),
		LastHeartbeat: s.LastHeartbeat.Format(time.RFC3339),
	}
}

// === Pushers ===

type RegisterPusherRequest struct {
	ID      string `json:"id"`
	Type    string `json:"type"` // "http", "kafka", "grpc", "nats"
	Address string `json:"address"`
}

type PusherResponse struct {
	ID            string `json:"id"`
	Type          string `json:"type"`
	Address       string `json:"address"`
	Status        string `json:"status"`
	RegisteredAt  string `json:"registered_at"`
	LastHeartbeat string `json:"last_heartbeat"`
}

func pusherToResponse(p *coordinator.PusherInfo) PusherResponse {
	return PusherResponse{
		ID:            p.ID,
		Type:          p.Type,
		Address:       p.Address,
		Status:        p.Status.String(),
		RegisteredAt:  p.RegisteredAt.Format(time.RFC3339),
		LastHeartbeat: p.LastHeartbeat.Format(time.RFC3339),
	}
}

// === Routing Rules ===

type CreateRoutingRuleRequest struct {
	ID        string `json:"id"`
	Pattern   string `json:"pattern"`
	MatchType int    `json:"match_type"` // 0=Exact, 1=Prefix, 2=Suffix, 3=Regex
	PusherID  string `json:"pusher_id"`
	Enabled   bool   `json:"enabled"`
}

type RoutingRuleResponse struct {
	ID        string `json:"id"`
	Pattern   string `json:"pattern"`
	MatchType int    `json:"match_type"`
	PusherID  string `json:"pusher_id"`
	Enabled   bool   `json:"enabled"`
}

func routingRuleToResponse(r *routingrule.RoutingRule) RoutingRuleResponse {
	return RoutingRuleResponse{
		ID:        r.ID,
		Pattern:   r.Pattern,
		MatchType: int(r.MatchType),
		PusherID:  r.PusherID,
		Enabled:   r.Enabled,
	}
}
