package models

import "github.com/google/uuid"

const (
	ReadersConnected     = "readers.connected"
	ReadersDisconnected  = "readers.disconnected"
	TokensLaunching      = "tokens.launching"
	TokensActive         = "tokens.active"
	MediaStopped         = "media.stopped"
	MediaStarted         = "media.started"
	MediaIndexing        = "media.indexing"
	MethodLaunch         = "launch"
	MethodStop           = "stop"
	MethodMediaIndex     = "media.index"
	MethodMediaSearch    = "media.search"
	MethodSettings       = "settings"
	MethodSettingsUpdate = "settings.update"
	MethodClients        = "clients"
	MethodClientsNew     = "clients.new"
	MethodClientsDelete  = "clients.delete"
	MethodSystems        = "systems"
	MethodHistory        = "history"
	MethodMappings       = "mappings"
	MethodMappingsNew    = "mappings.new"
	MethodMappingsDelete = "mappings.delete"
	MethodMappingsUpdate = "mappings.update"
	MethodReadersWrite   = "readers.write"
	MethodStatus         = "status"
	MethodVersion        = "version"
)

type Notification struct {
	Method string
	Params any
}

type RequestObject struct {
	JsonRpc string     `json:"jsonrpc"`
	Id      *uuid.UUID `json:"id,omitempty"`
	Method  string     `json:"method"`
	Params  any        `json:"params,omitempty"`
}

type ErrorObject struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ResponseObject struct {
	JsonRpc string       `json:"jsonrpc"`
	Id      uuid.UUID    `json:"id"`
	Result  any          `json:"result,omitempty"`
	Error   *ErrorObject `json:"error,omitempty"`
}

type ClientResponse struct {
	Id      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Address string    `json:"address"`
	Secret  string    `json:"secret"`
}

type MediaStartedParams struct {
	SystemId   string `json:"systemId"`
	SystemName string `json:"systemName"`
	MediaPath  string `json:"mediaPath"`
	MediaName  string `json:"mediaName"`
}
