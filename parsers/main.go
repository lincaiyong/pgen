package main

import (
	"github.com/lincaiyong/log"
	"github.com/lincaiyong/pgen"
	"os"
	"time"
)

func main() {
	start := time.Now()
	grammar, err := pgen.PreProcess("go/go.txt")
	if err != nil {
		log.ErrorLog("fail to preprocess: %v", err)
		return
	}
	output, err := pgen.Run(grammar)
	if err != nil {
		log.ErrorLog("fail to run: %v", err)
		return
	}
	_ = os.Mkdir("goparser", os.ModePerm)
	err = os.WriteFile("goparser/goparser.go", []byte(output), 0644)
	if err != nil {
		log.ErrorLog("fail to write file: %v", err)
		return
	}
	log.InfoLog("finished in %s\n", time.Since(start))
}
