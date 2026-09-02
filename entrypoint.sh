#!/bin/sh

: ${LOG_LEVEL:='ERROR'}

echo
date
echo "Quit the server with CONTROL-C."
echo

if [ ! -f /opt/gitverse_notifier/server.key ]; then
  openssl req -x509 -nodes -days 3650 -newkey rsa:2048 -keyout /opt/gitverse_notifier/server.key -out /opt/gitverse_notifier/server.crt -subj "/C=CN/ST=Moscow/L=Moscow/O=FIT/OU=FIT/CN=FIT"
fi

chmod 700 /opt/gitverse_notifier

exec "$@"
