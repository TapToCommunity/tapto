/*
TapTo
Copyright (C) 2023, 2024 Callan Barrett
Copyright (C) 2023 Gareth Jones

This file is part of TapTo.

TapTo is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

TapTo is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with TapTo.  If not, see <http://www.gnu.org/licenses/>.
*/

package config

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/rs/zerolog"
	"gopkg.in/ini.v1"
)

const UserConfigEnv = "TAPTO_CONFIG"
const UserAppPathEnv = "TAPTO_APP_PATH"

type TapToConfig struct {
	ConnectionString  string   `ini:"connection_string"` // DEPRECATED
	Reader            []string `ini:"reader,omitempty,allowshadow"`
	AllowCommands     bool     `ini:"allow_commands"` // TODO: rename to allow_shell
	DisableSounds     bool     `ini:"disable_sounds"`
	ProbeDevice       bool     `ini:"probe_device"`
	ExitGame          bool     `ini:"exit_game"` // TODO: rename to insert_mode
	ExitGameBlocklist []string `ini:"exit_game_blocklist"`
	ExitGameDelay     int8     `ini:"exit_game_delay"`
	Debug             bool     `ini:"debug"`
	ApiPort           int      `ini:"api_port"`
	ApiBasicAuth      string   `ini:"api_basic_auth"`
}

type SystemsConfig struct {
	GamesFolder []string `ini:"games_folder,omitempty,allowshadow"`
	SetCore     []string `ini:"set_core,omitempty,allowshadow"`
}

type UserConfig struct {
	mu      sync.RWMutex
	AppPath string        `ini:"-"`
	IniPath string        `ini:"-"`
	TapTo   TapToConfig   `ini:"tapto"`
	Systems SystemsConfig `ini:"systems,omitempty"`
}

func (c *UserConfig) GetConnectionString() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.TapTo.ConnectionString
}

func (c *UserConfig) SetConnectionString(connectionString string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.TapTo.ConnectionString = connectionString
}

func (c *UserConfig) GetReader() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.TapTo.Reader
}

func (c *UserConfig) SetReader(reader []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.TapTo.Reader = reader
}

func (c *UserConfig) GetAllowCommands() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.TapTo.AllowCommands
}

func (c *UserConfig) SetAllowCommands(allowCommands bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.TapTo.AllowCommands = allowCommands
}

func (c *UserConfig) GetDisableSounds() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.TapTo.DisableSounds
}

func (c *UserConfig) SetDisableSounds(disableSounds bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.TapTo.DisableSounds = disableSounds
}

func (c *UserConfig) GetProbeDevice() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.TapTo.ProbeDevice
}

func (c *UserConfig) SetProbeDevice(probeDevice bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.TapTo.ProbeDevice = probeDevice
}

func (c *UserConfig) GetExitGame() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.TapTo.ExitGame
}

func (c *UserConfig) SetExitGameDelay(exitGameDelay int8) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.TapTo.ExitGameDelay = exitGameDelay
}

func (c *UserConfig) SetExitGame(exitGame bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.TapTo.ExitGame = exitGame
}

func (c *UserConfig) GetExitGameBlocklist() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.TapTo.ExitGameBlocklist
}

func (c *UserConfig) GetExitGameDelay() int8 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.TapTo.ExitGameDelay
}

func (c *UserConfig) SetExitGameBlocklist(exitGameBlocklist []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.TapTo.ExitGameBlocklist = exitGameBlocklist
}

func (c *UserConfig) GetDebug() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.TapTo.Debug
}

func (c *UserConfig) SetDebug(debug bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.TapTo.Debug = debug
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

func (c *UserConfig) LoadConfig() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cfg, err := ini.ShadowLoad(c.IniPath)
	if err != nil {
		return err
	}

	err = cfg.StrictMapTo(c)
	if err != nil {
		return err
	}

	return nil
}

func (c *UserConfig) SaveConfig() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cfg := ini.Empty()

	ini.PrettyEqual = false
	ini.PrettyFormat = false

	err := cfg.ReflectFrom(c)
	if err != nil {
		return err
	}

	err = cfg.SaveTo(c.IniPath)
	if err != nil {
		return err
	}

	return nil
}

func NewUserConfig(name string, defaultConfig *UserConfig) (*UserConfig, error) {
	iniPath := os.Getenv(UserConfigEnv)

	exePath, err := os.Executable()
	if err != nil {
		return defaultConfig, err
	}

	appPath := os.Getenv(UserAppPathEnv)
	if appPath != "" {
		exePath = appPath
	}

	if iniPath == "" {
		iniPath = filepath.Join(filepath.Dir(exePath), name+".ini")
	}

	defaultConfig.AppPath = exePath
	defaultConfig.IniPath = iniPath

	if _, err := os.Stat(iniPath); os.IsNotExist(err) {
		return defaultConfig, nil
	}

	err = defaultConfig.LoadConfig()
	if err != nil {
		return defaultConfig, err
	}

	return defaultConfig, nil
}
