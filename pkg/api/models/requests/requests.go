package requests

import (
	"github.com/ZaparooProject/zaparoo-core/pkg/config"
	"github.com/ZaparooProject/zaparoo-core/pkg/database"
	"github.com/ZaparooProject/zaparoo-core/pkg/platforms"
	"github.com/ZaparooProject/zaparoo-core/pkg/service/state"
	"github.com/ZaparooProject/zaparoo-core/pkg/service/tokens"
	"github.com/google/uuid"
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
