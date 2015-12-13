Java implementation usage:

To start the server: java -jar tftp-java.jar
To stop serving: Hit 'q' followed by enter on cmd

To run tests with multiple client connections:
python tester.py <host> <port> <dir where files are located> <comma separated file names> where
host is the server address,
port is the server port,
dir is the location of the files to be uploaded or where received files are to be placed and
the last arg is the actual filenames comma separated.

Received files are placed with .recv extension

The python script creates multiple threads and just invokes the linux tftp client (tftp-hpa version 5.2)

My test env:
Two ubuntu 14.04 LTS machines
java 1.8
python 2.7
tftp-hpa 5.2, without readline (for client)


Features:
- Multi-threaded server. Can serve clients concurrently. Files being written to the server are visible only after the write is complete
- Stores files (text only in this release) in memory as a key value pair <filename, text>
- Only octet mode is supported
- Everything else as per RFC 1350 - http://tools.ietf.org/html/rfc1350

Future functionality:
1) Ability to pick a specific port - currently randomly chosen
2) Serves only text files for now - all formats in the future
3) Better logging and exception handling