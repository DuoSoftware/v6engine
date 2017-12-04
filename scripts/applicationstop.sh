#!/bin/bash

isExistAuth=$(docker ps -a | grep duo-auth)
isExistObstore=$(docker ps -a | grep duo-obstore)
isExistNotifier=$(docker ps -a | grep duo-notifier)

if [ "$isExistAuth" ]
then
    docker rm -f duo-auth
    docker rmi duo-auth
fi

if [ "$isExistObstore" ]
then
  docker rm -f duo-obstore
  docker rmi duo-obstore
fi

if [ "$isExistNotifier" ]
then
  docker rm -f duo-notifier
  docker rmi duo-notifier
fi
