package utils

import (
	"fmt"
	. "github.com/mchirico/firewall/include"
	"github.com/mchirico/firewall/set"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRead(t *testing.T) {

	c := ReadConfig("../fixtures/config.json")
	fw := &Firewall{Config: c}
	fw.Read()

	if "Apr 15 06:26:29 t sshd[12223]:" != string(fw.LogData[0][0:30]) {
		t.Fatalf("Cannot read log file")
	}

}

func TestParse(t *testing.T) {

	c := ReadConfig("../fixtures/config.json")
	fw := &Firewall{Config: c}
	fw.Read()
	fw.Parse()

	if fw.BadIP[0]["163.172.119.161"] != 2 {
		t.Errorf(("Didn't read entries"))
	}

}

func TestCreateIpRec(t *testing.T) {
	c := ReadConfig("../fixtures/config.json")
	fw := &Firewall{Config: c}
	fw.Read()
	fw.Parse()

	iprec := fw.CreateIpRec()

	if iprec[0].Ports[2] != 80 {
		t.Fatalf("Didn't get port")
	}

	fmt.Printf("iprec[9]: %v\n", iprec[9])

}

func TestWriteRecs(t *testing.T) {

	fw := CreateFirewall("../fixtures/config.json")
	fw.Read()
	fw.Parse()
	iprecs := fw.CreateIpRec()
	//log.Println(iprecs)
	fw.WriteRecs(iprecs)

}

func TestReadRecs(t *testing.T) {
	c := ReadConfig("../fixtures/config.json")
	fw := &Firewall{Config: c}
	fw.Read()
	fw.Parse()
	iprecs := fw.CreateIpRec()
	fw.WriteRecs(iprecs)

	recs := fw.ReadRecs()
	fmt.Printf("here %v\n", recs[0].IP)

}

func TestCmd(t *testing.T) {

	Cmd("date > /tmp/junk.txt")

}

func TestReadConfig(t *testing.T) {

	c := ReadConfig("../fixtures/config.json")
	if c.WhiteListIPS[1] != "102.29.34.4" {
		t.Errorf(("Error can't get ip"))
	}

	if c.SearchLogs[1].Log != "../fixtures/mail.log" {
		t.Errorf(("Error in search logs"))
	}
	if c.SearchLogs[1].Ports[0] != 25 {
		t.Errorf(("Error in search log port"))
	}
	if c.SearchLogs[1].Regex != ".*Invalid user.* "+
		"([0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]+)" {
		t.Errorf(("Error in search log regex"))
	}

}

func TestLogging(t *testing.T) {

	str := SetLogging()
	c := ReadConfig("../fixtures/config.json")
	fw := &Firewall{Config: c}
	fw.Read()
	fw.Parse()
	iprecs := fw.CreateIpRec()
	fw.WriteRecs(iprecs)

	fmt.Printf("->%s<-", str.String())

}

// Example of Command Structure (or maybe broker)
type CmdS struct {
	sync.Mutex
	cmd     string
	status  map[int]bool
	createS *set.Set
}

func CreateCmdS() *CmdS {
	c := &CmdS{}
	c.createS = set.CreateS()
	c.status = map[int]bool{}

	return c
}

func (cmdS *CmdS) Build(i int, iprec IpRec) {
	cmdS.Lock()
	defer cmdS.Unlock()

	tmpSet := set.CreateS()
	tmpSet.Add(set.CreateIpRec(iprec.IP, iprec.Ports))
	tmpSet = tmpSet.Difference(cmdS.createS)

	s := ""
	for k, v := range tmpSet.Values() {
		s += fmt.Sprintf("echo  %v:  %v >>/tmp/firewall.cmd\n",
			k, v)
	}
	cmdS.createS.Union(tmpSet)

	cmdS.cmd = s
	cmdS.status[i] = false
}

func (cmdS *CmdS) Exe(i int) {
	cmdS.Lock()
	defer cmdS.Unlock()
	Cmd(cmdS.cmd)
	cmdS.status[i] = true
}

func (cmdS *CmdS) ExeEnd(s string) {
	cmdS.Lock()
	defer cmdS.Unlock()

}

func (cmdS *CmdS) Tick(s string) {
	cmdS.Lock()
	defer cmdS.Unlock()

}

func checkForRepeats(file string) bool {
	f, err := os.OpenFile(file, os.O_RDONLY, 0600)
	if err != nil {
		return true
	}
	b := make([]byte, 19000)
	n, err := f.Read(b)
	s := string(b[0:n])
	strings.Split(s, "\n")
	m := map[string]int{}
	for _, v := range strings.Split(s, "\n") {
		ip := strings.Split(v, ":")[0]
		v, found := m[ip]
		if found {
			m[ip] = v + 1
			return true
		}
		m[ip] = 0
	}
	return false

}

func TestStagedRun(t *testing.T) {

	os.Remove("/tmp/firewall.cmd")
	str := SetLogging()
	c := ReadConfig("../fixtures/config.json")
	fw := &Firewall{Config: c}

	cmd := CreateCmdS()
	fw.SetCmdSlave(cmd)

	fw.Read()
	fw.Parse()
	iprecs := fw.CreateIpRec()
	fw.WriteRecs(iprecs)
	fmt.Printf("->%s<-", str.String())
	fmt.Printf("->%v<-", iprecs[0:3])

	fw.FireCommand()
	time.Sleep(1 * time.Second)

	f, err := os.OpenFile("/tmp/firewall.cmd", os.O_RDONLY, 0600)
	if err != nil {
		t.Errorf("/tmp/firewall.cmd not readable\n")
	}
	b := make([]byte, 9000)
	n, err := f.Read(b)

	if n > 100 {
		fmt.Printf("\nb=%v\n", string(b[0:100]))
	} else {
		fmt.Printf("\nb=%v\n", string(b[0:n]))
	}

	if checkForRepeats("/tmp/firewall.cmd") {
		t.Errorf("repeats found in /tmp/firewall.cmd")
	}

}
