<h1 align="left">
  <img width="60%" title="TapTo" src="assets/images/logo/tapto_gitbhub_logo.png" />
</h1>

TapTo is an open source system for launching games and scripted actions using physical objects like NFC cards. It's a great way to make playing games more accessible and add some fun to your gaming setup!

See the [Getting Started](#getting-started) section below for everything you need to get up and running. Additional hardware is required but is aimed to be cheap and easily available. Please [join the Discord](https://wizzo.dev/discord) if you need any help or want to show off your work!

TapTo is currently supported on these platforms:
- [MiSTer FPGA](https://mister-devel.github.io/MkDocs_MiSTer/)
- [Commodore 64 (via TeensyROM)](https://github.com/SensoriumEmbedded/TeensyROM/blob/main/docs/NFC_Loader.md)

***
[Download](https://github.com/wizzomafizzo/tapto/releases/latest/) | [TapTo Designer](https://tapto-designer.netlify.app/) | [Labels](docs/labels.md) | [Reader Cases](docs/community.md#cases) | [Vendors](docs/vendors.md) | [Community Projects](docs/community.md) | [API](docs/api.md)
***

## Getting Started

1. Buy or build a compatible NFC reader: [Reader Hardware](docs/readers.md)
2. Buy some compatible NFC cards or tags: [Token Hardware](docs/tokens.md)
3. Set up the software for your platform: [MiSTer FPGA](docs/mister.md) | [Commodore 64](https://github.com/SensoriumEmbedded/TeensyROM/blob/main/docs/NFC_Loader.md)
4. Set up your cards or tags to launch games: [Token Commands](docs/commands.md)
5. Make or buy some awesome labels: [Custom Labels](docs/labels.md)
6. Fabricate or buy some cool PCB cards: [PCB cards](docs/PCB_cards.md)

## Contributors

TapTo has been a community effort from day one. Everyone's contributions are appreciated and encouraged! Want to contribute too? Check out the [Developer Guide](docs/developers.md).

- **Andrea Bogazzi** &mdash; developer &mdash; [GitHub](https://github.com/asturur) | [Twitter](https://twitter.com/AndreaBogazzi)
- **batty** &mdash; writer &mdash; [GitHub](https://github.com/protogem2) | [Twitter](https://twitter.com/goddamnbathead)
- **BedroomNinja** &mdash; case designer &mdash; [Printables](https://www.printables.com/@bedroom_ninj_1665215) | [Twitter](https://twitter.com/Bedroom_Ninja)
- **Gaz** &mdash; developer &mdash; [GitHub](https://github.com/symm) | [Twitter](https://twitter.com/gazj)
- **Ranny Snice** &mdash; logo creator &mdash; [GitHub](https://github.com/Ranny-Snice) | [Twitter](https://twitter.com/RannySnice)
- **RetroCastle** &mdash; PCB designer &mdash; [Store](https://www.aliexpress.com/store/912024455) | [Twitter](https://twitter.com/zhangch93067765)
- **Sensorium** &mdash; C64 support &mdash; [GitHub](https://www.github.com/SensoriumEmbedded) | [Twitter](https://twitter.com/SensoriumEmb)
- **theypsilon** &mdash; developer &mdash; [GitHub](https://www.github.com/theypsilon) | [Patreon](https://www.patreon.com/theypsilon) | [Twitter](https://twitter.com/josembarroso)
- **Tim Wilsie** &mdash; template designer &mdash; [GitHub](https://github.com/timwilsie) | [Twitter](https://twitter.com/timwilsie)
- **wizzo** &mdash; developer &mdash; [GitHub](https://github.com/wizzomafizzo) | [Patreon](https://patreon.com/wizzo) | [Twitter](https://twitter.com/wizzomafizzo)
- **Ziggurat** &mdash; developer &mdash; [Github](https://github.com/sigboe)
- **TheTrain** &mdash; PCB designer &mdash; [GP2040-CE GitHub](https://github.com/OpenStickCommunity/GP2040-CE) | [Twitter](https://twitter.com/thetrain24)

Special thanks to **Gaz** for starting the project originally, and to **[javiwwweb](https://github.com/javiwwweb/MisTerRFID)** and **[Conner](https://github.com/ElRojo/MiSTerRFID)** for their existing MiSTerRFID projects.

## License

This project's source code is licensed under the [GNU General Public License v3](/LICENSE).

This project's [assets](/assets) (e.g. image templates, 3D models, PCB designs) and [documentation](/docs) is licensed under the [Creative Commons Attribution-ShareAlike 4.0 International](/assets/LICENSE) license, unless explicitly noted otherwise.

The TapTo logo was designed by and is © Ranny Snice. The terms of use for the logo are as follows:

- The logo **MAY** be used on designs and artwork for tokens (e.g. label designs, printed stickers, pre-printed NFC cards), including those sold and held commercially, with the intent to show they're compatible with TapTo software and hardware.
- The logo **MAY** be used on open source community hardware.
- The logo **MAY** be used to link back to this repository or for similar promotional purposes of a strictly non-commercial nature (e.g. blog posts, social media, YouTube videos).
- The logo **MUST NOT** be used on or for the marketing of closed source or commercial hardware (e.g. case designs, PCBs), without express permission from the project.
- The logo **MUST NOT** be used for any other commercial products or purposes, without express permission from the project.
- The shape and overall design of the logo **MUST NOT** be modified or distorted. You **MAY** change the colors if required.

Please contact the project if you are unsure about any of these licensing arrangements, or have any licensing requests.
