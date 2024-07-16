package gamesdb

import (
	"fmt"
	"path/filepath"
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
	Id         string
	Aliases    []string
	Folders    []string
	Extensions []string
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

// MatchSystemFile returns true if a given file's extension is valid for a system.
func MatchSystemFile(system System, path string) bool {
	// ignore dot files
	if strings.HasPrefix(filepath.Base(path), ".") {
		return false
	}

	for _, ext := range system.Extensions {
		if strings.HasSuffix(strings.ToLower(path), ext) {
			return true
		}
	}

	return false
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
		Id:         SystemAdventureVision,
		Aliases:    []string{"AVision"},
		Folders:    []string{"AVision"},
		Extensions: []string{".bin"},
	},
	SystemArcadia: {
		Id:         SystemArcadia,
		Folders:    []string{"Arcadia"},
		Extensions: []string{".bin"},
	},
	SystemAstrocade: {
		Id:         SystemAstrocade,
		Folders:    []string{"Astrocade"},
		Extensions: []string{".bin"},
	},
	SystemAtari2600: {
		Id:         SystemAtari2600,
		Folders:    []string{"ATARI7800", "Atari2600"},
		Extensions: []string{".a26"},
	},
	SystemAtari5200: {
		Id:         SystemAtari5200,
		Folders:    []string{"ATARI5200"},
		Extensions: []string{".a52"},
	},
	SystemAtari7800: {
		Id:         SystemAtari7800,
		Folders:    []string{"ATARI7800"},
		Extensions: []string{".a78"},
	},
	SystemAtariLynx: {
		Id:         SystemAtariLynx,
		Folders:    []string{"AtariLynx"},
		Extensions: []string{".lnx"},
	},
	SystemCasioPV1000: {
		Id:         SystemCasioPV1000,
		Aliases:    []string{"Casio_PV-1000"},
		Folders:    []string{"Casio_PV-1000"},
		Extensions: []string{".bin"},
	},
	SystemChannelF: {
		Id:         SystemChannelF,
		Folders:    []string{"ChannelF"},
		Extensions: []string{".rom", ".bin"},
	},
	SystemColecoVision: {
		Id:         SystemColecoVision,
		Aliases:    []string{"Coleco"},
		Folders:    []string{"Coleco"},
		Extensions: []string{".col", ".bin", ".rom"},
	},
	SystemCreatiVision: {
		Id:         SystemCreatiVision,
		Folders:    []string{"CreatiVision"},
		Extensions: []string{".rom", ".bin", ".bas"},
	},
	SystemFDS: {
		Id:         SystemFDS,
		Aliases:    []string{"FamicomDiskSystem"},
		Folders:    []string{"NES", "FDS"},
		Extensions: []string{".fds"},
	},
	SystemGamate: {
		Id:         SystemGamate,
		Folders:    []string{"Gamate"},
		Extensions: []string{".bin"},
	},
	SystemGameboy: {
		Id:         SystemGameboy,
		Aliases:    []string{"GB"},
		Folders:    []string{"GAMEBOY"},
		Extensions: []string{".gb"},
	},
	SystemGameboyColor: {
		Id:         SystemGameboyColor,
		Aliases:    []string{"GBC"},
		Folders:    []string{"GAMEBOY", "GBC"},
		Extensions: []string{".gbc"},
	},
	SystemGameboy2P: {
		// TODO: Split 2P core into GB and GBC?
		Id:         SystemGameboy2P,
		Folders:    []string{"GAMEBOY2P"},
		Extensions: []string{".gb", ".gbc"},
	},
	SystemGameGear: {
		Id:         SystemGameGear,
		Aliases:    []string{"GG"},
		Folders:    []string{"SMS", "GameGear"},
		Extensions: []string{".gg"},
	},
	SystemGameNWatch: {
		Id:         SystemGameNWatch,
		Folders:    []string{"GameNWatch"},
		Extensions: []string{".bin"},
	},
	SystemGBA: {
		Id:         SystemGBA,
		Aliases:    []string{"GameboyAdvance"},
		Folders:    []string{"GBA"},
		Extensions: []string{".gba"},
	},
	SystemGBA2P: {
		Id:         SystemGBA2P,
		Folders:    []string{"GBA2P"},
		Extensions: []string{".gba"},
	},
	SystemGenesis: {
		Id:         SystemGenesis,
		Aliases:    []string{"MegaDrive"},
		Folders:    []string{"MegaDrive", "Genesis"},
		Extensions: []string{".gen", ".bin", ".md"},
	},
	SystemIntellivision: {
		Id:         SystemIntellivision,
		Folders:    []string{"Intellivision"},
		Extensions: []string{".int", ".bin"},
	},
	// TODO: Jaguar
	SystemMasterSystem: {
		Id:         SystemMasterSystem,
		Aliases:    []string{"SMS"},
		Folders:    []string{"SMS"},
		Extensions: []string{".sms"},
	},
	SystemMegaCD: {
		Id:         SystemMegaCD,
		Aliases:    []string{"SegaCD"},
		Folders:    []string{"MegaCD"},
		Extensions: []string{".cue", ".chd"},
	},
	SystemMegaDuck: {
		Id:         SystemMegaDuck,
		Folders:    []string{"GAMEBOY", "MegaDuck"},
		Extensions: []string{".bin"},
	},
	SystemNeoGeo: {
		Id:         SystemNeoGeo,
		Folders:    []string{"NEOGEO"},
		Extensions: []string{".neo"}, // TODO: .zip and folder support
	},
	SystemNeoGeoCD: {
		Id:         SystemNeoGeo,
		Folders:    []string{"NeoGeo-CD", "NEOGEO"},
		Extensions: []string{".cue", ".chd"},
	},
	SystemNES: {
		Id:         SystemNES,
		Folders:    []string{"NES"},
		Extensions: []string{".nes"},
	},
	SystemNESMusic: {
		Id:         SystemNESMusic,
		Folders:    []string{"NES"},
		Extensions: []string{".nsf"},
	},
	SystemNintendo64: {
		Id:         SystemNintendo64,
		Aliases:    []string{"N64"},
		Folders:    []string{"N64"},
		Extensions: []string{".n64", ".z64"},
	},
	SystemOdyssey2: {
		Id:         SystemOdyssey2,
		Folders:    []string{"ODYSSEY2"},
		Extensions: []string{".bin"},
	},
	SystemPocketChallengeV2: {
		Id:         SystemPocketChallengeV2,
		Folders:    []string{"WonderSwan", "PocketChallengeV2"},
		Extensions: []string{".pc2"},
	},
	SystemPokemonMini: {
		Id:         SystemPokemonMini,
		Folders:    []string{"PokemonMini"},
		Extensions: []string{".min"},
	},
	SystemPSX: {
		Id:         SystemPSX,
		Aliases:    []string{"Playstation", "PS1"},
		Folders:    []string{"PSX"},
		Extensions: []string{".cue", ".chd", ".exe"},
	},
	SystemSega32X: {
		Id:         SystemSega32X,
		Aliases:    []string{"S32X", "32X"},
		Folders:    []string{"S32X"},
		Extensions: []string{".32x"},
	},
	SystemSG1000: {
		Id:         SystemSG1000,
		Folders:    []string{"SG1000", "Coleco", "SMS"},
		Extensions: []string{".sg"},
	},
	SystemSuperGameboy: {
		Id:         SystemSuperGameboy,
		Aliases:    []string{"SGB"},
		Folders:    []string{"SGB"},
		Extensions: []string{".sgb", ".gb", ".gbc"},
	},
	SystemSuperVision: {
		Id:         SystemSuperVision,
		Folders:    []string{"SuperVision"},
		Extensions: []string{".bin", ".sv"},
	},
	SystemSaturn: {
		Id:         SystemSaturn,
		Folders:    []string{"Saturn"},
		Extensions: []string{".cue", ".chd"},
	},
	SystemSNES: {
		Id:         SystemSNES,
		Aliases:    []string{"SuperNintendo"},
		Folders:    []string{"SNES"},
		Extensions: []string{".sfc", ".smc", ".bin", ".bs"},
	},
	SystemSNESMusic: {
		Id:         SystemSNESMusic,
		Folders:    []string{"SNES"},
		Extensions: []string{".spc"},
	},
	SystemSuperGrafx: {
		Id:         SystemSuperGrafx,
		Folders:    []string{"TGFX16"},
		Extensions: []string{".sgx"},
	},
	SystemTurboGrafx16: {
		Id:         SystemTurboGrafx16,
		Aliases:    []string{"TGFX16", "PCEngine"},
		Folders:    []string{"TGFX16"},
		Extensions: []string{".pce", ".bin"},
	},
	SystemTurboGrafx16CD: {
		Id:         SystemTurboGrafx16CD,
		Aliases:    []string{"TGFX16-CD", "PCEngineCD"},
		Folders:    []string{"TGFX16-CD"},
		Extensions: []string{".cue", ".chd"},
	},
	SystemVC4000: {
		Id:         SystemVC4000,
		Folders:    []string{"VC4000"},
		Extensions: []string{".bin"},
	},
	SystemVectrex: {
		Id:         SystemVectrex,
		Folders:    []string{"VECTREX"},
		Extensions: []string{".vec", ".bin", ".rom"}, // TODO: overlays (.ovr)
	},
	SystemWonderSwan: {
		Id:         SystemWonderSwan,
		Folders:    []string{"WonderSwan"},
		Extensions: []string{".ws"},
	},
	SystemWonderSwanColor: {
		Id:         SystemWonderSwanColor,
		Folders:    []string{"WonderSwan", "WonderSwanColor"},
		Extensions: []string{".wsc"},
	},
	// Computers
	SystemAcornAtom: {
		Id:         SystemAcornAtom,
		Folders:    []string{"AcornAtom"},
		Extensions: []string{".vhd"},
	},
	SystemAcornElectron: {
		Id:         SystemAcornElectron,
		Folders:    []string{"AcornElectron"},
		Extensions: []string{".vhd"},
	},
	SystemAliceMC10: {
		Id:         SystemAliceMC10,
		Folders:    []string{"AliceMC10"},
		Extensions: []string{".c10"},
	},
	SystemAmiga: {
		Id:         SystemAmiga,
		Folders:    []string{"Amiga"},
		Aliases:    []string{"Minimig"},
		Extensions: []string{".adf"},
	},
	SystemAmstrad: {
		Id:         SystemAmstrad,
		Folders:    []string{"Amstrad"},
		Extensions: []string{".dsk", ".cdt"}, // TODO: globbing support? for .e??
	},
	SystemAmstradPCW: {
		Id:         SystemAmstradPCW,
		Aliases:    []string{"Amstrad-PCW"},
		Folders:    []string{"Amstrad PCW"},
		Extensions: []string{".dsk"},
	},
	SystemAo486: {
		Id:         SystemAo486,
		Folders:    []string{"AO486"},
		Extensions: []string{".img", ".ima", ".vhd", ".vfd", ".iso", ".cue", ".chd"},
	},
	SystemApogee: {
		Id:         SystemApogee,
		Folders:    []string{"APOGEE"},
		Extensions: []string{".rka", ".rkr", ".gam"},
	},
	SystemAppleI: {
		Id:         SystemAppleI,
		Aliases:    []string{"Apple-I"},
		Folders:    []string{"Apple-I"},
		Extensions: []string{".txt"},
	},
	SystemAppleII: {
		Id:         SystemAppleII,
		Aliases:    []string{"Apple-II"},
		Folders:    []string{"Apple-II"},
		Extensions: []string{".dsk", ".do", ".po", ".nib", ".hdv"},
	},
	SystemAquarius: {
		Id:         SystemAquarius,
		Folders:    []string{"AQUARIUS"},
		Extensions: []string{".bin", ".caq"},
	},
	SystemAtari800: {
		Id:         SystemAtari800,
		Folders:    []string{"ATARI800"},
		Extensions: []string{".atr", ".xex", ".xfd", ".atx", ".car", ".rom", ".bin"},
	},
	SystemBBCMicro: {
		Id:         SystemBBCMicro,
		Folders:    []string{"BBCMicro"},
		Extensions: []string{".ssd", ".dsd", ".vhd"},
	},
	SystemBK0011M: {
		Id:         SystemBK0011M,
		Folders:    []string{"BK0011M"},
		Extensions: []string{".bin", ".dsk", ".vhd"},
	},
	SystemC16: {
		Id:         SystemC16,
		Folders:    []string{"C16"},
		Extensions: []string{".d64", ".g64", ".prg", ".tap", ".bin"},
	},
	SystemC64: {
		Id:         SystemC64,
		Folders:    []string{"C64"},
		Extensions: []string{".d64", ".g64", ".t64", ".d81", ".prg", ".crt", ".reu", ".tap"},
	},
	SystemCasioPV2000: {
		Id:         SystemCasioPV2000,
		Aliases:    []string{"Casio_PV-2000"},
		Folders:    []string{"Casio_PV-2000"},
		Extensions: []string{".bin"},
	},
	SystemCoCo2: {
		Id:         SystemCoCo2,
		Folders:    []string{"CoCo2"},
		Extensions: []string{".dsk", ".cas", ".ccc", ".rom"},
	},
	SystemEDSAC: {
		Id:         SystemEDSAC,
		Folders:    []string{"EDSAC"},
		Extensions: []string{".tap"},
	},
	SystemGalaksija: {
		Id:         SystemGalaksija,
		Folders:    []string{"Galaksija"},
		Extensions: []string{".tap"},
	},
	SystemInteract: {
		Id:         SystemInteract,
		Folders:    []string{"Interact"},
		Extensions: []string{".cin", ".k7"},
	},
	SystemJupiter: {
		Id:         SystemJupiter,
		Folders:    []string{"Jupiter"},
		Extensions: []string{".ace"},
	},
	SystemLaser: {
		Id:         SystemLaser,
		Aliases:    []string{"Laser310"},
		Folders:    []string{"Laser"},
		Extensions: []string{".vz"},
	},
	SystemLynx48: {
		Id:         SystemLynx48,
		Folders:    []string{"Lynx48"},
		Extensions: []string{".tap"},
	},
	SystemMacPlus: {
		Id:         SystemMacPlus,
		Folders:    []string{"MACPLUS"},
		Extensions: []string{".dsk", ".img", ".vhd"},
	},
	SystemMSX: {
		Id:         SystemMSX,
		Folders:    []string{"MSX"},
		Extensions: []string{".vhd"},
	},
	SystemMultiComp: {
		Id:         SystemMultiComp,
		Folders:    []string{"MultiComp"},
		Extensions: []string{".img"},
	},
	SystemOrao: {
		Id:         SystemOrao,
		Folders:    []string{"ORAO"},
		Extensions: []string{".tap"},
	},
	SystemOric: {
		Id:         SystemOric,
		Folders:    []string{"Oric"},
		Extensions: []string{".dsk"},
	},
	SystemPCXT: {
		Id:         SystemPCXT,
		Folders:    []string{"PCXT"},
		Extensions: []string{".img", ".vhd", ".ima", ".vfd"},
	},
	SystemPDP1: {
		Id:         SystemPDP1,
		Folders:    []string{"PDP1"},
		Extensions: []string{".bin", ".rim", ".pdp"},
	},
	SystemPET2001: {
		Id:         SystemPET2001,
		Folders:    []string{"PET2001"},
		Extensions: []string{".prg", ".tap"},
	},
	SystemPMD85: {
		Id:         SystemPMD85,
		Folders:    []string{"PMD85"},
		Extensions: []string{".rmm"},
	},
	SystemQL: {
		Id:         SystemQL,
		Folders:    []string{"QL"},
		Extensions: []string{".mdv", ".win"},
	},
	SystemRX78: {
		Id:         SystemRX78,
		Folders:    []string{"RX78"},
		Extensions: []string{".bin"},
	},
	SystemSAMCoupe: {
		Id:         SystemSAMCoupe,
		Folders:    []string{"SAMCOUPE"},
		Extensions: []string{".dsk", ".mgt", ".img"},
	},
	SystemSordM5: {
		Id:         SystemSordM5,
		Aliases:    []string{"Sord M5"},
		Folders:    []string{"Sord M5"},
		Extensions: []string{".bin", ".rom", ".cas"},
	},
	SystemSpecialist: {
		Id:         SystemSpecialist,
		Aliases:    []string{"SPMX"},
		Folders:    []string{"SPMX"},
		Extensions: []string{".rks", ".odi"},
	},
	SystemSVI328: {
		Id:         SystemSVI328,
		Folders:    []string{"SVI328"},
		Extensions: []string{".cas", ".bin", ".rom"},
	},
	SystemTatungEinstein: {
		Id:         SystemTatungEinstein,
		Folders:    []string{"TatungEinstein"},
		Extensions: []string{".dsk"},
	},
	SystemTI994A: {
		Id:         SystemTI994A,
		Aliases:    []string{"TI-99_4A"},
		Folders:    []string{"TI-99_4A"},
		Extensions: []string{".bin", ".m99"},
	},
	SystemTomyTutor: {
		Id:         SystemTomyTutor,
		Folders:    []string{"TomyTutor"},
		Extensions: []string{".bin", ".cas"},
	},
	SystemTRS80: {
		Id:         SystemTRS80,
		Folders:    []string{"TRS-80"},
		Extensions: []string{".jvi", ".dsk", ".cas"},
	},
	SystemTSConf: {
		Id:         SystemTSConf,
		Folders:    []string{"TSConf"},
		Extensions: []string{".vhf"},
	},
	SystemUK101: {
		Id:         SystemUK101,
		Folders:    []string{"UK101"},
		Extensions: []string{".txt", ".bas", ".lod"},
	},
	SystemVector06C: {
		Id:         SystemVector06C,
		Aliases:    []string{"Vector06"},
		Folders:    []string{"VECTOR06"},
		Extensions: []string{".rom", ".com", ".c00", ".edd", ".fdd"},
	},
	SystemVIC20: {
		Id:         SystemVIC20,
		Folders:    []string{"VIC20"},
		Extensions: []string{".d64", ".g64", ".prg", ".tap", ".crt"},
	},
	SystemX68000: {
		Id:         SystemX68000,
		Folders:    []string{"X68000"},
		Extensions: []string{".d88", ".hdf"},
	},
	SystemZX81: {
		Id:         SystemZX81,
		Folders:    []string{"ZX81"},
		Extensions: []string{".p", ".0"},
	},
	SystemZXSpectrum: {
		Id:         SystemZXSpectrum,
		Aliases:    []string{"Spectrum"},
		Folders:    []string{"Spectrum"},
		Extensions: []string{".tap", ".csw", ".tzx", ".sna", ".z80", ".trd", ".img", ".dsk", ".mgt"},
	},
	SystemZXNext: {
		Id:         SystemZXNext,
		Folders:    []string{"ZXNext"},
		Extensions: []string{".vhd"},
	},
	// Other
	SystemArcade: {
		Id:         SystemArcade,
		Folders:    []string{"_Arcade"},
		Extensions: []string{".mra"},
	},
	SystemArduboy: {
		Id:         SystemArduboy,
		Folders:    []string{"Arduboy"},
		Extensions: []string{".hex", ".bin"},
	},
	SystemChip8: {
		Id:         SystemChip8,
		Folders:    []string{"Chip8"},
		Extensions: []string{".ch8"},
	},
}
