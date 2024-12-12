package launcher

import (
	"fmt"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/ZaparooProject/zaparoo-core/pkg/platforms"
)

// DEPRECATED
func cmdKey(pl platforms.Platform, env platforms.CmdEnv) error {
	return pl.KeyboardInput(env.Args)
}

// converts a string to a list of key symbols. long names are named inside
// curly braces and characters can be escaped with a backslash
func readKeys(keys string) ([]string, error) {
	var names []string
	inEscape := false
	inName := false
	var name string

	for _, c := range keys {
		if inEscape {
			name += string(c)
			inEscape = false
			continue
		}

		if c == '\\' {
			inEscape = true
			continue
		}

		if c == '{' {
			if inName {
				return nil, fmt.Errorf("unexpected {")
			}

			inName = true
			continue
		}

		if c == '}' {
			if !inName {
				return nil, fmt.Errorf("unexpected }")
			}

			names = append(names, name)
			name = ""
			inName = false
			continue
		}

		if inName {
			name += string(c)
		} else {
			names = append(names, string(c))
		}
	}

	if inName {
		return nil, fmt.Errorf("missing }")
	}

	return names, nil
}

func cmdKeyboard(pl platforms.Platform, env platforms.CmdEnv) error {
	log.Info().Msgf("keyboard input: %s", env.Args)

	// TODO: stuff like adjust delay, only press, etc.
	//	     basically a filled out mini macro language for key presses

	names, err := readKeys(env.Args)
	if err != nil {
		return err
	}

	for _, name := range names {
		if err := pl.KeyboardPress(name); err != nil {
			return err
		}
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

func cmdGamepad(pl platforms.Platform, env platforms.CmdEnv) error {
	log.Info().Msgf("gamepad input: %s", env.Args)

	names, err := readKeys(env.Args)
	if err != nil {
		return err
	}

	for _, name := range names {
		if err := pl.GamepadPress(name); err != nil {
			return err
		}
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

func insertCoin(pl platforms.Platform, env platforms.CmdEnv, key string) error {
	amount, err := strconv.Atoi(env.Args)
	if err != nil {
		return err
	}

	for i := 0; i < amount; i++ {
		pl.KeyboardInput(key)
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

func cmdCoinP1(pl platforms.Platform, env platforms.CmdEnv) error {
	log.Info().Msgf("inserting coin for player 1: %s", env.Args)
	return insertCoin(pl, env, "6")
}

func cmdCoinP2(pl platforms.Platform, env platforms.CmdEnv) error {
	log.Info().Msgf("inserting coin for player 2: %s", env.Args)
	return insertCoin(pl, env, "7")
}
