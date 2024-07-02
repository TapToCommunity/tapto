package daemon

import (
	"fmt"
	"net"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/daemon/state"
	"github.com/wizzomafizzo/tapto/pkg/platforms/mister"
)

// TODO: i want to move away from the socket server entirely, it won't work on
//       windows and can be entirely replaced with an http server which will
//       also allow for remote access

const (
	ReaderTypePN532   = "PN532"
	ReaderTypeACR122U = "ACR122U"
	ReaderTypeUnknown = "Unknown"
)

func StartSocketServer(st *state.State) (net.Listener, error) {
	socket, err := net.Listen("unix", mister.SocketFile)
	if err != nil {
		log.Error().Msgf("error creating socket: %s", err)
		return nil, err
	}

	go func() {
		for {
			if st.ShouldStopService() {
				break
			}

			conn, err := socket.Accept()
			if err != nil {
				log.Error().Msgf("error accepting connection: %s", err)
				return
			}

			go func(conn net.Conn) {
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

				payload := ""

				switch strings.TrimSpace(string(buf[:n])) {
				case "status":
					lastScanned := st.GetLastScanned()
					if lastScanned.UID != "" {
						payload = fmt.Sprintf(
							"%d,%s,%t,%s",
							lastScanned.ScanTime.Unix(),
							lastScanned.UID,
							!st.IsLauncherDisabled(),
							lastScanned.Text,
						)
					} else {
						payload = fmt.Sprintf("0,,%t,", !st.IsLauncherDisabled())
					}
				case "disable":
					st.DisableLauncher()
					log.Info().Msg("launcher disabled")
				case "enable":
					st.EnableLauncher()
					log.Info().Msg("launcher enabled")
				case "connection":
					connected, rt := false, ""

					rs := st.ListReaders()
					if len(rs) > 0 {
						rid := rs[0]

						lt := st.GetLastScanned()
						if !lt.ScanTime.IsZero() && !lt.Remote {
							rid = lt.Source
						}

						reader, ok := st.GetReader(rid)
						if ok && reader != nil {
							connected = reader.Connected()
							info := reader.Info()

							var connProto string
							conn := strings.SplitN(strings.ToLower(reader.Device()), ":", 2)
							if len(connProto) == 2 {
								connProto = conn[0]
							}

							if connProto == "pn532_uart" {
								rt = ReaderTypePN532
							} else if strings.Contains(info, "ACR122U") {
								rt = ReaderTypeACR122U
							} else {
								rt = ReaderTypeUnknown
							}
						}
					}

					payload = fmt.Sprintf("%t,%s", connected, rt)
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
