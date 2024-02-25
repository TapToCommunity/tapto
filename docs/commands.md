# Token Commands

Tokens are set up to work with TapTo by writing a small piece of text to them, telling TapTo what it should do when it's read. The most common action is to launch a game, but it can perform other actions like launching a random game or even making an HTTP request.

This text can be as simple as a path to a game file, or perform multiple custom comands in a row.

- [Token Commands](#token-commands)
  - [Setting Up Tokens](#setting-up-tokens)
    - [NFCUI](#nfcui)
    - [Remote](#remote)
    - [Phone](#phone)
    - [Desktop](#desktop)
  - [Combining Commands](#combining-commands)
  - [Launching Games and Cores](#launching-games-and-cores)
  - [Custom Commands](#custom-commands)
    - [Launch a System (launch.system)](#launch-a-system-launchsystem)
    - [Launch a Random Game (launch.random)](#launch-a-random-game-launchrandom)
    - [Change the Actve MiSTer.ini File (mister.ini)](#change-the-actve-misterini-file-misterini)
    - [Make an HTTP GET Request to a URL (http.get)](#make-an-http-get-request-to-a-url-httpget)
    - [Make an HTTP POST Request to a URL (http.post)](#make-an-http-post-request-to-a-url-httppost)
    - [Press a Keyboard Key (input.key)](#press-a-keyboard-key-inputkey)
    - [Insert a Coin/Credit (input.coinp1/input.coinp2)](#insert-a-coincredit-inputcoinp1inputcoinp2)
    - [Run a System/Linux Command (shell)](#run-a-systemlinux-command-shell)
    - [Delay Command Execution (delay)](#delay-command-execution-delay)


## Setting Up Tokens

### NFCUI

The `nfcui.sh` script for MiSTer is included for download in the [TapTo releases](https://github.com/wizzomafizzo/tapto/releases/latest/). This is a frontend to the TapTo service which lets you write and interact with tokens in many advanced ways. It can browse and search games on the MiSTer to be written, and has an awesome command palette to build up more complex commands.

Download it to the `Scripts` folder on your MiSTer's SD card and run it from the `Scripts` menu.

### Remote

The [Remote](https://github.com/wizzomafizzo/mrext/blob/main/docs/remote.md) app has basic support for writing games to tokens through the TapTo service. If you have TapTo running, a button to write a game to a token will be displayed when selected games in search results and the games browser.

### Phone

The NFC Tools app is highly recommended for this. It's free and supports both
[iOS](https://apps.apple.com/us/app/nfc-tools/id1252962749) and 
[Android](https://play.google.com/store/apps/details?id=com.wakdev.wdnfc&hl=en&gl=US).

You'll want to write a *Text record* with it for all the supported NFC service features.

### Desktop

NFC Tools also has a free version for [Windows, Mac and Linux](https://www.wakdev.com/en/apps/nfc-tools-pc-mac.html) that works well. ACR122U readers will also work natively with most desktop operating systems.

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

## Launching Games and Cores

The NFC script supports launching game files, core .RBF files, arcade .MRA files and .MGL shortcut files. This is
done by simply writing the path to the file to the token.

For example, to launch a game, write something like this to the token:
```
/media/fat/games/Genesis/1 US - Q-Z/Road Rash (USA, Europe).md
```

To save space and to handle games moving between storage devices, you can also use a relative path:
```
Genesis/1 US - Q-Z/Road Rash (USA, Europe).md
```

This will search for the file in all standard MiSTer game folder paths including CIFS.

Some other examples:
```
_Arcade/1942 (Revision B).mra
```
```
_@Favorites/Super Metroid.mgl
```

Because core filenames often change, it's supported to use the same short name as in a .MGL file to launch it:
```
_Console/PSX
```

.ZIP files are also supported natively, same as they are in MiSTer itself. Just treat the .ZIP file as a folder name:
```
Genesis/@Genesis - MegaSD Mega EverDrive 2022-05-18.zip/1 US - Q-Z/Road Rash (USA, Europe).md
```

## Custom Commands

There are a small set of special commands that can be written to tokens to perform dynamic actions. These are marked in
a token by putting `**` at the start of the stored text.

### Launch a System (launch.system)

This command will launch a system, based on MiSTer Extensions own internal list of system IDs
[here](https://github.com/wizzomafizzo/mrext/blob/main/docs/systems.md). This can be useful for "meta systems" such as
Atari 2600 and WonderSwan Color which don't have their own core .RBF file.

For example:
```
**launch.system:Atari2600
```
```
**launch.system:WonderSwanColor
```

It also works for any other system if you prefer this method over the standard core .RBF file one.

### Launch a Random Game (launch.random)

This command will launch a game a random for the given system. For example:
```
**launch.random:snes
```
This will launch a random SNES game each time you read the token.

You can also select all systems with `**launch.random:all`.

### Change the Actve MiSTer.ini File (mister.ini)

Loads the specified MiSTer.ini file and relaunches the menu core if open.

Specify the .ini file with its index in the list shown in the MiSTer menu. Numbers `1` to `4`.

For example:
```
**mister.ini:1
```

This switch will not persist after a reboot, same as loading it through the OSD.

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

### Make an HTTP POST Request to a URL (http.post)

Perform an HTTP POST request to the specified URL. For example:
```
**http.post:https://example.com|application/json|{"key":"value"}
```

Or with Remote, to launch the Update All script:
```
**http.post:http://localhost:8182/api/scripts/launch/update_all.sh|application/json|
```

The command is in the format `URL|Content-Type|Body`.

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

**This feature is intentionally disabled for security reasons when run straight from a token. You can still use it, but only via the `nfc.csv` file [explained here](mister.md#mappings-database) or by enabling the `allow_commands` option in `tapto.ini`.**

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
