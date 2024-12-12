//go:build linux || darwin

package mister

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ZaparooProject/zaparoo-core/pkg/config"
	"github.com/rs/zerolog/log"
)

type ServiceEntry func() (func() error, error)

type Service struct {
	Name   string
	daemon bool
	start  ServiceEntry
	stop   func() error
}

type ServiceArgs struct {
	Entry    ServiceEntry
	NoDaemon bool
}

func NewService(args ServiceArgs) (*Service, error) {
	return &Service{
		Name:   config.AppName,
		daemon: !args.NoDaemon,
		start:  args.Entry,
	}, nil
}

// Create new PID file using current process PID.
func (s *Service) createPidFile() error {
	pid := os.Getpid()
	err := os.WriteFile(PidFile, []byte(fmt.Sprintf("%d", pid)), 0644)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) removePidFile() error {
	err := os.Remove(PidFile)
	if err != nil {
		return err
	}
	return nil
}

// Return the process ID of the current running service daemon.
func (s *Service) Pid() (int, error) {
	pid := 0

	if _, err := os.Stat(PidFile); err == nil {
		pidFile, err := os.ReadFile(PidFile)
		if err != nil {
			return pid, fmt.Errorf("error reading pid file: %w", err)
		}

		pidInt, err := strconv.Atoi(string(pidFile))
		if err != nil {
			return pid, fmt.Errorf("error parsing pid: %w", err)
		}

		pid = pidInt
	}

	return pid, nil
}

// Returns true if the service is running.
func (s *Service) Running() bool {
	pid, err := s.Pid()
	if err != nil {
		return false
	}

	if pid == 0 {
		return false
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = process.Signal(syscall.Signal(0))

	return err == nil
}

func (s *Service) stopService() error {
	log.Info().Msgf("stopping %s service", s.Name)

	err := s.stop()
	if err != nil {
		log.Error().Msgf("error stopping %s service: %s", s.Name, err)
		return err
	}

	err = s.removePidFile()
	if err != nil {
		log.Error().Msgf("error removing pid file: %s", err)
		return err
	}

	// remove temporary binary
	tempPath, err := os.Executable()
	if err != nil {
		log.Error().Msgf("error getting executable path: %s", err)
	} else if strings.HasPrefix(tempPath, TempFolder) {
		err = os.Remove(tempPath)
		if err != nil {
			log.Error().Msgf("error removing temporary binary: %s", err)
		}
	}

	return nil
}

// Set up signal handler to stop service on SIGINT or SIGTERM. Exits the application on signal.
func (s *Service) setupStopService() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs

		err := s.stopService()
		if err != nil {
			os.Exit(1)
		}

		os.Exit(0)
	}()
}

// Starts the service and blocks until the service is stopped.
func (s *Service) startService() {
	if s.Running() {
		log.Error().Msgf("%s service already running", s.Name)
		os.Exit(1)
	}

	log.Info().Msgf("starting %s service", s.Name)

	err := s.createPidFile()
	if err != nil {
		log.Error().Msgf("error creating pid file: %s", err)
		os.Exit(1)
	}

	err = syscall.Setpriority(syscall.PRIO_PROCESS, 0, 1)
	if err != nil {
		log.Error().Msgf("error setting nice level: %s", err)
	}

	stop, err := s.start()
	if err != nil {
		log.Error().Msgf("error starting service: %s", err)

		err = s.removePidFile()
		if err != nil {
			log.Error().Msgf("error removing pid file: %s", err)
		}

		os.Exit(1)
	}

	s.setupStopService()
	s.stop = stop

	if s.daemon {
		<-make(chan struct{})
	} else {
		err := s.stopService()
		if err != nil {
			os.Exit(1)
		}

		os.Exit(0)
	}
}

// Start a new service daemon in the background.
func (s *Service) Start() error {
	if s.Running() {
		return fmt.Errorf("%s service already running", s.Name)
	}

	// create a copy in binary in tmp so the original can be updated
	binPath := ""
	appPath := os.Getenv(config.UserAppPathEnv)
	if appPath != "" {
		binPath = appPath
	} else {
		exePath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("error getting absolute binary path: %w", err)
		}
		binPath = exePath
	}

	binFile, err := os.Open(binPath)
	if err != nil {
		return fmt.Errorf("error opening binary: %w", err)
	}

	tempPath := filepath.Join(TempFolder, filepath.Base(binPath))
	tempFile, err := os.OpenFile(tempPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("error creating temp binary: %w", err)
	}

	_, err = io.Copy(tempFile, binFile)
	if err != nil {
		return fmt.Errorf("error copying binary to temp: %w", err)
	}

	tempFile.Close()
	binFile.Close()

	cmd := exec.Command(tempPath, "-service", "exec", "&")
	env := os.Environ()
	cmd.Env = env

	// point new binary to existing config file
	configPath := filepath.Join(filepath.Dir(binPath), s.Name+".ini")

	if _, err := os.Stat(configPath); err == nil {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", config.UserConfigEnv, configPath))
	}
	cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", config.UserAppPathEnv, binPath))

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("error starting %s service: %w", s.Name, err)
	}

	return nil
}

// Stop the service daemon.
func (s *Service) Stop() error {
	if !s.Running() {
		return fmt.Errorf("%s service not running", s.Name)
	}

	pid, err := s.Pid()
	if err != nil {
		return err
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Restart() error {
	if s.Running() {
		err := s.Stop()
		if err != nil {
			return err
		}
	}

	for s.Running() {
		time.Sleep(1 * time.Second)
	}

	err := s.Start()
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) ServiceHandler(cmd *string) {
	if *cmd == "exec" {
		s.startService()
		os.Exit(0)
	} else if *cmd == "start" {
		err := s.Start()
		if err != nil {
			log.Error().Msg(err.Error())
			os.Exit(1)
		}

		os.Exit(0)
	} else if *cmd == "stop" {
		err := s.Stop()
		if err != nil {
			log.Error().Msg(err.Error())
			os.Exit(1)
		}

		os.Exit(0)
	} else if *cmd == "restart" {
		err := s.Restart()
		if err != nil {
			log.Error().Msg(err.Error())
			os.Exit(1)
		}

		os.Exit(0)
	} else if *cmd == "status" {
		if s.Running() {
			fmt.Printf("%s service running\n", s.Name)
		} else {
			fmt.Printf("%s service not running\n", s.Name)
		}

		os.Exit(0)
	} else if *cmd != "" {
		fmt.Printf("Invalid service command: %s", *cmd)
		os.Exit(1)
	}
}
