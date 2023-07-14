#!/usr/bin/env bash
set -eo pipefail

# Create mount directory for service.
mkdir -p $MNT_EVENTS_LOG_DIR/uploaded 2>&1
#mkdir -p /mnt/gyroscope/events/uploaded

echo "Mounting Cloud Filestore."
mount -o nolock $FILESTORE_IP_ADDRESS:/$FILE_SHARE_NAME $MNT_EVENTS_LOG_DIR 2>&1
#mount -o nolock 10.66.52.210:/gyroscope_shared /mnt/gyroscope/events
echo "Mounting completed."


# Run the web service on container startup. Here we use the gunicorn
# webserver, with one worker process and 8 threads.
# For environments with multiple CPU cores, increase the number of workers
# to be equal to the cores available.
# Timeout is set to 0 to disable the timeouts of the workers to allow Cloud Run to handle instance scaling.
#exec gunicorn --bind :$PORT --workers 1 --threads 8 --timeout 0 main:app

# Exit immediately when one of the background processes terminate.
#wait -n 2>&1

/app/bin/logging-jobs