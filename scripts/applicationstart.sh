#!/bin/bash
cd /home/DUO_V6_AUTH
docker build -t "duo-auth" .
cd /home/DUO_V6_OBSTORE
docker build -t "duo-obstore" .
cd /home/DUO_V6_NOTIFIER
docker build -t "duo-notifier" .

docker run -d -t -p 3000:3000 -p 3001:3001 --log-opt max-size=10m --log-opt max-file=10 --restart=always --name duo-obstore duo-obstore
docker run -d -t -p 3048:3048 --log-opt max-size=10m --log-opt max-file=10 --restart=always --name duo-auth duo-auth
docker run -d -t -p 7000:7000 --log-opt max-size=10m --log-opt max-file=10 --restart=always --name duo-notifier duo-notifier
