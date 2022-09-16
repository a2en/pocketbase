package core

import (

     "fmt"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/pocketbase/dbx"
)

func connectDB(dbPath string) (*dbx.DB, error) {
	var server = "192.168.29.225"
	var port = 1433
	var user = "sa"
	var password = "gitman"
	var database = "pocketbase"
	db, openErr := dbx.MustOpen("mssql",fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;",
	server, user, password, port, database))
	if openErr != nil {
		return nil, openErr
	}

	// additional pragmas not supported through the dsn string
	// _, err := db.NewQuery(`go get github.com/minus5/gofreetds

	// 	pragma journal_size_limit = 100000000;
	// `).Execute()

	return db, openErr
}