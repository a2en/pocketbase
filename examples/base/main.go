package main

import (
	"fmt"
	"log"
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
				file, fileErr := c.FormFile("file")
				tableName := c.FormValueDefault("tableName","data_csv")
				append := c.FormValueDefault("append", "false")
				if fileErr == nil {
					fileContent, _ := file.Open()
					reader, err := csvdict.NewReader(fileContent)
					if err != nil {
						log.Println("1 " + err.Error())
						return c.String(500, err.Error())
					}
					hasTable := app.Dao().HasTable(tableName)
					println(append+" value")
					if hasTable && append == "false" {
						col, _ := app.Dao().FindCollectionByNameOrId(tableName)
						delErr := app.Dao().DeleteCollection(col)
						if delErr != nil {
							log.Println(".1 " + delErr.Error())
							return c.String(500, delErr.Error())
						}
					}
					count := -1
					for {
						count++
						row, err := reader.Read()

						if err != nil {
							break
						}
						hasTable = app.Dao().HasTable(tableName)
						if !hasTable {
							collection := &models.Collection{}
							collection.Name = tableName
							collection.Schema = schema.NewSchema()

							for k, _ := range row {
								key := sanitizeString(k)
								collection.Schema.AddField(&schema.SchemaField{
									Name: strings.ToLower(key),
									Type: schema.FieldTypeText,
								})
							}

							if err := app.Dao().SaveCollection(collection); err != nil {
								log.Println("2 " + err.Error())
								return c.String(500, err.Error())
							}

						}
						collection, _ := app.Dao().FindCollectionByNameOrId(tableName)
						r1 := models.NewRecord(collection)
						for k, v := range row {
							key := sanitizeString(k)
							r1.SetDataValue(key, v)
						}
						err1 := app.Dao().SaveRecord(r1)
						if err1 != nil {
							log.Println("3 " + err1.Error())
							return c.String(500, err1.Error())
						}

					}

					return c.String(200,fmt.Sprintf("%d records added to the collection "+tableName, count))
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
	re, err := regexp.Compile(`[^\w]`)
	if err != nil {
		log.Fatal(err)
	}
	s = re.ReplaceAllString(s, "")
	s = strings.Replace(s, ".", "", 5)
	s = strings.ToLower(strings.Replace(s, " ", "_", 5))
	return s;
}
