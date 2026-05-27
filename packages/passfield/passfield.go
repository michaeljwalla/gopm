package passfield

import (
	"database/sql"
)

//for ref:
// type NullString struct {
//     String string // The actual text value
//     Valid  bool   // True if the value is NOT NULL, false if it is NULL
// }

type PassField interface {
	Identify() string
	// Populate() //unused rn
}

// considered invalid / will delete if no password
type PassFieldBasic struct {
	Username  sql.NullString
	Email     sql.NullString
	Phone     sql.NullString
	Password  sql.NullString
	Notes     sql.NullString
	Timestamp uint
}

func (p PassFieldBasic) Identify() string { return "Basic" }

type PassFieldSite struct {
	PassFieldBasic
	Website sql.NullString
}

func (p PassFieldSite) Identify() string { return "Site" }
