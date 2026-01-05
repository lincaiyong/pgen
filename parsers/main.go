package main

import (
	"fmt"
	"github.com/lincaiyong/pgen"
	"os"
	"time"
)

func main() {
	start := time.Now()
	b, err := os.ReadFile("go.txt")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	output, err := pgen.Run(string(b))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	_ = os.Mkdir("goparser", os.ModePerm)
	err = os.WriteFile("goparser/goparser.go", []byte(output), 0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("finished in %s\n", time.Since(start))
}
