package migrations

import "github.com/pocketbase/dbx"

func init() {
	AppMigrations.Register(func(db dbx.Builder) error {
		_, createErr := db.NewQuery(`
			CREATE TABLE {{_externalAuths}} (
				[[id]]         VARCHAR(100) PRIMARY KEY,
				[[userId]]     VARCHAR(100) NOT NULL,
				[[provider]]   VARCHAR(100) NOT NULL,
				[[providerId]] VARCHAR(100) NOT NULL,
				[[created]]    VARCHAR(100) DEFAULT '' NOT NULL,
				[[updated]]    VARCHAR(100) DEFAULT '' NOT NULL,
				---
				FOREIGN KEY ([[userId]]) REFERENCES {{_users}} ([[id]]) ON UPDATE CASCADE ON DELETE CASCADE
			);

			CREATE UNIQUE INDEX _externalAuths_userId_provider_idx on {{_externalAuths}} ([[userId]], [[provider]]);
			CREATE UNIQUE INDEX _externalAuths_provider_providerId_idx on {{_externalAuths}} ([[provider]], [[providerId]]);
		`).Execute()
		if createErr != nil {
			return createErr
		}

		return nil
	}, func(db dbx.Builder) error {
		if _, err := db.DropTable("_externalAuths").Execute(); err != nil {
			return err
		}

		// drop the partial email unique index and replace it with normal unique index
		_, indexErr := db.NewQuery(`
			DROP INDEX IF EXISTS _users_email_idx;
			CREATE UNIQUE INDEX _users_email_idx on {{_users}} ([[email]]);
		`).Execute()
		if indexErr != nil {
			return indexErr
		}

		return nil
	})
}
