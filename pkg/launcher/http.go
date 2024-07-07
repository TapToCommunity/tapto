package launcher

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/wizzomafizzo/tapto/pkg/platforms"
)

func cmdHttpGet(pl platforms.Platform, env platforms.CmdEnv) error {
	go func() {
		resp, err := http.Get(env.Args)
		if err != nil {
			log.Error().Msgf("error getting url: %s", err)
			return
		}
		resp.Body.Close()
	}()

	return nil
}

func cmdHttpPost(pl platforms.Platform, env platforms.CmdEnv) error {
	parts := strings.SplitN(env.Args, ",", 3)
	if len(parts) < 3 {
		return fmt.Errorf("invalid post format: %s", env.Args)
	}

	url, format, data := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), strings.TrimSpace(parts[2])

	go func() {
		resp, err := http.Post(url, format, strings.NewReader(data))
		if err != nil {
			log.Error().Msgf("error posting to url: %s", err)
			return
		}
		resp.Body.Close()
	}()

	return nil
}
