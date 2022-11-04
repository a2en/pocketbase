package core

import (

     "fmt"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/pocketbase/dbx"
)

func connectDB(dbPath string) (*dbx.DB, error) {
	var server = "127.0.0.1"
	var port = 1433
	var user = "sa"
	var password = "gitman"
	var database = "pocketbase"
	db, openErr := dbx.MustOpen("mssql",fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;",
	server, user, password, port, database))
	if openErr != nil {
		return nil, openErr
	}
	return db, openErr
}