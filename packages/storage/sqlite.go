package storage

import (
	"database/sql"
	"log"
	"main/packages/encrypt"
	pf "main/packages/passfield"
	"os"
	"path/filepath"
	"strconv"

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
			timestamp 	BLOB NOT NULL,
			username 	BLOB NOT NULL,
			email 		BLOB NOT NULL,
			phone 		BLOB NOT NULL,
			password 	BLOB NOT NULL,
			notes 		BLOB NOT NULL,

			website 	BLOB NOT NULL
		);
		CREATE TABLE IF NOT EXISTS vault (
			id		INTEGER PRIMARY KEY,
			salt	BLOB NOT NULL,
			wdek	BLOB NOT NULL
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
func throwPassWrapString(key []byte, data sql.NullString) []byte {
	field, err := encrypt.SealField(key, data)
	if err != nil {
		log.Fatal(err)
	}
	return field
}
func Save(key []byte, field *pf.PassFieldBasic) error {
	assertInitAndOwned()
	timestamp := throwPassWrapString(key, sql.NullString{
		Valid:  true,
		String: strconv.FormatInt(field.Timestamp, 10),
	})
	_, err := db.Exec(`
		INSERT INTO entries (id, timestamp, username, email, phone, password, notes, website)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			timestamp = excluded.timestamp,
			username  = excluded.username,
			email     = excluded.email,
			phone     = excluded.phone,
			password  = excluded.password,
			notes     = excluded.notes,
			website   = excluded.website
	`,
		field.UUID, timestamp,
		throwPassWrapString(key, field.Username),
		throwPassWrapString(key, field.Email),
		throwPassWrapString(key, field.Phone),
		throwPassWrapString(key, field.Password),
		throwPassWrapString(key, field.Notes),

		throwPassWrapString(key, field.Website),
	)

	return err
}

func Close() error {
	if db == nil {
		return nil
	}
	err := db.Close()
	db = nil
	owned = false
	return err
}

type VaultBlobs struct {
	Wdek []byte
	Salt []byte
}

// dek still wrapped
func GetVault() VaultBlobs {
	assertInitAndOwned()
	data := db.QueryRow("SELECT wdek, salt FROM vault LIMIT 1")
	var wdek, salt []byte
	if err := data.Scan(&wdek, &salt); err != nil {
		log.Fatal(err)
	}

	return VaultBlobs{Wdek: wdek, Salt: salt}
}

// wrap dek beforehand
func SetVault(in VaultBlobs) {
	assertInitAndOwned()
	_, err := db.Exec(`INSERT INTO vault (id, wdek, salt)
		VALUES (1, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			wdek = excluded.wdek,
			salt = excluded.salt
		`,
		in.Wdek,
		in.Salt,
	)
	if err != nil {
		log.Fatal(err)
	}
}

func throwPassUnwrapString(key []byte, data []byte) sql.NullString {
	field, err := encrypt.OpenField(key, data)
	if err != nil {
		log.Fatal(err)
	}
	return field
}
func throwPassInt64Eval(num string) int64 {
	field, err := strconv.ParseInt(num, 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	return field
}
func createPassFieldbasic(key []byte, q *sql.Rows) pf.PassFieldBasic {
	var uuid string
	var timestamp, username, email, phone, password, notes, website []byte

	if err := q.Scan(&uuid, &timestamp, &username, &email, &phone, &password, &notes, &website); err != nil {
		log.Fatal(err)
	}
	return pf.PassFieldBasic{
		UUID:      uuid,
		Timestamp: throwPassInt64Eval(throwPassUnwrapString(key, timestamp).String),
		Username:  throwPassUnwrapString(key, username),
		Email:     throwPassUnwrapString(key, email),
		Phone:     throwPassUnwrapString(key, phone),
		Password:  throwPassUnwrapString(key, password),
		Notes:     throwPassUnwrapString(key, notes),
		Website:   throwPassUnwrapString(key, website),
	}
}
func GetEntries(key []byte) ([]pf.PassFieldBasic, error) {
	assertInitAndOwned()
	var passfields []pf.PassFieldBasic
	data, err := db.Query(`SELECT * FROM entries`)
	if err != nil {
		return passfields, err
	}

	for data.Next() {
		passfields = append(passfields, createPassFieldbasic(key, data))
	}

	return passfields, data.Err()
}
