package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"go.bug.st/serial"
)

func getLinuxList() ([]string, error) {
	path := "/dev/serial/by-id"

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	files, err := f.Readdir(0)
	if err != nil {
		return nil, err
	}

	var devices []string

	for _, v := range files {
		if v.IsDir() {
			continue
		}

		devices = append(devices, filepath.Join(path, v.Name()))
	}

	return devices, nil
}

func GetSerialDeviceList() ([]string, error) {
	if runtime.GOOS == "linux" {
		return getLinuxList()
	} else if runtime.GOOS == "darwin" {
		var devices []string
		ports, err := serial.GetPortsList()
		if err != nil {
			return nil, err
		}

		for _, v := range ports {
			if !strings.HasPrefix(v, "/dev/tty.") {
				continue
			}

			devices = append(devices, v)
		}

		return devices, nil
	} else if runtime.GOOS == "windows" {
		var devices []string
		ports, err := serial.GetPortsList()
		if err != nil {
			return nil, err
		}

		for _, v := range ports {
			if !strings.HasPrefix(v, "COM") {
				continue
			}

			devices = append(devices, v)
		}

		return devices, nil
	}

	return serial.GetPortsList()
}
