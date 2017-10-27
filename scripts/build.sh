#!/bin/bash
BuildOutPath="$PWD"
export PATH=$PATH:/usr/local/go/bin;
export GOPATH=$PWD;
export PATH=$PATH:$GOPATH/bin;

echo "$PWD"
echo "$GOPATH"
echo "BEGIN duoauth build"
cd duoauth
go build > /var/www/html/buildtest/duoauth.txt
go install
cd ../
echo "END duoauth build"
echo "BEGIN objectstore build"
cd objectstore
go build > /var/www/html/buildtest/objectstore.txt
go install
