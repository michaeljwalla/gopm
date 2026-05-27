package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

func ptr[T any](in T) *T {
	return &in
}

// output message with "Enter to continue" prompt, padded by newlines
func termNotify(message *string) {
	out := "\n"
	if message != nil {
		out += *message
	}
	out += "\nPress enter to continue."
	fmt.Print(out)
	fmt.Scan()
	fmt.Println()
	//return
}

// nullstring iff len(s) == 0
func stringToSQLNullString(s string) sql.NullString {
	if len(s) > 0 {
		return sql.NullString{
			String: s,
			Valid:  true,
		}
	}
	return sql.NullString{
		Valid: false,
	}
}

// allowDefault included to not confuse between allowing default ""
// and tried-to-default-input-when-not-allowed
func readField(prompt string, retryMsg *string, allowDefault bool, def *string) string {
	reader := bufio.NewReader(os.Stdin)

	//default value
	var defaultValue string = ""
	if def != nil {
		defaultValue = strings.TrimSpace(*def)
	} else if allowDefault {
		panic("readField: 'allowDefault' enabled when 'def' is nil")
	}
	//retry msg
	var retry string = ""
	if retryMsg != nil {
		retry = *retryMsg
	}
	for {
		if defaultValue != "" {
			fmt.Printf("%s [default \"%s\"]: ", prompt, defaultValue)
		} else {
			fmt.Printf("%s: ", prompt)
		}

		response, _ := reader.ReadString('\n')

		response = strings.TrimSpace(response)
		if response != "" {
			return response
		} else if allowDefault {
			return defaultValue
		}
		//
		if retry != "" {
			fmt.Println(retry)
		}
	}
}

// get password using term library
func readPassword(prompt string, minLength int, requireRetype bool, allowOuterWhitespace bool) string {
	for {
		fmt.Printf("%s: ", prompt)
		bytes, _ := term.ReadPassword(int(os.Stdin.Fd()))
		response := string(bytes)
		fmt.Println()
		//
		if !allowOuterWhitespace && len(strings.TrimSpace(response)) != len(response) {
			termNotify(ptr("Outer whitespace is not allowed. Please try again."))
			continue //retry
		}
		if len(response) < minLength {
			termNotify(ptr(fmt.Sprintf("Password is too short, min. length %d. Please try again.", minLength)))
			continue //retry
		}
		if !requireRetype {
			return response
		}
		//
		fmt.Print("Confirm by retyping: ")
		bytesRetyped, _ := term.ReadPassword(int(os.Stdin.Fd()))
		responseRetyped := string(bytesRetyped)
		fmt.Println()

		if responseRetyped != response {
			termNotify(ptr("Password mismatch. Please try again."))
			continue //retry
		}
		return response
	}
}

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
func (p *PassFieldBasic) Populate() {
	no_str := ptr("")
	p.Username = stringToSQLNullString(readField("Username? ", nil, true, no_str))
	p.Email = stringToSQLNullString(readField("Email? ", nil, true, no_str))
	p.Phone = stringToSQLNullString(readField("Phone #? ", nil, true, no_str))
	p.Password = stringToSQLNullString(readPassword("Password", 6, true, false)) //readPassword
	p.Notes = stringToSQLNullString(readField("Notes? ", nil, true, no_str))
	p.Timestamp = uint(time.Now().Unix())
	// return
}

type PassFieldSite struct {
	PassFieldBasic
	Website sql.NullString
}

func (p PassFieldSite) Identify() string { return "Site" }
func (p *PassFieldSite) Populate() {
	p.Website = stringToSQLNullString(readField("Website", ptr("You must provide a website."), false, nil))
	p.PassFieldBasic.Populate()
	// return
}

func main() {
	fmt.Println("Core")

	my_entry := PassFieldSite{}
	my_entry.Populate()

	fmt.Println("\nFinal entry:\n", my_entry)

	// return
}
