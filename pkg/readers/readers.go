package readers

import (
	"github.com/wizzomafizzo/tapto/pkg/tokens"
)

type Reader interface {
	// Open any necessary connections to the device and start polling.
	// Takes a device connection string.
	Open(string) error
	// Close any open connections to the device and stop polling.
	Close() error
	// Detect attempts to search for a connected device and returns the device
	// connection string. If no device is found, an empty string is returned.
	// Takes a list of currently connected device strings.
	Detect([]string) string
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
