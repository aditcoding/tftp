<h3>Java implementation usage:</h3>

To start the server: <code>java -jar tftp.jar</code> <br>
To stop serving: Hit 'q' followed by enter on cmd

<b>To run tests with multiple client connections:</b> <br>
(an existing tftp-hpa installion is required)
<code>python tester.py<i> host  port  dir  filenames</i></code> where <br>
<i>host</i> is the server address,<br>
<i>port</i> is the server port,<br>
<i>dir</i> is the location of the files to be uploaded or where received files are to be placed and <br>
the last arg is the actual filenames comma separated.
Received files are placed with .recv extension

The python script creates multiple threads and just invokes the linux tftp client (tftp-hpa version 5.2)

<b>My test env:</b>
- Two ubuntu 14.04 LTS machines
- java 1.8
- python 2.7
- tftp-hpa 5.2, without readline (for client)


<b>Features:</b>
- Multi-threaded server. Can serve clients concurrently. Files being written to the server are visible only after the write is complete
- Stores files (text only in this release) in memory as a key value pair <filename, text>
- Only octet mode is supported
- Everything else as per RFC 1350 - http://tools.ietf.org/html/rfc1350

<b>Future functionality:</b>
- Ability to pick a specific port - currently randomly chosen
- Serves only text files for now - all formats in the future
- Better logging and exception handling

<h3>Go implementation usage:</h3>

<code> go get install 