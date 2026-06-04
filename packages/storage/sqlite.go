package storage

import (
	"database/sql"
	"log"
	pf "main/packages/passfield"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var db *sql.DB
var owned bool = false

/*
singleton

main should open & close, pass pointers
*/
func doInit() {
	var err error
	path, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	path = filepath.Join(filepath.Dir(path), "storage.db")

	db, err = sql.Open("sqlite", path)
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
}

/*
immediately assumes TryInit() == true means the caller will own & closew

if s.TryInit() { defer s.Close() }
*/
func TryInit() bool {
	if owned {
		return false
	}
	owned = true
	doInit()
	return true
}
func assertInitAndOwned() {
	if db == nil {
		panic("DB not initialized")
	}
	if !owned {
		panic("DB has no explicit owner")
	}
}
func Save(field *pf.PassFieldBasic) error {
	assertInitAndOwned()
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
	err := db.Close()
	db = nil
	owned = false
	return err
}

func createPassFieldbasic(q *sql.Rows) pf.PassFieldBasic {
	var uuid string
	var timestamp int64
	var username, email, phone, password, notes, website sql.NullString

	q.Scan(&uuid, &timestamp, &username, &email, &phone, &password, &notes, &website)
	return pf.PassFieldBasic{
		UUID:      uuid,
		Timestamp: timestamp,
		Username:  username,
		Email:     email,
		Phone:     phone,
		Password:  password,
		Notes:     notes,
		Website:   website,
	}
}
func GetEntries() ([]pf.PassFieldBasic, error) {
	assertInitAndOwned()
	var passfields []pf.PassFieldBasic
	data, err := db.Query(`SELECT * FROM entries`)
	if err != nil {
		return passfields, err
	}

	for data.Next() {
		passfields = append(passfields, createPassFieldbasic(data))
	}

	return passfields, data.Err()
}
