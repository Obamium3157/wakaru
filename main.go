package main

import (
	"fmt"
	"log"

	"wakaru/internal/jmdict"
)

func main() {
	dict, err := jmdict.InitDB("../jmdict/jmdict-examples-eng-3.6.2.json")
	if err != nil {
		log.Fatalf("%v", err)
	}

	fmt.Println(dict.Words[0])

	for idx, val := range dict.Tags {
		fmt.Printf("[%v]: %v\n", idx, val)
	}
}
