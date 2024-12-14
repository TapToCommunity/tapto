/*
TapTo
Copyright (C) 2023 Gareth Jones
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

package service

import (
	"github.com/ZaparooProject/zaparoo-core/pkg/service/tokens"
	"regexp"
	"strings"

	"github.com/ZaparooProject/zaparoo-core/pkg/database"
	"github.com/ZaparooProject/zaparoo-core/pkg/platforms"
	"github.com/rs/zerolog/log"
)

func checkMappingUid(m database.Mapping, t tokens.Token) bool {
	uid := database.NormalizeUid(t.UID)

	switch {
	case m.Match == database.MatchTypeExact:
		return uid == m.Pattern
	case m.Match == database.MatchTypePartial:
		return strings.Contains(uid, m.Pattern)
	case m.Match == database.MatchTypeRegex:
		re, err := regexp.Compile(m.Pattern)
		if err != nil {
			log.Error().Err(err).Msgf("error compiling regex")
			return false
		}
		return re.MatchString(uid)
	}

	return false
}

func checkMappingText(m database.Mapping, t tokens.Token) bool {
	switch {
	case m.Match == database.MatchTypeExact:
		return t.Text == m.Pattern
	case m.Match == database.MatchTypePartial:
		return strings.Contains(t.Text, m.Pattern)
	case m.Match == database.MatchTypeRegex:
		re, err := regexp.Compile(m.Pattern)
		if err != nil {
			log.Error().Err(err).Msgf("error compiling regex")
			return false
		}
		return re.MatchString(t.Text)
	}

	return false
}

func checkMappingData(m database.Mapping, t tokens.Token) bool {
	switch {
	case m.Match == database.MatchTypeExact:
		return t.Data == m.Pattern
	case m.Match == database.MatchTypePartial:
		return strings.Contains(t.Data, m.Pattern)
	case m.Match == database.MatchTypeRegex:
		re, err := regexp.Compile(m.Pattern)
		if err != nil {
			log.Error().Err(err).Msgf("error compiling regex")
			return false
		}
		return re.MatchString(t.Data)
	}

	return false
}

func getMapping(db *database.Database, pl platforms.Platform, token tokens.Token) (string, bool) {
	// check db mappings
	ms, err := db.GetEnabledMappings()
	if err != nil {
		log.Error().Err(err).Msgf("error getting db mappings")
	}

	for _, m := range ms {
		switch {
		case m.Type == database.MappingTypeUID:
			if checkMappingUid(m, token) {
				log.Info().Msg("launching with db uid match override")
				return m.Override, true
			}
		case m.Type == database.MappingTypeText:
			if checkMappingText(m, token) {
				log.Info().Msg("launching with db text match override")
				return m.Override, true
			}
		case m.Type == database.MappingTypeData:
			if checkMappingData(m, token) {
				log.Info().Msg("launching with db data match override")
				return m.Override, true
			}
		}
	}

	// check platform mappings
	return pl.LookupMapping(token)
}
