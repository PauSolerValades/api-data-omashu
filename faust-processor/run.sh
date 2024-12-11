#!/bin/sh

# Exit immediately if a command exits with a non-zero status
set -e

# Start the Faust worker
faust -A app.app worker --web-port=${WORKER_PORT:-6066} -l info
