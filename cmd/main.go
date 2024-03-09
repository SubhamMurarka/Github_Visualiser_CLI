package main

import (
	"flag"
	"fmt"
)

func main() {
	var folder string
	var email string

	flag.StringVar(&folder, "add", "", "add new folder to scan for git repo")
	flag.StringVar(&email, "email", "example@gmail.com", "the email to scan")

	flag.Parse()

	fmt.Println(folder)

	if folder != "" {
		Scan(folder)
		return
	}

	Stats(email)
}
