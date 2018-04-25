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
	RemoveActiveStageDirs()

	CopyStageFiles()
	c := UpdateConfigSettings()
	watch.FileExist(c.SearchLogs[0].Log)
	watch.FileExist(c.SearchLogs[1].Log)

	f, _ := os.OpenFile(c.SearchLogs[0].Log, os.O_RDONLY, 0600)
	b := make([]byte, 500)
	f.Read(b)
	//fmt.Println(string(b))
	s := string(b)
	count := strings.Count(s, "Invalid user supervisor from 87.138.66.123")
	if count != 1 {
		t.Errorf("Could not read log")
	}
}

func TestCopyFileBeginEnd(t *testing.T) {

	file := "./testF"
	fileOut := "./testF.out"
	os.Remove(file)
	os.Remove(fileOut)

	f, _ := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0600)
	f.WriteString("line 1\nline 2\nline 3\nline 4\n")
	f.Close()

	CopyFileBeginEnd(file, fileOut, 2, 3)

	f, _ = os.OpenFile(fileOut, os.O_RDONLY, 0600)
	b := make([]byte, 300)
	n, err := f.Read(b)
	if n == 0 || err != nil {
		t.Errorf("File could not be read: %v",
			file+".out")
	}
	if string(b[0:n]) != "line 3\n" {
		t.Errorf("Expected value ->%v<-, value returned ->%v<-",
			"line 3", string(b[0:n]))
	}

}

func TestCopyStageFilesBeginEnd(t *testing.T) {

	// Setup -- use for real test
	RemoveActiveStageDirs()
	CopyStageFilesBeginEnd(0, 2)

	c := UpdateConfigSettings()
	if !watch.FileExist(c.SearchLogs[0].Log) {
		t.Errorf("file not created")
	}
	if !watch.FileExist(c.SearchLogs[1].Log) {
		t.Errorf("file not created")
	}

	// Begin test

	s := LogRead(c, 500, 0)

	count := strings.Count(s, "Invalid user supervisor from 87.138.66.123")
	if count != 1 {
		t.Errorf("Could not read log")
	}

	count = strings.Count(s, "error: maximum authentication "+
		"attempts exceeded ")
	if count >= 1 {
		t.Errorf("Read too many lines")
	}

	CopyStageFilesBeginEnd(0, 16)
	s = LogRead(c, 500, 0)

	count = strings.Count(s, "error: maximum authentication "+
		"attempts exceeded ")
	if count == 0 {
		t.Errorf("Read too few lines: %v", count)
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

func TestReadConfigFirewall(t *testing.T) {

	c := UpdateConfigSettings()
	if watch.FileExist(c.StatusLog) {
		t.Errorf("File should exist:%s", c.StatusLog)
	}
	CreateStageFilesBeginEnd(0, 20)

	fw := &utils.Firewall{Config: c}
	fw.Read()
	fw.Parse()

	iprecs := fw.CreateIpRec()

	if iprecs[0].Ports[2] != 80 {
		t.Fatalf("Didn't get port: %v", iprecs[0])
	}

}

func TestFirewallLogging(t *testing.T) {

	c := UpdateConfigSettings()
	if watch.FileExist(c.StatusLog) {
		t.Errorf("File should exist:%s", c.StatusLog)
	}
	CreateStageFilesBeginEnd(0, 200)

	fw := &utils.Firewall{Config: c}
	fw.Read()
	fw.Parse()

	// Writing to log: GetOutLog()
	iprecs := fw.CreateIpRec()
	fw.WriteRecs(iprecs)

	log.Printf("Does file exist?")
	if !watch.FileExist(fw.GetOutLog()) {
		t.Errorf("File does not exist: %v\n\n", fw.GetOutLog())
	} else {
		log.Printf("Yes...file exists..\n")
	}

	// Travis has memory limitation
	utils.MaxFileSize = 50000
	iprecsFromJsonLog := fw.ReadRecs()
	log.Printf("got results from fw.ReadRecs()\n ")

	testResult := false

	for n, rec := range iprecsFromJsonLog {
		if n%3 == 0 {
			log.Printf("count: %v\n", n)
		}

		if rec.IP == "171.244.10.79" {
			testResult = true
			if rec.Count != 1 {
				t.Errorf("Count incorrect: %v\n", rec.Count)
			}
		}
	}
	if !testResult {
		t.Fatalf(" Could not find ip")
	}

	rec := iprecsFromJsonLog[0]
	log.Printf("r %v %v\n", rec.IP, rec.Count)
}

func TestCmdWatherWithFirewall(t *testing.T) {

}
