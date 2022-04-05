package excel2sql

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/kanguki/doExcel"
)

var (
	table          = getEnv(TABLE_NAME, "`tradex-user-utility`.`t_favorite_list_test`")
	excelFilePath  = getEnv(EXCEL_FILE_PATH, "doc/test2_out.xlsx")
	excelFileSheet = getEnv(EXCEL_FILE_SHEET, "Sheet1")
	outPath        = getEnv(SQL_OUT_PATH, "doc/test_out.sql")
	insertBatch    = 50
)

//read excel file
//convert data to sql
//write to sql file
func Do() {
	printUsage()
	data, err := doExcel.ReadSheet(excelFilePath, excelFileSheet)
	if err != nil {
		doExcel.Log("error reading xlsx file: %v", err)
		os.Exit(1)
	}
	limitNGoroutines := make(chan interface{}, 100000)
	result := make(chan string)
	for _, v := range data {
		limitNGoroutines <- struct{}{}
		go convert(limitNGoroutines, result, v)
	}

	//open file once
	if err := os.Truncate(outPath, 0); err != nil {
		doExcel.Debug("Failed to truncate: %v", err)
	}

	f, err := os.OpenFile(outPath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		doExcel.Log("error tructcate text file %v", err)
	}
	defer f.Close()

	//process out data
	batch, countSuccess := []string{}, 0
	for i := 0; i < len(data); i++ {
		info := <-result
		if info != "" {
			// insert multiple records each time
			batch = append(batch, info)
			if len(batch) == insertBatch || i == (len(data)-1) {
				w := "INSERT INTO " + table + " (`name`, `username`, `order`, `count`, `max_count`, `symbol_list`, `created_at`, `updated_at`) VALUES " + strings.Join(batch, ",") + ";\n"
				if _, err := f.WriteString(w); err != nil {
					doExcel.Debug("error write data to sql file: %v", err)
				}
				batch = []string{}
			}
			countSuccess++
		}
	}
	doExcel.Log("Handle %v records", countSuccess)
}

//data that has inconstent format are skipped
//wg is to close chan in, != limitNGoroutines
func convert(limitNGoroutines chan interface{}, in chan string, data []string) {
	var info string
	defer func() {
		in <- info
		<-limitNGoroutines
	}()
	l := len(data)
	if l < 4 { //4 when list is empty
		fmt.Printf("incorrect data format :( %v\n", data)
		return
	}
	order, err := strconv.Atoi(data[1])
	if err != nil {
		doExcel.Debug("error convert data (order) to info: %v\n", err)
		return
	}
	var list, listStr []string
	if l == 5 {
		list = strings.Split(data[4], ",")
		for _, v := range list {
			if s, isStock := isAStock(v); isStock {
				listStr = append(listStr, fmt.Sprintf("\\\"{\\\\\\\"data\\\\\\\":\\\\\\\"%v\\\\\\\",\\\\\\\"isNote\\\\\\\":false}\\\"", s))
			}
		}
	}
	//no need to convert to struct first, but it's more readable
	userId := sanitize(data[0], " ", "'", "\"", "\n")
	name := sanitize(data[2], " ", "'", "\"", "\n")
	info = fmt.Sprintf("('%v', '%v', %v, %v, 50, '[%v]', now(), now())", name, userId, order, len(listStr), strings.Join(listStr, ","))
}

//remove multiple characters in string
func sanitize(s string, cutSets ...string) string {
	for _, v := range cutSets {
		s = strings.ReplaceAll(s, v, "")
	}
	return s
}

func isAStock(s string) (cleaned string, isStock bool) {
	s = strings.ToUpper(sanitize(s, " ", "'", "\"", "\n"))
	match, err := regexp.MatchString("^[A-Z]+[A-Z0-9]*", s)
	return s, err == nil && match && len(s) < 13
}

func getEnv(envName, defaultVal string) string {
	v := os.Getenv(envName)
	if v == "" {
		v = defaultVal
	}
	return v
}
func printUsage() {
	usage := `
	parsing excel file to sql
	- flags:
		-d: enable debug mode
	- envs:
		TABLE_NAME: db_name.table_name
		EXCEL_FILE_PATH: absolute path of excel file
		EXCEL_FILE_SHEET:
		SQL_OUT_PATH:

	`
	fmt.Print(usage)
}
