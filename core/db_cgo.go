package core

import (

     "fmt"
	 "os"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/pocketbase/dbx"
)

func connectDB(dbPath string) (*dbx.DB, error) {
	var server = os.Getenv("DBSERVER")//192.168.29.225:1433
	var port = os.Getenv("DBPORT")//1433
	var user = os.Getenv("DBUSER")//"sa"
	var password = os.Getenv("DBPASS")//"gitman"
	var database = os.Getenv("DBNAME")//"pocketbase"
	db, openErr := dbx.MustOpen("mssql",fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s;",
	server, user, password, port, database))
	if openErr != nil {
		return nil, openErr
	}
	return db, openErr
}