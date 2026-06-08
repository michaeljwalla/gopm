package termutils

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"main/packages/encrypt"
	"main/packages/passfield"
	"main/packages/storage"
	"os"
	"path/filepath"
	"strconv"
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
	//
	p.Website = stringToSQLNullString(readField("Website?", nil, true, no_str))

	p.Username = stringToSQLNullString(readField("Username? ", nil, true, no_str))
	p.Email = stringToSQLNullString(readField("Email? ", nil, true, no_str))
	p.Phone = stringToSQLNullString(readField("Phone #? ", nil, true, no_str))
	p.Password = stringToSQLNullString(readPassword("Password", 6, true, false)) //readPassword
	p.Notes = stringToSQLNullString(readField("Notes? ", nil, true, no_str))
	p.Timestamp = time.Now().Unix()

	// return
}

func storeOldVault(path string) (string, error) {
	otherspath := filepath.Join(path, "others")
	if err := os.MkdirAll(otherspath, 0700); err != nil {
		return "", err
	}
	ts := strconv.FormatInt(time.Now().Unix(), 10)

	file := filepath.Join(path, "storage.db")
	newfile := filepath.Join(otherspath, ts+".db")
	if err := os.Rename(file, newfile); err != nil {
		if errors.Is(err, os.ErrNotExist) { // nothing to save
			return "", nil
		}
		return "", err
	}

	return newfile, nil
}

func RequestDEK(vault storage.VaultBlobs) []byte {
	fmt.Print("Hint: ")
	if vault.Hint.Valid {
		fmt.Println("\"" + vault.Hint.String + "\"")
	} else {
		fmt.Println("n/a")
	}
	//
	password := readPassword("Enter master password", 6, false, false)
	kek := encrypt.DeriveKEK(password, vault.Salt)

	dek, err := encrypt.Open(kek, vault.Wdek)
	if err != nil {
		log.Fatal(err)
	}
	return dek
}

// need reownership TryInit() after
func UpdateVault() []byte {
	no_str := ptr("")
	fmt.Print(`
 - Your password will not be saved, only used to generate a key.
 - You should remember it.
`)
	password := readPassword("Enter a master password", 6, true, false)
	hint := stringToSQLNullString(readField("Enter a hint", nil, true, no_str))
	fmt.Println()

	salt, err := encrypt.GenSalt()
	if err != nil {
		log.Fatal(err)
	}
	kek := encrypt.DeriveKEK(password, salt)
	dek, err := encrypt.GenKey()
	if err != nil {
		log.Fatal(err)
	}
	wdek, err := encrypt.Seal(kek, dek)
	if err != nil {
		log.Fatal(err)
	}
	//
	storage.SetVault(storage.VaultBlobs{Wdek: wdek, Salt: salt, Hint: hint})

	return dek
}

func NewVault() []byte {
	no_str := ptr("")
	resp := strings.TrimSpace(readField("\nCreate a new vault? [y/N]", nil, true, no_str))
	if resp != "y" && resp != "Y" {
		fmt.Print("Canceled.\n\n")
		return nil
	}
	// save old
	var err error
	path, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	savedTo, err := storeOldVault(filepath.Dir(path))

	storage.Close()
	storage.TryInit()
	defer storage.Close()

	if err != nil {
		log.Fatal(err)
	} else if savedTo != "" {
		fmt.Println("Saved old vault to", savedTo)
	}

	return UpdateVault()
}
