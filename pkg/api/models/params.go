package models

type SearchParams struct {
	Query      string    `json:"query"`
	Systems    *[]string `json:"systems"`
	MaxResults *int      `json:"maxResults"`
}

type MediaIndexParams struct {
	Systems *[]string `json:"systems"`
}

type LaunchParams struct {
	Type *string `json:"type"`
	UID  *string `json:"uid"`
	Text *string `json:"text"`
	Data *string `json:"data"`
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
	Id int `json:"id"`
}

type UpdateMappingParams struct {
	Id       int     `json:"id"`
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
	AudioFeedback   *bool `json:"audioFeedback"`
	DebugLogging    *bool `json:"debugLogging"`
	LaunchingActive *bool `json:"launchingActive"`
}

type UpdateSettingsReadersParams struct {
	AutoDetect *bool `json:"autoDetect"`
}

type UpdateSettingsReadersScanParams struct {
	Mode         *string   `json:"mode"`
	ExitDelay    *float32  `json:"exitDelay"`
	IgnoreSystem *[]string `json:"ignoreSystem"`
}

type NewClientParams struct {
	Name string `json:"name"`
}

type DeleteClientParams struct {
	Id string `json:"id"`
}
