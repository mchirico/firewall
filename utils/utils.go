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

	"bytes"
	"encoding/json"
	"sort"
)

var MaxFileSize = 500000000

// Set
type Set struct {
	set map[string]bool
}

// CreateS
func CreateS() *Set {
	return &Set{map[string]bool{}}
}

// Insert
func (set *Set) Add(s string) bool {
	_, found := set.set[s]
	set.set[s] = true
	return found
}

// In
func (set *Set) In(s string) bool {
	_, found := set.set[s]
	return found
}

func (s *Set) Union(s2 *Set) *Set {
	t := CreateS()
	for k, v := range s2.set {
		t.set[k] = v
	}
	for k, v := range s.set {
		t.set[k] = v
	}
	return t

}

func (s *Set) Diff(s2 *Set) *Set {
	t := CreateS()

	for k, v := range s2.set {
		if !s.In(k) {
			t.set[k] = v
		}
	}
	return t

}

func (s *Set) Keys() []string {
	t := []string{}
	for k := range s.set {
		t = append(t, k)
	}
	sort.Strings(t)
	return t

}

// IpRec -- ip's to block
type IpRec struct {
	IP    string
	Count int
	Ports []int
}

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
	StatusLog            string
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

// ReadConfig - Reads info from config file
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

// CreateFirewall
func CreateFirewall(file string) *Firewall {
	c := ReadConfig(file)
	return &Firewall{Config: c}
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

// CreateIpRec --
func (fw *Firewall) CreateIpRec() []IpRec {

	iprecs := []IpRec{}

	for idx, m := range fw.BadIP {

		ports := fw.Config.SearchLogs[idx].Ports
		for k, v := range m {
			t := &IpRec{IP: k, Count: v, Ports: ports}
			iprecs = append(iprecs, *t)
		}
	}
	return iprecs
}

// WriteRecs
func (fw *Firewall) WriteRecs(iprecs []IpRec) {

	f, err := os.OpenFile(fw.Config.OutputLog, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Printf("WriteRecs openFile error %v", err)
	}
	defer f.Close()

	iprecsJson, _ := json.Marshal(iprecs)
	f.Write(iprecsJson)
	return
}

// ReadRecs
func (fw *Firewall) ReadRecs() []IpRec {

	f, err := os.OpenFile(fw.Config.OutputLog, os.O_RDONLY, 0600)
	if err != nil {
		log.Printf("Can't open file %e", err)
	}
	b := make([]byte, MaxFileSize)
	n, err := f.Read(b)
	if err != nil {
		log.Printf("ReadRecs: Can't read file %e", err)
	}
	iprecs := []IpRec{}

	err = json.Unmarshal(b[0:n], &iprecs)
	if err != nil {
		log.Printf("ReadRecs: json.Unmarshal %e", err)
	}

	return iprecs
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

func SetLogging() bytes.Buffer {

	var str bytes.Buffer
	log.SetOutput(&str)
	return str

}
