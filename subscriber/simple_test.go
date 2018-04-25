package subscriber

import (
	"fmt"

	"time"

	. "github.com/mchirico/firewall/fixtures"
	"github.com/mchirico/firewall/utils"

	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {

	UnEncryptedFiles = "../fixtures/stage/access.log.stage"
	if StageCheck() {
		return
	}
	DeleteConfig()
	CreateActiveStageDirs()
	// call flag.Parse() here if TestMain uses flags
	os.Exit(m.Run())

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

func TestSubscriber_StagedRun(t *testing.T) {

	os.Remove("/tmp/firewall.cmd")
	str := utils.SetLogging()
	c := utils.ReadConfig("../fixtures/config.json")

	fw := &utils.Firewall{Config: c}

	// Example of push in command
	stageCmd := "echo  %v:  %v >>/tmp/firewall.cmd\n"
	cmd := CreateCmdS(stageCmd)
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
