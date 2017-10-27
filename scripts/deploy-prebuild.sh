#!/bin/bash

BuildOutPath="$PWD"
username="duobuilduser"
emailaddress="duobuilduser@duosoftware.com"
password="DuoS12345"

cd /usr/local

rm -r "pkg"

if [ ! -d "bin" ]; then 
	mkdir "bin" 
fi
if [ ! -d "pkg" ]; then 
	mkdir "pkg" 
fi
if [ ! -d "src" ]; then 
	mkdir "src" 
fi
export PATH=$PATH:/usr/local/go/bin;
export GOPATH=$PWD;
export PATH=$PATH:$GOPATH/bin;
cd bin
rm *
cd ../
cd src

git config --global user.name $username
git config --global user.email $emailaddress
echo ""
echo "BEGIN REPO PULL v6engine-deps"
if [ ! -d "depo" ]; then
	mkdir "depo"
	cd depo
	git clone https://github.com/DuoSoftware/v6engine-deps
	cd ../
fi


cd depo/v6engine-deps
git pull
cp * -r $GOPATH/src
cd ../
cd ../
echo "END REPO PULL v6engine-deps"
echo ""
echo "BEGIN REPO PULL v6engine"
if [ ! -d "duov6.com" ]; then
	git clone -b development https://github.com/DuoSoftware/v6engine
	mv v6engine duov6.com
fi
cd duov6.com
git pull
git log --pretty=format:"%h%x09%an%x09%ad%x09%s" > /var/www/html/build/history.txt
#mv v6engine duov6.com
echo "END REPO PULL v6engine"
