#!/bin/sh

: ${LOG_LEVEL:='ERROR'}

echo
date
echo "Quit the server with CONTROL-C."
echo

chmod 700 /opt/gitverse_notifier

exec "$@"
