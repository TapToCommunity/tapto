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

var Systems = map[string]System{
	// Consoles
	"AdventureVision": {
		Id:         "AdventureVision",
		Aliases:    []string{"AVision"},
		Folders:    []string{"AVision"},
		Extensions: []string{".bin"},
	},
	"Arcadia": {
		Id:         "Arcadia",
		Folders:    []string{"Arcadia"},
		Extensions: []string{".bin"},
	},
	"Astrocade": {
		Id:         "Astrocade",
		Folders:    []string{"Astrocade"},
		Extensions: []string{".bin"},
	},
	"Atari2600": {
		Id:         "Atari2600",
		Folders:    []string{"ATARI7800", "Atari2600"},
		Extensions: []string{".a26"},
	},
	"Atari5200": {
		Id:         "Atari5200",
		Folders:    []string{"ATARI5200"},
		Extensions: []string{".a52"},
	},
	"Atari7800": {
		Id:         "Atari7800",
		Folders:    []string{"ATARI7800"},
		Extensions: []string{".a78"},
	},
	"AtariLynx": {
		Id:         "AtariLynx",
		Folders:    []string{"AtariLynx"},
		Extensions: []string{".lnx"},
	},
	"CasioPV1000": {
		Id:         "CasioPV1000",
		Aliases:    []string{"Casio_PV-1000"},
		Folders:    []string{"Casio_PV-1000"},
		Extensions: []string{".bin"},
	},
	"ChannelF": {
		Id:         "ChannelF",
		Folders:    []string{"ChannelF"},
		Extensions: []string{".rom", ".bin"},
	},
	"ColecoVision": {
		Id:         "ColecoVision",
		Aliases:    []string{"Coleco"},
		Folders:    []string{"Coleco"},
		Extensions: []string{".col", ".bin", ".rom"},
	},
	"CreatiVision": {
		Id:         "CreatiVision",
		Folders:    []string{"CreatiVision"},
		Extensions: []string{".rom", ".bin", ".bas"},
	},
	"FDS": {
		Id:         "FDS",
		Aliases:    []string{"FamicomDiskSystem"},
		Folders:    []string{"NES", "FDS"},
		Extensions: []string{".fds"},
	},
	"Gamate": {
		Id:         "Gamate",
		Folders:    []string{"Gamate"},
		Extensions: []string{".bin"},
	},
	"Gameboy": {
		Id:         "Gameboy",
		Aliases:    []string{"GB"},
		Folders:    []string{"GAMEBOY"},
		Extensions: []string{".gb"},
	},
	"GameboyColor": {
		Id:         "GameboyColor",
		Aliases:    []string{"GBC"},
		Folders:    []string{"GAMEBOY", "GBC"},
		Extensions: []string{".gbc"},
	},
	"Gameboy2P": {
		// TODO: Split 2P core into GB and GBC?
		Id:         "Gameboy2P",
		Folders:    []string{"GAMEBOY2P"},
		Extensions: []string{".gb", ".gbc"},
	},
	"GameGear": {
		Id:         "GameGear",
		Aliases:    []string{"GG"},
		Folders:    []string{"SMS", "GameGear"},
		Extensions: []string{".gg"},
	},
	"GameNWatch": {
		Id:         "GameNWatch",
		Folders:    []string{"GameNWatch"},
		Extensions: []string{".bin"},
	},
	"GBA": {
		Id:         "GBA",
		Aliases:    []string{"GameboyAdvance"},
		Folders:    []string{"GBA"},
		Extensions: []string{".gba"},
	},
	"GBA2P": {
		Id:         "GBA2P",
		Folders:    []string{"GBA2P"},
		Extensions: []string{".gba"},
	},
	"Genesis": {
		Id:         "Genesis",
		Aliases:    []string{"MegaDrive"},
		Folders:    []string{"MegaDrive", "Genesis"},
		Extensions: []string{".gen", ".bin", ".md"},
	},
	"Intellivision": {
		Id:         "Intellivision",
		Folders:    []string{"Intellivision"},
		Extensions: []string{".int", ".bin"},
	},
	// TODO: Jaguar
	"MasterSystem": {
		Id:         "MasterSystem",
		Aliases:    []string{"SMS"},
		Folders:    []string{"SMS"},
		Extensions: []string{".sms"},
	},
	"MegaCD": {
		Id:         "MegaCD",
		Aliases:    []string{"SegaCD"},
		Folders:    []string{"MegaCD"},
		Extensions: []string{".cue", ".chd"},
	},
	"MegaDuck": {
		Id:         "MegaDuck",
		Folders:    []string{"GAMEBOY", "MegaDuck"},
		Extensions: []string{".bin"},
	},
	"NeoGeo": {
		Id:         "NeoGeo",
		Folders:    []string{"NEOGEO"},
		Extensions: []string{".neo"}, // TODO: .zip and folder support
	},
	"NeoGeoCD": {
		Id:         "NeoGeo",
		Folders:    []string{"NeoGeo-CD", "NEOGEO"},
		Extensions: []string{".cue", ".chd"},
	},
	"NES": {
		Id:         "NES",
		Folders:    []string{"NES"},
		Extensions: []string{".nes"},
	},
	"NESMusic": {
		Id:         "NESMusic",
		Folders:    []string{"NES"},
		Extensions: []string{".nsf"},
	},
	"Nintendo64": {
		Id:         "Nintendo64",
		Aliases:    []string{"N64"},
		Folders:    []string{"N64"},
		Extensions: []string{".n64", ".z64"},
	},
	"Odyssey2": {
		Id:         "Odyssey2",
		Folders:    []string{"ODYSSEY2"},
		Extensions: []string{".bin"},
	},
	"PocketChallengeV2": {
		Id:         "PocketChallengeV2",
		Folders:    []string{"WonderSwan", "PocketChallengeV2"},
		Extensions: []string{".pc2"},
	},
	"PokemonMini": {
		Id:         "PokemonMini",
		Folders:    []string{"PokemonMini"},
		Extensions: []string{".min"},
	},
	"PSX": {
		Id:         "PSX",
		Aliases:    []string{"Playstation", "PS1"},
		Folders:    []string{"PSX"},
		Extensions: []string{".cue", ".chd", ".exe"},
	},
	"Sega32X": {
		Id:         "Sega32X",
		Aliases:    []string{"S32X", "32X"},
		Folders:    []string{"S32X"},
		Extensions: []string{".32x"},
	},
	"SG1000": {
		Id:         "SG1000",
		Folders:    []string{"SG1000", "Coleco", "SMS"},
		Extensions: []string{".sg"},
	},
	"SuperGameboy": {
		Id:         "SuperGameboy",
		Aliases:    []string{"SGB"},
		Folders:    []string{"SGB"},
		Extensions: []string{".sgb", ".gb", ".gbc"},
	},
	"SuperVision": {
		Id:         "SuperVision",
		Folders:    []string{"SuperVision"},
		Extensions: []string{".bin", ".sv"},
	},
	"Saturn": {
		Id:         "Saturn",
		Folders:    []string{"Saturn"},
		Extensions: []string{".cue", ".chd"},
	},
	"SNES": {
		Id:         "SNES",
		Aliases:    []string{"SuperNintendo"},
		Folders:    []string{"SNES"},
		Extensions: []string{".sfc", ".smc", ".bin", ".bs"},
	},
	"SNESMusic": {
		Id:         "SNESMusic",
		Folders:    []string{"SNES"},
		Extensions: []string{".spc"},
	},
	"SuperGrafx": {
		Id:         "SuperGrafx",
		Folders:    []string{"TGFX16"},
		Extensions: []string{".sgx"},
	},
	"TurboGrafx16": {
		Id:         "TurboGrafx16",
		Aliases:    []string{"TGFX16", "PCEngine"},
		Folders:    []string{"TGFX16"},
		Extensions: []string{".pce", ".bin"},
	},
	"TurboGrafx16CD": {
		Id:         "TurboGrafx16CD",
		Aliases:    []string{"TGFX16-CD", "PCEngineCD"},
		Folders:    []string{"TGFX16-CD"},
		Extensions: []string{".cue", ".chd"},
	},
	"VC4000": {
		Id:         "VC4000",
		Folders:    []string{"VC4000"},
		Extensions: []string{".bin"},
	},
	"Vectrex": {
		Id:         "Vectrex",
		Folders:    []string{"VECTREX"},
		Extensions: []string{".vec", ".bin", ".rom"}, // TODO: overlays (.ovr)
	},
	"WonderSwan": {
		Id:         "WonderSwan",
		Folders:    []string{"WonderSwan"},
		Extensions: []string{".ws"},
	},
	"WonderSwanColor": {
		Id:         "WonderSwanColor",
		Folders:    []string{"WonderSwan", "WonderSwanColor"},
		Extensions: []string{".wsc"},
	},
	// Computers
	"AcornAtom": {
		Id:         "AcornAtom",
		Folders:    []string{"AcornAtom"},
		Extensions: []string{".vhd"},
	},
	"AcornElectron": {
		Id:         "AcornElectron",
		Folders:    []string{"AcornElectron"},
		Extensions: []string{".vhd"},
	},
	"AliceMC10": {
		Id:         "AliceMC10",
		Folders:    []string{"AliceMC10"},
		Extensions: []string{".c10"},
	},
	"Amiga": {
		Id:         "Amiga",
		Folders:    []string{"Amiga"},
		Aliases:    []string{"Minimig"},
		Extensions: []string{".adf"},
	},
	"Amstrad": {
		Id:         "Amstrad",
		Folders:    []string{"Amstrad"},
		Extensions: []string{".dsk", ".cdt"}, // TODO: globbing support? for .e??
	},
	"AmstradPCW": {
		Id:         "AmstradPCW",
		Aliases:    []string{"Amstrad-PCW"},
		Folders:    []string{"Amstrad PCW"},
		Extensions: []string{".dsk"},
	},
	"ao486": {
		Id:         "ao486",
		Folders:    []string{"AO486"},
		Extensions: []string{".img", ".ima", ".vhd", ".vfd", ".iso", ".cue", ".chd"},
	},
	"Apogee": {
		Id:         "Apogee",
		Folders:    []string{"APOGEE"},
		Extensions: []string{".rka", ".rkr", ".gam"},
	},
	"AppleI": {
		Id:         "AppleI",
		Aliases:    []string{"Apple-I"},
		Folders:    []string{"Apple-I"},
		Extensions: []string{".txt"},
	},
	"AppleII": {
		Id:         "AppleII",
		Aliases:    []string{"Apple-II"},
		Folders:    []string{"Apple-II"},
		Extensions: []string{".dsk", ".do", ".po", ".nib", ".hdv"},
	},
	"Aquarius": {
		Id:         "Aquarius",
		Folders:    []string{"AQUARIUS"},
		Extensions: []string{".bin", ".caq"},
	},
	"Atari800": {
		Id:         "Atari800",
		Folders:    []string{"ATARI800"},
		Extensions: []string{".atr", ".xex", ".xfd", ".atx", ".car", ".rom", ".bin"},
	},
	"BBCMicro": {
		Id:         "BBCMicro",
		Folders:    []string{"BBCMicro"},
		Extensions: []string{".ssd", ".dsd", ".vhd"},
	},
	"BK0011M": {
		Id:         "BK0011M",
		Folders:    []string{"BK0011M"},
		Extensions: []string{".bin", ".dsk", ".vhd"},
	},
	"C16": {
		Id:         "C16",
		Folders:    []string{"C16"},
		Extensions: []string{".d64", ".g64", ".prg", ".tap", ".bin"},
	},
	"C64": {
		Id:         "C64",
		Folders:    []string{"C64"},
		Extensions: []string{".d64", ".g64", ".t64", ".d81", ".prg", ".crt", ".reu", ".tap"},
	},
	"CasioPV2000": {
		Id:         "CasioPV2000",
		Aliases:    []string{"Casio_PV-2000"},
		Folders:    []string{"Casio_PV-2000"},
		Extensions: []string{".bin"},
	},
	"CoCo2": {
		Id:         "CoCo2",
		Folders:    []string{"CoCo2"},
		Extensions: []string{".dsk", ".cas", ".ccc", ".rom"},
	},
	"EDSAC": {
		Id:         "EDSAC",
		Folders:    []string{"EDSAC"},
		Extensions: []string{".tap"},
	},
	"Galaksija": {
		Id:         "Galaksija",
		Folders:    []string{"Galaksija"},
		Extensions: []string{".tap"},
	},
	"Interact": {
		Id:         "Interact",
		Folders:    []string{"Interact"},
		Extensions: []string{".cin", ".k7"},
	},
	"Jupiter": {
		Id:         "Jupiter",
		Folders:    []string{"Jupiter"},
		Extensions: []string{".ace"},
	},
	"Laser": {
		Id:         "Laser",
		Aliases:    []string{"Laser310"},
		Folders:    []string{"Laser"},
		Extensions: []string{".vz"},
	},
	"Lynx48": {
		Id:         "Lynx48",
		Folders:    []string{"Lynx48"},
		Extensions: []string{".tap"},
	},
	"MacPlus": {
		Id:         "MacPlus",
		Folders:    []string{"MACPLUS"},
		Extensions: []string{".dsk", ".img", ".vhd"},
	},
	"MSX": {
		Id:         "MSX",
		Folders:    []string{"MSX"},
		Extensions: []string{".vhd"},
	},
	"MultiComp": {
		Id:         "MultiComp",
		Folders:    []string{"MultiComp"},
		Extensions: []string{".img"},
	},
	"Orao": {
		Id:         "Orao",
		Folders:    []string{"ORAO"},
		Extensions: []string{".tap"},
	},
	"Oric": {
		Id:         "Oric",
		Folders:    []string{"Oric"},
		Extensions: []string{".dsk"},
	},
	"PCXT": {
		Id:         "PCXT",
		Folders:    []string{"PCXT"},
		Extensions: []string{".img", ".vhd", ".ima", ".vfd"},
	},
	"PDP1": {
		Id:         "PDP1",
		Folders:    []string{"PDP1"},
		Extensions: []string{".bin", ".rim", ".pdp"},
	},
	"PET2001": {
		Id:         "PET2001",
		Folders:    []string{"PET2001"},
		Extensions: []string{".prg", ".tap"},
	},
	"PMD85": {
		Id:         "PMD85",
		Folders:    []string{"PMD85"},
		Extensions: []string{".rmm"},
	},
	"QL": {
		Id:         "QL",
		Folders:    []string{"QL"},
		Extensions: []string{".mdv", ".win"},
	},
	"RX78": {
		Id:         "RX78",
		Folders:    []string{"RX78"},
		Extensions: []string{".bin"},
	},
	"SAMCoupe": {
		Id:         "SAMCoupe",
		Folders:    []string{"SAMCOUPE"},
		Extensions: []string{".dsk", ".mgt", ".img"},
	},
	"SordM5": {
		Id:         "SordM5",
		Aliases:    []string{"Sord M5"},
		Folders:    []string{"Sord M5"},
		Extensions: []string{".bin", ".rom", ".cas"},
	},
	"Specialist": {
		Id:         "Specialist",
		Aliases:    []string{"SPMX"},
		Folders:    []string{"SPMX"},
		Extensions: []string{".rks", ".odi"},
	},
	"SVI328": {
		Id:         "SVI328",
		Folders:    []string{"SVI328"},
		Extensions: []string{".cas", ".bin", ".rom"},
	},
	"TatungEinstein": {
		Id:         "TatungEinstein",
		Folders:    []string{"TatungEinstein"},
		Extensions: []string{".dsk"},
	},
	"TI994A": {
		Id:         "TI994A",
		Aliases:    []string{"TI-99_4A"},
		Folders:    []string{"TI-99_4A"},
		Extensions: []string{".bin", ".m99"},
	},
	"TomyTutor": {
		Id:         "TomyTutor",
		Folders:    []string{"TomyTutor"},
		Extensions: []string{".bin", ".cas"},
	},
	"TRS80": {
		Id:         "TRS80",
		Folders:    []string{"TRS-80"},
		Extensions: []string{".jvi", ".dsk", ".cas"},
	},
	"TSConf": {
		Id:         "TSConf",
		Folders:    []string{"TSConf"},
		Extensions: []string{".vhf"},
	},
	"UK101": {
		Id:         "UK101",
		Folders:    []string{"UK101"},
		Extensions: []string{".txt", ".bas", ".lod"},
	},
	"Vector06C": {
		Id:         "Vector06C",
		Aliases:    []string{"Vector06"},
		Folders:    []string{"VECTOR06"},
		Extensions: []string{".rom", ".com", ".c00", ".edd", ".fdd"},
	},
	"VIC20": {
		Id:         "VIC20",
		Folders:    []string{"VIC20"},
		Extensions: []string{".d64", ".g64", ".prg", ".tap", ".crt"},
	},
	"X68000": {
		Id:         "X68000",
		Folders:    []string{"X68000"},
		Extensions: []string{".d88", ".hdf"},
	},
	"ZX81": {
		Id:         "ZX81",
		Folders:    []string{"ZX81"},
		Extensions: []string{".p", ".0"},
	},
	"ZXSpectrum": {
		Id:         "ZXSpectrum",
		Aliases:    []string{"Spectrum"},
		Folders:    []string{"Spectrum"},
		Extensions: []string{".tap", ".csw", ".tzx", ".sna", ".z80", ".trd", ".img", ".dsk", ".mgt"},
	},
	"ZXNext": {
		Id:         "ZXNext",
		Folders:    []string{"ZXNext"},
		Extensions: []string{".vhd"},
	},
	// Other
	"Arcade": {
		Id:         "Arcade",
		Folders:    []string{"_Arcade"},
		Extensions: []string{".mra"},
	},
	"Arduboy": {
		Id:         "Arduboy",
		Folders:    []string{"Arduboy"},
		Extensions: []string{".hex", ".bin"},
	},
	"Chip8": {
		Id:         "Chip8",
		Folders:    []string{"Chip8"},
		Extensions: []string{".ch8"},
	},
}
