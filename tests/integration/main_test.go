package integration

import (
	"fmt"
	. "github.com/mchirico/firewall/fixtures"
	"github.com/mchirico/firewall/watch"
	"os"
	"strings"
	"testing"
	"log"
)

func TestMain(m *testing.M) {

	UnEncryptedFiles = "../../fixtures/stage/access.log.stage"
	if StageCheck() {
		return
	}
	DeleteConfig()
	CreateActiveStageDirs()
	// call flag.Parse() here if TestMain uses flags
	os.Exit(m.Run())

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
	fmt.Println(string(b[0:n]))
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
