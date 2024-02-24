#include <SoftwareSerial.h>
#include "PN532_SWHSU.h"
#include "PN532.h"
#include "NfcAdapter.h"

// pins connected to module SDA and SCL
#define SDA_PIN 3
#define SCL_PIN 2
// pin connected to arduino status LED
#define LED_PIN LED_BUILTIN

const char *version = "0.1"; // firmware version
uint16_t readDelay = 300;    // ms between reads
uint8_t connectMax = 10;     // number of times to try to connect to module
uint8_t connectDelay = 500;  // ms between connection attempts

SoftwareSerial swserial(SDA_PIN, SCL_PIN);
PN532_SWHSU swhsu(swserial);
PN532 nfc(swhsu);
NfcAdapter nfcAdapter(swhsu);

boolean readTag()
{
  if (!nfcAdapter.tagPresent())
  {
    digitalWrite(LED_PIN, LOW);
    return;
  }

  NfcTag tag = nfcAdapter.read();

  digitalWrite(LED_PIN, HIGH);

  Serial.print("#read=");
  Serial.print(tag.getUidString());
  Serial.print(",");

  NdefMessage message = tag.getNdefMessage();

  if (message.getRecordCount() == 0)
  {
    Serial.println();
    return;
  }

  tag.print();
  Serial.println();
}

void queryModule()
{
  uint8_t connectTries = 0;
  uint32_t moduleVersion = 0;
  Serial.print("Querying PN53x board...");
  while (!moduleVersion)
  {
    Serial.print(".");
    moduleVersion = nfc.getFirmwareVersion();
    connectTries++;
    if (connectTries > connectMax)
    {
      Serial.println();
      Serial.println("#error=PN53x module not found");
      while (1)
      {
        digitalWrite(LED_PIN, HIGH);
        delay(connectDelay);
        digitalWrite(LED_PIN, LOW);
        delay(connectDelay);
      }
    }
    delay(connectDelay);
  }

  Serial.print("found chip PN5");
  Serial.print((moduleVersion >> 24) & 0xFF, HEX);
  Serial.print(" (firmware v");
  Serial.print((moduleVersion >> 16) & 0xFF, DEC);
  Serial.print('.');
  Serial.print((moduleVersion >> 8) & 0xFF, DEC);
  Serial.println(")");
}

void setup()
{
  Serial.begin(115200);
  while (!Serial)
    delay(10);

  Serial.println("TapTo Firmware v" + String(version));
  Serial.println("#version=tapto-" + String(version));

  pinMode(LED_PIN, OUTPUT);
  nfc.begin();

  queryModule();

  nfc.setPassiveActivationRetries(0xFF);
  nfc.SAMConfig();

  // TODO: this doubles up on earlier commands, but you can't query version
  //       info with just the nfcadapter
  nfcAdapter.begin();
}

void loop()
{
  while (1)
  {
    readTag();
    delay(readDelay);
  }
}
