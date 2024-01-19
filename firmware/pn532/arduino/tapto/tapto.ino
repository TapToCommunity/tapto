#include <SoftwareSerial.h>
#include <PN532_SWHSU.h>
#include <PN532.h>

int readDelay = 300;

SoftwareSerial SWSerial(2, 3);
PN532_SWHSU PN532SWHSU(SWSerial);
PN532 NFC(PN532SWHSU);

void readNFC()
{
  boolean wasRead;
  uint8_t uid[] = {0, 0, 0, 0, 0, 0, 0};
  uint8_t uidLength;

  wasRead = NFC.readPassiveTargetID(PN532_MIFARE_ISO14443A, &uid[0], &uidLength);
  if (!wasRead)
  {
    return;
  }

  char uidString[32] = {0};

  if (uidLength == 4)
  {
    sprintf(uidString, "%02X:%02X:%02X:%02X", uid[0], uid[1], uid[2], uid[3]);
  }
  else if (uidLength == 7)
  {
    sprintf(uidString, "%02X:%02X:%02X:%02X:%02X:%02X:%02X", uid[0], uid[1], uid[2], uid[3], uid[4], uid[5], uid[6]);
  }
  else
  {
    Serial.println("Invalid UID length");
  }

  Serial.print("**read:");
  Serial.print(uidString);
  Serial.print(",");
  Serial.println();

  delay(readDelay);
}

void setup()
{
  Serial.begin(115200);
  Serial.println("TapTo firmware v0.1");

  NFC.begin();

  uint32_t pn532_version = NFC.getFirmwareVersion();
  if (!pn532_version)
  {
    Serial.print("PN53X module not found");
    while (1)
      ;
  }

  Serial.print("Found module PN5");
  Serial.println((pn532_version >> 24) & 0xFF, HEX);
  Serial.print("Firmware v");
  Serial.print((pn532_version >> 16) & 0xFF, DEC);
  Serial.print('.');
  Serial.println((pn532_version >> 8) & 0xFF, DEC);

  NFC.SAMConfig();
}

void loop()
{
  while (1)
  {
    readNFC();
  }
}
