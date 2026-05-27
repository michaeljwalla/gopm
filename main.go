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

	fmt.Println("\nFinal entry:\n", my_entry)

	// return
}
