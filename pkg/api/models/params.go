package models

type SearchParams struct {
	Query      string    `json:"query"`
	Systems    *[]string `json:"systems"`
	MaxResults *int      `json:"maxResults"`
}

type MediaIndexParams struct {
	Systems *[]string `json:"systems"`
}

// TODO: not everything should be required
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

type NewClientParams struct {
	Name string `json:"name"`
}

type DeleteClientParams struct {
	Id string `json:"id"`
}
