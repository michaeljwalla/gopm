package passfield

import (
	"database/sql"
	"fmt"
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
	Username  sql.NullString
	Email     sql.NullString
	Phone     sql.NullString
	Password  sql.NullString
	Notes     sql.NullString
	Timestamp int64
}

func (p PassFieldBasic) Identify() string { return "Basic" }

// passfieldsite
type PassFieldSite struct {
	PassFieldBasic
	Website sql.NullString
}

func (p PassFieldSite) Identify() string { return "Site" }

// strings
func formatNullStr(n sql.NullString) string {
	if n.Valid {
		return n.String
	}
	return "< N/A >"
}

func (p PassFieldBasic) String() string {
	return fmt.Sprintf(
		"%-12s %s\n%-12s %s\n%-12s %s\n%-12s %s\n%-12s %s\n%-12s %d\n%-12s %s",
		"Username:", formatNullStr(p.Username),
		"Email:", formatNullStr(p.Email),
		"Phone:", formatNullStr(p.Phone),
		"Password:", "< OMITTED >",
		"Notes:", formatNullStr(p.Notes),
		"Timestamp:", p.Timestamp,
		"UUID:", p.UUID,
	)
}

func (p PassFieldSite) String() string {
	return fmt.Sprintf("%-12s %s\n%s", "Website:", formatNullStr(p.Website), p.PassFieldBasic.String())
}
