package watch

import (
	"testing"
	"os"
	"time"
	"fmt"
	"log"
)



func testCheckLog(file string)  {
	//log.Println("got file: ",file)

}


func TestBackground(t *testing.T) {

	file := "../fixtures/tempfoo"
	os.Remove(file)

	f,_ :=os.OpenFile(file,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,0600)


	m := NewMC(file)
	cmd := OpenWatcher(m.LogTest,testCheckLog,file)
	cmd.Watcher()

	if m.Count() != 0 {
		t.Errorf("Count should be zero")
	}

	time.Sleep(1 * time.Second)

	expectedString := "  Write 0"
	f.WriteString(expectedString)
	f.Sync()
	time.Sleep(3 * time.Second)

	log.Println("m.GetB()",string(m.GetB()))
	if string(m.GetB()) != expectedString {
		t.Errorf("Error b %v expected %v",
			m.GetB(),expectedString)
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
		t.Errorf("Count should be two: %d",m.Count())
	}

	log.Println("m.GetB()",string(m.GetB()))
	if string(m.GetB()) != expectedString {
		t.Errorf("Error b %v expected %v",
			m.GetB(),expectedString)
	}

	time.Sleep(4 * time.Second)
	cmd.Stop()


	time.Sleep(3 * time.Second)
	os.Remove(file)

}




func TestFileExist(t *testing.T) {
	file := "../fixtures/junktest"
	os.Remove(file)

	if FileExist(file) {
		t.Errorf("file should not exist")
	}

	f, err:=os.OpenFile(file,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,0600)
	if err != nil {
		t.Errorf("Can't create test file")
	}
	defer f.Close()
	f.WriteString("test")

	if !FileExist(file) {
		t.Errorf("file should exist")
	}

}
