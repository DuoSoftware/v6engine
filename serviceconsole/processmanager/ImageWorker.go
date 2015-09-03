package processmanager

import (
	"duov6.com/serviceconsole/messaging"
	"encoding/json"
	"fmt"
	"github.com/nfnt/resize"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
)

type ImageWorker struct {
}

func (worker ImageWorker) GetWorkerName() string {
	return "ImageWorker"
}

func (worker ImageWorker) ExecuteWorker(request *messaging.ServiceRequest) messaging.ServiceResponse {

	req := messaging.ServiceRequest{}

	fmt.Println("Starting Image Worker!")
	var temp = messaging.ServiceResponse{}
	if request.Body != nil {
		//log.Printf("Received a message: %s", request.Body)

		err := json.Unmarshal(request.Body, &req)

		if err != nil {
			fmt.Println(err.Error())
		}

		fileServerPath := ""
		//implement image logic here

		if runtime.GOOS == "linux" {
			fileServerPath = request.Configuration.ServerConfiguration["WindowsFileServer"]["SavePath"]
		} else {
			fileServerPath = request.Configuration.ServerConfiguration["LinuxFileServer"]["SavePath"]
		}

		if req.Parameters["OutFileName"] != "" {

			backupPath := fileServerPath + "backup/"
			currentname := req.Parameters["InFileName"]
			backupName := getFileName(currentname) + "-backup." + getFileExtension(req.Parameters["InFileName"])

			//read file from the current place
			file, _ := ioutil.ReadFile(fileServerPath + req.Parameters["InFileName"])
			//write the bacup file to the backup folder
			err := ioutil.WriteFile((backupPath + backupName), file, 0666)

			if err != nil {
				//fmt.Print("File Backup failed! Error : " + err.Error())
				request.Log("File Backup failed! Error : " + err.Error())
			}

		}

		if getFileExtension(req.Parameters["InFileName"]) != getFileExtension(req.Parameters["OutFileName"]) {
			filePath := fileServerPath + req.Parameters["InFileName"]
			imageFile := ImageRead(filePath, getFileExtension(req.Parameters["InFileName"]))

			if getFileExtension(req.Parameters["OutFileName"]) == "png" {
				_ = ConvertToPNG(imageFile, (fileServerPath + req.Parameters["OutFileName"]))
			} else if getFileExtension(req.Parameters["OutFileName"]) == "jpeg" {
				_ = ConvertToJPEG(imageFile, (fileServerPath + req.Parameters["OutFileName"]))
			} else {
				_ = ConvertToGIF(imageFile, (fileServerPath + req.Parameters["OutFileName"]))
			}
		}

		if req.Parameters["Height"] != "" && req.Parameters["Width"] != "" {

			filePath := fileServerPath + req.Parameters["InFileName"]
			imageData := ImageRead(filePath, getFileExtension(req.Parameters["InFileName"]))

			width1, _ := strconv.Atoi(req.Parameters["Width"])
			height1, _ := strconv.Atoi(req.Parameters["Height"])

			width := uint(width1)
			height := uint(height1)

			imageFile := resize.Resize(width, height, imageData, resize.NearestNeighbor)

			if getFileExtension(req.Parameters["OutFileName"]) == "png" {
				ConvertToPNG(imageFile, (fileServerPath + req.Parameters["OutFileName"]))
			} else if getFileExtension(req.Parameters["OutFileName"]) == "jpeg" {
				ConvertToJPEG(imageFile, (fileServerPath + req.Parameters["OutFileName"]))
			} else {
				ConvertToGIF(imageFile, (fileServerPath + req.Parameters["OutFileName"]))
			}

		}

		fmt.Println("End")

		temp.IsSuccess = true
		temp.Message = "Image Convertion successful!"
		request.Log("Image Convertion successful!")

	} else {
		temp.IsSuccess = false
		temp.Message = "Image Convertion unsuccessful!"
		request.Log("Image Convertion unsuccessful!")
	}
	return temp
}

func getFileExtension(fileName string) (fileType string) {

	var tempArray []string
	tempArray = strings.Split(fileName, ".")
	if len(tempArray) > 1 {
		fileType = tempArray[len(tempArray)-1]
	} else {
		fileType = "NAF"
		//request.Log("Requested file is not a file. Run DEBUG : Logical Error!")
	}
	return
}

func getFileName(path string) (fileName string) {
	subsets := strings.Split(path, "\\")
	subfilenames := strings.Split(subsets[len(subsets)-1], ".")
	fileName = subfilenames[0]
	return
}

func ImageRead(ImageFile string, format string) (image image.Image) {

	if format == "jpeg" {
		file, err := os.Open(ImageFile)
		if err != nil {
			log.Fatal(err)
		}
		image, err = jpeg.Decode(file)
		if err != nil {
			log.Fatal(err)
		}
		file.Close()
		//request.Log(ImageFile + " converted to Byte array successfully!")

	} else if format == "gif" {
		file, err := os.Open(ImageFile)
		if err != nil {
			log.Fatal(err)
		}
		image, err = gif.Decode(file)
		if err != nil {
			log.Fatal(err)
		}
		file.Close()
		//	request.Log(ImageFile + " converted to Byte array successfully!")
	} else if format == "png" {
		file, err := os.Open(ImageFile)
		if err != nil {
			log.Fatal(err)
		}
		image, err = png.Decode(file)
		if err != nil {
			log.Fatal(err)
		}
		file.Close()
		//request.Log(ImageFile + " converted to Byte array successfully!")
	}
	return
}

func ConvertToPNG(img image.Image, path string) (isOK bool) {
	isOK = true
	out, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
		isOK = false
	}
	err = png.Encode(out, img)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
		isOK = false
	}
	out.Close()
	return
}

func ConvertToGIF(img image.Image, path string) (isOK bool) {
	isOK = true
	out, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
		isOK = false
	}
	var opt gif.Options
	opt.NumColors = 256
	err = gif.Encode(out, img, &opt)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
		isOK = false
	}
	out.Close()
	return

}
func ConvertToJPEG(img image.Image, path string) (isOK bool) {
	isOK = true
	out, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
		isOK = false
	}

	err = jpeg.Encode(out, img, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)

		isOK = false
	}
	out.Close()
	return
}
