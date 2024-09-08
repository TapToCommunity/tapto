package models

const (
	ReaderChanged        = "state.readerChanged"
	ReaderRemoved        = "state.readerRemoved"
	LaunchingState       = "state.launching"
	ActiveCardState      = "state.activeCard"
	SystemStopped        = "system.stopped" // TODO: REMOVE
	SystemStarted        = "system.started" // TODO: REMOVE
	MediaStopped         = "media.stopped"
	MediaStarted         = "media.started"
	MediaIndexing        = "media.indexing"
	MethodLaunch         = "launch"
	MethodStop           = "stop"
	MethodMediaIndex     = "media.index"
	MethodMediaSearch    = "media.search"
	MethodSettings       = "settings"
	MethodSettingsUpdate = "settings.update"
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
	Method string `json:"method"`
	Params any    `json:"params,omitempty"`
}
