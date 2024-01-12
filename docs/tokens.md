# Token Hardware

Tokens tell the reader and software what action to take. They **do not** contain any games, but instead contain a reference to a game or instruction. This is in the form of a small piece of text stored on the token.

NFC tokens come in many form factors and standards. The form factor is entirely your preference, but the standard may affect compatibility with TapTo and your particular reader hardware.

:question: If in doubt, **NTAG215 NFC cards** are a very solid option with the best software and community label compatibility.

* [Where To Buy](#where-to-buy)
* [NTAG](#ntag)
* [MIFARE](#mifare)
* [Amiibos](#amiibos)
* [Lego Dimensions](#lego-dimensions)

## Where To Buy

NFC tokens are readily available on Amazon, eBay and AliExpress by searching for the standard, form factor and storage size if applicable (e.g. NTAG215 NFC card, NTAG213 NFC sticker). You'll also find them on many local "NFC" and "ID" related stores. At this stage no difference in quality has been noted between suppliers. AliExpress is a *great* place to get NFC tokens in bulk.

## NTAG

This NTAG standard has the best compatibility with TapTo and is readily available in form factors such as cards, stickers and keychains.

There are multiple NTAG types that have been confirmed working with TapTo. The only difference between them is storage size:

| Standard               | Storage                                            |
|------------------------|----------------------------------------------------|
| NTAG213                | 144 bytes                                          |
| NTAG215                | 504 bytes                                          |
| NTAG216                | 888 bytes                                          |

The NTAG215 is a generous amount of storage for TapTo and you shouldn't have any trouble with it. The main consideration is just needing to fit the full path of a game file on the token.

## MIFARE

:warning: MIFARE is only partially supported by TapTo. It can read cards just fine, but it can only write to them after the card has been NDEF formatted with a third party application.

MIFARE Classic 1K cards have been confirmed working with TapTo. It has 716 bytes of storage. If you order a commercial NFC reader that comes with tokens, this is often the standard you'll receive.

They're totally usable with TapTo, but generally not recommend at this stage if buying new tokens separately.

## Amiibos

Reading Amiibos is supported by TapTo. Internally they are NTAG215 tokens. They can't be written to, but their ID can be read and mapped to a custom action with TapTo.

:exclamation: Link to nfc.csv configuration.

## Lego Dimensions

Reading Lego Dimensions figurines is supported by TapTo. Internally they are NTAG213 tokens. They can't be written to, but their ID can be read and mapped to a custom action with TapTo.

:exclamation: Link to nfc.csv configuration.
