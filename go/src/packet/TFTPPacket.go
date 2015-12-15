package packet

import (
	"bytes"
	"encoding/binary"
)

const HEADER_SIZE = 4
const MAX_PACKET_SIZE uint16 = 512 + HEADER_SIZE
const RESET_BLOCK_NUM uint16 = 65535

const READ_REQUEST uint16 = 1
const WRITE_REQUEST uint16 = 2
const DATA uint16 = 3
const ACKNOWLEDGEMENT uint16 = 4
const ERROR uint16 = 5

const ERROR_NOT_DEFINED uint16 = 0
const ERROR_FILE_NOT_FOUND uint16 = 1
const ERROR_DISK_FULL uint16 = 3
const ERROR_ILLEGAL_TFTP_OPERATION uint16 = 4
const ERROR_UNKNOWN_TRANSFER_ID uint16 = 5
const ERROR_FILE_ALREADY_EXISTS uint16 = 6

var ERROR_MESSAGES = map[uint16]string{
	ERROR_NOT_DEFINED :  "Error not defined",
	ERROR_FILE_NOT_FOUND :  "File not found",
	ERROR_ILLEGAL_TFTP_OPERATION :  "Illegal TFTP operation",
	ERROR_UNKNOWN_TRANSFER_ID :  "Unknown transfer id",
	ERROR_FILE_ALREADY_EXISTS : "File already exists",
	ERROR_DISK_FULL :  "Disk full or allocation exceeded"}

const NULL_TERMINATOR byte = 0
const MODE string = "octet"

var blockNumber uint16
var errorCode uint16
var filename string
var data []byte
var length int
var packetType uint16
var port uint16
var address string
var errorMessage string

type TFTPPacket struct {
	packetType uint16
	address    string
	port       uint16
}

func makeTFTPPacket(packetType1 uint16, address1 string, port1 uint16) TFTPPacket {
	packetType = packetType1
	address = address1
	port = port1
	return TFTPPacket{packetType1, address1, port1}
}

func getLength() int {
	return length
}

func setLength(length1 int) {
	length = length1
}

func setErrorCode(eCode uint16) {
	switch (eCode) {
	case ERROR_NOT_DEFINED:
	case ERROR_FILE_NOT_FOUND:
	case ERROR_DISK_FULL:
	case ERROR_FILE_ALREADY_EXISTS:
	case ERROR_ILLEGAL_TFTP_OPERATION:
	case ERROR_UNKNOWN_TRANSFER_ID:
		errorCode = eCode
		break
	default:
		errorCode = 0
	}
}

func setErrorMessage(errorMessage1 string) {
	errorMessage = errorMessage1
}

func getType() uint16 {
	return packetType
}

func setAddress(address1 string) {
	address = address1
}

func getAddress() string {
	return address
}

func setPort(port1 uint16) {
	port = port1
}

func getPort() uint16 {
	return port
}

func getFilename() string {
	return filename
}

func setFilename(filename1 string) {
	filename = filename1
}

func getBlockNumber() uint16 {
	return blockNumber
}

func setBlockNumber(blockNumber1 uint16) {
	blockNumber = blockNumber1
}

func getData() []byte {
	return data
}

func setData(data []byte) {
	data = data
}

func buildDatagram() []byte {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, packetType)
	switch (packetType) {
	case READ_REQUEST:
	case WRITE_REQUEST:
		buffer.WriteString(filename)
		buffer.WriteByte(NULL_TERMINATOR)
		buffer.WriteString(MODE)
		buffer.WriteByte(NULL_TERMINATOR)
		return buffer.Bytes()
	case DATA:
		binary.Write(buffer, binary.BigEndian, blockNumber)
		buffer.Write(data)
		return buffer.Bytes()
	case ACKNOWLEDGEMENT:
		binary.Write(buffer, binary.BigEndian, blockNumber)
		return buffer.Bytes()
	case ERROR:
		binary.Write(buffer, binary.BigEndian, errorCode)
		buffer.WriteString(errorMessage)
		buffer.WriteByte(NULL_TERMINATOR)
		return buffer.Bytes()
	default:
		return nil
	}
	return nil
}

func parseTFTPPacket(buf []byte) TFTPPacket {
	packetType := binary.BigEndian.Uint16(buf[0:2])
	packet := makeTFTPPacket(packetType, getAddress(), getPort())
	switch (packetType) {
	case READ_REQUEST:
	case WRITE_REQUEST:
		filename := bytes.Split(buf[2:], []byte{0})[0]
		setFilename(string(filename))
		return packet
	case DATA:
		blockNumber := binary.BigEndian.Uint16(buf[2:4])
		setBlockNumber(blockNumber)
		setLength(len(buf[4:]))
		data := buf[4:]
		setData(data)
		return packet
	case ACKNOWLEDGEMENT:
		blockNumber := binary.BigEndian.Uint16(buf[2:4])
		setBlockNumber(blockNumber)
		return packet
	case ERROR:
		errorCode := binary.BigEndian.Uint16(buf[2:4])
		setErrorCode(errorCode)
		setLength(len(buf[4:]))
		errorMessage := string(buf[4:])
		setErrorMessage(errorMessage)
		return packet
	default:
		return TFTPPacket{}
	}
	return TFTPPacket{}
}
