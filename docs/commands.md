# Token Commands

Tokens are set up to work with TapTo by writing a small piece of text to them, telling TapTo what it should do when it's read. The most common action is to launch a game, but it can perform other actions like launching a random game or even making an HTTP request.

This text can be as simple as a path to a game file, or perform multiple custom comands in a row.

- [Token Commands](#token-commands)
  - [Setting Up Tokens](#setting-up-tokens)
    - [TapTUI](#taptui)
    - [Remote](#remote)
    - [Phone](#phone)
    - [Desktop](#desktop)
  - [Launching Games](#launching-games)
    - [Absolute Path](#absolute-path)
    - [Relative Path](#relative-path)
    - [System Lookup](#system-lookup)
    - [Special Cases](#special-cases)
      - [ao486](#ao486)
      - [AmigaVision (Amiga)](#amigavision-amiga)
      - [NeoGeo](#neogeo)
  - [Combining Commands](#combining-commands)
  - [Custom Commands](#custom-commands)
    - [Generic Launch (launch)](#generic-launch-launch)
    - [Launch a System (launch.system)](#launch-a-system-launchsystem)
    - [Launch a Random Game (launch.random)](#launch-a-random-game-launchrandom)
    - [Change the Active MiSTer.ini File (mister.ini)](#change-the-active-misterini-file-misterini)
    - [Launch a Core RBF File (mister.core)](#launch-a-core-rbf-file-mistercore)
    - [Make an HTTP GET Request to a URL (http.get)](#make-an-http-get-request-to-a-url-httpget)
    - [Make an HTTP POST Request to a URL (http.post)](#make-an-http-post-request-to-a-url-httppost)
    - [Press a Keyboard Key (input.key)](#press-a-keyboard-key-inputkey)
    - [Insert a Coin/Credit (input.coinp1/input.coinp2)](#insert-a-coincredit-inputcoinp1inputcoinp2)
    - [Run a System/Linux Command (shell)](#run-a-systemlinux-command-shell)
    - [Delay Command Execution (delay)](#delay-command-execution-delay)


## Setting Up Tokens

### TapTUI

The `taptui.sh` script for MiSTer is included for download in the [TapTo releases](https://github.com/wizzomafizzo/tapto/releases/latest/).

This is a frontend to the TapTo service which lets you write and interact with tokens in many advanced ways. It can browse and search games on the MiSTer to be written, and has an awesome command palette to build up more complex commands.

It's included with TapTo on MiSTer if you're using the Downloader or Update All scripts, or you can download it manually to the `Scripts` folder on your MiSTer's SD card, and run it from the `Scripts` menu.

### Remote

The [Remote](https://github.com/wizzomafizzo/mrext/blob/main/docs/remote.md) app has basic support for writing games to tokens through the TapTo service.

If you have TapTo running, a button to write a game to a token will be displayed when selecting games in search results and the games browser.

### Phone

The NFC Tools app is highly recommended for this. It's free and supports both
[iOS](https://apps.apple.com/us/app/nfc-tools/id1252962749) and 
[Android](https://play.google.com/store/apps/details?id=com.wakdev.wdnfc&hl=en&gl=US).

You'll want to write a `Text record` with it for all the supported NFC service features.

### Desktop

NFC Tools also has a free version for [Windows, Mac and Linux](https://www.wakdev.com/en/apps/nfc-tools-pc-mac.html) that works well. ACR122U readers will work natively with most desktop operating systems.

## Launching Games

For the most basic usage, a file path can be written to a token and TapTo will attempt to find the file on the device and launch it.

The required system and configuration for the game will be automatically selected based on the file path. On MiSTer, you can also launch arcade .mra files, shortcut .mgl files and core .rbf files with just their path.

TapTo uses the following rules, in order, to find the game file. Keep these rules in mind if you want a token to work well between different devices.

> [!TIP]
> If you're not sure what to do, it's recommended to use the [System Lookup](#system-lookup) method for the best portability between devices.

### Absolute Path

Any path starting with a `/` will be treated as an absolute path.

For example, to launch a game, write something like this to the token:
```
/media/fat/games/Genesis/1 US - Q-Z/Some Game (USA, Europe).md
```

This is the least portable method, as it will only work on the device with the exact same file path.

### Relative Path

It's also possible to use a file path relative to the games folder. This will search for the file in all standard MiSTer game folder paths including CIFS and USB.

For example:
```
Genesis/1 US - Q-Z/Some Game (USA, Europe).md
```

This saves storage space on the token and will work if one device has games stored on a USB drive and another on the SD card. There's no downside to using it compared to an absolute path.

Some other examples:
```
_Arcade/Some Arcade Game.mra
```
```
_@Favorites/My Favorite Game.mgl
```

.zip files are also supported natively, same as they are in MiSTer itself. Just treat the .zip file as a folder name:
```
Genesis/@Genesis - 2022-05-18.zip/1 US - Q-Z/Some Game (USA, Europe).md
```

### System Lookup

This is similar to a relative path, but the first "folder" will be treated as a reference to a system instead of a folder. Like this: `<System ID>/<Game Path>`.

Check the [Systems](https://github.com/wizzomafizzo/mrext/blob/main/docs/systems.md) documentation for a list of supported system IDs.

For example:
```
N64/1 US - A-M/Another Game (USA).z64
```

While this looks like a relative path, it will work on any device with the same system folder structure, even if the Nintendo 64 folder does not have the same name. TapTo will look up the system ID and find the game file based on that.

System ID aliases (listed in the [Systems](https://github.com/wizzomafizzo/mrext/blob/main/docs/systems.md) page as well) can also be used here.

For example, this will work:

```
TurboGrafx16/Another Game (USA).pce
```

Or this:
```
tgfx16/Another Game (USA).pce
```

Or even this:
```
PCEngine/Another Game (USA).pce
```

> [!NOTE]
> While this method is basically the same as a relative path right now, as TapTo develops, we will be adding more features to this System Lookup method so it's not necessary to use explcit paths. It will also be important going forward as more devices are supported with different file structures.

### Special Cases

Some systems have special cases and extra support implemented for their launching capabilities.

#### ao486

If a .vhd file is launched via TapTo, and this .vhd file is sitting in its own folder with an .iso file, that .iso file will also be automatically mounted alongside the .vhd file.

#### AmigaVision (Amiga)

Launching games in the [AmigaVision](https://amiga.vision/) image on the Amiga core is supported via the `games.txt` files and `demos.txt` files located in the `Amiga/listings` folder on your SD card.

For example, to launch Beneath a Stell Sky in AmigaVision:
```
Amiga/listings/games.txt/Beneath a Steel Sky (OCS)[en]
```

The `games.txt` and `demos.txt` files contain a listing of all supported games and demos, generated by AmigaVision, and can be treated as a virtual folder for launching via TapTo. Other games can be launched using the same format of `Amiga/listings/games.txt/<Game Name>`.

Opening the `games.txt` and `demos.txt` files in a text editor will show the full list of supported games and demos.

#### NeoGeo

NeoGeo also supports launching .zip files and folders directly with TapTo, as is supported with the MiSTer core itself.

## Combining Commands

All commands and game/core launches can be combined on a single token if space permits using the `||` separator.

For example, to switch to MiSTer.ini number 3 and launch the SNES core:
```
**mister.ini:3||_Console/SNES
```

Or launch a game and notify an HTTP service:
```
_Console/SNES||**http.get:https://example.com
```

As many of these can be combined as you like.

## Custom Commands

There are a small set of special commands that can be written to tokens to perform dynamic actions. These are marked in
a token by putting `**` at the start of the stored text.

### Generic Launch (launch)

This command works exactly the same as the basic [game launch](#launching-games-and-cores) behaviour. It's just
a more explicit way of doing it.

For example:
```
**launch:/media/fat/games/Genesis/1 US - Q-Z/Some Game (USA, Europe).md
```

### Launch a System (launch.system)

This command will launch a system, based on MiSTer Extensions own internal list of system IDs
[here](https://github.com/wizzomafizzo/mrext/blob/main/docs/systems.md). This can be useful for "meta systems" such as
Atari 2600 and WonderSwan Color which don't have their own core .rbf file.

For example:
```
**launch.system:Atari2600
```
```
**launch.system:WonderSwanColor
```

It also works for any other system if you prefer this method over the standard core .rbf file one.

### Launch a Random Game (launch.random)

This command will launch a game at random for the given system. For example:
```
**launch.random:snes
```
This will launch a random SNES game each time you read the token.

System IDs can also be combined with the `,` separator. For example:
```
**launch.random:snes,nes
```
This will launch a random game from either the SNES or NES systems.

You can also select all systems with `**launch.random:all`.

### Change the Active MiSTer.ini File (mister.ini)

Loads the specified MiSTer.ini file and relaunches the menu core if open.

Specify the .ini file with its index in the list shown in the MiSTer menu. Numbers `1` to `4`.

For example:
```
**mister.ini:1
```

This switch will not persist after a reboot, same as loading it through the OSD.

### Launch a Core RBF File (mister.core)

This command will launch a MiSTer core .rbf file directly. For example:
```
**mister.core:_Console/SNES
```

Or:
```
**mister.core:_Console/PSX_20220518
```

It uses the exact same format as the `rbf` tag in a .mgl file, where the ending of a filename can be omitted. The path is relative to the SD card.

### Make an HTTP GET Request to a URL (http.get)

Perform an HTTP GET request to the specified URL. For example:
```
**http.get:https://example.com
```

This is useful for triggering webhooks or other web services.

It can be combined with other commands using the `||` separator. For example:
```
**http.get:https://example.com||_Console/SNES
```

> [!IMPORTANT]
> If your URL contains any of the following characters, you must URL encode them by replacing them with the following:
> 
> - `,` with `%2C`
> - `||` with `%7C%7C`
> - `**` with `%2A%2A`

### Make an HTTP POST Request to a URL (http.post)

Perform an HTTP POST request to the specified URL. For example:
```
**http.post:https://example.com,application/json,{"key":"value"}
```

Or with Remote, to launch the Update All script:
```
**http.post:http://localhost:8182/api/scripts/launch/update_all.sh,application/json,
```

The command is in the format `URL,Content-Type,Body`.

> [!IMPORTANT]
> If your URL contains any of the following characters, you must URL encode them by replacing them with the following:
> 
> - `,` with `%2C`
> - `||` with `%7C%7C`
> - `**` with `%2A%2A`

### Press a Keyboard Key (input.key)

Press a key on the keyboard using its uinput code. For example (to press F12 to bring up the OSD):
```
**input.key:88
```

See a full list of key codes [here](https://pkg.go.dev/github.com/bendahl/uinput@v1.6.0#pkg-constants).

### Insert a Coin/Credit (input.coinp1/input.coinp2)

Insert a coin/credit for player 1 or 2. For example (to insert 1 coin for player 1):
```
**input.coinp1:1
```

This command presses the `5` and `6` key on the keyboard respectively, which is generally accepted as the coin insert
keys in MiSTer arcade cores. If it doesn't work, try manually mapping the coin insert keys in the OSD.

It also supports inserting multiple coins at once. For example (to insert 3 coins for player 2):
```
**input.coinp2:3
```

### Run a System/Linux Command (shell)

> [!WARNING]
> This feature is intentionally disabled for security reasons when run straight from a token. You can still use it, but only via the `nfc.csv` file [explained here](mister.md#mappings-database), or by enabling the `allow_commands` option in `tapto.ini`.

This command will run a MiSTer Linux command directly. For example:
```
**shell:reboot
```

### Delay Command Execution (delay)

This command will delay the execution of the next command by the specified number of milliseconds. For example:
```
**delay:500
```

Will delay the next command by 500ms (half a second). This is a *blocking command* and will delay the entire token read by the specified time.

It can be combined with other commands using the `||` separator. For example, to launch SNES, wait 10 seconds, then press F12:
```
_Console/SNES||**delay:10000||**input.key:88
```
