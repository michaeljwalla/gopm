package termutils

import (
	"bufio"
	"database/sql"
	"fmt"
	"main/packages/passfield"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/term"
)

func ptr[T any](in T) *T {
	return &in
}

var reader = bufio.NewReader(os.Stdin)

func flushReader() {
	reader.Discard(reader.Buffered())
}
func termResetLine() {
	fmt.Print("\r\033[K")
}

// output message with "Enter to continue" prompt, padded by newlines
func termNotify(message *string) {
	out := "\n"
	if message != nil {
		out += *message
	}
	out += "\nPress enter to continue."
	fmt.Print(out)
	flushReader()
	reader.ReadString('\n')
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

		flushReader()
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
		//
		if !allowOuterWhitespace && len(strings.TrimSpace(response)) != len(response) {
			termNotify(ptr("\nOuter whitespace is not allowed. Please try again."))
			continue //retry
		}
		if len(response) < minLength {
			termNotify(ptr(fmt.Sprintf("\nPassword is too short, min. length %d. Please try again.", minLength)))
			continue //retry
		}
		//
		termResetLine()
		if !requireRetype {
			fmt.Println("Password: (set)")
			return response
		}
		fmt.Print("Confirm by retyping: ")
		bytesRetyped, _ := term.ReadPassword(int(os.Stdin.Fd()))
		responseRetyped := string(bytesRetyped)

		if responseRetyped != response {
			termNotify(ptr("\nPassword mismatch. Please try again."))
			continue //retry
		}
		termResetLine()
		fmt.Println("Password: (set)")
		return response
	}
}

func PopulatePassFieldBasic(p *passfield.PassFieldBasic) {
	no_str := ptr("")
	p.UUID = uuid.NewString()
	p.Username = stringToSQLNullString(readField("Username? ", nil, true, no_str))
	p.Email = stringToSQLNullString(readField("Email? ", nil, true, no_str))
	p.Phone = stringToSQLNullString(readField("Phone #? ", nil, true, no_str))
	p.Password = stringToSQLNullString(readPassword("Password", 6, true, false)) //readPassword
	p.Notes = stringToSQLNullString(readField("Notes? ", nil, true, no_str))
	p.Timestamp = time.Now().Unix()
	// return
}

func PopulatePassFieldSite(p *passfield.PassFieldSite) {
	p.Website = stringToSQLNullString(readField("Website", ptr("You must provide a website.\n"), false, nil))
	PopulatePassFieldBasic(&p.PassFieldBasic)
	// return
}
