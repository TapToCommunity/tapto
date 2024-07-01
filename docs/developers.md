# Developer Guide

## Development Environment

The project is primarily written in Go, uses Task for build scripts and Docker for MiSTer builds. Development can be done on any platform, though the build system currently assumes a Linux environment and applications which use C binding libraries can be challenging to build on Windows.

### Dependencies

- [Go](https://go.dev/)

  The whole meat of the project. Version 1.19 or newer.

- [Task](https://taskfile.dev/)

  Used for all builds and automations in the project. Follow instructions for your OS on the installation page.

- [Docker](https://www.docker.com/)

  Used for building all the binaries. You also need to configure cross-compilation in Docker since ARM images are used for the build process. Podman should also work using the Docker commands wrapper.

  On Linux, enable cross-platform builds with something like this: `apt install qemu binfmt-support qemu-user-static`

  On Mac, Docker Desktop comes with everything you need already.

### Optional Dependencies

- [libnfc](https://github.com/nfc-tools/libnfc) (build only)

- [ncurses](https://tldp.org/HOWTO/NCURSES-Programming-HOWTO/) (build only)

- [Python](https://www.python.org/)

  Used for some scripts and older projects. Remember that MiSTer currently ships with version 3.9, so don't use any newer Python features.

## Building

To start, you can run `go mod download` from the root of the project folder. This will download all dependencies used by the project. Builds automatically do this, but running it now will stop your editor from complaining about missing modules.

All build steps are done with the `task` command run from the root of the project folder. Run `task --list-all` by itself to see a list of available commands.

Before building MiSTer binaries, you'll need to build the Docker image it uses. Just run `task build-mister-image` to add it to your system.

Built binaries will be created in the `_build` directory under its appropriate platform and architecture subdirectory.

These are the important commands:

- `task build-mister`

  Builds a MiSTer binary of TapTo and copies TapTUI to the build directory.

- `mage deploy-mister`

  Runs the build task, copies the binaries to the MiSTer directory on the SD card via SSH, then restarts the TapTo service on the MiSTer. This requires the `MISTER_IP` environment variable to be set to the IP address of the MiSTer. Add to `.env` in the root of the project or add `MISTER_IP=1.2.3.4` to the start of the command.

Builds may display warnings such as this:

```
/usr/bin/ld: /tmp/go-link-600559994/000029.o: in function `mygetgrouplist':
/_/os/user/getgrouplist_unix.go:15: warning: Using 'getgrouplist' in statically linked applications requires at runtime the shared libraries from the glibc version used for linking
```

They can be safely ignored. Some low-level things do not support static linking, but the Docker build environment matches the MiSTer image just to be safe.

### Testing

When changing the application behavior, in particular the reader loop, some testing is required. This [file](./scanner_behavior.md) contains a list of expected behavior for the application under certain conditions. It is useful to test them to ensure we didn't break any flow.
