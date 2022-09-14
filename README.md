# MiSTer Extensions

Extensions and utilities to make your [MiSTer](https://github.com/MiSTer-devel/Main_MiSTer/wiki) even better.

Make sure to check the linked documentation for each script you use. Most are simple and work out-of-the-box, but some require manual setup before they do anything useful.

## Install

All scripts listed can be installed by downloading the linked file, placing it in the `Scripts` folder on your SD card, and running it from the `Scripts` menu on your MiSTer. No scripts require a keyboard to use or any additional dependencies.

Add the following to your `downloader.ini` file to install everything at once through the `update` script:

```
[mrext/all]
db_url = https://github.com/wizzomafizzo/mrext/raw/main/releases/all.json
```

Each script also provides its own individual update file if you only want certain ones.

## BGM
Play your own music in the MiSTer menu. BGM is a highly configurable background music player that automatically pauses when you're playing games. Supports many common audio formats including internet radio streams.

[![Download BGM](docs/download.png "Download BGM")](https://github.com/wizzomafizzo/MiSTer_BGM/raw/main/bgm.sh)
[![Readme BGM](docs/readme.png "Readme BGM")](https://github.com/wizzomafizzo/MiSTer_BGM)

## Favorites
Create and manage shortcuts for your favorite games. Favorites allows you to pick any game or core from your system and automatically generate a shortcut to it in the MiSTer menu.

[![Download Favorites](docs/download.png "Download Favorites")](https://github.com/wizzomafizzo/MiSTer_Favorites/raw/main/favorites.sh)
[![Readme Favorites](docs/readme.png "Readme Favorites")](https://github.com/wizzomafizzo/MiSTer_Favorites)

## Games Menu
Browse your entire collection from the main MiSTer menu. Games Menu indexes all your games and generates a set of shortcuts in the menu mirroring your folder layout.

[![Download Games Menu](docs/download.png "Download Games Menu")](https://github.com/wizzomafizzo/MiSTer_GamesMenu/raw/main/games_menu.sh)
[![Readme Games Menu](docs/readme.png "Readme Games Menu")](https://github.com/wizzomafizzo/MiSTer_GamesMenu)

## Launch Sync
Create shareable and subscriptable game playlists. Launch Sync automatically generates working menu shortcuts from custom playlist files, with the ability to keep them up-to-date with the author's live version.

[![Download Launch Sync](docs/download.png "Download Launch Sync")](https://github.com/wizzomafizzo/mrext/raw/main/releases/launchsync/launchsync.sh)
[![Readme Launch Sync](docs/readme.png "Readme Launch Sync")](https://github.com/wizzomafizzo/mrext/tree/main/docs/launchsync.md)

## Random
Instantly launch a random game in your collection from the Scripts menu.

[![Download Random](docs/download.png "Download Random")](https://github.com/wizzomafizzo/mrext/raw/main/releases/random/random.sh)

## Search
Search for and launch games from your collection. Searching is *fast* and great for discovering games.

[![Download Search](docs/download.png "Download Search")](https://github.com/wizzomafizzo/mrext/raw/main/releases/search/search.sh)

## Other Projects

Other great projects that add heaps of functionality to your MiSTer.

Please [open an issue](https://github.com/wizzomafizzo/mrext/issues/new) if you'd like to suggest something for this list. Anything is welcome, though the focus is on software projects that work without custom hardware.

### Core

- [mister-boot-roms](https://github.com/uberyoji/mister-boot-roms)

  Adds high quality MiSTer-themed boot screens to cores which support loadable boot roms.

- [VIDEO PRESETS by Robby](https://github.com/RGarciaLago/VIDEO_PRESETS_by_Robby)

  A curated and extensive set of video presets for the MiSTer cores.

### Frontend

- [Coin-Op](https://github.com/funkycochise/Coin-Op)

  An alternative layout for browsing the Arcade folder.

- [MiSTer Super Attract Mode (SAM)](https://github.com/mrchrisster/MiSTer_SAM)

  Add an attract screen to your MiSTer. When idle, games will start to play at random and rotate after a set period. You can even jump in and start playing if a game looks fun! A mature project and highly configurable.

- [MiSTerWallpapers](https://github.com/RetroDriven/MiSTerWallpapers)

  Automatically download a large collection of high quality wallpapers.

- [MiSTer-CRT-Wallpapers](https://github.com/RetroDriven/MiSTer-CRT-Wallpapers)

  The same, but specifically for 4:3 CRT screens.

### Ports

- [MiSTer Basilisk II](https://github.com/bbond007/MiSTer_BasiliskII)

  A build of the [Basilisk II](https://basilisk.cebix.net/) project for MiSTer. A 68k Macintosh emulator.

- [MiSTer DOSBox](https://github.com/bbond007/MiSTer_DOSBox)

  A build of the [DOSBox](https://www.dosbox.com/) project for MiSTer. Play a huge range of DOS games.

- [MiSTer PrBoom-Plus](https://github.com/bbond007/MiSTer_PrBoom-Plus)

  A build of the [PrBoom](http://prboom.sourceforge.net/) project for MiSTer. An enhanced Doom engine with a massive number of expansions.

- [MiSTer ScummVM](https://github.com/bbond007/MiSTer_ScummVM)

  A build of the [ScummVM](https://www.scummvm.org/) project for MiSTer. Runs well and even works for games out of reach of the AO486 core.

### System

- [MiSTer FPGA Overclock Scripts](https://github.com/coolbho3k/MiSTer-Overclock-Scripts)

  A kernel patch that allows overclocking the MiSTer. An overclocked system can run Munt (MT32 emulator) at full speed and get extra performance out of software like ScummVM.

- [MiSTerArch](https://github.com/MiSTerArch/PKGBUILDs)

  A replacement image for the MiSTer with a full [Arch Linux](https://archlinux.org/) system.

- [MiSTerTools](https://github.com/morfeus77/MiSTerTools/)

  Scripts for custom aspect ratio calculation, modeline to video_mode conversion, video_mode to modeline conversion, ini profile switcher and to parse MRA files.

- [Official Scripts](https://github.com/MiSTer-devel/Scripts_MiSTer)

  The official MiSTer scripts repository. A miscellaneous collection of small scripts for various system tasks and configuration.

### Updaters

- [DOS Shareware Updater](https://github.com/flynnsbit/DOS_Shareware_MyMenu)

  Update script for the Flynn's DOS Shareware pack on the AO486 core.

- [Top 300 Updater](https://github.com/flynnsbit/Top300_updates)

  Update script for Flynn's Top 300 pack on the AO486 core.

- [Update All](https://github.com/theypsilon/Update_All_MiSTer)
  
  If you're reading this, you already use it. Don't forget to check the advanced options.

- [Yet another random MiSTer utility script (YARMUS)](https://github.com/jayp76/MiSTer_get_optional_installers)

  A script to install a lot of MiSTer scripts at once. It includes many things on this list, plus extra custom installers for software like [DevilutionX](https://github.com/diasurgical/devilutionX), [Cave Story](https://nxengine.sourceforge.io/) and homebrew packs.
