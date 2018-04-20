package fixtures

import (
	"fmt"
	"log"
	"os"
	"time"
)

var UnEncryptedFiles = "../fixtures/stage/access.log.stage"
var UseTemp = false

var stageFiles = "%s/src/github.com/mchirico/firewall"

var Dirs = []string{"/fixtures/var",
	"/fixtures/var/log",
	"/fixtures/var/etc",
	"/fixtures/etc"}

func FileExist(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}

func StageCheck() bool {
	if !FileExist(UnEncryptedFiles) {
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

func GetBaseDir() string {
	if !UseTemp {
		path := os.Getenv("GOPATH")
		path = fmt.Sprintf(stageFiles, path)
		return path
	}
	return "/tmp"
}

func CreateConfig() (string, string) {

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
	}
	return configContents, dst
}

func CreateActiveStageDirs() {

	parent := GetBaseDir()
	dirs := make([]string, len(Dirs))

	for i, d := range Dirs {
		dirs[i] = parent + d

	}

	for _, dir := range dirs {
		err := os.Mkdir(dir, os.ModePerm)
		if err != nil {
			log.Println(err.Error())
		}
	}

}
