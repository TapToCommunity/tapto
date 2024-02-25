# MiSTer Setup

- [MiSTer Setup](#mister-setup)
  - [Installation](#installation)
    - [Downloader and Update All](#downloader-and-update-all)
    - [Hardware Setup](#hardware-setup)
  - [Configuration File](#configuration-file)
    - [Connection String (connection\_string)](#connection-string-connection_string)
    - [Allow Shell Commands From Tokens (allow\_commands)](#allow-shell-commands-from-tokens-allow_commands)
    - [Disable Sounds After a Read (disable\_sounds)](#disable-sounds-after-a-read-disable_sounds)
    - [Probe for Serial Devices (probe\_device)](#probe-for-serial-devices-probe_device)
    - [Exit Game When Token Is Removed (exit\_game)](#exit-game-when-token-is-removed-exit_game)
    - [Exit Game Core Blocklist (exit\_game\_blocklist)](#exit-game-core-blocklist-exit_game_blocklist)
  - [Mappings Database](#mappings-database)

## Installation

Download [TapTo](https://github.com/wizzomafizzo/tapto/releases/latest/) and copy it to the `Scripts` folder on your MiSTer's SD card.

Once installed, run `tapto` from the MiSTer `Scripts` menu, a prompt will offer to enable TapTo as a startup service, then the service will be started in the background.

After the initial setup is complete, a status display will be shown. It's ok to exit this screen, the service will continue to run in the background.

### Downloader and Update All

TapTo is available in [Update All](https://github.com/theypsilon/Update_All_MiSTer) by enabling the `MiSTer Extensions` repository in the `Tools & Scripts` menu.

If you only want TapTo, add the following text to the `downloader.ini` file on your MiSTer:

```
[mrext/tapto]
db_url = https://github.com/wizzomafizzo/tapto/raw/main/scripts/mister/repo/tapto.json
```

### Hardware Setup

Your reader may work out of the box with no extra configuration. Run `tapto` from the `Scripts` menu, plug it in, and check if it shows as connected in the log view.

If you are using a PN532 NFC module connected with a USB to TTL cable, then the following config may be needed in `tapto.ini` in the `Scripts` folder:

```ini
[tapto]
probe_device=yes
allow_commands=no
```

Create this file if it doesn't exist.

If TapTo is unable to auto-detect your device, it may be necessary to manually configure the connection string:

```ini
[tapto]
connection_string="pn532_uart:/dev/ttyUSB0"
allow_commands=no
```

Be aware the ttyUSB0 part may be different if you have other devices connected such as tty2oled. For a list of possible devices try:

`ls /dev/serial/by-id` or `ls /dev | grep ttyUSB`

## Configuration File

TapTo supports a `tapto.ini` file in the `Scripts` folder. This file can be used to configure the TapTo service.

If one doesn't exist, create a new one. This example has all the default values:

```ini
[tapto]
connection_string=""
allow_commands=no
disable_sounds=no
probe_device=yes
exit_game=no
```

All lines except the `[tapto]` header are optional.

### Connection String (connection_string)

| Key                 | Default Value | 
|---------------------|---------------|
| `connection_string` |               |

See [Hardware Setup](#hardware-setup) for details. This option is for configuration of [libnfc](https://github.com/nfc-tools/libnfc).

### Allow Shell Commands From Tokens (allow_commands)

| Key                 | Default Value | 
|---------------------|---------------|
| `allow_commands`    | no            |

Enables the [command](commands.md#run-a-systemlinux-command-command) custom command to be triggered from a tag.

By default this is disabled and only works from the [Mappings Database](#mappings-database) described below.

### Disable Sounds After a Read (disable_sounds)

| Key                 | Default Value | 
|---------------------|---------------|
| `disable_sounds`    | no            |

Disables the success and fail sounds played when a tag is read by the reader.

### Probe for Serial Devices (probe_device)

| Key                 | Default Value | 
|---------------------|---------------|
| `probe_device`      | yes           |

Enables auto-detection of a serial based reader device.

### Exit Game When Token Is Removed (exit_game)

| Key                 | Default Value | 
|---------------------|---------------|
| `exit_game`         | no            |

Enables exiting the current game when a token is removed from the reader.

:warning: This does not trigger a save file to be written in MiSTer, you have to do that manually.

### Exit Game Core Blocklist (exit_game_blocklist)

| Key                   | Default Value |
|-----------------------|---------------|
| `exit_game_blocklist` |               |

A comma separated list of cores to ignore the `exit_game` option for. For example, to ignore the `exit_game` option for the NES and SNES cores:

```ini
[tapto]
exit_game=yes
exit_game_blocklist=NES,SNES
```

With this configuration, removing a token will not exit the game when using the NES or SNES cores, but will for all other cores.

The core name is the same as the name that shows on the left sidebar of the OSD when in a core.

## Mappings Database

TapTo supports an `nfc.csv` file in the top of the SD card. This file can be used to override the text read from a tag and map it to a different text value. This is useful for mapping Amiibos which are read-only, testing text values before actually writing them, and is necessary for using the `command` custom command by default.

Create a file called `nfc.csv` in the top of the SD card, with this as the header:
```csv
match_uid,match_text,text
```

You'll then need to either power cycle your MiSTer, or restart the TapTo service by running `tapto` from the `Scripts` menu, selecting the `Stop` button, then the `Start` button.

After the file is created, the service will automatically reload it every time it's updated.

Here's an example `nfc.csv` file that maps several Amiibos to different functions:
```csv
match_uid,match_text,text
04e5c7ca024980,,**command:reboot
04078e6a724c80,,_#Favorites/Final Fantasy VII.mgl
041e6d5a983c80,,_#Favorites/Super Metroid.mgl
041ff6ea973c81,,_#Favorites/Legend of Zelda.mgl
```

Only one `match_` column is required for an entry, and the `match_uid` can include colons and uppercase characters. You can get the UID of a tag by checking the output in the `tapto` Script display or on your phone.
