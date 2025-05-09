package migrations

import (
	"context"
	"database/sql"
	"time"

	"gitlab.com/clseibold/auragem_sis/db"
	"gitlab.com/clseibold/auragem_sis/migration/types"
)

func init() {
	registerMigration(LifeKeptTables{})
}

type LifeKeptTables struct{}

func (m LifeKeptTables) Version() types.MigrationVersion {
	return types.MigrationVersion(time.Date(2021, 5, 26, 23, 53, 54, 0, time.UTC))
}

func (m LifeKeptTables) Name() string {
	return "LifeKeptTables"
}

func (m LifeKeptTables) DB() db.DBType {
	return db.LifeKeptDB
}

func (m LifeKeptTables) Description() string {
	return "Make rest of LifeKept tables"
}

func (m LifeKeptTables) Up(tx *sql.Tx) error {
	_, err := tx.ExecContext(context.Background(), `
	CREATE TABLE collections (
		id integer generated by default as identity primary key,
		memberid integer references members,
		name character varying(255) NOT NULL,
		date_start timestamp with time zone NOT NULL,
		date_end timestamp with time zone NOT NULL,
		starred boolean NOT NULL,
		date_created timestamp with time zone NOT NULL
	);
	`)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(context.Background(), `
	CREATE TABLE bullets (
		id integer generated by default as identity primary key,
		collectionid integer references members,
		parent integer references bullets,
		content character varying(1024) NOT NULL,
		priority integer NOT NULL,
		date_start timestamp with time zone,
		date_end timestamp with time zone,
		recurring character varying(255) NOT NULL,
		date_created timestamp with time zone NOT NULL
	);
	`)
	if err != nil {
		return err
	}

	return nil
}

func (m LifeKeptTables) Down(tx *sql.Tx) error {
	panic("Dangerous")
	_, err := tx.ExecContext(context.Background(), `DROP TABLE bullets;`)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(context.Background(), `DROP TABLE collections;`)
	if err != nil {
		return err
	}

	return nil
}
