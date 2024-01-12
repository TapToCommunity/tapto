# Token Commands

## Setting up tags

The NFC Tools app is highly recommended for this. It's free and supports both
[iOS](https://apps.apple.com/us/app/nfc-tools/id1252962749) and 
[Android](https://play.google.com/store/apps/details?id=com.wakdev.wdnfc&hl=en&gl=US).

You'll want to write a *Text record* with it for all the supported NFC service features.

### Combining commands

All commands and game/core launches can be combined on a single tag if space permits using the `||` separator.

For example, to switch to MiSTer.ini number 3 and launch the SNES core:
```
**ini:3||_Console/SNES
```

Or launch a game and notify an HTTP service:
```
_Console/SNES||**get:https://example.com
```

As many of these can be combined as you like.

### Launching games and cores

The NFC script supports launching game files, core .RBF files, arcade .MRA files and .MGL shortcut files. This is
done by simply writing the path to the file to the tag.

For example, to launch a game, write something like this to the tag:
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

### Custom commands

There are a small set of special commands that can be written to tags to perform dynamic actions. These are marked in
a tag by putting `**` at the start of the stored text.

#### Launch a system (system)

This command will launch a system, based on MiSTer Extensions own internal list of system IDs
[here](https://github.com/wizzomafizzo/mrext/blob/main/docs/systems.md). This can be useful for "meta systems" such as
Atari 2600 and WonderSwan Color which don't have their own core .RBF file.

For example:
```
**system:Atari2600
```
```
**system:WonderSwanColor
```

It also works for any other system if you prefer this method over the standard core .RBF file one.

#### Launch a random game (random)

This command will launch a game a random for the given system. For example:
```
**random:snes
```
This will launch a random SNES game each time you scan the tag.

You can also select all systems with `**random:all`.

#### Change the actve MiSTer.ini file (ini)

Loads the specified MiSTer.ini file and relaunches the menu core if open.

Specify the .ini file with its index in the list shown in the MiSTer menu. Numbers `1` to `4`.

For example:
```
**ini:1
```

This switch will not persist after a reboot, same as loading it through the OSD.

#### Make an HTTP request to a URL (get)

Perform an HTTP GET request to the specified URL. For example:
```
**get:https://example.com
```

This is useful for triggering webhooks or other web services.

It can be combined with other commands using the `||` separator. For example:
```
**get:https://example.com||_Console/SNES
```

This does *not* check for any errors, and will not show any output. You send the request and off it goes into the ether.

#### Press a keyboard key (key)

Press a key on the keyboard using its uinput code. For example (to press F12 to bring up the OSD):
```
**key:88
```

See a full list of key codes [here](https://pkg.go.dev/github.com/bendahl/uinput@v1.6.0#pkg-constants).

#### Insert a coin/credit (coinp1/coinp2)

Insert a coin/credit for player 1 or 2. For example (to insert 1 coin for player 1):
```
**coinp1:1
```

This command presses the `5` and `6` key on the keyboard respectively, which is generally accepted as the coin insert
keys in MiSTer arcade cores. If it doesn't work, try manually mapping the coin insert keys in the OSD.

It also supports inserting multiple coins at once. For example (to insert 3 coins for player 2):
```
**coinp2:3
```

#### Run a system/Linux command (command)

**This feature is intentionally disabled for security reasons when run straight from a tag. You can still use it,
but only via the `nfc.csv` file explained below or by enabling the `allow_commands` option in `nfc.ini`.**

This command will run a MiSTer Linux command directly. For example:
```
**command:reboot
```
