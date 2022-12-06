package main

import (
	"fmt"
	"log"
	"os"
	"net/http"
	"regexp"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/sfomuseum/go-csvdict"
)

func main() {
	_, ok := os.LookupEnv("DBSERVER")
	if !ok {
		log.Fatal("DBSERVER environment variable is not set")
		return
	}
	_, ok = os.LookupEnv("DBPORT")
	if !ok {
		log.Fatal("DBPORT environment variable is not set")
		return
	}
	_, ok = os.LookupEnv("DBUSER")
	if !ok {
		log.Fatal("DBUSER environment variable is not set")
		return
	}
	_, ok = os.LookupEnv("DBPASS")
	if !ok {
		log.Fatal("DBPASS environment variable is not set")
		return
	}
	_, ok = os.LookupEnv("DBNAME")
	if !ok {
		log.Fatal("DBNAME environment variable is not set")
		return
	}

	app := pocketbase.New()

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		// serves static files from the provided public dir (if exists)
		subFs := echo.MustSubFS(e.Router.Filesystem, "pb_public")
		e.Router.GET("/*", apis.StaticDirectoryHandler(subFs, false))

		return nil
	})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.AddRoute(echo.Route{
			Method: http.MethodPost,
			Path:   "/api/uploadData",
			Handler: func(c echo.Context) error {
				// get the file and params
				file, fileErr := c.FormFile("file")
				tableName := c.FormValueDefault("tableName", "data_csv")
				append := c.FormValueDefault("append", "false")

				// check for errors in file upload
				if fileErr == nil {

					// open the file
					fileContent, _ := file.Open()
					reader, err := csvdict.NewReader(fileContent)

					// check for errors in opening the file
					if err != nil {
						log.Println("1 " + err.Error())
						return c.String(500, err.Error())
					}

					// check if the table exists
					hasTable := app.Dao().HasTable(tableName)

					// if the table exists and append is false, delete the table
					if hasTable && append == "false" {
						col, _ := app.Dao().FindCollectionByNameOrId(tableName)
						delErr := app.Dao().DeleteCollection(col)
						if delErr != nil {
							log.Println(".1 " + delErr.Error())
							return c.String(500, delErr.Error())
						}
					}

					// Loop through the rows in csv file
					count := -1
					for {

						// read the row
						count++
						row, err := reader.Read()
						if err != nil {
							break
						}

						// if the table doesn't exist, create it
						hasTable = app.Dao().HasTable(tableName)
						if !hasTable {
							collection := &models.Collection{}
							collection.Name = tableName
							collection.Schema = schema.NewSchema()

							// add access rules. empty string means full access
							rule := ""
							collection.ViewRule = &rule
							collection.ListRule = &rule

							// check csv integrity
							if _, ok := row["Fixed asset number"]; !ok {
								return c.String(500, "Missing \"Fixed asset number\" column")
							}
							if _, ok := row["Asset Name"]; !ok {
								return c.String(500, "Missing \"Asset Name\" column")
							}
							if _, ok := row["Department"]; !ok {
								return c.String(500, "Missing \"Department\" column")
							}
							if _, ok := row["Name"]; !ok {
								return c.String(500, "Missing \"Department\" column")
							}
							if _, ok := row["Serial number"]; !ok {
								return c.String(500, "Missing \"Serial number\" column")
							}
							if _, ok := row["Assigned Emp Code"]; !ok {
								return c.String(500, "Missing \"Assigned Emp Code\" column")
							}

							for k, _ := range row {
								key := sanitizeString(k)
								collection.Schema.AddField(&schema.SchemaField{
									Name: strings.ToLower(key),
									Type: schema.FieldTypeText,
								})
							}
							collection.Schema.AddField(&schema.SchemaField{
								Name: "verified",
								Type: schema.FieldTypeBool,
							})
							if err := app.Dao().SaveCollection(collection); err != nil {
								log.Println("2 " + err.Error())
								return c.String(500, err.Error())
							}

						}

						// create the document
						collection, _ := app.Dao().FindCollectionByNameOrId(tableName)
						r1 := models.NewRecord(collection)

						// add the fields
						for k, v := range row {
							key := sanitizeString(k)
							r1.SetDataValue(key, v)
						}
						r1.SetDataValue("verified", false)

						// save the document
						err1 := app.Dao().SaveRecord(r1)

						// check for errors in saving the document
						if err1 != nil {
							log.Println("3 " + err1.Error())
							return c.String(500, err1.Error())
						}

					}

					return c.String(200, fmt.Sprintf("%d records added to the collection "+tableName, count))
				}
				log.Println("4 " + fileErr.Error())
				return c.String(500, fileErr.Error())
			},
			Middlewares: []echo.MiddlewareFunc{
				// apis.RequireAdminOrUserAuth(),
			},
		})

		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

// Sanitize string
func sanitizeString(s string) string {
	reSp, err := regexp.Compile(`[&.]`)
	if err != nil {
		log.Fatal(err)
	}
	s = reSp.ReplaceAllString(s, "")
	re, err := regexp.Compile(`[^\w]`)
	if err != nil {
		log.Fatal(err)
	}
	s = re.ReplaceAllString(s, "_")
	s = strings.Replace(s, ".", "", 5)
	s = strings.ToLower(strings.Replace(s, " ", "_", 5))
	return s
}
