//go:build mage

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
	"os"
	"path/filepath"
	"runtime"

	_ "github.com/joho/godotenv/autoload"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var (
	cwd, _         = os.Getwd()
	binDir         = filepath.Join(cwd, "_bin")
	binReleasesDir = filepath.Join(binDir, "releases")
	upxBin         = os.Getenv("UPX_BIN")
	// docker mister arm build
	misterBuild          = filepath.Join(cwd, "scripts", "misterbuild")
	misterBuildImageName = "tapto/misterbuild"
	misterBuildCache     = filepath.Join(os.TempDir(), "tapto-mister-buildcache")
	misterModCache       = filepath.Join(os.TempDir(), "tapto-mister-modcache")
)

type app struct {
	name         string
	path         string
	bin          string
	ldFlags      string
	releaseId    string
	reboot       bool
	inAll        bool
	releaseFiles []string
}

var apps = []app{
	{
		name:         "tapto",
		path:         filepath.Join(cwd, "cmd", "nfc"),
		bin:          "tapto.sh",
		releaseId:    "mrext/tapto",
		ldFlags:      "-lnfc -lusb -lcurses",
		releaseFiles: []string{filepath.Join(cwd, "scripts", "nfcui", "nfcui.sh")},
	},
}

func getApp(name string) *app {
	for _, a := range apps {
		if a.name == name {
			return &a
		}
	}
	return nil
}

func cleanPlatform(name string) {
	_ = sh.Rm(filepath.Join(binDir, name))
}

func Clean() {
	_ = sh.Rm(binDir)
	_ = sh.Rm(misterBuildCache)
	_ = sh.Rm(misterModCache)
}

func buildApp(a app, out string) {
	if a.ldFlags == "" {
		env := map[string]string{
			"GOPROXY": "https://goproxy.io,direct",
		}
		_ = sh.RunWithV(env, "go", "build", "-o", out, a.path)
	} else {
		staticEnv := map[string]string{
			"GOPROXY":     "https://goproxy.io,direct",
			"CGO_ENABLED": "1",
			"CGO_LDFLAGS": a.ldFlags,
		}
		_ = sh.RunWithV(staticEnv, "go", "build", "--ldflags", "-linkmode external -extldflags -static", "-o", out, a.path)
	}
}

func Build(appName string) {
	platform := runtime.GOOS + "_" + runtime.GOARCH

	if appName == "all" {
		mg.Deps(func() { cleanPlatform(platform) })
		for _, app := range apps {
			fmt.Println("Building", app.name)
			buildApp(app, filepath.Join(binDir, platform, app.bin))
		}
	} else {
		app := getApp(appName)
		if app == nil {
			fmt.Println("Unknown app", appName)
			os.Exit(1)
		}
		buildApp(*app, filepath.Join(binDir, platform, app.bin))
	}
}

func MakeMisterImage() {
	if runtime.GOOS != "linux" {
		_ = sh.RunV("docker", "build", "--platform", "linux/arm/v7", "-t", misterBuildImageName, misterBuild)
	} else {
		_ = sh.RunV("sudo", "docker", "build", "--platform", "linux/arm/v7", "-t", misterBuildImageName, misterBuild)
	}
}

func Mister(appName string) {
	buildCache := fmt.Sprintf("%s:%s", misterBuildCache, "/home/build/.cache/go-build")
	_ = os.Mkdir(misterBuildCache, 0755)

	modCache := fmt.Sprintf("%s:%s", misterModCache, "/home/build/go/pkg/mod")
	_ = os.Mkdir(misterModCache, 0755)

	args := []string{
		"docker",
		"run",
		"--rm",
		"--platform",
		"linux/arm/v7",
		"-v",
		buildCache,
		"-v",
		modCache,
		"-v",
		fmt.Sprintf("%s:%s", cwd, "/build"),
		"--user",
		"1000:1000",
		misterBuildImageName,
		"mage",
		"build",
		appName,
	}

	if runtime.GOOS == "linux" {
		args = append([]string{"sudo"}, args...)
	}

	_ = sh.RunV(args[0], args[1:]...)
}

type updateDbFile struct {
	Hash   string `json:"hash"`
	Size   int64  `json:"size"`
	Url    string `json:"url"`
	Reboot bool   `json:"reboot,omitempty"`
}

type updateDbFolder struct {
	Tags []string `json:"tags,omitempty"`
}

type updateDb struct {
	DbId      string                    `json:"db_id"`
	Timestamp int64                     `json:"timestamp"`
	Files     map[string]updateDbFile   `json:"files"`
	Folders   map[string]updateDbFolder `json:"folders"`
}

func Release(name string) {
	a := getApp(name)
	if a == nil {
		fmt.Println("Unknown app", name)
		os.Exit(1)
	}

	Mister(name)

	_ = os.MkdirAll(binReleasesDir, 0755)
	releaseBin := filepath.Join(binReleasesDir, a.bin)
	err := sh.Copy(releaseBin, filepath.Join(binDir, "linux_arm", a.bin))
	if err != nil {
		fmt.Println("Error copying binary", err)
		os.Exit(1)
	}

	for _, f := range a.releaseFiles {
		err := sh.Copy(filepath.Join(binReleasesDir, filepath.Base(f)), f)
		if err != nil {
			fmt.Println("Error copying release file", err)
			os.Exit(1)
		}
	}

	if upxBin == "" {
		fmt.Println("UPX is required for releases")
		os.Exit(1)
	} else {
		if runtime.GOOS != "windows" {
			err := os.Chmod(releaseBin, 0755)
			if err != nil {
				fmt.Println("Error chmod release bin", err)
				os.Exit(1)
			}
		}

		err := sh.RunV(upxBin, "-9", releaseBin)
		if err != nil {
			fmt.Println("Error compressing binary", err)
			os.Exit(1)
		}
	}
}

func PrepRelease() {
	_ = sh.Rm(binReleasesDir)
	_ = os.MkdirAll(binReleasesDir, 0755)
	cleanPlatform("linux_arm")
	for _, app := range apps {
		if app.releaseId != "" {
			fmt.Println("Preparing release:", app.name)
			Release(app.name)
		}
	}
}

func MakeArmApp(name string) {
	buildScript := name + ".sh"
	if _, err := os.Stat(filepath.Join(misterBuild, buildScript)); os.IsNotExist(err) {
		fmt.Println("No build script for", name)
		os.Exit(1)
	}

	buildDir := filepath.Join(misterBuild, "_build")
	_ = os.MkdirAll(buildDir, 0755)

	err := sh.Copy(filepath.Join(buildDir, buildScript), filepath.Join(misterBuild, buildScript))
	if err != nil {
		fmt.Println("Error copying build script", err)
		os.Exit(1)
	}

	args := []string{
		"docker",
		"run",
		"--rm",
		"--platform", "linux/arm/v7",
		"-v", buildDir + ":/build",
		"--user", "1000:1000",
		misterBuildImageName,
		"bash",
		"./" + buildScript,
	}

	if runtime.GOOS == "linux" {
		args = append([]string{"sudo"}, args...)
	}

	_ = sh.RunV(args[0], args[1:]...)
}

func Test() {
	_ = sh.RunV("go", "test", "./...")
}

func Coverage() {
	_ = sh.RunV("go", "test", "-coverprofile", "coverage.out", "./...")
	_ = sh.RunV("go", "tool", "cover", "-html", "coverage.out")
	_ = sh.Rm("coverage.out")
}
