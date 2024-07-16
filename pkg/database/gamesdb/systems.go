package gamesdb

import (
	"fmt"
	"strings"

	"github.com/wizzomafizzo/tapto/pkg/utils"
)

// The Systems list contains all the supported systems such as consoles,
// computers and media types that are indexable by TapTo. This is the reference
// list of hardcoded system IDs used throughout TapTo. A platform can choose
// not to support any of them.
//
// This list also contains some basic heuristics which, given a file path, can
// be used to attempt to associate a file with a system.

type System struct {
	Id      string
	Aliases []string
}

// GetSystem looks up an exact system definition by ID.
func GetSystem(id string) (*System, error) {
	if system, ok := Systems[id]; ok {
		return &system, nil
	} else {
		return nil, fmt.Errorf("unknown system: %s", id)
	}
}

// LookupSystem case-insensitively looks up system ID definition including aliases.
func LookupSystem(id string) (*System, error) {
	for k, v := range Systems {
		if strings.EqualFold(k, id) {
			return &v, nil
		}

		for _, alias := range v.Aliases {
			if strings.EqualFold(alias, id) {
				return &v, nil
			}
		}
	}

	return nil, fmt.Errorf("unknown system: %s", id)
}

func AllSystems() []System {
	var systems []System

	keys := utils.AlphaMapKeys(Systems)
	for _, k := range keys {
		systems = append(systems, Systems[k])
	}

	return systems
}

const (
	// Consoles
	SystemAdventureVision   = "AdventureVision"
	SystemArcadia           = "Arcadia"
	SystemAstrocade         = "Astrocade"
	SystemAtari2600         = "Atari2600"
	SystemAtari5200         = "Atari5200"
	SystemAtari7800         = "Atari7800"
	SystemAtariLynx         = "AtariLynx"
	SystemCasioPV1000       = "CasioPV1000"
	SystemChannelF          = "ChannelF"
	SystemColecoVision      = "ColecoVision"
	SystemCreatiVision      = "CreatiVision"
	SystemFDS               = "FDS"
	SystemGamate            = "Gamate"
	SystemGameboy           = "Gameboy"
	SystemGameboyColor      = "GameboyColor"
	SystemGameboy2P         = "Gameboy2P"
	SystemGameGear          = "GameGear"
	SystemGameNWatch        = "GameNWatch"
	SystemGBA               = "GBA"
	SystemGBA2P             = "GBA2P"
	SystemGenesis           = "Genesis"
	SystemIntellivision     = "Intellivision"
	SystemMasterSystem      = "MasterSystem"
	SystemMegaCD            = "MegaCD"
	SystemMegaDuck          = "MegaDuck"
	SystemNeoGeo            = "NeoGeo"
	SystemNeoGeoCD          = "NeoGeoCD"
	SystemNES               = "NES"
	SystemNESMusic          = "NESMusic"
	SystemNintendo64        = "Nintendo64"
	SystemOdyssey2          = "Odyssey2"
	SystemPocketChallengeV2 = "PocketChallengeV2"
	SystemPokemonMini       = "PokemonMini"
	SystemPSX               = "PSX"
	SystemSega32X           = "Sega32X"
	SystemSG1000            = "SG1000"
	SystemSuperGameboy      = "SuperGameboy"
	SystemSuperVision       = "SuperVision"
	SystemSaturn            = "Saturn"
	SystemSNES              = "SNES"
	SystemSNESMusic         = "SNESMusic"
	SystemSuperGrafx        = "SuperGrafx"
	SystemTurboGrafx16      = "TurboGrafx16"
	SystemTurboGrafx16CD    = "TurboGrafx16CD"
	SystemVC4000            = "VC4000"
	SystemVectrex           = "Vectrex"
	SystemWonderSwan        = "WonderSwan"
	SystemWonderSwanColor   = "WonderSwanColor"
	// Computers
	SystemAcornAtom      = "AcornAtom"
	SystemAcornElectron  = "AcornElectron"
	SystemAliceMC10      = "AliceMC10"
	SystemAmiga          = "Amiga"
	SystemAmstrad        = "Amstrad"
	SystemAmstradPCW     = "AmstradPCW"
	SystemAo486          = "ao486" // TODO: should this be PC?
	SystemApogee         = "Apogee"
	SystemAppleI         = "AppleI"
	SystemAppleII        = "AppleII"
	SystemAquarius       = "Aquarius"
	SystemAtari800       = "Atari800"
	SystemBBCMicro       = "BBCMicro"
	SystemBK0011M        = "BK0011M"
	SystemC16            = "C16"
	SystemC64            = "C64"
	SystemCasioPV2000    = "CasioPV2000"
	SystemCoCo2          = "CoCo2"
	SystemEDSAC          = "EDSAC"
	SystemGalaksija      = "Galaksija"
	SystemInteract       = "Interact"
	SystemJupiter        = "Jupiter"
	SystemLaser          = "Laser"
	SystemLynx48         = "Lynx48"
	SystemMacPlus        = "MacPlus"
	SystemMSX            = "MSX"
	SystemMultiComp      = "MultiComp"
	SystemOrao           = "Orao"
	SystemOric           = "Oric"
	SystemPCXT           = "PCXT"
	SystemPDP1           = "PDP1"
	SystemPET2001        = "PET2001"
	SystemPMD85          = "PMD85"
	SystemQL             = "QL"
	SystemRX78           = "RX78"
	SystemSAMCoupe       = "SAMCoupe"
	SystemSordM5         = "SordM5"
	SystemSpecialist     = "Specialist"
	SystemSVI328         = "SVI328"
	SystemTatungEinstein = "TatungEinstein"
	SystemTI994A         = "TI994A"
	SystemTomyTutor      = "TomyTutor"
	SystemTRS80          = "TRS80"
	SystemTSConf         = "TSConf"
	SystemUK101          = "UK101"
	SystemVector06C      = "Vector06C"
	SystemVIC20          = "VIC20"
	SystemX68000         = "X68000"
	SystemZX81           = "ZX81"
	SystemZXSpectrum     = "ZXSpectrum"
	SystemZXNext         = "ZXNext"
	// Other
	SystemArcade  = "Arcade"
	SystemArduboy = "Arduboy"
	SystemChip8   = "Chip8"
)

