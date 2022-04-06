package excel2sql

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/kanguki/doExcel"
)

var (
	table           = getEnv(TABLE_NAME, "`tradex-user-utility`.`t_favorite_list`")
	excelFilePath   = getEnv(EXCEL_FILE_PATH, "doc/test2_out.xlsx")
	excelFileSheet  = getEnv(EXCEL_FILE_SHEET, "Sheet1")
	outPath         = getEnv(SQL_OUT_PATH, "doc/test_out.sql")
	insertBatch     = 50
	maxJobPerWorker = 100
	numWorkers      = 5
)

//read excel file
//convert data to sql
//write to sql file
func Do() {
	printUsage()
	data := getInputChan()
	outs := []<-chan string{}
	for i := 0; i < numWorkers; i++ {
		outs = append(outs, convert(data))
	}
	result := mergeOutput(outs...)
	countSuccess := make(chan int, 1)
	countSuccess <- 0
	//process out data
	if err := os.Truncate(outPath, 0); err != nil {
		doExcel.Debug("Failed to truncate: %v", err)
	}

	f, err := os.OpenFile(outPath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		doExcel.Log("error tructcate text file %v", err)
	}
	defer f.Close()

	batch := []string{}
	for info := range result {
		if info != "" {
			// insert multiple records each time
			batch = append(batch, info)
			if len(batch) == insertBatch || len(result) == 0 {
				w := "INSERT INTO " + table + " (`name`, `username`, `order`, `count`, `max_count`, `symbol_list`, `created_at`, `updated_at`) VALUES " + strings.Join(batch, ",") + ";\n"
				if _, err := f.WriteString(w); err != nil {
					doExcel.Debug("error write data to sql file: %v", err)
				}
				batch = []string{}
			}
			countSuccess <- <-countSuccess + 1
		}
	}

	doExcel.Log("Handle %v records", <-countSuccess)
}
func getInputChan() <-chan []string {
	data, err := doExcel.ReadSheet(excelFilePath, excelFileSheet)
	if err != nil {
		doExcel.Log("error reading xlsx file: %v", err)
		os.Exit(1)
	}
	input := make(chan []string, maxJobPerWorker)

	go func() {
		for _, d := range data {
			input <- d
		}
		close(input)
	}()

	return input
}

func mergeOutput(outs ...<-chan string) <-chan string {
	var wg sync.WaitGroup
	merged := make(chan string, maxJobPerWorker)
	wg.Add(len(outs))

	for _, o := range outs {
		go func(c <-chan string) {
			for data := range c {
				merged <- data
			}
			wg.Done()
		}(o)
	}

	go func() {
		wg.Wait()
		close(merged)
	}()

	return merged
}

//data that has inconstent format are skipped
//wg is to close chan in, != limitNGoroutines
func convert(in <-chan []string) <-chan string {
	out := make(chan string, maxJobPerWorker)
	go func() {
		for data := range in {
			var info, userId, name string
			var list, listStr []string
			var order int
			var err error
			l := len(data)
			if l < 4 { //4 when list is empty
				fmt.Printf("incorrect data format :( %v\n", data)
				goto end
			}
			order, err = strconv.Atoi(data[1])
			if err != nil {
				doExcel.Debug("error convert data (order) to info: %v\n", err)
				goto end
			}

			if l == 5 {
				list = strings.Split(data[4], ",")
				for _, v := range list {
					if s, isStock := isAStock(v); isStock {
						listStr = append(listStr, fmt.Sprintf("\\\"{\\\\\\\"data\\\\\\\":\\\\\\\"%v\\\\\\\",\\\\\\\"isNote\\\\\\\":false}\\\"", s))
					}
				}
			}
			//no need to convert to struct first, but it's more readable
			userId = sanitize(data[0], " ", "'", "\"", "\n")
			name = sanitize(data[2], "'", "\"", "\n")
			info = fmt.Sprintf("('%v', '%v', %v, %v, 50, '[%v]', now(), now())", name, userId, order, len(listStr), strings.Join(listStr, ","))
		end:
			out <- info
		}
		close(out)
	}()
	return out
}

//remove multiple characters in string
func sanitize(s string, cutSets ...string) string {
	for _, v := range cutSets {
		s = strings.ReplaceAll(s, v, "")
	}
	return strings.TrimSpace(s)
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
