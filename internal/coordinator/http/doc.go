// Package http provides HTTP API for the Coordinator service.
//
//	@title			Orbital Coordinator API
//	@version		1.0
//	@description	API для управления кластером Orbital: регистрация компонентов, правила маршрутизации, конфигурация.
//
//	@host		localhost:8080
//	@BasePath	/api/v1
//
//	@tag.name			Nodes
//	@tag.description	Управление нодами координатора
//
//	@tag.name			Gateways
//	@tag.description	Регистрация и управление Gateway инстансами
//
//	@tag.name			Storages
//	@tag.description	Регистрация и управление Storage инстансами
//
//	@tag.name			Pushers
//	@tag.description	Регистрация и управление Pusher инстансами
//
//	@tag.name			RoutingRules
//	@tag.description	Управление правилами маршрутизации
//
//	@tag.name			Config
//	@tag.description	Конфигурации кластера и координатора
package http
