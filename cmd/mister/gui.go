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

package main

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/rthornton128/goncurses"
	"github.com/wizzomafizzo/mrext/pkg/curses"
	mrextMister "github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/tapto/pkg/assets"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/database/gamesdb"
	"github.com/wizzomafizzo/tapto/pkg/platforms"
	"github.com/wizzomafizzo/tapto/pkg/platforms/mister"
	"github.com/wizzomafizzo/tapto/pkg/utils"
)

func tryAddStartup(stdscr *goncurses.Window) error {
	var startup mrextMister.Startup

	err := startup.Load()
	if err != nil {
		log.Error().Msgf("failed to load startup file: %s", err)
	}

	if !startup.Exists("mrext/" + appName) {
		win, err := curses.NewWindow(stdscr, 6, 43, "", -1)
		if err != nil {
			return err
		}
		defer func(win *goncurses.Window) {
			err := win.Delete()
			if err != nil {
				log.Error().Msgf("failed to delete window: %s", err)
			}
		}(win)

		var ch goncurses.Key
		selected := 0

		for {
			win.MovePrint(1, 3, "Add TapTo service to MiSTer startup?")
			win.MovePrint(2, 2, "This won't impact MiSTer's performance.")
			curses.DrawActionButtons(win, []string{"Yes", "No"}, selected, 10)

			win.NoutRefresh()
			err := goncurses.Update()
			if err != nil {
				return err
			}

			ch = win.GetChar()

			if ch == goncurses.KEY_LEFT {
				if selected == 0 {
					selected = 1
				} else if selected == 1 {
					selected = 0
				}
			} else if ch == goncurses.KEY_RIGHT {
				if selected == 0 {
					selected = 1
				} else if selected == 1 {
					selected = 0
				}
			} else if ch == goncurses.KEY_ENTER || ch == 10 || ch == 13 {
				break
			} else if ch == goncurses.KEY_ESC {
				selected = 1
				break
			}
		}

		if selected == 0 {
			err = startup.AddService("mrext/" + appName)
			if err != nil {
				return err
			}

			err = startup.Save()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func generateIndexWindow(pl platforms.Platform, cfg *config.UserConfig, stdscr *goncurses.Window) error {
	win, err := curses.NewWindow(stdscr, 4, 40, "", -1)
	if err != nil {
		return err
	}
	defer win.Delete()

	_, width := win.MaxYX()

	drawProgressBar := func(current int, total int) {
		pct := int(float64(current) / float64(total) * 100)
		progressWidth := width - 4
		progressPct := int(float64(pct) / float64(100) * float64(progressWidth))
		if progressPct < 1 {
			progressPct = 1
		}
		for i := 0; i < progressPct; i++ {
			win.MoveAddChar(2, 2+i, goncurses.ACS_BLOCK)
		}
		win.NoutRefresh()
	}

	clearText := func() {
		win.MovePrint(1, 2, strings.Repeat(" ", width-4))
	}

	status := struct {
		Step        int
		Total       int
		SystemName  string
		DisplayText string
		Complete    bool
		Error       error
	}{
		Step:        1,
		Total:       100,
		DisplayText: "Finding games folders...",
	}

	go func() {
		_, err = gamesdb.NewNamesIndex(pl, cfg, gamesdb.AllSystems(), func(is gamesdb.IndexStatus) {
			systemName := is.SystemId
			system, err := gamesdb.GetSystem(is.SystemId)
			if err == nil {
				md, err := assets.GetSystemMetadata(system.Id)
				if err == nil {
					systemName = md.Name
				}
			}

			text := fmt.Sprintf("Indexing %s...", systemName)
			if is.Step == 1 {
				text = "Finding games folders..."
			} else if is.Step == is.Total {
				text = "Writing database to disk..."
			}

			status.Step = is.Step
			status.Total = is.Total
			status.SystemName = systemName
			status.DisplayText = text
		})

		status.Error = err
		status.Complete = true
	}()

	spinnerSeq := []string{"|", "/", "-", "\\"}
	spinnerCount := 0

	for {
		if status.Complete || status.Error != nil {
			break
		}

		clearText()

		spinnerCount++
		if spinnerCount == len(spinnerSeq) {
			spinnerCount = 0
		}

		win.MovePrint(1, width-3, spinnerSeq[spinnerCount])

		win.MovePrint(1, 2, status.DisplayText)
		drawProgressBar(status.Step, status.Total)

		win.NoutRefresh()
		_ = goncurses.Update()
		goncurses.Nap(100)
	}

	return status.Error
}

func displayServiceInfo(pl platforms.Platform, cfg *config.UserConfig, stdscr *goncurses.Window, service *mister.Service) error {
	width := 50
	height := 7

	win, err := curses.NewWindow(stdscr, height, width, "", -1)
	if err != nil {
		return err
	}
	defer func(win *goncurses.Window) {
		err := win.Delete()
		if err != nil {
			log.Error().Msgf("failed to delete window: %s", err)
		}
	}(win)

	win.Timeout(300)

	printLeft := func(y int, text string) {
		win.MovePrint(y, 2, text)
	}

	printCenter := func(y int, text string) {
		win.MovePrint(y, (width-len(text))/2, text)
	}

	clearLine := func(y int) {
		win.MovePrint(y, 2, strings.Repeat(" ", width-4))
	}

	var ch goncurses.Key
	selected := 1

	for {
		var statusText string
		running := service.Running()
		if running {
			statusText = "Service:        RUNNING"
		} else {
			statusText = "Service:        NOT RUNNING"
		}

		printCenter(0, "TapTo v"+appVersion)

		clearLine(1)
		printLeft(1, statusText)

		ip, err := utils.GetLocalIp()
		var ipDisplay string
		if err != nil {
			ipDisplay = "Unknown"
		} else {
			ipDisplay = ip.String()
			if cfg.TapTo.ApiPort != config.DefaultApiPort {
				ipDisplay += ":" + cfg.TapTo.ApiPort
			}
		}

		clearLine(2)
		printLeft(2, "Device address: "+ipDisplay)

		clearLine(3)
		dbExistsDisplay := "NOT CREATED"
		if gamesdb.DbExists(pl) {
			dbExistsDisplay = "CREATED"
		}
		printLeft(3, "Games DB:       "+dbExistsDisplay)

		clearLine(height - 2)
		curses.DrawActionButtons(win, []string{"Update Games DB", "Exit"}, selected, 2)

		win.NoutRefresh()
		err = goncurses.Update()
		if err != nil {
			return err
		}

		ch = win.GetChar()

		if ch == goncurses.KEY_LEFT {
			if selected == 0 {
				selected = 1
			} else {
				selected = 0
			}
		} else if ch == goncurses.KEY_RIGHT {
			if selected == 0 {
				selected = 1
			} else {
				selected = 0
			}
		} else if ch == goncurses.KEY_ENTER || ch == 10 || ch == 13 {
			if selected == 0 {
				err := generateIndexWindow(pl, cfg, stdscr)
				if err != nil {
					log.Error().Msgf("failed to display gamesdb window: %s", err)
				}
			} else if selected == 1 {
				break
			}
		} else if ch == goncurses.KEY_ESC {
			break
		}
	}

	return nil
}
