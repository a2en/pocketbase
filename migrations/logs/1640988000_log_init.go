package logs

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/tools/migrate"
)

var LogsMigrations migrate.MigrationsList

func init() {
	LogsMigrations.Register(func(db dbx.Builder) (err error) {
		_, err = db.NewQuery(`
			CREATE TABLE {{_requests}} (
				[[id]]        VARCHAR(100) PRIMARY KEY,
				[[url]]       TEXT DEFAULT '' NOT NULL,
				[[methodName]] TEXT DEFAULT 'get' NOT NULL,
				[[status]]    INTEGER DEFAULT 200 NOT NULL,
				[[auth]]      VARCHAR(100) DEFAULT 'guest' NOT NULL,
				[[remoteIp]]        VARCHAR(100) DEFAULT '127.0.0.1' NOT NULL,
				[[userIp]]        VARCHAR(100) DEFAULT '127.0.0.1' NOT NULL,
				[[referer]]   TEXT DEFAULT '' NOT NULL,
				[[userAgent]] TEXT DEFAULT '' NOT NULL,
				[[meta]]      TEXT DEFAULT '{}' NOT NULL,
				[[created]]   VARCHAR(100) DEFAULT '' NOT NULL,
				[[updated]]   VARCHAR(100) DEFAULT '' NOT NULL
			);

			CREATE INDEX _request_status_idx on {{_requests}} ([[status]]);
			CREATE INDEX _request_auth_idx on {{_requests}} ([[auth]]);
			--CREATE INDEX _request_created_hour_idx on {{_requests}} (FORMAT([[created]],''yyyy-MM-dd HH-mm:ss'));
			CREATE INDEX _request_remote_ip_idx on {{_requests}} ([[remoteIp]]);
			CREATE INDEX _request_user_ip_idx on {{_requests}} ([[userIp]]);

		`).Execute()

		return err
	}, func(db dbx.Builder) error {
		_, err := db.DropTable("_requests").Execute()
		return err
	})
}
