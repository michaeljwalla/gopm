package passfield

import (
	"database/sql"
)

//for ref:
// type NullString struct {
//     String string // The actual text value
//     Valid  bool   // True if the value is NOT NULL, false if it is NULL
// }

// iface
type PassField interface {
	Identify() string
	// Populate() //unused rn
}

// passfieldbasic
// considered invalid / will delete if no password
type PassFieldBasic struct {
	UUID      string
	Timestamp int64
	Username  sql.NullString
	Email     sql.NullString
	Phone     sql.NullString
	Password  sql.NullString
	Notes     sql.NullString
	//
	Website sql.NullString
}

func (p PassFieldBasic) Identify() string { return "Basic" }

func (p PassFieldBasic) String() string {
	return p.Identify()
}
