#!/bin/bash

function checkport {
        if nc -zv -w30 $1 $2 <<< '' &> /dev/null
        then
                echo "[+] Port $1/$2 is open"
        else
                echo "[-] Port $1/$2 is closed"
                if [ $2 == 3000 ]; then
                        docker restart duo-obstore
                elif [ $2 == 3048 ]; then
                        docker restart duo-auth
                elif [ $2 == 7000 ]; then
                        docker restart duo-notifier
                fi
        fi
}

checkport '127.0.0.1' 3000
checkport '127.0.0.1' 3048
checkport '127.0.0.1' 7000
