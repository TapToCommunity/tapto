package daemon

import (
	"fmt"
	"net"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/platforms/mister"
)

// TODO: i want to move away from the socket server entirely, it won't work on
//       windows and can be entirely replaced with an http server which will
//       also allow for remote access

func StartSocketServer(state *State) (net.Listener, error) {
	socket, err := net.Listen("unix", mister.SocketFile)
	if err != nil {
		log.Error().Msgf("error creating socket: %s", err)
		return nil, err
	}

	go func() {
		for {
			if state.ShouldStopService() {
				break
			}

			conn, err := socket.Accept()
			if err != nil {
				log.Error().Msgf("error accepting connection: %s", err)
				return
			}

			go func(conn net.Conn) {
				log.Debug().Msg("new socket connection")

				defer func(conn net.Conn) {
					err := conn.Close()
					if err != nil {
						log.Warn().Msgf("error closing connection: %s", err)
					}
				}(conn)

				buf := make([]byte, 4096)

				n, err := conn.Read(buf)
				if err != nil {
					log.Error().Msgf("error reading from connection: %s", err)
					return
				}

				if n == 0 {
					return
				}
				log.Debug().Msgf("received %d bytes", n)

				payload := ""

				switch strings.TrimSpace(string(buf[:n])) {
				case "status":
					lastScanned := state.GetLastScanned()
					if lastScanned.UID != "" {
						payload = fmt.Sprintf(
							"%d,%s,%t,%s",
							lastScanned.ScanTime.Unix(),
							lastScanned.UID,
							!state.IsLauncherDisabled(),
							lastScanned.Text,
						)
					} else {
						payload = fmt.Sprintf("0,,%t,", !state.IsLauncherDisabled())
					}
				case "disable":
					state.DisableLauncher()
					log.Info().Msg("launcher disabled")
				case "enable":
					state.EnableLauncher()
					log.Info().Msg("launcher enabled")
				default:
					log.Warn().Msgf("unknown command: %s", strings.TrimRight(string(buf[:n]), "\n"))
				}

				_, err = conn.Write([]byte(payload))
				if err != nil {
					log.Error().Msgf("error writing to socket: %s", err)
					return
				}
			}(conn)
		}
	}()

	return socket, nil
}
