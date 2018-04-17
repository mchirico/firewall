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

// Configuration for logs
type Configuration struct {
	WhiteListIPS         []string
	OutputLog            string
	QueryIntervalSeconds string
	SearchLogs           []Ports
}

// Reads info from config file
func ReadConfig(file string) Configuration {

	f, _ := os.Open(file)
	defer f.Close()
	decoder := json.NewDecoder(f)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(configuration.WhiteListIPS)
	return configuration

}

// Read
func Read(file string) ([]byte, error, int) {

	f, err := os.Open(file)
	if err != nil {
		return []byte{}, err, 0

	}

	b := make([]byte, MaxFileSize)
	n, err := f.Read(b)

	return b[0:n], err, n
}

// Parse
func Parse(b []byte) map[string]int {

	re := regexp.MustCompile(
		".*Invalid user.* ([0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]+)")

	ips := map[string]int{}

	lines := strings.Split(string(b), "\n")
	for _, line := range lines {
		list := re.FindStringSubmatch(line)

		if len(list) == 2 {
			ips[list[1]] += 1
		}

	}
	return ips
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