var Systems = map[string]System{
	// Consoles
	SystemAdventureVision: {
		Id:      SystemAdventureVision,
		Aliases: []string{"AVision"},
	},
	SystemArcadia: {
		Id: SystemArcadia,
	},
	SystemAstrocade: {
		Id: SystemAstrocade,
	},
	SystemAtari2600: {
		Id: SystemAtari2600,
	},
	SystemAtari5200: {
		Id: SystemAtari5200,
	},
	SystemAtari7800: {
		Id: SystemAtari7800,
	},
	SystemAtariLynx: {
		Id: SystemAtariLynx,
	},
	SystemCasioPV1000: {
		Id:      SystemCasioPV1000,
		Aliases: []string{"Casio_PV-1000"},
	},
	SystemChannelF: {
		Id: SystemChannelF,
	},
	SystemColecoVision: {
		Id:      SystemColecoVision,
		Aliases: []string{"Coleco"},
	},
	SystemCreatiVision: {
		Id: SystemCreatiVision,
	},
	SystemFDS: {
		Id:      SystemFDS,
		Aliases: []string{"FamicomDiskSystem"},
	},
	SystemGamate: {
		Id: SystemGamate,
	},
	SystemGameboy: {
		Id:      SystemGameboy,
		Aliases: []string{"GB"},
	},
	SystemGameboyColor: {
		Id:      SystemGameboyColor,
		Aliases: []string{"GBC"},
	},
	SystemGameboy2P: {
		// TODO: Split 2P core into GB and GBC?
		Id: SystemGameboy2P,
	},
	SystemGameGear: {
		Id:      SystemGameGear,
		Aliases: []string{"GG"},
	},
	SystemGameNWatch: {
		Id: SystemGameNWatch,
	},
	SystemGBA: {
		Id:      SystemGBA,
		Aliases: []string{"GameboyAdvance"},
	},
	SystemGBA2P: {
		Id: SystemGBA2P,
	},
	SystemGenesis: {
		Id:      SystemGenesis,
		Aliases: []string{"MegaDrive"},
	},
	SystemIntellivision: {
		Id: SystemIntellivision,
	},
	// TODO: Jaguar
	SystemMasterSystem: {
		Id:      SystemMasterSystem,
		Aliases: []string{"SMS"},
	},
	SystemMegaCD: {
		Id:      SystemMegaCD,
		Aliases: []string{"SegaCD"},
	},
	SystemMegaDuck: {
		Id: SystemMegaDuck,
	},
	SystemNeoGeo: {
		Id: SystemNeoGeo,
	},
	SystemNeoGeoCD: {
		Id: SystemNeoGeo,
	},
	SystemNES: {
		Id: SystemNES,
	},
	SystemNESMusic: {
		Id: SystemNESMusic,
	},
	SystemNintendo64: {
		Id:      SystemNintendo64,
		Aliases: []string{"N64"},
	},
	SystemOdyssey2: {
		Id: SystemOdyssey2,
	},
	SystemPocketChallengeV2: {
		Id: SystemPocketChallengeV2,
	},
	SystemPokemonMini: {
		Id: SystemPokemonMini,
	},
	SystemPSX: {
		Id:      SystemPSX,
		Aliases: []string{"Playstation", "PS1"},
	},
	SystemSega32X: {
		Id:      SystemSega32X,
		Aliases: []string{"S32X", "32X"},
	},
	SystemSG1000: {
		Id: SystemSG1000,
	},
	SystemSuperGameboy: {
		Id:      SystemSuperGameboy,
		Aliases: []string{"SGB"},
	},
	SystemSuperVision: {
		Id: SystemSuperVision,
	},
	SystemSaturn: {
		Id: SystemSaturn,
	},
	SystemSNES: {
		Id:      SystemSNES,
		Aliases: []string{"SuperNintendo"},
	},
	SystemSNESMusic: {
		Id: SystemSNESMusic,
	},
	SystemSuperGrafx: {
		Id: SystemSuperGrafx,
	},
	SystemTurboGrafx16: {
		Id:      SystemTurboGrafx16,
		Aliases: []string{"TGFX16", "PCEngine"},
	},
	SystemTurboGrafx16CD: {
		Id:      SystemTurboGrafx16CD,
		Aliases: []string{"TGFX16-CD", "PCEngineCD"},
	},
	SystemVC4000: {
		Id: SystemVC4000,
	},
	SystemVectrex: {
		Id: SystemVectrex,
	},
	SystemWonderSwan: {
		Id: SystemWonderSwan,
	},
	SystemWonderSwanColor: {
		Id: SystemWonderSwanColor,
	},
	// Computers
	SystemAcornAtom: {
		Id: SystemAcornAtom,
	},
	SystemAcornElectron: {
		Id: SystemAcornElectron,
	},
	SystemAliceMC10: {
		Id: SystemAliceMC10,
	},
	SystemAmiga: {
		Id:      SystemAmiga,
		Aliases: []string{"Minimig"},
	},
	SystemAmstrad: {
		Id: SystemAmstrad,
	},
	SystemAmstradPCW: {
		Id:      SystemAmstradPCW,
		Aliases: []string{"Amstrad-PCW"},
	},
	SystemAo486: {
		Id: SystemAo486,
	},
	SystemApogee: {
		Id: SystemApogee,
	},
	SystemAppleI: {
		Id:      SystemAppleI,
		Aliases: []string{"Apple-I"},
	},
	SystemAppleII: {
		Id:      SystemAppleII,
		Aliases: []string{"Apple-II"},
	},
	SystemAquarius: {
		Id: SystemAquarius,
	},
	SystemAtari800: {
		Id: SystemAtari800,
	},
	SystemBBCMicro: {
		Id: SystemBBCMicro,
	},
	SystemBK0011M: {
		Id: SystemBK0011M,
	},
	SystemC16: {
		Id: SystemC16,
	},
	SystemC64: {
		Id: SystemC64,
	},
	SystemCasioPV2000: {
		Id:      SystemCasioPV2000,
		Aliases: []string{"Casio_PV-2000"},
	},
	SystemCoCo2: {
		Id: SystemCoCo2,
	},
	SystemEDSAC: {
		Id: SystemEDSAC,
	},
	SystemGalaksija: {
		Id: SystemGalaksija,
	},
	SystemInteract: {
		Id: SystemInteract,
	},
	SystemJupiter: {
		Id: SystemJupiter,
	},
	SystemLaser: {
		Id:      SystemLaser,
		Aliases: []string{"Laser310"},
	},
	SystemLynx48: {
		Id: SystemLynx48,
	},
	SystemMacPlus: {
		Id: SystemMacPlus,
	},
	SystemMSX: {
		Id: SystemMSX,
	},
	SystemMultiComp: {
		Id: SystemMultiComp,
	},
	SystemOrao: {
		Id: SystemOrao,
	},
	SystemOric: {
		Id: SystemOric,
	},
	SystemPCXT: {
		Id: SystemPCXT,
	},
	SystemPDP1: {
		Id: SystemPDP1,
	},
	SystemPET2001: {
		Id: SystemPET2001,
	},
	SystemPMD85: {
		Id: SystemPMD85,
	},
	SystemQL: {
		Id: SystemQL,
	},
	SystemRX78: {
		Id: SystemRX78,
	},
	SystemSAMCoupe: {
		Id: SystemSAMCoupe,
	},
	SystemSordM5: {
		Id:      SystemSordM5,
		Aliases: []string{"Sord M5"},
	},
	SystemSpecialist: {
		Id:      SystemSpecialist,
		Aliases: []string{"SPMX"},
	},
	SystemSVI328: {
		Id: SystemSVI328,
	},
	SystemTatungEinstein: {
		Id: SystemTatungEinstein,
	},
	SystemTI994A: {
		Id:      SystemTI994A,
		Aliases: []string{"TI-99_4A"},
	},
	SystemTomyTutor: {
		Id: SystemTomyTutor,
	},
	SystemTRS80: {
		Id: SystemTRS80,
	},
	SystemTSConf: {
		Id: SystemTSConf,
	},
	SystemUK101: {
		Id: SystemUK101,
	},
	SystemVector06C: {
		Id:      SystemVector06C,
		Aliases: []string{"Vector06"},
	},
	SystemVIC20: {
		Id: SystemVIC20,
	},
	SystemX68000: {
		Id: SystemX68000,
	},
	SystemZX81: {
		Id: SystemZX81,
	},
	SystemZXSpectrum: {
		Id:      SystemZXSpectrum,
		Aliases: []string{"Spectrum"},
	},
	SystemZXNext: {
		Id: SystemZXNext,
	},
	// Other
	SystemArcade: {
		Id: SystemArcade,
	},
	SystemArduboy: {
		Id: SystemArduboy,
	},
	SystemChip8: {
		Id: SystemChip8,
	},
}
