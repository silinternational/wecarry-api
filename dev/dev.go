package main

import (
	"fmt"
	"os"

	wecarryaws "github.com/silinternational/wecarry-api/aws"
)

// This utility establishes things in the dev environment that buffalo nor the app can set up on its own.
func main() {
	if err := wecarryaws.CreateS3Bucket(); err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
}
