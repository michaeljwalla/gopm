package main

import (
	"fmt"
	"main/packages/passfield"
	"main/packages/passfield/termutils"
)

func main() {
	fmt.Println("Core")

	my_entry := passfield.PassFieldSite{}
	termutils.PopulatePassFieldSite(&my_entry)

	fmt.Printf("\nFinal entry:\n%s\n", my_entry)

	// return
}
