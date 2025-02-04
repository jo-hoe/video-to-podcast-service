import os
import subprocess
import socket
import argparse


def get_local_ip():
    # Create a socket that connects to an external address
    # This helps us determine which network interface is used for external connections
    s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    try:
        # We don't need to actually send data
        # The connection attempt itself is enough to determine the local IP
        s.connect(("8.8.8.8", 80))
        local_ip = s.getsockname()[0]
        return local_ip
    except Exception as e:
        print(f"Error getting local IP: {e}")
        return None
    finally:
        s.close()


parser = argparse.ArgumentParser(
    description='starts docker compose and sets host IP as environment variable')
parser.add_argument('--rebuild', dest='rebuild', action=argparse.BooleanOptionalAction,
                    help='specifies if docker compose should be rebuilt')
parser.add_argument('--services', dest='services', type=str, nargs='*',
                    help='specify one or more services to run (space-separated)')
args = parser.parse_args()

# Get the local network IP
host_ip = get_local_ip()
if host_ip:
    print(f"Using local IP: {host_ip}")
    os.environ['BASE_URL'] = host_ip
else:
    print("Failed to get local IP address")
    exit(1)

# Base command
cmd = ["docker-compose", "up"]

# Add --build flag if rebuild is specified
if args.rebuild:
    cmd.append("--build")

# Add specific services if provided
if args.services:
    cmd.extend(args.services)

# Run the command
try:
    subprocess.run(cmd, check=True)
except subprocess.CalledProcessError as e:
    print(f"Error running docker-compose: {e}")
    exit(1)
