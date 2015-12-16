package main

import (
	"flag"
	"net"
	"bytes"
	"encoding/binary"
	"time"
	"fmt"
)

var port* string
const TIMEOUT = time.Second * 3
const MAX_RETRIES = 5
var content = make(map[string]string) // In memory file system

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

type Packet struct {
	BlockNumber  uint16
	ErrorCode    uint16
	Filename     string
	Data         []byte
	Length       int
	PacketType   uint16
	Port         uint16
	Address      string
	ErrorMessage string
}

func NewPacket(packetType uint16, blkNum uint16, Filename string, eCode uint16, eMsg string, data []byte) Packet {
	return Packet{
		PacketType: packetType,
		BlockNumber:blkNum,
		Filename:Filename,
		ErrorCode:eCode,
		ErrorMessage:eMsg,
		Data:data,
	}
}

func main() {
	port = flag.String("port", "8800", "specify a port to listen on")
	flag.Parse()
	panic(Start(":" + *port))
}

func Start(addr string) error {
	uaddr, _ := net.ResolveUDPAddr("udp", addr)
	uconn, _ := net.ListenUDP("udp", uaddr)
	fmt.Println("Server running on port: " + *port + ". Hit 'Ctrl-C' to stop it and all current transfers")
	for {
		buf := make([]byte, MAX_PACKET_SIZE)
		n, udpaddr, _ := uconn.ReadFromUDP(buf)
		buf = buf[:n]
		pkt := parseTFTPPacket(buf)
		go handleReq(udpaddr, pkt) // new thread
	}
}

func buildDatagram(p Packet) []byte {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, p.PacketType)
	switch (p.PacketType) {
	case READ_REQUEST:
		buffer.WriteString(p.Filename)
		buffer.WriteByte(NULL_TERMINATOR)
		buffer.WriteString(MODE)
		buffer.WriteByte(NULL_TERMINATOR)
		return buffer.Bytes()
	case WRITE_REQUEST:
		buffer.WriteString(p.Filename)
		buffer.WriteByte(NULL_TERMINATOR)
		buffer.WriteString(MODE)
		buffer.WriteByte(NULL_TERMINATOR)
		return buffer.Bytes()
	case DATA:
		binary.Write(buffer, binary.BigEndian, p.BlockNumber)
		buffer.Write(p.Data)
		return buffer.Bytes()
	case ACKNOWLEDGEMENT:
		binary.Write(buffer, binary.BigEndian, p.BlockNumber)
		return buffer.Bytes()
	case ERROR:
		binary.Write(buffer, binary.BigEndian, p.ErrorCode)
		buffer.WriteString(p.ErrorMessage)
		buffer.WriteByte(NULL_TERMINATOR)
		return buffer.Bytes()
	default:
		return nil
	}
	return nil
}

func parseTFTPPacket(buf []byte) Packet {
	pktType := binary.BigEndian.Uint16(buf[0:2])
	switch (pktType) {
	case READ_REQUEST:
		Filename := bytes.Split(buf[2:], []byte{0})[0]
		return Packet{
			PacketType: pktType,
			Filename: string(Filename),
		}
	case WRITE_REQUEST:
		Filename := bytes.Split(buf[2:], []byte{0})[0]
		return Packet{
			PacketType: pktType,
			Filename: string(Filename),
		}
	case DATA:
		blockNumber := binary.BigEndian.Uint16(buf[2:4])
		fmt.Println("data length: ", len(buf[:]), " content ", buf[4:], " string ", string(buf[4:]))
		return Packet{
			PacketType: pktType,
			BlockNumber: blockNumber,
			Length: len(buf[4:]),
			Data: buf[4:],
		}
	case ACKNOWLEDGEMENT:
		blockNumber := binary.BigEndian.Uint16(buf[2:4])
		return Packet{
			PacketType: pktType,
			BlockNumber: blockNumber,
		}
	case ERROR:
		errorCode := binary.BigEndian.Uint16(buf[2:4])
		errorMessage := string(buf[4:])
		return Packet{
			PacketType: pktType,
			ErrorMessage: errorMessage,
			ErrorCode:errorCode,
			Length: len(buf[4:]),
		}
	default:
		return Packet{
		}
	}
	return Packet{
	}
}

