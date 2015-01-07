package processmanager

import (
	"duov6.com/objectstore/client"
	"duov6.com/serviceconsole/messaging"
	//"fmt"
	"github.com/tealeg/xlsx"
	"log"
	"strings"
)

type ExcelWorker struct {
}

func (worker ExcelWorker) GetWorkerName() string {
	return "ExcelWorker"
}

func (worker ExcelWorker) ExecuteWorker(request *messaging.ServiceRequest) messaging.ServiceResponse {
	var temp = messaging.ServiceResponse{}
	rowcount := 0
	colunmcount := 0
	var exceldata []map[string]interface{}
	var colunName []string

	if request.Body != nil {
		log.Printf("Received a message: %s", request.Body)
		excelFileName := string(request.Body[:])
		//file read
		xlFile, error := xlsx.OpenFile(excelFileName)

		if error == nil {
			for _, sheet := range xlFile.Sheets {
				rowcount = sheet.MaxRow
				colunmcount = sheet.MaxCol
				colunName = make([]string, colunmcount)
				for _, row := range sheet.Rows {
					for j, cel := range row.Cells {
						colunName[j] = cel.String()
					}
					break
				}

				exceldata = make(([]map[string]interface{}), rowcount)

				if error == nil {
					for _, sheet := range xlFile.Sheets {
						for rownumber, row := range sheet.Rows {
							currentRow := make(map[string]interface{})
							exceldata[rownumber] = currentRow
							for cellnumber, cell := range row.Cells {
								exceldata[rownumber][colunName[cellnumber]] = cell.String()
							}
						}
					}
				}

				client.Go("token", "com.duosoftware.com", getExcelFileName(excelFileName)+"."+sheet.Name).StoreObject().WithKeyField("Id").AndStoreMapInterface(exceldata).Ok()

			}
		}

		temp.IsSuccess = true
	} else {
		temp.IsSuccess = false
	}

	return temp
}

func getExcelFileName(path string) (fileName string) {
	subsets := strings.Split(path, "\\")
	subfilenames := strings.Split(subsets[len(subsets)-1], ".")
	fileName = subfilenames[0]
	return
}
