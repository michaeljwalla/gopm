package storage

import (
	"database/sql"
	"log"
	pf "main/packages/passfield"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

var db *sql.DB
var once sync.Once

/*
singleton

main should open & close, pass pointers
*/
func StartInstance() {
	once.Do(func() {
		var err error
		path, err := os.Executable()
		if err != nil {
			log.Fatal(err)
		}
		path = filepath.Join(filepath.Dir(path), "storage.db")

		db, err = sql.Open("sqlite", path)
		if err != nil {
			once = sync.Once{}
			log.Fatal(err)
		}

		if _, err := db.Exec(`
		PRAGMA foreign_keys = ON;
		PRAGMA user_version = 1;
		CREATE TABLE IF NOT EXISTS entries (
			id 			TEXT PRIMARY KEY,
			timestamp 	INTEGER NOT NULL,
			username 	TEXT,
			email 		TEXT,
			phone 		TEXT,
			password 	TEXT,
			notes 		TEXT,

			website 	TEXT
		);
		`); err != nil {
			log.Fatal(err)
		}

		db.SetMaxOpenConns(1)
	})
}

func Save(field *pf.PassFieldBasic) error {
	_, err := db.Exec(`
		INSERT INTO entries (id, username, email, phone, password, notes, timestamp, website)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			timestamp = excluded.timestamp,
			username  = COALESCE(excluded.username,  username),
			email     = COALESCE(excluded.email,     email),
			phone     = COALESCE(excluded.phone,     phone),
			password  = COALESCE(excluded.password,  password),
			notes     = COALESCE(excluded.notes,     notes),
			website   = COALESCE(excluded.website,   website)
	`,
		field.UUID,
		field.Username, field.Email, field.Phone, field.Password, field.Notes,
		field.Timestamp,
		field.Website,
	)

	return err
}

func Close() error {
	return db.Close()
}
func GetEntries() ([]pf.PassField, error) {
	var passfields []pf.PassField
	data, err := db.Query(`SELECT * FROM entries`)
	if err != nil {
		return passfields, err
	}

	for data.Next() {

	}
	return passfields, nil
}
