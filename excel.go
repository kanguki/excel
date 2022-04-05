package doExcel

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

//return [[row1][row2][...][rowN]]
func ReadSheet(filePath string, sheetName string) ([][]string, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	rows, err := f.GetRows(sheetName)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return rows, nil

}
