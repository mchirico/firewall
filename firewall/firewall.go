package main

import (
	"fmt"
	"github.com/mchirico/firewall/utils"
	"log"
	"os"
)

func init() {
	//	c := utils.ReadConfig("../fixtures/config.json")
	//	fw := &utils.Firewall{Config: c}
	//	fw.Read()
	////	log.Println(fw.LogData[0][0:30])
}

func w() {
	f, err := os.OpenFile("../fixtures/junk.spock",
		os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {

	}
	f.WriteString("test")
	f.Close()
}

func b() {
	c := utils.ReadConfig("../fixtures/config.json")
	fw := &utils.Firewall{Config: c}
	fw.Read()
	log.Println(fw.LogData[0][0:30])
}

func main() {

	fmt.Println(os.Getenv("GOPATH"))

}
