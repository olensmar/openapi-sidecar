#!/bin/bash

if ! [ "$SOURCE_PORT" -gt 0 ]; then
  echo "Missing valid SOURCE_PORT"
  exit
fi

if ! [ "$DEST_PORT" -gt 0 ]; then
  echo "Missing valid DEST_PORT"
  exit
fi

echo "Mapping $SOURCE_PORT to $DEST_PORT"

iptables -t nat -A PREROUTING -p tcp -i eth0 --dport "$SOURCE_PORT" -j REDIRECT --to-port "$DEST_PORT"

# List all iptables rules.
iptables -t nat --list