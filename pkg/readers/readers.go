package readers

import (
	"github.com/wizzomafizzo/tapto/pkg/tokens"
)

type Reader interface {
	// Open any necessary connections to the device and start polling.
	Open(string) error
	// Close any open connections to the device and stop polling.
	Close() error
	// Device returns the device connection string.
	Device() string
	// Connected returns true if the device is connected and active.
	Connected() bool
	// Info returns a string with information about the connected device.
	Info() string
	// History returns a list of tokens that have been read by the device.
	History() []tokens.Token
	// Read returns the active token being read by the device. Non-blocking.
	Read() (*tokens.Token, error)
	// Write sends a string to the device, if supported. Blocking.
	Write(string) error
}
