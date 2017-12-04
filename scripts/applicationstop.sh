#!/bin/bash

isExistAuth=$(docker ps -a | grep duo-auth)
isExistObstore=$(docker ps -a | grep duo-obstore)
isExistNotifier=$(docker ps -a | grep duo-notifier)

if [ $isExistAuth ];
then
    docker rm -f duo-auth
    docker rmi duo-auth
elif [ isExistObstore ];
then
  docker rm -f duo-obstore
  docker rmi duo-obstore

elif [ isExistNotifier ];
then
  docker rm -f duo-notifier
  docker rmi duo-notifier
fi

