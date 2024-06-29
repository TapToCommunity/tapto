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
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/rthornton128/goncurses"
	"github.com/wizzomafizzo/mrext/pkg/curses"
	mrextMister "github.com/wizzomafizzo/mrext/pkg/mister"
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

func displayServiceInfo(stdscr *goncurses.Window, service *mister.Service) error {
	width := 40
	height := 6

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
	selected := 0

	for {
		var statusText string
		//var toggleText string
		running := service.Running()
		if running {
			statusText = "Service:        RUNNING"
			//toggleText = "Stop Service"
		} else {
			statusText = "Service:        NOT RUNNING"
			//toggleText = "Start Service"
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
		}

		clearLine(2)
		printLeft(2, "Device address: "+ipDisplay)

		clearLine(height - 2)
		curses.DrawActionButtons(win, []string{"Exit"}, selected, 0)

		win.NoutRefresh()
		err = goncurses.Update()
		if err != nil {
			return err
		}

		ch = win.GetChar()

		if ch == goncurses.KEY_LEFT {
			selected = 0
		} else if ch == goncurses.KEY_RIGHT {
			selected = 0
		} else if ch == goncurses.KEY_ENTER || ch == 10 || ch == 13 {
			if selected == 0 {
				// if service.Running() {
				// 	err := service.Stop()
				// 	if err != nil {
				// 		log.Error().Msgf("could not stop service: %s", err)
				// 	}
				// } else {
				// 	err := service.Start()
				// 	if err != nil {
				// 		log.Error().Msgf("could not start service: %s", err)
				// 	}
				// }
				// time.Sleep(1 * time.Second)
				break
			}
		} else if ch == goncurses.KEY_ESC {
			break
		}
	}

	return nil
}
