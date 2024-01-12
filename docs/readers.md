# Reader Hardware

These readers are currently confirmed to work with TapTo. Want to add a device or vendor? Please [open an issue](https://github.com/wizzomafizzo/tapto/issues/new).

TapTo uses the [libnfc](https://nfc-tools.github.io/projects/libnfc/) library, so there may be other devices that work with the correct configuration.

* [DIY NFC Reader](#diy-nfc-reader)
* [ACR122U](#acr122u)
* [PN532 Module](#pn532-module)

## DIY NFC Reader

:question: Are you a potential vendor? This reader is fully open source and we're always looking for more vendors to offer it for sale. Check out the [DIY NFC Reader Build Guide](diyreader.md) for details.

![](/assets/images/readers/diyreader.jpg)

The DIY NFC Reader is a TapTo community project aiming to offer a high quality and reliable reader at a low price, which is guaranteed to work with TapTo software.

- Can be built yourself with the [DIY NFC Reader Build Guide](diyreader.md)!
- Plug and play.
- Depending on vendor or batch size, comes at a very competitive price compared to ACR122U.
- Very small footprint with case.
- Used by other TapTo projects as a base.

Internally, this reader uses a [PN532 module](#pn532-module) and has all the same pros and cons associated with it. Having trusted vendors for this reader means none of the quality related issues with the module should be of concern.

### Where To Buy

- **[Grumpy Old Gamer](https://www.ebay.co.uk/usr/grumpyoldgamer)** &mdash; £24.99, UK based, ships worldwide &mdash; this is a great complete starter kit including USB cable, NFC cards, NFC stickers and rubber feet.

## ACR122U

:exclamation: Add a photo of reader.

- Plug and play.
- Reasonably priced from around $30 to $50 USD depending on vendor.
- Comes with a nice injection molded case and integrated USB cable.
- Has excellent short range reading but poor range beyond that.
- The PCB is quite large which can make it difficult to use in custom projects.

### Known Issues

There is a version of the ACR122U reader which is not compatible with TapTo. It's not possible to tell from a shop listing, and we currently have no ETA on support.

Most listings are fine, but be aware of the risk. Your best bet is to not buy the literal cheapest listing available.

### Where To Buy

The ACR122U has been cloned for years and is readily available by searching for "ACR122U" on sites like Amazon, eBay and AliExpress.

These are some confirmed working listings submitted by users:

- https://www.amazon.com/dp/B07KRKPWYC
- https://www.ebay.co.uk/itm/145044206870

## PN532 Module

:exclamation: More informaton required about compatible converters.

:warning: The PN532 module is **not** a ready-to-use reader, it's a bare PCB module intended for custom projects. It can used directly with TapTo via a *USB to TTL converter*.

![](/assets/images/readers/pn532.jpg)

- Very cheap price, sometimes as little as $1 USD.
- Very small footprint and is great for custom projects.
- High quality modules have excellent short and long range reading (up to 5cm without interference).
- Plug and play when paired with a USB to TTL converter.
- Has no case and will not function at all without additonal components.
- Build quality can be very variable depending on vendor, resulting in non-functioning modules or poor read range.
- Requires soldering equipment (but is a great beginner project).

### Known Issues

Build quality of these modules can vary a lot between vendors. In general they are fine, but be cautious about choosing the cheapest possible option. One concern is shipping, sometimes modules can be shipped loose and unprotected, and are easy to be damaged.

:exclamation: Add photos of inductor comparison.

Some modules may use low quality inductors connecting the antenna. This can result in a module that appears to function fine, but has extremely poor read range. It is possible to replace the inductors with new ones that are to spec.

Not necessarily related to the module, but some USB to TTL converters can have poor quality voltage regulators which are not enough to power the module properly.

### Where To Buy

The PN532 module is extremely common in the hobby electronics space and very easy to find. You can search for "PN532 module" on sites such as Amazon, eBay and AliExpress, or check out your local hobby electronics shops and websites.
