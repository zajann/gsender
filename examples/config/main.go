package main

import (
	"fmt"
	"log"

	"github.com/zajann/gsender/internal/config"
)

func main() {
	c, err := config.Load("/home/zajan/git/gsender/configs/gsender_config.yml")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(c)
}
