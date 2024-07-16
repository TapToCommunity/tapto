//go:build linux || darwin

package mister

import (
	"github.com/wizzomafizzo/tapto/pkg/database/gamesdb"
	"github.com/wizzomafizzo/tapto/pkg/platforms"
)

var Launchers = map[string]platforms.Launcher{
	// Consoles
	gamesdb.SystemAdventureVision: {
		SystemId:   gamesdb.SystemAdventureVision,
		Folders:    []string{"AVision"},
		Extensions: []string{".bin"},
	},
	gamesdb.SystemArcadia: {
		SystemId:   gamesdb.SystemArcadia,
		Folders:    []string{"Arcadia"},
		Extensions: []string{".bin"},
	},
	gamesdb.SystemAstrocade: {
		SystemId:   gamesdb.SystemAstrocade,
		Folders:    []string{"Astrocade"},
		Extensions: []string{".bin"},
	},
	gamesdb.SystemAtari2600: {
		SystemId:   gamesdb.SystemAtari2600,
		Folders:    []string{"ATARI7800", "Atari2600"},
		Extensions: []string{".a26"},
	},
	gamesdb.SystemAtari5200: {
		SystemId:   gamesdb.SystemAtari5200,
		Folders:    []string{"ATARI5200"},
		Extensions: []string{".a52"},
	},
	gamesdb.SystemAtari7800: {
		SystemId:   gamesdb.SystemAtari7800,
		Folders:    []string{"ATARI7800"},
		Extensions: []string{".a78"},
	},
	gamesdb.SystemAtariLynx: {
		SystemId:   gamesdb.SystemAtariLynx,
		Folders:    []string{"AtariLynx"},
		Extensions: []string{".lnx"},
	},
	gamesdb.SystemCasioPV1000: {
		SystemId:   gamesdb.SystemCasioPV1000,
		Folders:    []string{"Casio_PV-1000"},
		Extensions: []string{".bin"},
	},
	gamesdb.SystemChannelF: {
		SystemId:   gamesdb.SystemChannelF,
		Folders:    []string{"ChannelF"},
		Extensions: []string{".rom", ".bin"},
	},
	gamesdb.SystemColecoVision: {
		SystemId:   gamesdb.SystemColecoVision,
		Folders:    []string{"Coleco"},
		Extensions: []string{".col", ".bin", ".rom"},
	},
	gamesdb.SystemCreatiVision: {
		SystemId:   gamesdb.SystemCreatiVision,
		Folders:    []string{"CreatiVision"},
		Extensions: []string{".rom", ".bin", ".bas"},
	},
	gamesdb.SystemFDS: {
		SystemId:   gamesdb.SystemFDS,
		Folders:    []string{"NES", "FDS"},
		Extensions: []string{".fds"},
	},
	gamesdb.SystemGamate: {
		SystemId:   gamesdb.SystemGamate,
		Folders:    []string{"Gamate"},
		Extensions: []string{".bin"},
	},
	gamesdb.SystemGameboy: {
		SystemId:   gamesdb.SystemGameboy,
		Folders:    []string{"GAMEBOY"},
		Extensions: []string{".gb"},
	},
	gamesdb.SystemGameboyColor: {
		SystemId:   gamesdb.SystemGameboyColor,
		Folders:    []string{"GAMEBOY", "GBC"},
		Extensions: []string{".gbc"},
	},
	gamesdb.SystemGameboy2P: {
		SystemId:   gamesdb.SystemGameboy2P,
		Folders:    []string{"GAMEBOY2P"},
		Extensions: []string{".gb", ".gbc"},
	},
	gamesdb.SystemGameGear: {
		SystemId:   gamesdb.SystemGameGear,
		Folders:    []string{"SMS", "GameGear"},
		Extensions: []string{".gg"},
	},
	gamesdb.SystemGameNWatch: {
		SystemId:   gamesdb.SystemGameNWatch,
		Folders:    []string{"GameNWatch"},
		Extensions: []string{".bin"},
	},
	gamesdb.SystemGBA: {
		SystemId:   gamesdb.SystemGBA,
		Folders:    []string{"GBA"},
		Extensions: []string{".gba"},
	},
	gamesdb.SystemGBA2P: {
		SystemId:   gamesdb.SystemGBA2P,
		Folders:    []string{"GBA2P"},
		Extensions: []string{".gba"},
	},
	gamesdb.SystemGenesis: {
		SystemId:   gamesdb.SystemGenesis,
		Folders:    []string{"MegaDrive", "Genesis"},
		Extensions: []string{".gen", ".bin", ".md"},
	},
	gamesdb.SystemIntellivision: {
		SystemId:   gamesdb.SystemIntellivision,
		Folders:    []string{"Intellivision"},
		Extensions: []string{".int", ".bin"},
	},
	gamesdb.SystemMasterSystem: {
		SystemId:   gamesdb.SystemMasterSystem,
		Folders:    []string{"SMS"},
		Extensions: []string{".sms"},
	},
	gamesdb.SystemMegaCD: {
		SystemId:   gamesdb.SystemMegaCD,
		Folders:    []string{"MegaCD"},
		Extensions: []string{".cue", ".chd"},
	},
	gamesdb.SystemMegaDuck: {
		SystemId:   gamesdb.SystemMegaDuck,
		Folders:    []string{"GAMEBOY", "MegaDuck"},
		Extensions: []string{".bin"},
	},
	gamesdb.SystemNeoGeo: {
		SystemId:   gamesdb.SystemNeoGeo,
		Folders:    []string{"NEOGEO"},
		Extensions: []string{".neo"}, // TODO: .zip and folder support
	},
	gamesdb.SystemNeoGeoCD: {
		SystemId:   gamesdb.SystemNeoGeo,
		Folders:    []string{"NeoGeo-CD", "NEOGEO"},
		Extensions: []string{".cue", ".chd"},
	},
	gamesdb.SystemNES: {
		SystemId:   gamesdb.SystemNES,
		Folders:    []string{"NES"},
		Extensions: []string{".nes"},
	},
	gamesdb.SystemNESMusic: {
		SystemId:   gamesdb.SystemNESMusic,
		Folders:    []string{"NES"},
		Extensions: []string{".nsf"},
	},
	gamesdb.SystemNintendo64: {
		SystemId:   gamesdb.SystemNintendo64,
		Folders:    []string{"N64"},
		Extensions: []string{".n64", ".z64"},
	},
	gamesdb.SystemOdyssey2: {
		SystemId:   gamesdb.SystemOdyssey2,
		Folders:    []string{"ODYSSEY2"},
		Extensions: []string{".bin"},
	},
	gamesdb.SystemPocketChallengeV2: {
		SystemId:   gamesdb.SystemPocketChallengeV2,
		Folders:    []string{"WonderSwan", "PocketChallengeV2"},
		Extensions: []string{".pc2"},
	},
	gamesdb.SystemPokemonMini: {
		SystemId:   gamesdb.SystemPokemonMini,
		Folders:    []string{"PokemonMini"},
		Extensions: []string{".min"},
	},
	gamesdb.SystemPSX: {
		SystemId:   gamesdb.SystemPSX,
		Folders:    []string{"PSX"},
		Extensions: []string{".cue", ".chd", ".exe"},
	},
	gamesdb.SystemSega32X: {
		SystemId:   gamesdb.SystemSega32X,
		Folders:    []string{"S32X"},
		Extensions: []string{".32x"},
	},
	gamesdb.SystemSG1000: {
		SystemId:   gamesdb.SystemSG1000,
		Folders:    []string{"SG1000", "Coleco", "SMS"},
		Extensions: []string{".sg"},
	},
	gamesdb.SystemSuperGameboy: {
		SystemId:   gamesdb.SystemSuperGameboy,
		Folders:    []string{"SGB"},
		Extensions: []string{".sgb", ".gb", ".gbc"},
	},
	gamesdb.SystemSuperVision: {
		SystemId:   gamesdb.SystemSuperVision,
		Folders:    []string{"SuperVision"},
		Extensions: []string{".bin", ".sv"},
	},
	gamesdb.SystemSaturn: {
		SystemId:   gamesdb.SystemSaturn,
		Folders:    []string{"Saturn"},
		Extensions: []string{".cue", ".chd"},
	},
	gamesdb.SystemSNES: {
		SystemId:   gamesdb.SystemSNES,
		Folders:    []string{"SNES"},
		Extensions: []string{".sfc", ".smc", ".bin", ".bs"},
	},
	gamesdb.SystemSNESMusic: {
		SystemId:   gamesdb.SystemSNESMusic,
		Folders:    []string{"SNES"},
		Extensions: []string{".spc"},
	},
	gamesdb.SystemSuperGrafx: {
		SystemId:   gamesdb.SystemSuperGrafx,
		Folders:    []string{"TGFX16"},
		Extensions: []string{".sgx"},
	},
	gamesdb.SystemTurboGrafx16: {
		SystemId:   gamesdb.SystemTurboGrafx16,
		Folders:    []string{"TGFX16"},
		Extensions: []string{".pce", ".bin"},
	},
	gamesdb.SystemTurboGrafx16CD: {
		SystemId:   gamesdb.SystemTurboGrafx16CD,
		Folders:    []string{"TGFX16-CD"},
		Extensions: []string{".cue", ".chd"},
	},
	gamesdb.SystemVC4000: {
		SystemId:   gamesdb.SystemVC4000,
		Folders:    []string{"VC4000"},
		Extensions: []string{".bin"},
	},
	gamesdb.SystemVectrex: {
		SystemId:   gamesdb.SystemVectrex,
		Folders:    []string{"VECTREX"},
		Extensions: []string{".vec", ".bin", ".rom"}, // TODO: overlays (.ovr)
	},
	gamesdb.SystemWonderSwan: {
		SystemId:   gamesdb.SystemWonderSwan,
		Folders:    []string{"WonderSwan"},
		Extensions: []string{".ws"},
	},
	gamesdb.SystemWonderSwanColor: {
		SystemId:   gamesdb.SystemWonderSwanColor,
		Folders:    []string{"WonderSwan", "WonderSwanColor"},
		Extensions: []string{".wsc"},
	},
	// Computers
	gamesdb.SystemAcornAtom: {
		SystemId:   gamesdb.SystemAcornAtom,
		Folders:    []string{"AcornAtom"},
		Extensions: []string{".vhd"},
	},
	gamesdb.SystemAcornElectron: {
		SystemId:   gamesdb.SystemAcornElectron,
		Folders:    []string{"AcornElectron"},
		Extensions: []string{".vhd"},
	},
	gamesdb.SystemAliceMC10: {
		SystemId:   gamesdb.SystemAliceMC10,
		Folders:    []string{"AliceMC10"},
		Extensions: []string{".c10"},
	},
	gamesdb.SystemAmiga: {
		SystemId:   gamesdb.SystemAmiga,
		Folders:    []string{"Amiga"},
		Extensions: []string{".adf"},
	},
	gamesdb.SystemAmstrad: {
		SystemId:   gamesdb.SystemAmstrad,
		Folders:    []string{"Amstrad"},
		Extensions: []string{".dsk", ".cdt"}, // TODO: globbing support? for .e??
	},
	gamesdb.SystemAmstradPCW: {
		SystemId:   gamesdb.SystemAmstradPCW,
		Folders:    []string{"Amstrad PCW"},
		Extensions: []string{".dsk"},
	},
	gamesdb.SystemAo486: {
		SystemId:   gamesdb.SystemAo486,
		Folders:    []string{"AO486"},
		Extensions: []string{".img", ".ima", ".vhd", ".vfd", ".iso", ".cue", ".chd"},
	},
	gamesdb.SystemApogee: {
		SystemId:   gamesdb.SystemApogee,
		Folders:    []string{"APOGEE"},
		Extensions: []string{".rka", ".rkr", ".gam"},
	},
	gamesdb.SystemAppleI: {
		SystemId:   gamesdb.SystemAppleI,
		Folders:    []string{"Apple-I"},
		Extensions: []string{".txt"},
	},
	gamesdb.SystemAppleII: {
		SystemId:   gamesdb.SystemAppleII,
		Folders:    []string{"Apple-II"},
		Extensions: []string{".dsk", ".do", ".po", ".nib", ".hdv"},
	},
	gamesdb.SystemAquarius: {
		SystemId:   gamesdb.SystemAquarius,
		Folders:    []string{"AQUARIUS"},
		Extensions: []string{".bin", ".caq"},
	},
	gamesdb.SystemAtari800: {
		SystemId:   gamesdb.SystemAtari800,
		Folders:    []string{"ATARI800"},
		Extensions: []string{".atr", ".xex", ".xfd", ".atx", ".car", ".rom", ".bin"},
	},
	gamesdb.SystemBBCMicro: {
		SystemId:   gamesdb.SystemBBCMicro,
		Folders:    []string{"BBCMicro"},
		Extensions: []string{".ssd", ".dsd", ".vhd"},
	},
	gamesdb.SystemBK0011M: {
		SystemId:   gamesdb.SystemBK0011M,
		Folders:    []string{"BK0011M"},
		Extensions: []string{".bin", ".dsk", ".vhd"},
	},
	gamesdb.SystemC16: {
		SystemId:   gamesdb.SystemC16,
		Folders:    []string{"C16"},
		Extensions: []string{".d64", ".g64", ".prg", ".tap", ".bin"},
	},
	gamesdb.SystemC64: {
		SystemId:   gamesdb.SystemC64,
		Folders:    []string{"C64"},
		Extensions: []string{".d64", ".g64", ".t64", ".d81", ".prg", ".crt", ".reu", ".tap"},
	},
	gamesdb.SystemCasioPV2000: {
		SystemId:   gamesdb.SystemCasioPV2000,
		Folders:    []string{"Casio_PV-2000"},
		Extensions: []string{".bin"},
	},
	gamesdb.SystemCoCo2: {
		SystemId:   gamesdb.SystemCoCo2,
		Folders:    []string{"CoCo2"},
		Extensions: []string{".dsk", ".cas", ".ccc", ".rom"},
	},
	gamesdb.SystemEDSAC: {
		SystemId:   gamesdb.SystemEDSAC,
		Folders:    []string{"EDSAC"},
		Extensions: []string{".tap"},
	},
	gamesdb.SystemGalaksija: {
		SystemId:   gamesdb.SystemGalaksija,
		Folders:    []string{"Galaksija"},
		Extensions: []string{".tap"},
	},
	gamesdb.SystemInteract: {
		SystemId:   gamesdb.SystemInteract,
		Folders:    []string{"Interact"},
		Extensions: []string{".cin", ".k7"},
	},
	gamesdb.SystemJupiter: {
		SystemId:   gamesdb.SystemJupiter,
		Folders:    []string{"Jupiter"},
		Extensions: []string{".ace"},
	},
	gamesdb.SystemLaser: {
		SystemId:   gamesdb.SystemLaser,
		Folders:    []string{"Laser"},
		Extensions: []string{".vz"},
	},
	gamesdb.SystemLynx48: {
		SystemId:   gamesdb.SystemLynx48,
		Folders:    []string{"Lynx48"},
		Extensions: []string{".tap"},
	},
	gamesdb.SystemMacPlus: {
		SystemId:   gamesdb.SystemMacPlus,
		Folders:    []string{"MACPLUS"},
		Extensions: []string{".dsk", ".img", ".vhd"},
	},
	gamesdb.SystemMSX: {
		SystemId:   gamesdb.SystemMSX,
		Folders:    []string{"MSX"},
		Extensions: []string{".vhd"},
	},
	gamesdb.SystemMultiComp: {
		SystemId:   gamesdb.SystemMultiComp,
		Folders:    []string{"MultiComp"},
		Extensions: []string{".img"},
	},
	gamesdb.SystemOrao: {
		SystemId:   gamesdb.SystemOrao,
		Folders:    []string{"ORAO"},
		Extensions: []string{".tap"},
	},
	gamesdb.SystemOric: {
		SystemId:   gamesdb.SystemOric,
		Folders:    []string{"Oric"},
		Extensions: []string{".dsk"},
	},
	gamesdb.SystemPCXT: {
		SystemId:   gamesdb.SystemPCXT,
		Folders:    []string{"PCXT"},
		Extensions: []string{".img", ".vhd", ".ima", ".vfd"},
	},
	gamesdb.SystemPDP1: {
		SystemId:   gamesdb.SystemPDP1,
		Folders:    []string{"PDP1"},
		Extensions: []string{".bin", ".rim", ".pdp"},
	},
	gamesdb.SystemPET2001: {
		SystemId:   gamesdb.SystemPET2001,
		Folders:    []string{"PET2001"},
		Extensions: []string{".prg", ".tap"},
	},
	gamesdb.SystemPMD85: {
		SystemId:   gamesdb.SystemPMD85,
		Folders:    []string{"PMD85"},
		Extensions: []string{".rmm"},
	},
	gamesdb.SystemQL: {
		SystemId:   gamesdb.SystemQL,
		Folders:    []string{"QL"},
		Extensions: []string{".mdv", ".win"},
	},
	gamesdb.SystemRX78: {
		SystemId:   gamesdb.SystemRX78,
		Folders:    []string{"RX78"},
		Extensions: []string{".bin"},
	},
	gamesdb.SystemSAMCoupe: {
		SystemId:   gamesdb.SystemSAMCoupe,
		Folders:    []string{"SAMCOUPE"},
		Extensions: []string{".dsk", ".mgt", ".img"},
	},
	gamesdb.SystemSordM5: {
		SystemId:   gamesdb.SystemSordM5,
		Folders:    []string{"Sord M5"},
		Extensions: []string{".bin", ".rom", ".cas"},
	},
	gamesdb.SystemSpecialist: {
		SystemId:   gamesdb.SystemSpecialist,
		Folders:    []string{"SPMX"},
		Extensions: []string{".rks", ".odi"},
	},
	gamesdb.SystemSVI328: {
		SystemId:   gamesdb.SystemSVI328,
		Folders:    []string{"SVI328"},
		Extensions: []string{".cas", ".bin", ".rom"},
	},
	gamesdb.SystemTatungEinstein: {
		SystemId:   gamesdb.SystemTatungEinstein,
		Folders:    []string{"TatungEinstein"},
		Extensions: []string{".dsk"},
	},
	gamesdb.SystemTI994A: {
		SystemId:   gamesdb.SystemTI994A,
		Folders:    []string{"TI-99_4A"},
		Extensions: []string{".bin", ".m99"},
	},
	gamesdb.SystemTomyTutor: {
		SystemId:   gamesdb.SystemTomyTutor,
		Folders:    []string{"TomyTutor"},
		Extensions: []string{".bin", ".cas"},
	},
	gamesdb.SystemTRS80: {
		SystemId:   gamesdb.SystemTRS80,
		Folders:    []string{"TRS-80"},
		Extensions: []string{".jvi", ".dsk", ".cas"},
	},
	gamesdb.SystemTSConf: {
		SystemId:   gamesdb.SystemTSConf,
		Folders:    []string{"TSConf"},
		Extensions: []string{".vhf"},
	},
	gamesdb.SystemUK101: {
		SystemId:   gamesdb.SystemUK101,
		Folders:    []string{"UK101"},
		Extensions: []string{".txt", ".bas", ".lod"},
	},
	gamesdb.SystemVector06C: {
		SystemId:   gamesdb.SystemVector06C,
		Folders:    []string{"VECTOR06"},
		Extensions: []string{".rom", ".com", ".c00", ".edd", ".fdd"},
	},
	gamesdb.SystemVIC20: {
		SystemId:   gamesdb.SystemVIC20,
		Folders:    []string{"VIC20"},
		Extensions: []string{".d64", ".g64", ".prg", ".tap", ".crt"},
	},
	gamesdb.SystemX68000: {
		SystemId:   gamesdb.SystemX68000,
		Folders:    []string{"X68000"},
		Extensions: []string{".d88", ".hdf"},
	},
	gamesdb.SystemZX81: {
		SystemId:   gamesdb.SystemZX81,
		Folders:    []string{"ZX81"},
		Extensions: []string{".p", ".0"},
	},
	gamesdb.SystemZXSpectrum: {
		SystemId:   gamesdb.SystemZXSpectrum,
		Folders:    []string{"Spectrum"},
		Extensions: []string{".tap", ".csw", ".tzx", ".sna", ".z80", ".trd", ".img", ".dsk", ".mgt"},
	},
	gamesdb.SystemZXNext: {
		SystemId:   gamesdb.SystemZXNext,
		Folders:    []string{"ZXNext"},
		Extensions: []string{".vhd"},
	},
	// Other
	gamesdb.SystemArcade: {
		SystemId:   gamesdb.SystemArcade,
		Folders:    []string{"_Arcade"},
		Extensions: []string{".mra"},
	},
	gamesdb.SystemArduboy: {
		SystemId:   gamesdb.SystemArduboy,
		Folders:    []string{"Arduboy"},
		Extensions: []string{".hex", ".bin"},
	},
	gamesdb.SystemChip8: {
		SystemId:   gamesdb.SystemChip8,
		Folders:    []string{"Chip8"},
		Extensions: []string{".ch8"},
	},
}
