package storage

import (
	"database/sql"
	"fmt"
	"log"
	pf "main/packages/passfield"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

var instance *sql.DB
var once sync.Once

/*
singleton

main should open & close, pass pointers
*/
func GetInstance() *sql.DB {
	once.Do(func() {
		path, err := os.Executable()
		if err != nil {
			log.Fatal(err)
		}
		path = filepath.Join(filepath.Dir(path), "storage.db")

		db, err := sql.Open("sqlite", path)
		if err != nil {
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
		instance = db
	})
	return instance
}

func Save(db *sql.DB, field pf.PassField) error {
	switch p := field.(type) {
	case *pf.PassFieldSite:
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
			p.UUID,
			p.Username, p.Email, p.Phone, p.Password, p.Notes,
			p.Timestamp,
			p.Website,
		)
		return err
	case *pf.PassFieldBasic:
		_, err := db.Exec(`
          INSERT INTO entries (id, username, email, phone, password, notes, timestamp, website)
          VALUES (?, ?, ?, ?, ?, ?, ?, ?)
          ON CONFLICT(id) DO UPDATE SET
              timestamp = excluded.timestamp
              username  = COALESCE(excluded.username,  username),
              email     = COALESCE(excluded.email,     email),
              phone     = COALESCE(excluded.phone,     phone),
              password  = COALESCE(excluded.password,  password),
              notes     = COALESCE(excluded.notes,     notes),
      `,
			p.UUID, p.Timestamp,
			p.Username, p.Email, p.Phone, p.Password, p.Notes,
		)
		return err
	}

	return fmt.Errorf("unknown PassField: %s", field.Identify())
}
