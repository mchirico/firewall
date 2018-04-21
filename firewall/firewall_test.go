package main

import (
	"github.com/mchirico/firewall/fixtures"
	"os"
	"testing"
)

func TestMain(m *testing.M) {

	if fixtures.StageCheck() {
		return
	}
	fixtures.CreateActiveStageDirs()
	// call flag.Parse() here if TestMain uses flags
	os.Exit(m.Run())

}

func TestFileExist(t *testing.T) {

}

func TestFileExist2(t *testing.T) {

}
