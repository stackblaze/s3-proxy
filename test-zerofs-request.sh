#!/bin/bash
# Capture what ZeroFS is actually sending to the proxy

echo "=== Starting tcpdump to capture ZeroFS requests ==="
echo "Run ZeroFS in another terminal, then press Ctrl+C here"
echo ""

sudo tcpdump -i lo -A -s 0 'tcp port 8080 and (((ip[2:2] - ((ip[0]&0xf)<<2)) - ((tcp[12]&0xf0)>>2)) != 0)' 2>&1 | grep -A 20 "PUT\|GET\|POST\|Authorization"
