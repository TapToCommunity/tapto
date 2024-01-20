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
	"encoding/json"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/rthornton128/goncurses"
	"github.com/wizzomafizzo/mrext/pkg/curses"
	mrextMister "github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/utils"
	"github.com/wizzomafizzo/tapto/pkg/platforms/mister"
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

type logEntry struct {
	Level   string `json:"level"`
	Time    int    `json:"time"`
	Message string `json:"message"`
}

func displayServiceInfo(stdscr *goncurses.Window, service *mister.Service) error {
	width := 57
	height := 18

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

	clearLine := func(y int) {
		win.MovePrint(y, 2, strings.Repeat(" ", width-4))
	}

	var ch goncurses.Key
	selected := 1

	for {
		var statusText string
		var toggleText string
		running := service.Running()
		if running {
			statusText = "Service:   RUNNING"
			toggleText = "Stop"
		} else {
			statusText = "Service:   NOT RUNNING"
			toggleText = "Start"
		}

		scanTime := "never"
		tagUid := ""
		tagText := ""
		conn, err := net.Dial("unix", mister.SocketFile)
		if err != nil {
			log.Debug().Msgf("could not connect to nfc service: %s", err)
		} else {
			_, err := conn.Write([]byte("status"))
			if err != nil {
				log.Debug().Msgf("could not write to nfc service: %s", err)
			} else {
				buf := make([]byte, 4096)
				_, err := conn.Read(buf)
				if err != nil {
					log.Debug().Msgf("could not read from nfc service: %s", err)
				} else {
					parts := strings.SplitN(string(buf), ",", 5)
					if parts[0] != "0" {
						scanTime = parts[0]
					}
					tagUid = parts[1]
					tagText = parts[3]
				}
			}
		}

		var logLines []logEntry
		logFile, err := os.ReadFile(mister.LogFile)
		if err != nil {
			log.Error().Msgf("could not read log file: %s", err)
		} else {
			lines := strings.Split(string(logFile), "\n")
			for i := len(lines) - 1; i >= 0; i-- {
				line := lines[i]

				if len(line) == 0 {
					continue
				}

				entry := logEntry{}
				err := json.Unmarshal([]byte(line), &entry)
				if err != nil {
					log.Error().Str("value", line).Msgf("could not unmarshal log entry: %s", err)
					continue
				}

				if entry.Level != "debug" {
					logLines = append(logLines, entry)
				}

				if len(logLines) >= 10 {
					break
				}
			}
			utils.Reverse(logLines)
		}

		clearLine(1)
		printLeft(1, statusText)

		if scanTime != "never" {
			t, err := strconv.ParseInt(scanTime, 10, 64)
			if err != nil {
				log.Debug().Msgf("could not parse scan time: %s", err)
			} else {
				scanTime = time.Unix(t, 0).Format("2006-01-02 15:04:05")
			}
		}
		clearLine(2)
		printLeft(2, "Last scan: "+scanTime)

		if tagUid != "" {
			tagUid = strings.ToUpper(tagUid)
			parts := make([]string, 0)
			for i := 0; i < len(tagUid); i += 2 {
				parts = append(parts, tagUid[i:i+2])
			}
			tagUid = strings.Join(parts, ":")
		}
		clearLine(3)
		printLeft(3, "Tag UID:   "+tagUid)

		if len(tagText) > 42 {
			tagText = tagText[:42-3] + "..."
		}
		clearLine(4)
		printLeft(4, "Tag text:  "+tagText)

		win.HLine(5, 1, goncurses.ACS_HLINE, width-2)
		win.MoveAddChar(5, 0, goncurses.ACS_LTEE)
		win.MoveAddChar(5, width-1, goncurses.ACS_RTEE)

		errorOn := func(level string) {
			if level != "info" {
				win.ColorOn(1)
			}
		}

		errorOff := func() {
			win.ColorOff(1)
		}

		// maximum 10 log lines, from line 6 to 15 of the window
		// print from bottom to top, if a line is over the width (53), split it
		// to a second line. if it's still over the width (106), truncate the second
		// line with a "..." on the end
		// TODO: this doesn't quite capture every edge case and it would make a good
		// 		 general purpose function in the curses package
		winLine := 15
		logLine := len(logLines) - 1
		for i := 0; i < 10; i++ {
			if logLine < 0 || winLine < 6 {
				break
			}

			line := logLines[logLine].Message
			level := logLines[logLine].Level

			logLine--

			if len(line) == 0 {
				continue
			}

			if len(line) > 53 {
				if winLine < 6 {
					break
				}

				if winLine == 6 {
					// just truncate the line
					line = line[:53-3] + "..."
					clearLine(winLine)
					errorOn(level)
					win.MovePrint(winLine, 2, line)
					errorOff()
					break
				}

				line1 := line[:53]
				line2 := line[53:]
				if len(line2) > 53 {
					line2 = line2[:53-3] + "..."
				}
				clearLine(winLine)
				errorOn(level)
				win.MovePrint(winLine, 2, line2)
				errorOff()
				winLine--
				clearLine(winLine)
				errorOn(level)
				win.MovePrint(winLine, 2, line1)
				errorOff()
				winLine--
			} else {
				clearLine(winLine)
				errorOn(level)
				win.MovePrint(winLine, 2, line)
				errorOff()
				winLine--
			}
		}

		clearLine(height - 2)
		curses.DrawActionButtons(win, []string{toggleText, "Exit"}, selected, 8)

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
				selected--
			}
		} else if ch == goncurses.KEY_RIGHT {
			if selected == 1 {
				selected = 0
			} else {
				selected++
			}
		} else if ch == goncurses.KEY_ENTER || ch == 10 || ch == 13 {
			if selected == 0 {
				if service.Running() {
					err := service.Stop()
					if err != nil {
						log.Error().Msgf("could not stop service: %s", err)
					}
				} else {
					err := service.Start()
					if err != nil {
						log.Error().Msgf("could not start service: %s", err)
					}
				}
				time.Sleep(1 * time.Second)
			} else if selected == 1 {
				break
			}
		} else if ch == goncurses.KEY_ESC {
			break
		}
	}

	return nil
}
