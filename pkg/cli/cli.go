package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/google/uuid"
	"github.com/mdp/qrterminal/v3"
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
	Write        *string
	Launch       *string
	Api          *string
	Clients      *bool
	NewClient    *string
	DeleteClient *string
	Qr           *bool
	Version      *bool
}

// SetupFlags defines all common CLI flags between platforms.
func SetupFlags() *Flags {
	return &Flags{
		Write: flag.String(
			"write",
			"",
			"write text to tag using connected reader",
		),
		Launch: flag.String(
			"launch",
			"",
			"launch text as if it were a scanned token",
		),
		Api: flag.String(
			"api",
			"",
			"send method and params to API and print response",
		),
		Clients: flag.Bool(
			"clients",
			false,
			"list all registered API clients and secrets",
		),
		NewClient: flag.String(
			"new-client",
			"",
			"register new API client with given display name",
		),
		DeleteClient: flag.String(
			"delete-client",
			"",
			"revoke access to API for given client ID",
		),
		Qr: flag.Bool(
			"qr",
			false,
			"output a connection QR code along with client details",
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

type ConnQr struct {
	Id      uuid.UUID `json:"id"`
	Secret  string    `json:"sec"`
	Address string    `json:"addr"`
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

	// clients
	if *f.Clients {
		resp, err := client.LocalClient(cfg, models.MethodClients, "")
		if err != nil {
			log.Error().Err(err).Msg("error calling API")
			_, _ = fmt.Fprintf(os.Stderr, "Error calling API: %v\n", err)
			os.Exit(1)
		}

		var clients []models.ClientResponse
		err = json.Unmarshal([]byte(resp), &clients)
		if err != nil {
			log.Error().Err(err).Msg("error decoding API response")
			_, _ = fmt.Fprintf(os.Stderr, "Error decoding API response: %v\n", err)
		}

		for _, c := range clients {
			fmt.Println("---")
			if c.Name != "" {
				fmt.Printf("- Name:   %s\n", c.Name)
			}
			if c.Address != "" {
				fmt.Printf("- Address: %s\n", c.Address)
			}
			fmt.Printf("- ID:     %s\n", c.Id)
			fmt.Printf("- Secret: %s\n", c.Secret)

			if *f.Qr {
				ip, err := utils.GetLocalIp()
				if err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "Error getting local IP: %v\n", err)
					os.Exit(1)
				}

				cq := ConnQr{
					Id:      c.Id,
					Secret:  c.Secret,
					Address: ip.String(),
				}
				respQr, err := json.Marshal(cq)
				if err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "Error encoding QR code: %v\n", err)
					os.Exit(1)
				}

				qrterminal.Generate(
					string(respQr),
					qrterminal.L,
					os.Stdout,
				)
			}
		}

		os.Exit(0)
	} else if *f.NewClient != "" {
		data, err := json.Marshal(&models.NewClientParams{
			Name: *f.NewClient,
		})
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error encoding params: %v\n", err)
			os.Exit(1)
		}

		resp, err := client.LocalClient(
			cfg,
			models.MethodClientsNew,
			string(data),
		)
		if err != nil {
			log.Error().Err(err).Msg("error calling API")
			_, _ = fmt.Fprintf(os.Stderr, "Error calling API: %v\n", err)
			os.Exit(1)
		}

		var c models.ClientResponse
		err = json.Unmarshal([]byte(resp), &c)
		if err != nil {
			log.Error().Err(err).Msg("error decoding API response")
			_, _ = fmt.Fprintf(os.Stderr, "Error decoding API response: %v\n", err)
		}

		fmt.Println("New client registered:")
		fmt.Printf("- ID:     %s\n", c.Id)
		fmt.Printf("- Name:   %s\n", c.Name)
		fmt.Printf("- Secret: %s\n", c.Secret)

		if *f.Qr {
			ip, err := utils.GetLocalIp()
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Error getting local IP: %v\n", err)
				os.Exit(1)
			}

			cq := ConnQr{
				Id:      c.Id,
				Secret:  c.Secret,
				Address: ip.String(),
			}
			respQr, err := json.Marshal(cq)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Error encoding QR code: %v\n", err)
				os.Exit(1)
			}

			qrterminal.Generate(
				string(respQr),
				qrterminal.L,
				os.Stdout,
			)
		}

		os.Exit(0)
	} else if *f.DeleteClient != "" {
		data, err := json.Marshal(&models.DeleteClientParams{
			Id: *f.DeleteClient,
		})
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error encoding params: %v\n", err)
			os.Exit(1)
		}

		_, err = client.LocalClient(
			cfg,
			models.MethodClientsDelete,
			string(data),
		)
		if err != nil {
			log.Error().Err(err).Msg("error calling API")
			_, _ = fmt.Fprintf(os.Stderr, "Error calling API: %v\n", err)
			os.Exit(1)
		}

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