func handleReq(addr *net.UDPAddr, reqPacket Packet) {
	fmt.Println("serving client req from: " + addr.String())
	if reqPacket.PacketType != READ_REQUEST && reqPacket.PacketType != WRITE_REQUEST {
		fmt.Println("wrong packet type for new connection, ignoring")
		return
	}
	clientaddr, _ := net.ResolveUDPAddr("udp", addr.String())
	switch reqPacket.PacketType {
	case READ_REQUEST:
		handleReadReq(reqPacket, clientaddr)
	case WRITE_REQUEST:
		handleWriteReq(reqPacket, clientaddr)
	default:
		fmt.Println("unknown packet type")
	}
}

func handleWriteReq(writeReq Packet, addr *net.UDPAddr) {
	fmt.Println("Storing file: " + writeReq.Filename)
	listaddr, _ := net.ResolveUDPAddr("udp", ":0")
	con, _ := net.DialUDP("udp", listaddr, addr)
	if _, ok := content[writeReq.Filename] ; ok {
		errPkt := NewPacket(ERROR, 0, "", ERROR_FILE_ALREADY_EXISTS, ERROR_MESSAGES[ERROR_FILE_ALREADY_EXISTS], nil)
		con.Write(buildDatagram(errPkt))
		return
	}
	buffer := new(bytes.Buffer)
	ackPkt := NewPacket(ACKNOWLEDGEMENT, 0, "", 0, "", nil)
	con.Write(buildDatagram(ackPkt))
	curblk := uint16(1)
	buf := make([]byte, MAX_PACKET_SIZE)
	for {
		n, _, _ := con.ReadFromUDP(buf)
		dataPacket := parseTFTPPacket(buf[:n])
		if dataPacket.PacketType != DATA {
			errPkt := NewPacket(ERROR, 0, "", ERROR_NOT_DEFINED, "unexpected packet type", nil)
			con.Write(buildDatagram(errPkt))
			return
		}
		if dataPacket.BlockNumber == curblk - 1 {
			ackPkt := NewPacket(ACKNOWLEDGEMENT, curblk - 1, "", 0, "", nil)
			con.Write(buildDatagram(ackPkt))
			continue
		} else if dataPacket.BlockNumber != curblk {
			fmt.Println("unexpected blocknum, aborting transfer")
			return
		}
		buffer.Write(dataPacket.Data)
		ackPkt := NewPacket(ACKNOWLEDGEMENT, curblk, "", 0, "", nil)
		_, _ = con.Write(buildDatagram(ackPkt))
		if len(dataPacket.Data) < 512 {
			fmt.Println("Last packet, updating map...")
			fmt.Println("storing in map::", buffer.String())
			content[writeReq.Filename] = buffer.String()
			return
		}
		curblk++
	}
}

func handleReadReq(readReq Packet, addr *net.UDPAddr)  {
	fmt.Println("Sending file " + readReq.Filename + " to " + addr.String())
	listaddr, _ := net.ResolveUDPAddr("udp", ":0")
	con, _ := net.DialUDP("udp", listaddr, addr)
	if _, ok := content[readReq.Filename] ; !ok {
		errPkt := NewPacket(ERROR, 0, "", ERROR_FILE_NOT_FOUND, ERROR_MESSAGES[ERROR_FILE_NOT_FOUND], nil)
		con.Write(buildDatagram(errPkt))
		return
	}
	data := bytes.NewBufferString(content[readReq.Filename])
	buf := make([]byte, MAX_PACKET_SIZE)
	blknum := uint16(1)
	for len(buf) == 516 {
		n, _ := data.Read(buf)
		buf = buf[:n]
		dataPacket := NewPacket(DATA, blknum, readReq.Filename, 0, "", buf)
		sendDataPacket(dataPacket, con)
		blknum++
	}
	fmt.Println("done with transfer")
}

func sendDataPacket(d Packet, con *net.UDPConn) {
	con.Write(buildDatagram(d))
	ack := make([]byte, MAX_PACKET_SIZE)
	for retries := 0; retries <= MAX_RETRIES; retries++  {
		con.SetReadDeadline(time.Now().Add(TIMEOUT))
		n, _, err := con.ReadFromUDP(ack)
		if err != nil {
			fmt.Println("retransmit")
			con.Write(buildDatagram(d))
			continue
		}
		ackPkt := parseTFTPPacket(ack[:n])
		if ackPkt.PacketType != ACKNOWLEDGEMENT {
			fmt.Println("ack expected but got something else, ignoring")
			return
		}
		if ackPkt.BlockNumber != d.BlockNumber {
			fmt.Println("got ack", d.BlockNumber, " but expected ack ", ackPkt.BlockNumber)
			continue
		}
		break
	}
}