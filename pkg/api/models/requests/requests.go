package requests

import (
	"github.com/google/uuid"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/database"
	"github.com/wizzomafizzo/tapto/pkg/platforms"
	"github.com/wizzomafizzo/tapto/pkg/service/state"
	"github.com/wizzomafizzo/tapto/pkg/service/tokens"
)

type RequestEnv struct {
	Platform   platforms.Platform
	Config     *config.UserConfig
	State      *state.State
	Database   *database.Database
	TokenQueue chan<- tokens.Token
	IsLocal    bool
	Id         uuid.UUID
	Params     []byte
}
