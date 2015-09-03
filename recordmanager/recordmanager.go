package main

import (
	//"bufio"
	"duov6.com/recordmanager/processes"
	"fmt"
	//"os"
)

func main() {
	paint()
	fmt.Println("Starting Elastic Record Manager!")
	fmt.Println()
	fmt.Println("Please Enter choice and press ENTER")
	fmt.Println("		[a] : Create Data BACKUP of a data store.")
	fmt.Println("		[b] : RESTORE Backup to a new instance.")
	fmt.Println()

	var ipAddress string
	var response string
	print("Enter your Selection : ")
	_, err := fmt.Scanln(&response)

	if err != nil {
		fmt.Println(err.Error())
	}

	if response == "a" {
		fmt.Println("Create Backup option selected.")
		fmt.Println("Specify IP Address : ")
		_, err := fmt.Scanln(&ipAddress)
		fmt.Println("Please wait! Data being transferred! :)")
		if err != nil {
			fmt.Println(err.Error())
		} else {
			status := processes.BackupServer(ipAddress)
			if status {
				fmt.Println("Records Successfully Recieved! :D ")
			} else {
				fmt.Println("An error occured! :( ")
			}
		}
	}

	if response == "b" {
		fmt.Println("Restore Backup option selected. Make sure all Backup file are in the same folder.")
		fmt.Println("Specify IP Address : ")
		_, err := fmt.Scanln(&ipAddress)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			status := processes.RestoreServer(ipAddress)
			if status {
				fmt.Println("Records Successfully Restored! :D ")
			} else {
				fmt.Println("An error occured! :( ")
			}
		}
	}

}

func paint() {
	fmt.Println()
	fmt.Println("______            ______            _                ")
	fmt.Println("|  _  \\           | ___ \\          | |               ")
	fmt.Println("| | | |_   _  ___ | |_/ / __ _  ___| | ___   _ _ __  ")
	fmt.Println("| | | | | | |/ _ \\| ___ \\/ _` |/ __| |/ / | | | '_ \\ ")
	fmt.Println("| |/ /| |_| | (_) | |_/ / (_| | (__|   <| |_| | |_) |")
	fmt.Println("|___/  \\__,_|\\___/\\____/ \\__,_|\\___|_|\\_ \\__,_| .__/ ")
	fmt.Println("                                              | |    ")
	fmt.Println("                                              |_| ")
	fmt.Println()
	fmt.Println()

}
