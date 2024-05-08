import os
import subprocess
import socket

# Get the host IP address dynamically
# This will get the IP address assigned to the default network interface
host_ip = socket.gethostbyname(socket.gethostname())

# Set the environment variable
os.environ['BASE_URL'] = host_ip

# Start Docker Compose
subprocess.run(["docker-compose", "up"])
