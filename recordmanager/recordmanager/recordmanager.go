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
	fmt.Println("		[a] : BACKUP data store.")
	fmt.Println("		[b] : RESTORE new instance.")
	fmt.Println("		[c] : Export to Couchbase")
	fmt.Println("		[d] : Export to MySQL")
	fmt.Println()

	var ipAddress string
	var username string
	var password string
	var bucket string
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

	if response == "c" {
		fmt.Println("Export to COUCHBASE option selected. Make sure all Backup file are in the same folder.")
		fmt.Println("Specify URL : ")
		_, err := fmt.Scanln(&ipAddress)
		fmt.Println("Specify URL : ")
		_, err2 := fmt.Scanln(&bucket)
		if err != nil || err2 != nil {
			fmt.Println(err.Error())
		} else {
			status := processes.ExportToCouchServer(ipAddress, bucket)
			if status {
				fmt.Println("Records Successfully Restored! :D ")
			} else {
				fmt.Println("An error occured! :( ")
			}
		}
	}

	if response == "d" {
		fmt.Println("Export to MySQL option selected. Make sure all Backup file are in the same folder.")
		fmt.Println("Specify URL : ")
		_, err := fmt.Scanln(&ipAddress)
		fmt.Println("Specify Username : ")
		_, err = fmt.Scanln(&username)
		fmt.Println("Specify Password : ")
		_, err = fmt.Scanln(&password)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			status := processes.ExportToMySQLServer(ipAddress, username, password)
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
