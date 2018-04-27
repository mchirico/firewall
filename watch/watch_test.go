package watch

import (
	"fmt"
	"github.com/mchirico/firewall/subscriber"
	"github.com/mchirico/firewall/utils"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

var file = "../fixtures/tempfoo"

type T struct {
	i int
}

func (t *T) WriteEvent(event string) {
	//log.Printf("WriteEvent: %v", event)
	t.OtherStuff()
}
func (t *T) AllEvents(event string) {

}
func (t *T) Tick(event string) {

}
func (t *T) OtherStuff() {
	log.Printf("Other Stuff\n")
}

func TestFileExist(t *testing.T) {
	file := "../fixtures/junktest"

	f, err := os.OpenFile(file,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_TRUNC, 0600)
	if err != nil {
		t.Errorf("Can't create test file")
	}
	defer f.Close()
	f.WriteString("test")

	if !FileExist(file) {
		t.Errorf("file should exist")
	}

}

func TestGeneralSlave(t *testing.T) {

	f, _ := os.OpenFile(file,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_TRUNC, 0600)

	slave := &T{}
	m := NewMC(file, slave)
	cmd := OpenWatcher(m.WriteEvent, m.AllEvents,
		m.Tick, file)

	cmd.Watcher()

	if m.Count() != 0 {
		t.Errorf("Count should be zero")
	}

	time.Sleep(1 * time.Second)

	expectedString := "  Write 0"
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

	cmd.Stop()
	fmt.Println("see done above?")
	time.Sleep(3 * time.Second)
	cmd.Watcher()
	time.Sleep(1 * time.Second)

	expectedString = "  test 2...."
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

	time.Sleep(4 * time.Second)
	cmd.Stop()

	time.Sleep(3 * time.Second)
	os.Remove(file)

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

// TODO: Only full test need another...
func TestFireWallSlave(t *testing.T) {

	os.Remove("/tmp/firewall.cmd")

	f, _ := os.OpenFile(file,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_TRUNC, 0600)

	c := utils.ReadConfig("../fixtures/config.json")
	// Changing log for our test
	c.SearchLogs[0].Log = file

	fw := &utils.Firewall{Config: c}

	stageCmd := "echo  %v:  %v >>/tmp/firewall.cmd\n"
	fwcmd := subscriber.CreateCmdS(stageCmd)
	fw.SetCmdSlave(fwcmd)

	fw.Read()
	fw.Parse()

	m := NewMC(file, fw)
	cmd := OpenWatcher(m.WriteEvent, m.AllEvents,
		m.Tick, file)

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

	time.Sleep(4 * time.Second)
	cmd.Stop()

	time.Sleep(3 * time.Second)
	os.Remove(file)

	mapResult := mapOfOuput("/tmp/firewall.cmd")

	expectedMap := map[string]int{
		"8.199.139.46":   0,
		"28.199.139.46":  0,
		"128.199.139.46": 0,
		"":               0, // Extra Return
	}

	if !reflect.DeepEqual(mapResult, expectedMap) {

		t.Errorf("Output incorrect: "+
			"mapResult: %v,  expectedMap: %v\n", mapResult,
			expectedMap)
	}

}

func TestFileDeletedAndRestored(t *testing.T) {

	f, _ := os.OpenFile(file,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_TRUNC, 0600)

	slave := &T{}
	m := NewMC(file, slave)

	cmd := OpenWatcher(m.WriteEvent, m.AllEvents,
		m.Tick, file)

	cmd.Watcher()

	f.WriteString("test")
	f.Close()

	if m.Count() != 0 {
		t.Errorf("Count should be  0 %v ", m.Count())
	}

	if m.StatusRemoveRename() == true {
		t.Errorf("Status should be false")
	}

	t.Log("\n -- Now Removing File --\n")

	os.Remove(file)
	time.Sleep(2 * time.Second)

	if m.StatusRemoveRename() != true {
		t.Errorf("File was remove  " +
			"Not showing up")
	}

	f, _ = os.OpenFile(file,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)

	time.Sleep(10 * time.Second)
	//cmd.Watcher()

	expectedString := "  test 2...."
	f.WriteString(expectedString)
	f.Sync()

	time.Sleep(1 * time.Second)
	if string(m.GetB()) != expectedString {
		t.Errorf("Error b %v expected %v",
			m.GetB(), expectedString)
	}

	if m.Count() == 0 {
		t.Errorf("Count should be > 0 %v ", m.Count())
	}

	cmd.Stop()
	f.Close()

}

func TestMC_LastEvent(t *testing.T) {

	f, _ := os.OpenFile(file,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_TRUNC, 0600)

	slave := &T{}
	m := NewMC(file, slave)

	cmd := OpenWatcher(m.WriteEvent, m.AllEvents,
		m.Tick, file)

	cmd.Watcher()

	lastEvent, err := m.LastEvent()
	if err == nil {
		t.Errorf("LastEvent should "+
			"not have any entries: %v", lastEvent)
	}

	f.WriteString("test")
	f.Sync()
	time.Sleep(1 * time.Second)
	lastEvent, err = m.LastEvent()
	if err != nil {
		t.Errorf("LastEvent should "+
			"have any entries: %v", lastEvent)
	}
	if lastEvent.event != "WRITE" {
		t.Errorf("Event: %v, expected: %v\n",
			lastEvent.event, "WRITE")
	}

	expectedB := "test"
	if string(m.GetB()) != expectedB {
		t.Errorf("GetB(): ->%v<-  Expected value: "+
			"->%v<-", string(m.GetB()), expectedB)
	}

	for i := 0; i < 12; i++ {
		expectedB += "\n more Data\n"
	}
	f.WriteString(expectedB)
	f.Sync()
	time.Sleep(1 * time.Second)
	if string(m.GetB()) != expectedB {
		t.Errorf("GetB(): ->%v<-  Expected value: "+
			"->%v<-", string(m.GetB()), expectedB)
	}

}
