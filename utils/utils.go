package utils

import (
	"os"
	"os/exec"
	"regexp"
	"strings"

	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/mchirico/firewall/include"
	"log"
	"sync"
	"time"
)

var MaxFileSize = 5000000

// Cmd to execute
func Cmd(cmd string) {

	ctx, cancel := context.WithTimeout(context.Background(), 10000*time.Millisecond)
	defer cancel()

	if err := exec.CommandContext(ctx, "sh", "-c", cmd).Run(); err != nil {
		log.Printf("Command failed")
	}

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
	sync.Mutex
	Config   Config
	LogData  [][]byte
	Events   []string
	BadIP    []map[string]int
	MarkedIP []include.IpRec
	cmdSlave include.CmdSlave
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

func (fw *Firewall) SetCmdSlave(c include.CmdSlave) {
	fw.Lock()
	defer fw.Unlock()
	fw.cmdSlave = c
}

// Read
func (fw *Firewall) Read() {
	fw.Lock()
	defer fw.Unlock()

	fw.LogData = [][]byte{}
	for _, slog := range fw.Config.SearchLogs {

		f, err := os.Open(slog.Log)
		if err != nil {
			// TODO: Do we need this?
			// fw.LogData = append(fw.LogData, []byte{})
			//log.Printf("Error Opening file: %v"+
			//	" in read Read", slog.Log)
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
	fw.Lock()
	defer fw.Unlock()
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
				ips[list[1]]++
			}

		}

		fw.BadIP = append(fw.BadIP, ips)
	}

}

// CreateIpRec --
func (fw *Firewall) CreateIpRec() []include.IpRec {
	fw.Lock()
	defer fw.Unlock()
	iprecs := []include.IpRec{}

	for idx, m := range fw.BadIP {

		ports := fw.Config.SearchLogs[idx].Ports
		for k, v := range m {
			t := &include.IpRec{IP: k, Count: v, Ports: ports}
			iprecs = append(iprecs, *t)
		}
	}
	return iprecs
}

// createIpRec -- No locking
func (fw *Firewall) createIpRec() []include.IpRec {
	iprecs := []include.IpRec{}
	for idx, m := range fw.BadIP {

		ports := fw.Config.SearchLogs[idx].Ports
		for k, v := range m {
			t := &include.IpRec{IP: k, Count: v, Ports: ports}
			iprecs = append(iprecs, *t)
		}
	}
	return iprecs
}

func (fw *Firewall) GetOutLog() string {
	fw.Lock()
	defer fw.Unlock()
	return fw.Config.OutputLog
}

// WriteRecs
func (fw *Firewall) WriteRecs(iprecs []include.IpRec) {
	fw.Lock()
	defer fw.Unlock()
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
func (fw *Firewall) ReadRecs() []include.IpRec {
	fw.Lock()
	defer fw.Unlock()

	f, err := os.OpenFile(fw.Config.OutputLog, os.O_RDONLY, 0600)
	if err != nil {
		log.Printf("Can't open file %e", err)
	}
	log.Printf("Maxfile size: %v\n", MaxFileSize)
	b := make([]byte, MaxFileSize)
	n, err := f.Read(b)
	if err != nil {
		log.Printf("ReadRecs: Can't read file %e", err)
	}
	iprecs := []include.IpRec{}

	err = json.Unmarshal(b[0:n], &iprecs)
	if err != nil {
		log.Printf("ReadRecs: json.Unmarshal %e", err)
	}

	return iprecs
}

// FireCommand -- you'll need to take this out
func (fw *Firewall) FireCommand() {
	fw.Lock()
	defer fw.Unlock()
	iprecs := fw.createIpRec()
	for i, j := range iprecs {

		if fw.cmdSlave != nil {
			fw.cmdSlave.Build(i, j)
			fw.cmdSlave.Exe(i)
		} else {
			//log.Printf("fw.cmdSlave nil -- " +
			//	"nothing to fire")
		}

	}
}

// Do not lock these events

func (fw *Firewall) WriteEvent(event string) {
	fw.Read()
	fw.Parse()
	fw.FireCommand()
	//log.Printf("Firewall WriteEvent: %v", event)
}

func (fw *Firewall) AllEvents(event string) {
	//log.Printf("Firewall AllEvents: %v", event)
}

func (fw *Firewall) Tick(event string) {
	//log.Printf("Tick Tick: %v", event)
}

// SetLogging --
func SetLogging() bytes.Buffer {

	var str bytes.Buffer
	log.SetOutput(&str)
	return str

}
