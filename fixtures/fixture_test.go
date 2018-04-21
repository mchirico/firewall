package fixtures

import (
	"encoding/json"
	"fmt"
	"github.com/mchirico/firewall/utils"
	"github.com/mchirico/firewall/watch"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {

	if StageCheck() {
		return
	}
	CreateActiveStageDirs()
	// call flag.Parse() here if TestMain uses flags
	os.Exit(m.Run())

}

func TestGetBaseDir(t *testing.T) {
	UseTemp = true
	if GetBaseDir() != "/tmp" {
		t.Errorf(" BaseDir incorrect: %s", GetBaseDir())
	}

	UseTemp = !UseTemp
	path := os.Getenv("GOPATH")
	expected := fmt.Sprintf(stageFiles, path)
	if GetBaseDir() != expected {
		t.Errorf(" BaseDir incorrect: %s "+
			"expected: %s", GetBaseDir(), expected)
	}

}

func TestFileExist(t *testing.T) {

	_, file := CreateConfig()
	if !watch.FileExist(file) {
		t.Errorf(" File was not created")
	}
}

func TestRemoveActiveStageDirs(t *testing.T) {

	parent := GetBaseDir()
	dirs := make([]DirStruct, len(Dirs))

	for i, d := range Dirs {
		dirs[i].file = parent + d.file

	}

	for _, dir := range dirs {
		if !watch.FileExist(dir.file) {
			t.Errorf(" File was not created")
		}
	}

	RemoveActiveStageDirs()

	for _, dir := range dirs {
		if watch.FileExist(dir.file) {
			t.Errorf(" Directory should have been " +
				"deleted")
		}
		log.Println(dir)
	}

}

func TestCreateConfig(t *testing.T) {

	expectedValue := "junkTest"

	RemoveActiveStageDirs()
	_, dst := CreateConfig()

	c := utils.ReadConfig(dst)
	fmt.Printf("c: %v\n", c)
	c.OutputLog = expectedValue

	cJson, _ := json.Marshal(c)
	err := ioutil.WriteFile(dst, cJson, 0644)
	if err != nil {
		log.Println(err)
	}

	c = utils.ReadConfig(dst)

	if c.OutputLog != expectedValue {
		t.Errorf("Ouput not saved %v, %v\n"+
			"  ", c.OutputLog, expectedValue)
	}

}

func TestUpdateConfigLogs(t *testing.T) {
	_, dst := CreateConfig()
	c := utils.ReadConfig(dst)
	UpdateConfigLogs(c, dst)

}

func TestUpdateConfigSettings(t *testing.T) {
	c := UpdateConfigSettings()
	if watch.FileExist(c.StatusLog) {
		t.Errorf("File should exist:%s", c.StatusLog)
	}

	log.Println(c.SearchLogs[0].Log)
}

func TestCopyStageFiles(t *testing.T) {
	CopyStageFiles()
	c := UpdateConfigSettings()
	watch.FileExist(c.SearchLogs[0].Log)
	watch.FileExist(c.SearchLogs[1].Log)

	f, _ := os.OpenFile(c.SearchLogs[0].Log, os.O_RDONLY, 0600)
	b := make([]byte, 500)
	f.Read(b)
	fmt.Println(string(b))
	s := string(b)
	count := strings.Count(s, "Invalid user supervisor from 87.138.66.123")
	if count != 1 {
		t.Errorf("Could not read log")
	}
}
