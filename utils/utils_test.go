package utils

import (
	"fmt"
	"reflect"
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

func TestSets(t *testing.T) {
	s := CreateS()
	s.Add("a")
	s.Add("b")

	s2 := CreateS()
	s2.Add("c")
	s2.Add("d")

	a := s.Union(s2).Keys()
	b := []string{"a", "b", "c", "d"}
	if !reflect.DeepEqual(a, b) {

		t.Error("missing values")
	}

	a = s.Diff(s2).Keys()
	b = []string{"c", "d"}
	if !reflect.DeepEqual(a, b) {

		t.Error("missing values")
	}

}

func TestCreateIpRec(t *testing.T) {
	c := ReadConfig("../fixtures/config.json")
	fw := &Firewall{Config: c}
	fw.Read()
	fw.Parse()

	iprec := fw.CreateIpRec()

	if iprec[0].Ports[2] != 80 {
		t.Fatalf("Didn't get port")
	}

	fmt.Printf("iprec[9]: %v\n", iprec[9])

}

func TestWriteRecs(t *testing.T) {

	fw := CreateFirewall("../fixtures/config.json")
	fw.Read()
	fw.Parse()
	iprecs := fw.CreateIpRec()
	fw.WriteRecs(iprecs)

}

func TestReadRecs(t *testing.T) {
	c := ReadConfig("../fixtures/config.json")
	fw := &Firewall{Config: c}
	fw.Read()
	fw.Parse()
	iprecs := fw.CreateIpRec()
	fw.WriteRecs(iprecs)

	recs := fw.ReadRecs()
	fmt.Printf("here %v\n", recs[0].IP)

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

func TestLogging(t *testing.T) {

	str := SetLogging()
	c := ReadConfig("../fixtures/config.json")
	fw := &Firewall{Config: c}
	fw.Read()
	fw.Parse()
	iprecs := fw.CreateIpRec()
	fw.WriteRecs(iprecs)
	fmt.Printf("->%s<-", str.String())

}
