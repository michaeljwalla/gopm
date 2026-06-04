package main

import (
	"fmt"
	"log"
	"main/packages/passfield"
	"main/packages/passfield/termutils"
	"main/packages/storage"
)

func main() {
	fmt.Println("Core")
	if storage.TryInit() {
		defer storage.Close()
	}

	my_entry := passfield.PassFieldBasic{}
	termutils.PopulatePassFieldBasic(&my_entry)

	fmt.Printf("\nFinal entry:\n%s\n", my_entry)

	err := storage.Save(&my_entry)
	if err != nil {
		log.Fatal(err)
	}
	// return
}
