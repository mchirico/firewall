package utils

import (
	"testing"
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
