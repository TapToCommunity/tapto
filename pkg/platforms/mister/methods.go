package mister

import (
	"os"
	"os/exec"

	"github.com/rs/zerolog/log"
	mrextMister "github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/tapto/pkg/assets"
	"github.com/wizzomafizzo/tapto/pkg/config"
)

// TODO: don't want to depend on external aplay command, but i'm out of
//       time to keep messing with this. oto/beep would not work for me
//       and are annoying to compile statically

func Setup() error {
	// copy success sound to temp
	sf, err := os.Create(SuccessSoundFile)
	if err != nil {
		log.Error().Msgf("error creating success sound file: %s", err)
	}
	_, err = sf.Write(assets.SuccessSound)
	if err != nil {
		log.Error().Msgf("error writing success sound file: %s", err)
	}
	_ = sf.Close()

	// copy fail sound to temp
	ff, err := os.Create(FailSoundFile)
	if err != nil {
		log.Error().Msgf("error creating fail sound file: %s", err)
	}
	_, err = ff.Write(assets.FailSound)
	if err != nil {
		log.Error().Msgf("error writing fail sound file: %s", err)
	}
	_ = ff.Close()

	return nil
}

func PlaySuccess(cfg *config.UserConfig) {
	if cfg.TapTo.DisableSounds {
		return
	}

	err := exec.Command("aplay", SuccessSoundFile).Start()
	if err != nil {
		log.Error().Msgf("error playing success sound: %s", err)
	}
}

func PlayFail(cfg *config.UserConfig) {
	if cfg.TapTo.DisableSounds {
		return
	}

	err := exec.Command("aplay", FailSoundFile).Start()
	if err != nil {
		log.Error().Msgf("error playing fail sound: %s", err)
	}
}

func ExitGame() {
	_ = mrextMister.LaunchMenu()
}

func GetActiveCoreName() string {
	coreName, err := mrextMister.GetActiveCoreName()
	if err != nil {
		log.Error().Msgf("error trying to get the core name: %s", err)
	}
	return coreName
}
