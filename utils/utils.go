package utils

import (
	"os"
	"os/exec"
	"regexp"
	"strings"

	"context"
	"fmt"
	"log"
	"time"

	"encoding/json"
)

var MaxFileSize = 500000000

// Ports -- used for search logs
type Ports struct {
	Log   string
	Ports []int
	Regex string
}

// Config for logs
type Config struct {
	WhiteListIPS         []string
	OutputLog            string
	QueryIntervalSeconds string
	SearchLogs           []Ports
}

// Firewall
type Firewall struct {
	Config  Config
	LogData [][]byte
	Events  []string
	BadIP   []map[string]int
}

// Reads info from config file
func ReadConfig(file string) Config {

	f, _ := os.Open(file)
	defer f.Close()
	decoder := json.NewDecoder(f)
	Config := Config{}
	err := decoder.Decode(&Config)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(Config.WhiteListIPS)
	return Config

}

// Read
func (fw *Firewall) Read() {

	fw.LogData = [][]byte{}
	for _, slog := range fw.Config.SearchLogs {

		f, err := os.Open(slog.Log)
		if err != nil {

			fw.LogData = append(fw.LogData, []byte{})
			log.Printf("Error Opening file" +
				" in read Read")
			continue

		}

		b := make([]byte, MaxFileSize)
		n, err := f.Read(b)
		if err != nil {
			fw.LogData = append(fw.LogData, []byte{})
			log.Printf("Error Reading file" +
				" in read Read")
			continue
		}
		fw.LogData = append(fw.LogData, b[0:n])
	}

}

// Parse
func (fw *Firewall) Parse() {

	fw.BadIP = []map[string]int{}

	for idx, b := range fw.LogData {

		regxString := fw.Config.SearchLogs[idx].Regex
		re := regexp.MustCompile(
			regxString)

		ips := map[string]int{}
		lines := strings.Split(string(b), "\n")
		for _, line := range lines {
			list := re.FindStringSubmatch(line)

			if len(list) == 2 {
				ips[list[1]] += 1
			}

		}

		fw.BadIP = append(fw.BadIP, ips)
	}

}

// Cmd to execute
func Cmd(cmd string) {

	ctx, cancel := context.WithTimeout(context.Background(), 10000*time.Millisecond)
	defer cancel()

	if err := exec.CommandContext(ctx, "sh", "-c", cmd).Run(); err != nil {
		log.Printf("Command failed")
		fmt.Printf("f")
	}

}
