import os
import subprocess
import socket
import argparse

parser = argparse.ArgumentParser(description='starts docker compose and sets host IP as environment variable')
parser.add_argument('--rebuild', dest='rebuild', action=argparse.BooleanOptionalAction,
                        help='specifies if docker compose should be rebuilt')
args = parser.parse_args()

# This will get the IP address assigned to the default network interface
host_ip = socket.gethostbyname(socket.gethostname())

os.environ['BASE_URL'] = host_ip
if args.rebuild:
    subprocess.run(["docker-compose", "up", "--build"])
else: 
    subprocess.run(["docker-compose", "up"])