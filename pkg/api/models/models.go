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

type SearchParams struct {
	Query      string `json:"query"`
	System     string `json:"system"`
	MaxResults *int   `json:"maxResults"`
}

type LaunchParams struct {
	Type string `json:"type"`
	UID  string `json:"uid"`
	Text string `json:"text"`
	Data string `json:"data"`
}

type AddMappingParams struct {
	Label    string `json:"label"`
	Enabled  bool   `json:"enabled"`
	Type     string `json:"type"`
	Match    string `json:"match"`
	Pattern  string `json:"pattern"`
	Override string `json:"override"`
}

type DeleteMappingParams struct {
	Id string `json:"id"`
}

type UpdateMappingParams struct {
	Id       string  `json:"id"`
	Label    *string `json:"label"`
	Enabled  *bool   `json:"enabled"`
	Type     *string `json:"type"`
	Match    *string `json:"match"`
	Pattern  *string `json:"pattern"`
	Override *string `json:"override"`
}

type ReaderWriteParams struct {
	Text string `json:"text"`
}

type UpdateSettingsParams struct {
	ConnectionString  *string   `json:"connectionString"`
	AllowCommands     *bool     `json:"allowCommands"`
	DisableSounds     *bool     `json:"disableSounds"`
	ProbeDevice       *bool     `json:"probeDevice"`
	ExitGame          *bool     `json:"exitGame"`
	ExitGameDelay     *int      `json:"exitGameDelay"`
	ExitGameBlocklist *[]string `json:"exitGameBlocklist"`
	Debug             *bool     `json:"debug"`
	Launching         *bool     `json:"launching"`
}
