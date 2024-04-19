/*
TapTo
Copyright (C) 2023 Gareth Jones

This file is part of TapTo.

TapTo is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

TapTo is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with TapTo.  If not, see <http://www.gnu.org/licenses/>.
*/

package tokens

import (
	"encoding/hex"
	"bytes"
	"errors"
	"fmt"

	"github.com/clausecker/nfc/v2"
	"github.com/rs/zerolog/log"
)

const (
	NTAG_213_CAPACITY_BYTES = 114
	NTAG_213_IDENTIFIER     = 0x12

	NTAG_215_CAPACITY_BYTES = 496
	NTAG_215_IDENTIFIER     = 0x3E

	NTAG_216_CAPACITY_BYTES = 872
	NTAG_216_IDENTIFIER     = 0x6D
)

// Can be identified by matching blocks 0x03-0x07
// https://github.com/RfidResearchGroup/proxmark3/blob/master/client/src/cmdhfmfu.c
var LEGO_DIMENSIONS_MATCHER = []byte{
	//0xE1, 0x10, 0x12, 0x00, // Skip as we never read 0x03
	0x01, 0x03, 0xA0, 0x0C,
	0x34, 0x03, 0x13, 0xD1,
	0x01, 0x0F, 0x54, 0x02,
	0x65, 0x6E}

// Can be identified by matching address 0x09-0x0F
var AMIIBO_MATCHER = []byte{
	      0x48, 0x0F, 0xE0,
	0xF1, 0x10, 0xFF, 0xEE}

func ReadNtag(pnd nfc.Device) ([]byte, error) {
	blockCount, err := getNtagBlockCount(pnd)
	if err != nil {
		return []byte{}, err
	}

	log.Debug().Msgf("NTAG has %d blocks", blockCount)

	header, _ := comm(pnd, []byte{READ_COMMAND, byte(0)}, 16)
	if bytes.Equal(header[9:], AMIIBO_MATCHER) {
		log.Info().Msg("found Amiibo")
		amiibo, _ := comm(pnd, []byte{READ_COMMAND, byte(21)}, 16)
		amiibo = amiibo[:8]
		log.Info().Msg("Amiibo identifier:" + hex.EncodeToString(amiibo))
		return amiibo, nil
	}

	allBlocks := make([]byte, 0)
	currentBlock := 4

	for i := 0; i <= (blockCount / 4); i++ {
		blocks, err := comm(pnd, []byte{READ_COMMAND, byte(currentBlock)}, 16)
		if err != nil {
			return nil, err
		}

		if byte(currentBlock) == 0x04 && bytes.Equal(blocks[0:14], LEGO_DIMENSIONS_MATCHER) {
			log.Info().Msg("found Lego Dimensions tag")
			return []byte{}, nil
		}

		allBlocks = append(allBlocks, blocks...)
		currentBlock = currentBlock + 4

		if bytes.Contains(allBlocks, NDEF_END) {
			// Once we find the end of the NDEF text record there is no need to
			// continue reading the rest of the card.
			// This should make things "load" quicker
			log.Debug().Msg("found end of ndef record")
			break
		}
	}

	return allBlocks, nil
}

func WriteNtag(pnd nfc.Device, text string) ([]byte, error) {
	var payload, err = BuildMessage(text)
	if err != nil {
		return nil, err
	}

	cardCapacity, err := getNtagCapacity(pnd)
	if err != nil {
		return nil, err
	}

	if len(payload) > cardCapacity {
		return nil, errors.New(fmt.Sprintf("Payload too big for card: [%d/%d] bytes used\n", len(payload), cardCapacity))
	}

	var startingBlock byte = 0x04
	for i, chunk := range chunkBy(payload, 4) {
		for len(chunk) < 4 {
			chunk = append(chunk, []byte{0x00}...)
		}
		var tx = []byte{WRITE_COMMAND, startingBlock + byte(i)}
		tx = append(tx, chunk...)
		_, err := comm(pnd, tx, 1)
		if err != nil {
			return nil, err
		}
	}

	return payload, nil
}

func getNtagBlockCount(pnd nfc.Device) (int, error) {
	// Find tag capacity by looking in block 3 (capability container)
	tx := []byte{READ_COMMAND, 0x03}
	rx := make([]byte, 16)

	timeout := 0
	_, err := pnd.InitiatorTransceiveBytes(tx, rx, timeout)
	if err != nil {
		return 0, err
	}

	switch rx[2] {
	case 0x12:
		// NTAG213. (144 -4) / 4
		return 35, nil
	case 0x3E:
		// NTAG215. (504 - 4) / 4
		return 125, nil
	case 0x6D:
		// NTAG216. (888 -4) / 4
		return 221, nil
	default:
		// fallback to NTAG213
		return 35, nil
	}
}

func getNtagCapacity(pnd nfc.Device) (int, error) {
	// Find tag capacity by looking in block 3 (capability container)
	tx := []byte{READ_COMMAND, 0x03}
	rx := make([]byte, 16)

	timeout := 0
	_, err := pnd.InitiatorTransceiveBytes(tx, rx, timeout)
	if err != nil {
		return 0, err
	}

	// https://github.com/adafruit/Adafruit_MFRC630/blob/master/docs/NTAG.md#capability-container
	switch rx[2] {
	case NTAG_213_IDENTIFIER:
		return NTAG_213_CAPACITY_BYTES, nil
	case NTAG_215_IDENTIFIER:
		return NTAG_215_CAPACITY_BYTES, nil
	case NTAG_216_IDENTIFIER:
		return NTAG_216_CAPACITY_BYTES, nil
	default:
		// fallback
		return NTAG_213_CAPACITY_BYTES, nil
	}
}

func chunkBy[T any](items []T, chunkSize int) (chunks [][]T) {
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}
	return append(chunks, items)
}
