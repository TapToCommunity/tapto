package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/api/client"
	"github.com/wizzomafizzo/tapto/pkg/api/models"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/platforms"
	"github.com/wizzomafizzo/tapto/pkg/utils"
	"os"
	"strings"
)

type Flags struct {
	Write   *string
	Launch  *string
	Api     *string
	Version *bool
}

// SetupFlags defines all common CLI flags between platforms.
func SetupFlags() *Flags {
	return &Flags{
		Write: flag.String(
			"write",
			"",
			"write text to tag using connected reader (via API)",
		),
		Launch: flag.String(
			"launch",
			"",
			"launch text as if it were a scanned token (via API)",
		),
		Api: flag.String(
			"api",
			"",
			"send method and params to API and print response",
		),
		Version: flag.Bool(
			"version",
			false,
			"print version and exit",
		),
	}
}

// Pre runs flag parsing and actions any immediate flags that don't
// require environment setup. Add any custom flags before running this.
func (f *Flags) Pre(pl platforms.Platform) {
	flag.Parse()

	if *f.Version {
		fmt.Printf("TapTo v%s (%s)\n", config.Version, pl.Id())
		os.Exit(0)
	}
}

// Post actions all remaining common flags that require the environment to be
// set up. Logging is allowed.
func (f *Flags) Post(cfg *config.UserConfig) {
	if *f.Write != "" {
		data, err := json.Marshal(&models.ReaderWriteParams{
			Text: *f.Write,
		})
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error encoding params: %v\n", err)
			os.Exit(1)
		}

		_, err = client.LocalClient(cfg, models.MethodReadersWrite, string(data))
		if err != nil {
			log.Error().Err(err).Msg("error writing tag")
			_, _ = fmt.Fprintf(os.Stderr, "Error writing tag: %v\n", err)
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	} else if *f.Launch != "" {
		data, err := json.Marshal(&models.LaunchParams{
			Text: *f.Launch,
		})
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error encoding params: %v\n", err)
			os.Exit(1)
		}

		_, err = client.LocalClient(cfg, models.MethodLaunch, string(data))
		if err != nil {
			log.Error().Err(err).Msg("error launching")
			_, _ = fmt.Fprintf(os.Stderr, "Error launching: %v\n", err)
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	} else if *f.Api != "" {
		ps := strings.SplitN(*f.Api, ":", 2)
		method := ps[0]
		params := ""
		if len(ps) > 1 {
			params = ps[1]
		}

		resp, err := client.LocalClient(cfg, method, params)
		if err != nil {
			log.Error().Err(err).Msg("error calling API")
			_, _ = fmt.Fprintf(os.Stderr, "Error calling API: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(resp)
		os.Exit(0)
	}
}

// Setup initializes the user config and logging. Returns a user config object.
func Setup(pl platforms.Platform, defaultConfig *config.UserConfig) *config.UserConfig {
	cfg, err := config.NewUserConfig(defaultConfig)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	err = utils.InitLogging(cfg, pl)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error initializing logging: %v\n", err)
		os.Exit(1)
	}

	return cfg
}
