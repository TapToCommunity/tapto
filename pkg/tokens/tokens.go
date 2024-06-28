package tokens

import "time"

type Token struct {
	Type     string
	UID      string
	Text     string
	Data     string
	ScanTime time.Time
	FromApi  bool
}

const (
	TypeNTAG           = "NTAG"
	TypeMifare         = "MIFARE"
	TypeAmiibo         = "Amiibo"
	TypeLegoDimensions = "LegoDimensions"
)
