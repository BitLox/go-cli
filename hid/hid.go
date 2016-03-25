package hid

import (
	hid "github.com/GeertJohan/go.hid"

	"bytes"
	"encoding/binary"

	"bitlox/logger"
	"errors"
	"fmt"
	"time"
)

const CHUNK_SIZE = 32

const MAX_PAYLOAD = ^uint32(0)

var ERR_PAYLOAD_TOO_LARGE = errors.New(fmt.Sprintf("Payload exceeds %d bytes", MAX_PAYLOAD))

func GetDevice(vendorId, productId uint16) (*hid.Device, error) {
	infoList, err := hid.Enumerate(vendorId, productId)
	if err != nil {
		return nil, err
	}
	device, err := infoList[0].Device()
	if err != nil {
		return nil, err
	}
	return device, nil
}

func Write(dev *hid.Device, data []byte) error {

	var thisWrite, remainder []byte

	// if the data has no terminator, then this is the first run add
	// the magic bytes and the terminatopr to the data
	if bytes.Index(data, BITLOX_TERMINATOR) < 0 {
		data = append(BITLOX_MAGIC, data...)
		data = append(data, BITLOX_TERMINATOR...)
	}
	// prepend the report byte every time
	data = append(BITLOX_REPORT, data...)

	// test to see if the data is below the chunk size
	if len(data) < CHUNK_SIZE {
		// if so, make thisWrite be the full data slice and set the
		// remainder to an empty slice
		thisWrite = data
		remainder = []byte{}
	} else {
		logger.Debug("Breaking into chunks")
		// otherwise, take the front CHUNK_SIZE bytes off the front of
		// the data and set the rest to the remainder slice
		thisWrite = data[0:CHUNK_SIZE]
		remainder = data[CHUNK_SIZE:len(data)]
	}
	logger.Debugf("thisWrite %x\n", thisWrite)
	logger.Debugf("remainder %x\n", remainder)
	// then write this chunk to the device
	_, err := dev.Write(thisWrite)
	if err != nil {
		logger.Error("write error", err)
		return err
	}
	// if we have a remainder, then recurse
	if len(remainder) > 0 {
		time.Sleep(50 * time.Millisecond)
		return Write(dev, remainder)
	}
	// otherwise, return with no errors
	return nil
}

func WriteVariable(dev *hid.Device, cmd, data []byte) error {
	dataLen := len(data)
	if dataLen > int(MAX_PAYLOAD) {
		return ERR_PAYLOAD_TOO_LARGE
	}
	len := make([]byte, 4)
	binary.BigEndian.PutUint32(len, uint32(dataLen))
	cmd = append(cmd, len...)
	cmd = append(cmd, data...)
	return Write(dev, cmd)
}

func Read(dev *hid.Device, buf *bytes.Buffer) error {
	// make a buffer for this response
	response := make([]byte, CHUNK_SIZE)

	// and read into it from the device
	_, err := dev.Read(response)
	if err != nil {
		return err
	}

	// write the data we got to the buffer passed in
	_, err = buf.Write(response)
	if err != nil {
		return err
	}

	bufBytes := buf.Bytes()
	// test to see if the last byte is 0x23, if so, we may have cut
	// off magic bytes, so read again
	if bufBytes[len(bufBytes)-1] == 0x23 {
		logger.Debug("last byte is 0x23, reading more data")
		return Read(dev, buf)
	}
	//test to see if we have the magic bytes in our buffer
	magicIndex := bytes.Index(bufBytes, BITLOX_MAGIC)
	if magicIndex < 0 {
		logger.Debug("no magic bytes, reading again")
		return Read(dev, buf)
	}
	// now that we know we have magic bytes, strip off the leading
	// bullshit for the rest of the checks
	bufBytes = bufBytes[magicIndex:len(bufBytes)]
	// next, make sure there is enough room to have full command and
	// payload length bytes
	if len(bufBytes) < 8 {
		logger.Debug("Not enough room for command and payload size, reading again")
		return Read(dev, buf)
	}
	// next, get the payload length and make sure we have read the
	// full payload
	pSize := getPayloadSize(bufBytes)
	if len(bufBytes) < (8 + int(pSize)) {
		logger.Debug("More payload to read, reading again")
		return Read(dev, buf)
	}
	logger.Debug("finished read")
	// finally, we have all the data, return nil for no error
	return nil
}

func getCommand(data []byte) []byte {
	return data[2:4]
}

func getPayloadSizeBytes(data []byte) []byte {
	return data[4:8]
}

func getPayload(data []byte, size int32) []byte {
	return data[8 : 8+size]
}

func getPayloadSize(data []byte) int32 {
	sizeB := getPayloadSizeBytes(data)
	sizeBuf := bytes.NewReader(sizeB)
	var size int32
	binary.Read(sizeBuf, binary.BigEndian, &size)
	return size
}

func ParseResponse(buf *bytes.Buffer) (cmd byte, size int32, payload []byte) {
	// make a byte slice and read the byte buffer into it (this kills
	// the buffer)
	data := make([]byte, buf.Len())
	buf.Read(data)

	magicIndex := bytes.Index(data, BITLOX_MAGIC)
	// strip any leading bullshit
	data = data[magicIndex : len(data)-1]

	// extract the command
	command := getCommand(data)
	// all commands have 0x00, get second byte
	cmd = command[1]

	// get the payload size
	size = getPayloadSize(data)

	// extract it
	payload = getPayload(data, size)

	logger.Debugf("parsed: cmd: %#04x size: %d (%#08x) payload: %x\n", cmd, size, size, payload)
	// return predeclared returns
	return
}
