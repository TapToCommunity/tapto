package models

const (
	ReaderChanged   = "state.readerChanged"
	ReaderRemoved   = "state.readerRemoved"
	LaunchingState  = "state.launching"
	ActiveCardState = "state.activeCard"
	SystemStopped   = "system.stopped" // TODO: REMOVE
	SystemStarted   = "system.started" // TODO: REMOVE
	MediaStopped    = "media.stopped"
	MediaStarted    = "media.started"
	MediaIndexing   = "media.indexing"
)

type Notification struct {
	Method string `json:"method"`
	Params any    `json:"params,omitempty"`
}
