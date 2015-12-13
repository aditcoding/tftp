import threading
import sys
import os
import time
import logging

if len(sys.argv) < 5:
    print 'Usage: python tester.py <host> <port> <dir where files are located> <comma separated file names>'
    exit()

host = sys.argv[1]
port = sys.argv[2]
dir = sys.argv[3]
files = sys.argv[4]

file_arr = files.split(',')

def run(file):
    os.system('tftp ' + host + ' ' + port + ' -m octet -v -c put ' + dir + '/' + file)
    os.system('tftp ' + host + ' ' + port + ' -m octet -v -c get ' + dir + '/' + file + ' ' + dir + '/' + file + '.recv')

threads = []

try:
    for i in range(len(file_arr)):
        thread = threading.Thread(target=run, args=(file_arr[i],))
        threads.append(thread)
        thread.start()

    for thread in threads:
        thread.join()
except Exception as e:
    logging.exception(e)

print "Done."