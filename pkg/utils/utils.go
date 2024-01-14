/*
TapTo
Copyright (C) 2023, 2024 Callan Barrett

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

package utils

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"

	mrextConfig "github.com/wizzomafizzo/mrext/pkg/config"
	config "github.com/wizzomafizzo/tapto/pkg/config"
)

func GetMd5Hash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	hash := md5.New()
	_, _ = io.Copy(hash, file)
	_ = file.Close()
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func GetFileSize(path string) (int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}

	stat, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return 0, err
	}

	size := stat.Size()
	_ = file.Close()

	return size, nil
}

func UserConfigToMrext(cfg *config.UserConfig) *mrextConfig.UserConfig {
	return &mrextConfig.UserConfig{
		AppPath: cfg.AppPath,
		IniPath: cfg.IniPath,
		Nfc: mrextConfig.NfcConfig{
			ConnectionString: cfg.TapTo.ConnectionString,
			AllowCommands:    cfg.TapTo.AllowCommands,
			DisableSounds:    cfg.TapTo.DisableSounds,
			ProbeDevice:      cfg.TapTo.ProbeDevice,
		},
		Systems: mrextConfig.SystemsConfig{
			GamesFolder: cfg.Systems.GamesFolder,
			SetCore:     cfg.Systems.SetCore,
		},
	}
}
