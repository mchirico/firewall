package utils

import (
	"fmt"
	"testing"
)

func TestRead(t *testing.T) {

	b, err, n := Read("../fixtures/auth.log")
	if err != nil {
		t.Errorf("Error reading file")
	}

	if n != 2667332 {
		t.Errorf("Can't read all bytes. "+
			"Expecing %d got %d", 2667332, n)
	}
	fmt.Printf("b:%d  %s\n", n, string(b[0:3]))
}

func TestParse(t *testing.T) {

	b, _, _ := Read("../fixtures/auth.log")
	m := Parse(b)
	//for k,v := range m {
	//	fmt.Printf("%v:%v\n",k,v)
	//}
	if m["163.172.119.161"] != 2 {
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

	if c.SearchLogs[1].Log != "/var/log/mail.log" {
		t.Errorf(("Error in search logs"))
	}
	if c.SearchLogs[1].Ports[0] != 25 {
		t.Errorf(("Error in search log port"))
	}
	if c.SearchLogs[1].Regex != ".*Invalid user.* " +
		"([0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]+)" {
		t.Errorf(("Error in search log regex"))
	}

}
