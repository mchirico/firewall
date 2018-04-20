package fixtures

import (
	"fmt"
	"os"
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
	if !FileExist(file) {
		t.Errorf(" File was not created")
	}
}
