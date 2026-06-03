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

	db := storage.GetInstance()
	defer db.Close()

	my_entry := passfield.PassFieldSite{}
	termutils.PopulatePassFieldSite(&my_entry)

	fmt.Printf("\nFinal entry:\n%s\n", my_entry)

	err := storage.Save(db, &my_entry)
	if err != nil {
		log.Fatal(err)
	}
	// return
}
