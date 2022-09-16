// Package migrations contains the system PocketBase DB migrations.
package migrations

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/pocketbase/pocketbase/tools/migrate"
)

var AppMigrations migrate.MigrationsList

// Register is a short alias for `AppMigrations.Register()`
// that is usually used in external/user defined migrations.
func Register(
	up func(db dbx.Builder) error,
	down func(db dbx.Builder) error,
	optFilename ...string,
) {
	var optFiles []string
	if len(optFilename) > 0 {
		optFiles = optFilename
	} else {
		_, path, _, _ := runtime.Caller(1)
		optFiles = append(optFiles, filepath.Base(path))
	}
	AppMigrations.Register(up, down, optFiles...)
}

func init() {
	AppMigrations.Register(func(db dbx.Builder) error {
		_, tablesErr := db.NewQuery(`
		CREATE TABLE {{_admins}} (
			[[id]]              VARCHAR(100) PRIMARY KEY,
			[[avatar]]          INTEGER DEFAULT 0 NOT NULL,
			[[email]]           VARCHAR(200) UNIQUE NOT NULL,
			[[tokenKey]]        VARCHAR(200) UNIQUE NOT NULL,
			[[passwordHash]]    TEXT NOT NULL,
			[[lastResetSentAt]] TEXT DEFAULT '' NOT NULL,
			[[created]]         VARCHAR(200) DEFAULT '' NOT NULL,
			[[updated]]         VARCHAR(200) DEFAULT '' NOT NULL
		);
	`).Execute()
		if tablesErr != nil {
			return tablesErr
		}

		_, tablesUsersErr := db.NewQuery(`
	CREATE TABLE {{_users}} (
		[[id]]                     VARCHAR(100) PRIMARY KEY,
		[[verified]]               BIT NOT NULL,
		[[email]]                  VARCHAR(200) NOT NULL,
		[[tokenKey]]               VARCHAR(200) NOT NULL,
		[[passwordHash]]           TEXT NOT NULL,
		[[lastResetSentAt]]        TEXT DEFAULT '' NOT NULL,
		[[lastVerificationSentAt]] TEXT DEFAULT '' NOT NULL,
		[[created]]                VARCHAR(100) DEFAULT '' NOT NULL,
		[[updated]]                VARCHAR(100) DEFAULT '' NOT NULL
	);
	CREATE UNIQUE INDEX _users_email_idx ON {{_users}} ([[email]]) WHERE [[email]] != '';
	CREATE UNIQUE INDEX _users_tokenKey_idx ON {{_users}} ([[tokenKey]]);
	`).Execute()

		if tablesUsersErr != nil {
			return tablesUsersErr
		}
		_, tablesColErr := db.NewQuery(`
	CREATE TABLE {{_collections}} (
		[[id]]         VARCHAR(100) PRIMARY KEY,
		[[system]]     BIT NOT NULL,
		[[name]]       VARCHAR(100) UNIQUE NOT NULL,
		[[schema]]     TEXT NOT NULL,
		[[listRule]]   TEXT DEFAULT NULL,
		[[viewRule]]   TEXT DEFAULT NULL,
		[[createRule]] TEXT DEFAULT NULL,
		[[updateRule]] TEXT DEFAULT NULL,
		[[deleteRule]] TEXT DEFAULT NULL,
		[[created]]    VARCHAR(100) DEFAULT '' NOT NULL,
		[[updated]]    VARCHAR(100) DEFAULT '' NOT NULL
	);

	`).Execute()
		if tablesColErr != nil {
			return tablesColErr
		}
		db.NewQuery(`
	CREATE TABLE {{_params}} (
		[[id]]      VARCHAR(100) PRIMARY KEY,
		[[key]]     VARCHAR(100) UNIQUE NOT NULL,
		[[value]]   TEXT DEFAULT NULL,
		[[created]] VARCHAR(100) DEFAULT '' NOT NULL,
		[[updated]] VARCHAR(100) DEFAULT '' NOT NULL
	);
	`).Execute()

		// inserts the system profiles collection
		// -----------------------------------------------------------
		profileOwnerRule := fmt.Sprintf("%s = @request.user.id", models.ProfileCollectionUserFieldName)
		collection := &models.Collection{
			Name:       models.ProfileCollectionName,
			System:     true,
			CreateRule: &profileOwnerRule,
			ListRule:   &profileOwnerRule,
			ViewRule:   &profileOwnerRule,
			UpdateRule: &profileOwnerRule,
			Schema: schema.NewSchema(
				&schema.SchemaField{
					Id:       "pbfielduser",
					Name:     models.ProfileCollectionUserFieldName,
					Type:     schema.FieldTypeUser,
					Unique:   true,
					Required: true,
					System:   true,
					Options: &schema.UserOptions{
						MaxSelect:     1,
						CascadeDelete: true,
					},
				},
				&schema.SchemaField{
					Id:      "pbfieldname",
					Name:    "name",
					Type:    schema.FieldTypeText,
					Options: &schema.TextOptions{},
				},
				&schema.SchemaField{
					Id:   "pbfieldavatar",
					Name: "avatar",
					Type: schema.FieldTypeFile,
					Options: &schema.FileOptions{
						MaxSelect: 1,
						MaxSize:   5242880,
						MimeTypes: []string{
							"image/jpg",
							"image/jpeg",
							"image/png",
							"image/svg+xml",
							"image/gif",
						},
					},
				},
			),
		}
		collection.Id = "systemprofiles0"
		collection.MarkAsNew()

		return daos.New(db).SaveCollection(collection)
	}, func(db dbx.Builder) error {
		tables := []string{
			"_params",
			"_collections",
			"_users",
			"_admins",
			models.ProfileCollectionName,
		}

		for _, name := range tables {
			if _, err := db.DropTable(name).Execute(); err != nil {
				return err
			}
		}

		return nil
	})
}
