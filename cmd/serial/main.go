package main

import (
	"strings"

	mrextConfig "github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/input"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/service"
	"go.bug.st/serial"
)

var logger = service.NewLogger("tapto")

func main() {
	ports, err := serial.GetPortsList()
	if err != nil {
		panic(err)
	}

	for _, port := range ports {
		println(port)
	}

	kbd, err := input.NewKeyboard()
	if err != nil {
		panic(err)
	}

	state := ServiceState{
		uidMap:  make(map[string]string),
		textMap: make(map[string]string),
	}
	loadDatabase(&state)

	port, err := serial.Open("/dev/ttyUSB0", &serial.Mode{
		BaudRate: 115200,
	})
	if err != nil {
		panic(err)
	}

	// read buffer until a newline and print it in a loop
	sBuf := make([]byte, 128)
	lBuf := make([]byte, 0)
	uid := ""
	for {
		n, err := port.Read(sBuf)
		if err != nil {
			panic(err)
		}

		if strings.Contains(string(sBuf[:n]), "\n") {
			lBuf = append(lBuf, sBuf[:n]...)
			line := strings.TrimSpace(string(lBuf))

			println(line)

			if strings.HasPrefix(line, "**read:") {
				value := strings.TrimPrefix(line, "**read:")
				parts := strings.SplitN(value, ",", 2)
				if parts[0] != uid {
					uid = parts[0]
					println("UID:", uid)
					mister.LaunchToken(&mrextConfig.UserConfig{}, true, kbd, "/media/fat/games/N64/1 US - N-Z/Super Mario 64 (USA).z64")
				}
			}

			lBuf = make([]byte, 0)
		} else {
			lBuf = append(lBuf, sBuf[:n]...)
		}
	}
}
