package main

import (
	"flag"
	"io"
	"os"
	"net"
	"log"
	"github.com/aditcoding/tftp/packet"
)

var port uint16

func reader(path string) (r io.Reader, err error) {
	r, err = os.Open(path)
	return
}

func writer(path string) (w io.Writer, err error) {
	w, err = os.Create(path)
	return
}

func main() {
	port = flag.String("port", "8800", "specify a port to listen on")
	flag.Parse()
	panic(Serve(":" + *port))
}

func Serve(addr string) error {
	uaddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}

	uconn, err := net.ListenUDP("udp", uaddr)
	if err != nil {
		return err
	}

	for { // read in new requests
		buf := make([]byte, packet.MAX_PACKET_SIZE)
		n, ua, err := uconn.ReadFromUDP(buf)
		if err != nil {
			return err
		}

		log.Println("Server running on port: " + port + ". Enter 'q' to stop it and all current transfers")

		buf = buf[:n]
		packet := packet.parseTFTPPacket(buf)
		if packet == nil {
			log.Printf("Invalid packet")
			continue
		}

		//go s.HandleClient(ua, packet)
	}
}

/*
private final List<Transfer> transfers = new ArrayList<>();
private final DatagramSocket socket;
private boolean stopServer;
private DatagramPacket receiveDatagram;
private final Map<String, String> content = new HashMap<>(); // In memory file system

public TFTPServer() throws IOException {
int portForServer = new ServerSocket(0).getLocalPort();
socket = new DatagramSocket(portForServer);
socket.setSoTimeout(0); // Listen forever
receiveDatagram = new DatagramPacket(new byte[TFTPPacket.MAX_PACKET_SIZE], TFTPPacket.MAX_PACKET_SIZE);
Thread serverThread = new Thread(this);
serverThread.setDaemon(true);
serverThread.start();
}

public static void main(String[] args) throws Exception {
TFTPServer ts = new TFTPServer();
System.out.println("Server running on port: " + ts.socket.getLocalPort() + ". Enter 'q' to stop it and all current transfers");
BufferedReader br = new BufferedReader(new InputStreamReader(System.in));
while (true) {
if (br.readLine().equals("q")) {
ts.stop();
System.exit(0);
}
}
}

@Override
public void run() {
try {
while (!stopServer) {
socket.receive(receiveDatagram);
int tranferPort = new ServerSocket(0).getLocalPort();
Transfer newTransfer = new Transfer(receiveDatagram, tranferPort);
transfers.add(newTransfer);
Thread thread = new Thread(newTransfer);
thread.setDaemon(true);
thread.start();
}
} catch (Exception e) {
if (!stopServer) LOGGER.log(Level.SEVERE, e.getMessage(), e);
} finally {
stopServer = true;
}
}

public void stop() {
System.out.println("Stopping server and aborting all transfers...");
stopServer = true;
Iterator<Transfer> it = transfers.iterator();
while (it.hasNext()) {
it.next().stop();
it.remove();
}
}

private class Transfer implements Runnable {
private static final int DEFAULT_TIMEOUT = 3000;
private static final int DEFAULT_MAX_TIMEOUTS = 5;
private final DatagramSocket transferSocket;
private TFTPPacket packet;
private boolean stopTransfer;
private DatagramPacket transferDatagram;

public Transfer(DatagramPacket datagramPacket, int port) throws SocketException {
transferDatagram = new DatagramPacket(datagramPacket.getData(), datagramPacket.getData().length,
datagramPacket.getAddress(), datagramPacket.getPort());
packet = TFTPPacket.parseTFTPPacket(transferDatagram);
transferSocket = new DatagramSocket(port);
transferSocket.setSoTimeout(DEFAULT_TIMEOUT);
}

public void stop() {
stopTransfer = true;
transferSocket.close();
}

@Override
public void run() {
try {
if (packet.getType() == 1) handleRead(packet);
else if (packet.getType() == 2) handleWrite(packet);
} catch (Exception e) {
if (!stopTransfer) LOGGER.log(Level.SEVERE, e.getMessage(), e);
} finally {
transferSocket.close();
transfers.remove(this);
}
}

private TFTPPacket receive() throws IOException {
transferSocket.receive(transferDatagram);
return TFTPPacket.parseTFTPPacket(transferDatagram);
}

private void send(TFTPPacket packet) throws IOException {
transferSocket.send(packet.buildDatagram());
}

private void handleRead(TFTPPacket packet) throws IOException {
InputStream is = null;
System.out.println("Sending file: " + packet.getFilename());
StringBuilder sb = new StringBuilder();
try {
try {
if (!content.containsKey(packet.getFilename())) throw new FileNotFoundException("File: " + packet.getFilename() + " not found");
is = new ByteArrayInputStream(content.get(packet.getFilename()).getBytes());
} catch (FileNotFoundException e) {
LOGGER.log(Level.SEVERE, e.getMessage(), e);
TFTPPacket error = new TFTPPacket(TFTPPacket.ERROR, packet.getAddress(), packet.getPort());
error.setErrorCode(TFTPPacket.ERROR_FILE_NOT_FOUND);
error.setErrorMessage(TFTPPacket.ERROR_MESSAGES.get(TFTPPacket.ERROR_FILE_NOT_FOUND));
send(error);
return;
}
TFTPPacket ack;
int block = 1;
boolean sendNext = true;
int readLength = TFTPPacket.MAX_PACKET_SIZE - TFTPPacket.HEADER_SIZE;
TFTPPacket data = null;
while (readLength == TFTPPacket.MAX_PACKET_SIZE - TFTPPacket.HEADER_SIZE && !stopTransfer) { // End is reached when data read is less than 512 bytes
if (sendNext) {
readLength = Math.min(readLength, is.available());
byte[] temp = new byte[readLength];
is.read(temp);
data = new TFTPPacket(TFTPPacket.DATA, packet.getAddress(), packet.getPort());
int numLength = BigInteger.valueOf(block).toByteArray().length;
if (numLength == 2) data.setBlockNumber(BigInteger.valueOf(block).toByteArray());
else data.setBlockNumber(new byte[] {0, (byte) block});
data.setLength(readLength);
data.setData(temp);
sb.append(new String(temp));
send(data);
}
ack = null;
int timeOuts = 0;
while (!stopTransfer && (ack == null || !ack.getAddress().equals(packet.getAddress()) || ack.getPort() != packet.getPort())) {
if (ack != null) {
System.out.println("Wrong tid, needed: " + packet.getPort() + " got: " + data.getPort());
TFTPPacket error = new TFTPPacket(TFTPPacket.ERROR, ack.getAddress(), ack.getPort()); // Unknown tid
error.setErrorCode(TFTPPacket.ERROR_UNKNOWN_TRANSFER_ID);
error.setErrorMessage(TFTPPacket.ERROR_MESSAGES.get(TFTPPacket.ERROR_UNKNOWN_TRANSFER_ID));
send(error);
}
try {
ack = receive();
} catch (SocketTimeoutException e) {
if (timeOuts++ >= DEFAULT_MAX_TIMEOUTS) {
System.out.println("Max timeouts reached");
throw e;
}
send(data); // no ack received, resend data
}
}
if (ack == null || !(ack.getType() == TFTPPacket.ACKNOWLEDGEMENT)) {
System.out.println("Aborting transfer since we got an invalid packet instead of ack");
break;
} else {
if (new BigInteger(ack.getBlockNumber()).intValue() != block)
sendNext = false; // takes care of the SAS syndrome
else {
if (++block > TFTPPacket.RESET_BLOCK_NUM) block = 0;
sendNext = true;
}
}
}
} finally {
try {
if (is != null) is.close();
transferSocket.close();
} catch (IOException e) {
LOGGER.log(Level.SEVERE, e.getMessage(), e);
}
}
}

private void handleWrite(TFTPPacket packet) throws IOException {
String key = packet.getFilename();
StringBuilder sb = new StringBuilder();
System.out.println("Storing file: " + key);
try {
int lastBlock = 0;
try {
if (content.containsKey(key)) {
System.out.println("File already exists, aborting");
TFTPPacket error = new TFTPPacket(TFTPPacket.ERROR, packet.getAddress(), packet.getPort());
error.setErrorCode(TFTPPacket.ERROR_FILE_ALREADY_EXISTS);
error.setErrorMessage(TFTPPacket.ERROR_MESSAGES.get(TFTPPacket.ERROR_FILE_ALREADY_EXISTS));
send(error);
return;
}
} catch (Exception e) {
LOGGER.log(Level.SEVERE, e.getMessage(), e);
TFTPPacket error = new TFTPPacket(TFTPPacket.ERROR, packet.getAddress(), packet.getPort());
error.setErrorCode(TFTPPacket.ERROR_NOT_DEFINED);
error.setErrorMessage(TFTPPacket.ERROR_MESSAGES.get(TFTPPacket.ERROR_NOT_DEFINED));
send(error);
return;
}
TFTPPacket ack = new TFTPPacket(TFTPPacket.ACKNOWLEDGEMENT, packet.getAddress(), packet.getPort());
send(ack);
while (true) {
TFTPPacket data = null;
int timeOuts = 0;
while (!stopTransfer && (data == null || !data.getAddress().equals(packet.getAddress()) || data.getPort() != packet.getPort())) {
if (data != null) {
System.out.println("Wrong tid, needed: " + packet.getPort() + " got: " + data.getPort());
TFTPPacket error = new TFTPPacket(TFTPPacket.ERROR, data.getAddress(), data.getPort()); // wrong tid
error.setErrorCode(TFTPPacket.ERROR_UNKNOWN_TRANSFER_ID);
error.setErrorMessage(TFTPPacket.ERROR_MESSAGES.get(TFTPPacket.ERROR_UNKNOWN_TRANSFER_ID));
send(error);
}
try {
data = receive();
System.out.println("data received from " + data.getAddress() + " port: " + data.getPort());
} catch (SocketTimeoutException e) {
if (timeOuts++ >= DEFAULT_MAX_TIMEOUTS) {
System.out.println("Max timeouts reached");
throw e;
}
System.out.println("Resending ack as no data is received");
send(ack); // no data received, resend ack
}
}
if (data != null && data.getType() == TFTPPacket.WRITE_REQUEST) {
ack = new TFTPPacket(TFTPPacket.ACKNOWLEDGEMENT, packet.getAddress(), packet.getPort()); // resend initial ack
System.out.println("Resending initial ack");
send(ack);
} else if (data == null || !(data.getType() == TFTPPacket.DATA)) {
System.out.println("Aborting transfer since either we got an invalid packet instead of ack or no data is received");
break;
} else {
int block = new BigInteger(data.getBlockNumber()).intValue();
byte[] bytes = data.getData();
int dataLength = data.getLength() - TFTPPacket.HEADER_SIZE;
System.out.println("writing data of length: " + dataLength);
if (block > lastBlock || (lastBlock == TFTPPacket.RESET_BLOCK_NUM && block == 0)) {
sb.append(new String(bytes).substring(0, dataLength));
lastBlock = block;
}
ack = new TFTPPacket(TFTPPacket.ACKNOWLEDGEMENT, packet.getAddress(), packet.getPort());
ack.setBlockNumber(data.getBlockNumber());
send(ack);
if (dataLength < TFTPPacket.MAX_PACKET_SIZE - TFTPPacket.HEADER_SIZE) {
content.put(key, sb.toString());
break;
}
}
}
} finally {
transferSocket.close();
}
}

}*/