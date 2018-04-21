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
	"time"
)

var UnEncryptedFiles = "../fixtures/stage/access.log.stage"
var UseTemp = false

var stageFiles = "%s/src/github.com/mchirico/firewall"

type DirStruct struct {
	name string
	file string
}

var Dirs = []DirStruct{
	{name: "var", file: "/fixtures/var"},
	{name: "log", file: "/fixtures/var/log"},
	{name: "etc", file: "/fixtures/var/etc"},
}

// StageCheck --
func StageCheck() bool {
	if !watch.FileExist(UnEncryptedFiles) {
		log.Printf("These files must be unEncrypted " +
			"for further test.")
		return true
	}
	return false
}

func myLog(s string) {
	f, err := os.OpenFile("/tmp/gostuff", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Println("Can't open file /tmp/gostuff")
	}
	msg := fmt.Sprintf("%s:%s\n", time.Now(), s)
	f.WriteString(msg)
	f.Close()
}

// GetBaseDir --
func GetBaseDir() string {
	if !UseTemp {
		path := os.Getenv("GOPATH")
		path = fmt.Sprintf(stageFiles, path)
		return path
	}
	return "/tmp"
}

func getFile(s string) string {
	for _, d := range Dirs {
		if d.name == s {
			return d.file
		}
	}
	return ""
}

func UpdateConfigSettings() utils.Config {
	_, dst := CreateConfig()
	c := utils.ReadConfig(dst)
	UpdateConfigLogs(c, dst)

	return utils.ReadConfig(dst)
}

func UpdateConfigLogs(c utils.Config, file string) {

	c.SearchLogs[0].Log =
		GetBaseDir() + getFile("var") + "/auth.log"

	c.SearchLogs[1].Log =
		GetBaseDir() + getFile("var") + "/mail.log"

	c.OutputLog = GetBaseDir() + getFile("var") +
		"/firewall.json"

	c.StatusLog = GetBaseDir() + getFile("var") +
		"/firewallStatus.log"

	cJson, _ := json.Marshal(c)
	err := ioutil.WriteFile(file, cJson, 0644)
	if err != nil {
		log.Println(err)
	}

}

// CreateConfig --
func CreateConfig() (string, string) {

	CreateActiveStageDirs()

	dst := GetBaseDir() + "/fixtures/var/etc/config.cfg"
	configContents := ""
	src := GetBaseDir() + "/fixtures/stage/stage.config.json"

	_, err := os.Stat(dst)

	if err != nil {
		f, _ := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY, 0600)
		defer f.Close()
		s, _ := os.OpenFile(src, os.O_CREATE|os.O_RDONLY, 0600)
		defer s.Close()

		b := make([]byte, 500000)
		n, err := s.Read(b)
		if err != nil {
			log.Println("can't read source config")
		}

		f.Write(b[0:n])
		f.Sync()
		configContents = string(b[0:n])
	} else {
		log.Println("Config exists.... will not overwrite")
	}
	return configContents, dst
}

// CreateActiveStageDirs --
func CreateActiveStageDirs() {

	dirs := make([]DirStruct, len(Dirs))
	parent := GetBaseDir()

	for i, d := range Dirs {
		dirs[i].file = parent + d.file

	}

	for _, dir := range dirs {
		err := os.Mkdir(dir.file, os.ModePerm)
		if err != nil {
			log.Println(err.Error())
		}
	}

}

// CreateActiveStageDirs --
func RemoveActiveStageDirs() {

	var Dirs = []DirStruct{
		{name: "var", file: "/fixtures/var"},
		{name: "log", file: "/fixtures/var/log"},
		{name: "etc", file: "/fixtures/var/etc"},
	}

	dirs := make([]DirStruct, len(Dirs))
	parent := GetBaseDir()

	for i, d := range Dirs {
		dirs[i].file = parent + d.file

	}

	for _, dir := range dirs {

		// Safety Check
		sIndex := strings.LastIndex(dir.file, "/firewall/fixtures/")
		if sIndex < 10 {
			return
		}
		if watch.FileExist(dir.file) {
			err := os.RemoveAll(dir.file)
			if err != nil {
				log.Println(err)
			}
		}

	}

}
