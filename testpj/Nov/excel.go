package main

import (
	"fmt"
	"github.com/tealeg/xlsx"
	"strconv"
)

func main() {
	rowcount := 0
	colunmcount := 0
	var exceldata []map[string]interface{}
	var colunName []string
	blockSizeRecords := 50000
	blockindex := 0

	xlFile, err := xlsx.OpenFile("OrderItems.xlsx")
	fmt.Println("File Opened")
	if err == nil {
		for _, sheet := range xlFile.Sheets {
			rowcount = (sheet.MaxRow - 1)
			colunmcount = sheet.MaxCol
			colunName = make([]string, colunmcount)
			for _, row := range sheet.Rows {
				for j, cel := range row.Cells {
					colunName[j] = cel.String()
				}
				break
			}
			wholeRowIndex := 1
			if err == nil {
				for _, sheet := range xlFile.Sheets {
					rowIndex := 1
					for rownumber, row := range sheet.Rows {
						currentRow := make(map[string]interface{})
						if rownumber != 0 {
							for cellnumber, cell := range row.Cells {
								currentRow[colunName[cellnumber]] = cell.String()
							}
							exceldata = append(exceldata, currentRow)
							if rowIndex == blockSizeRecords || wholeRowIndex == rowcount {
								createExcelFile(("OrderItems" + strconv.Itoa(blockindex) + ".xlsx"), colunName, exceldata)
								rowIndex = 1
								exceldata = nil
								blockindex++
							} else {
								rowIndex++
							}
							wholeRowIndex++
						}

					}
				}
			}
		}

	}
}

func createExcelFile(fileName string, columns []string, data []map[string]interface{}) {
	var file *xlsx.File
	var sheet *xlsx.Sheet
	var row *xlsx.Row
	var cell *xlsx.Cell

	file = xlsx.NewFile()
	sheet, err := file.AddSheet("Sheet1")
	if err != nil {
		fmt.Printf(err.Error())
	}

	row = sheet.AddRow()
	for x := 0; x < len(columns); x++ {
		cell = row.AddCell()
		cell.Value = columns[x]
	}

	for x := 0; x < len(data); x++ {
		row = sheet.AddRow()
		for y := 0; y < len(columns); y++ {
			cell = row.AddCell()
			cell.Value = data[x][columns[y]].(string)
		}
	}

	err = file.Save(fileName)
	if err != nil {
		fmt.Printf(err.Error())
	}
}
