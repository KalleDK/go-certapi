package main

import (
	"fmt"

	"github.com/KalleDK/go-certapi/certapi"
)

func main() {
	k := certapi.APIKey{}
	k[0] = 245
	fmt.Println(k)
}
