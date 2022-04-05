package main

import (
	"flag"

	excel2sql "github.com/kanguki/doExcel/examples/excel2Sql"
)

func main() {
	flag.Parse()
	excel2sql.Do()
}
