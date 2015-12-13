package com.igneous.hw.main;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.math.BigInteger;
import java.net.DatagramPacket;
import java.net.InetAddress;
import java.util.HashMap;
import java.util.Map;
import java.util.logging.Level;
import java.util.logging.Logger;

public class TFTPPacket {
    public static final int HEADER_SIZE = 4;
    public static final int MAX_PACKET_SIZE = 512 + HEADER_SIZE;
    public static final int RESET_BLOCK_NUM = 65535;

    public static final byte READ_REQUEST = 1;
    public static final byte WRITE_REQUEST = 2;
    public static final byte DATA = 3;
    public static final byte ACKNOWLEDGEMENT = 4;
    public static final byte ERROR = 5;

    public static final byte ERROR_NOT_DEFINED = 0;
    public static final byte ERROR_FILE_NOT_FOUND = 1;
    public static final byte ERROR_DISK_FULL = 3;
    public static final byte ERROR_ILLEGAL_TFTP_OPERATION = 4;
    public static final byte ERROR_UNKNOWN_TRANSFER_ID = 5;
    public static final byte ERROR_FILE_ALREADY_EXISTS = 6;
    public static final Map<Byte, String> ERROR_MESSAGES = new HashMap<>(3);
    static {
        ERROR_MESSAGES.put(ERROR_NOT_DEFINED, "Error not defined");
        ERROR_MESSAGES.put(ERROR_FILE_NOT_FOUND, "File not found");
        ERROR_MESSAGES.put(ERROR_ILLEGAL_TFTP_OPERATION, "Illegal TFTP operation");
        ERROR_MESSAGES.put(ERROR_UNKNOWN_TRANSFER_ID, "Unknown transfer id");
        ERROR_MESSAGES.put(ERROR_FILE_ALREADY_EXISTS, "File already exists");
        ERROR_MESSAGES.put(ERROR_DISK_FULL, "Disk full or allocation exceeded");
    }
    public static final byte NULL_TERMINATOR = 0;
    public static final String MODE = "octet";

    private static final Logger LOGGER = Logger.getLogger(TFTPPacket.class.getSimpleName());
    private byte[] opCode = new byte[2];
    private byte[] blockNumber = new byte[2];
    private byte[] errorCode = new byte[2];
    private String filename;
    private byte[] data;
    private int length;
    private int type;
    private int port;
    private InetAddress address;
    private String errorMessage;

    public TFTPPacket(int type, InetAddress address, int port) {
        this.type = type;
        this.address = address;
        this.port = port;
    }

    public int getLength() {
        return length;
    }

    public void setLength(int length) {
        this.length = length;
    }

    public void setErrorCode(byte eCode) {
        this.errorCode[0] = 0;
        switch (eCode) {
            case ERROR_NOT_DEFINED:
            case ERROR_FILE_NOT_FOUND:
            case ERROR_DISK_FULL:
            case ERROR_FILE_ALREADY_EXISTS:
            case ERROR_ILLEGAL_TFTP_OPERATION:
            case ERROR_UNKNOWN_TRANSFER_ID:
                this.errorCode[1] = eCode;
                break;
            default:
                this.errorCode[1] = -1;
        }
    }

    public void setErrorMessage(String errorMessage) {
        this.errorMessage = errorMessage;
    }

    public int getType() {
        return type;
    }

    public InetAddress getAddress() {
        return address;
    }

    public int getPort() {
        return port;
    }

    public String getFilename() {
        return filename;
    }

    public void setFilename(String filename) {
        this.filename = filename;
    }

    public byte[] getBlockNumber() {
        return blockNumber;
    }

    public void setBlockNumber(byte[] blockNumber) {
        this.blockNumber = blockNumber;
    }

    public byte[] getData() {
        return data;
    }

    public void setData(byte[] data) {
        this.data = data;
    }

    public DatagramPacket buildDatagram() {
        try {
            byte[] buffer = new byte[length + HEADER_SIZE];
            ByteArrayOutputStream bos = new ByteArrayOutputStream();
            DatagramPacket sendPacket = new DatagramPacket(buffer, buffer.length, address, port);
            opCode[0] = 0;
            opCode[1] = BigInteger.valueOf(type).toByteArray()[0];
            bos.write(opCode, 0, 2);
            switch (opCode[1]) {
                case READ_REQUEST:
                case WRITE_REQUEST:
                    bos.write(filename.getBytes());
                    bos.write(NULL_TERMINATOR);
                    bos.write(MODE.getBytes());
                    bos.write(NULL_TERMINATOR);
                    sendPacket.setData(bos.toByteArray());
                    return sendPacket;
                case DATA:
                    bos.write(blockNumber);
                    bos.write(data);
                    sendPacket.setData(bos.toByteArray());
                    return sendPacket;
                case ACKNOWLEDGEMENT:
                    bos.write(blockNumber);
                    sendPacket.setData(bos.toByteArray());
                    return sendPacket;
                case ERROR:
                    bos.write(errorCode);
                    bos.write(errorMessage.getBytes());
                    bos.write(NULL_TERMINATOR);
                    sendPacket.setData(bos.toByteArray());
                    return sendPacket;
                default:
                    return null;
            }
        } catch (Exception e) {
            LOGGER.log(Level.SEVERE, e.getMessage(), e);
        }
        return null;
    }

    public static TFTPPacket parseTFTPPacket(DatagramPacket datagramPacket) {
        try {
            ByteArrayInputStream bis = new ByteArrayInputStream(datagramPacket.getData(), 0, datagramPacket.getData().length);
            byte[] opCode = new byte[2];
            bis.read(opCode);
            TFTPPacket packet = new TFTPPacket(opCode[1], datagramPacket.getAddress(), datagramPacket.getPort());
            switch (opCode[1]) {
                case READ_REQUEST:
                case WRITE_REQUEST:
                    byte[] arr = new byte[bis.available()];
                    byte next = (byte) bis.read();
                    int count = 0;
                    while (next != NULL_TERMINATOR) {
                        arr[count++] = next;
                        next = (byte) bis.read();
                    }
                    packet.setFilename(new String(arr));
                    return packet;
                case DATA:
                    bis.read(packet.blockNumber);
                    packet.setLength(datagramPacket.getLength());
                    packet.data = new byte[bis.available()];
                    bis.read(packet.data);
                    return packet;
                case ACKNOWLEDGEMENT:
                    bis.read(packet.blockNumber);
                    return packet;
                case ERROR:
                    bis.read(packet.errorCode);
                    packet.setLength(datagramPacket.getLength());
                    packet.data = new byte[bis.available() - 1];
                    bis.read(packet.data);
                    String errorMessage = new String(packet.data);
                    packet.setErrorMessage(errorMessage);
                    return packet;
                default:
                    return null;
            }
        } catch (Exception e) {
            LOGGER.log(Level.SEVERE, e.getMessage(), e);
        }
        return null;
    }
}
