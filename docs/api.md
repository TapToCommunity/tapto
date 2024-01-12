# API

## Writing to tags

The NFC script currently supports writing to NTAG tags through the command line option `-write <text>`.

For example, from the console or SSH:
```bash
/media/fat/Scripts/nfc.sh -write "_Console/SNES"
```
This will write the text `_Console/SNES` to the next detected tag.

This is available to any script or application on the MiSTer.

## Reading tags

Whenever a tag is successfully scanned, its UID and text contents (if available) will be written to the
file `/tmp/NFCSCAN`. The contents of the file is in the format `<uid>,<text>`.

You can monitor the file for changes to detect when a tag is scanned with the `inotifywait` command that is
shipped on the MiSTer Linux image. For example:
```bash
while inotifywait -e modify /tmp/NFCSCAN; do
    echo "Tag scanned"
done
```

## Service socket

When the NFC service is active, a Unix socket is created at `/tmp/nfc.sock`. This socket can be used to send
commands to the service.

Commands can be sent in a shell script like this:
```bash
echo "status" | socat - UNIX-CONNECT:/tmp/nfc.sock
```

### status

Returns the current status of the service with the following information:

- Last card scanned date in Unix epoch format
- Last card scanned UID
- Whether launching is enabled
- Last card scanned text

Each value is separated by a comma. For example:

```
1695650197,04faa9d2295880,true,**random:psx
```

### enable

Enables launching from tags (the default state).

This command has no output.

### disable

Disables launching from tags. Cards will scan and log, but no action will be triggered.

This command has no output.
