package websocket

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func send(c *websocket.Conn, msg string) error {
	if c == nil {
		return fmt.Errorf("websocket connection is nil")
	}

	err := c.WriteMessage(websocket.TextMessage, []byte(msg))
	if err != nil {
		return err
	}

	return nil
}

type connGroup struct {
	mu     sync.Mutex
	logger zerolog.Logger
	conns  []*websocket.Conn
}

func (cg *connGroup) Add(c *websocket.Conn) int {
	cg.mu.Lock()
	defer cg.mu.Unlock()
	cg.conns = append(cg.conns, c)
	return len(cg.conns) - 1
}

func (cg *connGroup) Remove(i int) {
	cg.mu.Lock()
	defer cg.mu.Unlock()
	cg.conns = append(cg.conns[:i], cg.conns[i+1:]...)
}

func (cg *connGroup) All() []*websocket.Conn {
	cg.mu.Lock()
	defer cg.mu.Unlock()
	return cg.conns
}

func (cg *connGroup) Clean() {
	cg.mu.Lock()
	defer cg.mu.Unlock()
	fresh := make([]*websocket.Conn, 0)
	for _, c := range cg.conns {
		if c != nil {
			fresh = append(fresh, c)
		}
	}
	cg.conns = fresh
}

func (cg *connGroup) Broadcast(msg string) {
	cg.Clean()
	cg.mu.Lock()
	defer cg.mu.Unlock()
	for i, c := range cg.conns {
		err := send(c, msg)
		if err != nil {
			cg.logger.Error().Err(err).Msg("failed to write to websocket")
			c.Close()
			cg.conns[i] = nil
		}
	}
}

var conns = &connGroup{}

func Handle(
	connectPayload func() []string,
	msgHandler func(msg string) string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			conns.logger.Error().Err(err).Msg("failed to upgrade websocket")
			return
		}

		id := conns.Add(c)

		defer func(c *websocket.Conn) {
			err := c.Close()
			if err != nil {
				conns.logger.Error().Err(err).Msg("failed to close websocket")
			}
			conns.Remove(id)
		}(c)

		for _, msg := range connectPayload() {
			err = send(c, msg)
			if err != nil {
				conns.logger.Error().Err(err).Msg("failed to write to websocket during connect")
				return
			}
		}

		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseGoingAway) {
					return
				} else if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					return
				}
				conns.logger.Error().Err(err).Msg("failed to read from websocket")
				return
			}

			conns.logger.Info().Msgf("received message: %s", msg)
			response := msgHandler(string(msg))

			if response == "" {
				continue
			}

			err = send(c, response)
			if err != nil {
				conns.logger.Error().Err(err).Msg("failed to write to websocket")
				return
			}
		}
	}
}

func Broadcast(msg string) {
	conns.Broadcast(msg)
}
