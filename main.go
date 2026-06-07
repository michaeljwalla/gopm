package main

import (
	"fmt"
	"log"
	"main/packages/passfield"
	"main/packages/passfield/termutils"
	term "main/packages/passfield/termutils"
	"main/packages/storage"
)

func genEntry(dek []byte) {
	if storage.TryInit() {
		defer storage.Close()
	}
	my_entry := passfield.PassFieldBasic{}
	termutils.PopulatePassFieldBasic(&my_entry)

	err := storage.Save(dek, &my_entry)
	if err != nil {
		log.Fatal(err)
	}
}

func printEntries(dek []byte) {
	if storage.TryInit() {
		defer storage.Close()
	}
	entries, err := storage.GetEntries(dek)
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range entries {
		fmt.Println(v)
	}
}
func main() {
	fmt.Println("Core")
	if storage.TryInit() {
		defer storage.Close()
	}
	//
	vault := storage.GetVault()
	dek := term.RequestDEK(vault)

	genEntry(dek)
	printEntries(dek)
	// return
}
