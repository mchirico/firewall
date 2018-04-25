package simple

import (
	. "github.com/mchirico/firewall/fixtures"
	"github.com/mchirico/firewall/utils"
	"github.com/mchirico/firewall/watch"
	. "github.com/mchirico/firewall/watch"
	"log"
	"os"
	"strings"
	"testing"
)

type T struct {
	i int
}

func (t *T) WriteEvent(event string) {
	log.Printf("WriteEvent: %v", event)
	t.OtherStuff()
}
func (t *T) AllEvents(event string) {

}
func (t *T) Tick(event string) {

}
func (t *T) OtherStuff() {
	log.Printf("Other Stuff\n")
}

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

func TestFirewallParse(t *testing.T) {

	RemoveActiveStageDirs()
	UpdateConfigSettings()
	CopyStageFilesBeginEnd(0, 2)
	c := UpdateConfigSettings()

	fw := &utils.Firewall{Config: c}
	fw.Read()
	fw.Parse()

	if fw.BadIP[0]["87.138.66.123"] != 1 {
		t.Errorf("Didn't read entries: %v", fw.BadIP[0])
	}
}

func TestCopyStageFilesBeginEnd(t *testing.T) {

	// Setup -- use for real test
	RemoveActiveStageDirs()
	UpdateConfigSettings()

	CopyStageFilesBeginEnd(0, 2)

	c := UpdateConfigSettings()
	if !watch.FileExist(c.SearchLogs[0].Log) {
		t.Errorf("file not created: %v",
			c.SearchLogs[0].Log)
	}
	if !watch.FileExist(c.SearchLogs[1].Log) {
		t.Errorf("file not created")
	}

	// Begin test

	f, _ := os.OpenFile(c.SearchLogs[0].Log, os.O_RDONLY, 0600)
	b := make([]byte, 500)
	n, _ := f.Read(b)
	// fmt.Println(string(b[0:n]))
	s := string(b[0:n])
	count := strings.Count(s, "Invalid user supervisor from 87.138.66.123")
	if count != 1 {
		t.Errorf("Could not read log")
	}

	count = strings.Count(s, "error: maximum authentication "+
		"attempts exceeded ")
	if count >= 1 {
		t.Errorf("Read too many lines")
	}

}

func TestCreateStageFilesBeginEnd(t *testing.T) {

	CreateStageFilesBeginEnd(0, 2)
	c := UpdateConfigSettings()
	if !watch.FileExist(c.SearchLogs[0].Log) {
		t.Errorf("file not created")
	}
	if !watch.FileExist(c.SearchLogs[1].Log) {
		t.Errorf("file not created")
	}
	s := LogRead(c, 50000, 0)
	log.Printf(".\n\n.. %s", s)

}

func TestWatchfile(t *testing.T) {

	CopyStageFilesBeginEnd(0, 2)
	c := UpdateConfigSettings()
	if !watch.FileExist(c.SearchLogs[0].Log) {
		t.Errorf("file not created")
	}
	if !watch.FileExist(c.SearchLogs[1].Log) {
		t.Errorf("file not created")
	}
	s := LogRead(c, 50000, 0)
	log.Printf(".\n\n.. %s", s)

	file := c.SearchLogs[0].Log
	log.Printf("Log name: %v\n", file)

	slave := &T{}
	m := NewMC(file, slave)

	cmd := OpenWatcher(m.WriteEvent, m.AllEvents,
		m.Tick, file)

	cmd.Watcher()

	if m.Count() != 0 {
		t.Errorf("Count should be zero")
	}

	CopyStageFilesBeginEnd(2, 5)
	if m.Count() < 1 {
		t.Errorf("Count: %v\n", m.Count())
	}
	CopyStageFilesBeginEnd(5, 15)
	log.Printf("Count: %v\n", m.Count())

}
