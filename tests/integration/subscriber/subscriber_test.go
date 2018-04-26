package subscriber

import (
	. "github.com/mchirico/firewall/fixtures"
	"os"
	"testing"
	"reflect"
	"github.com/mchirico/firewall/utils"
	"github.com/mchirico/firewall/subscriber"
	"time"
	"log"
	"fmt"
	"strings"
	"github.com/mchirico/firewall/watch"
)

var alertLogTestfile = "../../../fixtures/tempfoo"

func TestMain(m *testing.M) {

	UnEncryptedFiles = "../../../fixtures/stage/access.log.stage"
	if StageCheck() {
		return
	}
	DeleteConfig()
	CreateActiveStageDirs()
	// call flag.Parse() here if TestMain uses flags
	os.Exit(m.Run())

}

func mapOfOuput(file string) map[string]int {

	m := map[string]int{}
	f, err := os.OpenFile(file, os.O_RDONLY, 0600)
	if err != nil {
		log.Printf("Could not open file: %v\n", file)
		return m
	}
	b := make([]byte, 19000)
	n, err := f.Read(b)
	s := string(b[0:n])
	strings.Split(s, "\n")

	for _, v := range strings.Split(s, "\n") {
		ip := strings.Split(v, ":")[0]
		v, found := m[ip]
		if found {
			m[ip] = v + 1
		}
		m[ip] = 0
	}
	return m

}

// Finally getting close to full prototype...
func TestFirewallWatch(t *testing.T) {

	os.Remove("/tmp/firewall.cmd")

	f, _ := os.OpenFile(alertLogTestfile,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_TRUNC, 0600)

	c := utils.ReadConfig("../../../fixtures/config.json")
	// Changing log for our test
	c.SearchLogs[0].Log = alertLogTestfile

	fw := &utils.Firewall{Config: c}

	stageCmd := "echo  %v:  %v >>/tmp/firewall.cmd\n"
	fwcmd := subscriber.CreateCmdS(stageCmd)

	// Normally command below; however, for test directories
	// are relative
	//     fwcmd.SetWriteLog(c.OutputLog)
	fwcmd.SetWriteLog("../../../fixtures/firewall.json")
	fw.SetCmdSlave(fwcmd)

	fw.Read()
	fw.Parse()

	m := watch.NewMC(alertLogTestfile, fw)
	cmd := watch.OpenWatcher(m.WriteEvent, m.AllEvents,
		m.Tick, alertLogTestfile)

	cmd.Watcher()

	if m.Count() != 0 {
		t.Errorf("Count should be zero")
	}

	time.Sleep(1 * time.Second)

	expectedString := "Apr 15 06:26:38 t sshd[12253]: Invalid user " +
		"api from 8.199.139.46\n"

	f.WriteString(expectedString)
	f.Sync()
	time.Sleep(3 * time.Second)

	log.Println("m.GetB()", string(m.GetB()))
	if string(m.GetB()) != expectedString {
		t.Errorf("Error b %v expected %v",
			m.GetB(), expectedString)
	}

	time.Sleep(1 * time.Second)
	if m.Count() != 1 {
		t.Errorf("Count should be one")
	}

	// Stop restart test
	cmd.Stop()
	fmt.Println("see done above?")
	time.Sleep(3 * time.Second)
	cmd.Watcher()
	time.Sleep(1 * time.Second)

	expectedString = "Apr 15 06:26:38 t sshd[12253]: Invalid user " +
		"api from 28.199.139.46\n"

	f.WriteString(expectedString)
	f.Sync()
	time.Sleep(1 * time.Second)

	if m.Count() != 2 {
		t.Errorf("Count should be two: %d", m.Count())
	}

	log.Println("m.GetB()", string(m.GetB()))
	if string(m.GetB()) != expectedString {
		t.Errorf("Error b %v expected %v",
			m.GetB(), expectedString)
	}

	// Extra to investigate
	expectedString = "Apr 15 06:26:38 t sshd[12253]: Invalid user " +
		"api from 128.199.139.46\n"

	f.WriteString(expectedString)
	f.Sync()
	time.Sleep(1 * time.Second)


	// Delete file and restart test
	time.Sleep(3 * time.Second)
	os.Remove(alertLogTestfile)
	f.Close()
	time.Sleep(1 * time.Second)

	f, _ = os.OpenFile(alertLogTestfile,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)

	for i := 0; i < 10; i++ {

		expectedString = "Apr 15 06:26:38 t sshd[12253]: Invalid user " +
			"api from 228.1.139.45\n"

		f.WriteString(expectedString)
		f.Sync()
		time.Sleep(1 * time.Second)
	}

	time.Sleep(1 * time.Second)

	mapResult := mapOfOuput("/tmp/firewall.cmd")

	expectedMap := map[string]int{
		"8.199.139.46":   0,
		"28.199.139.46":  0,
		"128.199.139.46": 0,
		"228.1.139.45":   0,
		"":               0, // Extra Return
	}

	if !reflect.DeepEqual(mapResult, expectedMap) {

		t.Errorf("Output incorrect: "+
			"mapResult: %v,  expectedMap: %v\n", mapResult,
			expectedMap)
	}



}
