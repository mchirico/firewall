package watch

import (
	"github.com/fsnotify/fsnotify"
	"log"
	"os"
	"fmt"
	"time"
	"sync"
	"io"
)

// MaxFileSize to open
var MaxFileSize = 20000000

// FileExists --
func FileExist(file string) bool {

	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}

// CMD --
type CMD struct {
	done       chan struct{}
	Cmd        func(string) string
	TickCmd    func(string)
	File       string
	tickTime   time.Duration
	singleTime time.Duration
	mu         sync.Mutex
}

// OpenWatcher --
func OpenWatcher(cmd func(string) string, tickcmd func(string), file string) *CMD {
	return &CMD{make(chan struct{}, 1), cmd,
		tickcmd, file, 600,
		1000, sync.Mutex{}}
}

// Stop --
func (cmd *CMD) Stop() {
	cmd.done <- struct{}{}
}

// Watcher
func (cmd *CMD) Watcher() {
	tick := time.Tick(cmd.tickTime * time.Millisecond)
	single := time.After(cmd.singleTime * time.Millisecond)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		defer watcher.Close()
		for {
			select {
			case event := <-watcher.Events:
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
					go cmd.Cmd(event.Name)
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)

			case <-tick:
				//fmt.Println("tick.")
				go cmd.TickCmd(cmd.File)

			case <-cmd.done:
				fmt.Println("done")

				return

			case <-single:
				fmt.Println("single call")



			}

		}
	}()

	err = watcher.Add(cmd.File)
	if err != nil {
		log.Fatal(err)
	}

}

// MC --
type MC struct {
	sync.Mutex
	count int
	file  string
	n     int
	ret  int64
	f     *os.File
	b []byte
}

// NewMc --
func NewMC(file string) *MC {
	return &MC{file: file}
}

// Inc
func (m *MC) Inc() int {
	m.Lock()
	defer m.Unlock()
	m.count += 1
	return m.count
}

// Count --
func (m *MC) Count() int {
	m.Lock()
	defer m.Unlock()
	return m.count
}

// File --
func (m *MC) File() string {
	m.Lock()
	defer m.Unlock()
	return m.file
}

// SetFile --
func (m *MC) SetFile(file string)  {
	m.Lock()
	defer m.Unlock()
	m.file = file
}

// GetB --
func (m *MC)GetB() []byte {
	m.Lock()
	defer m.Unlock()
	return m.b

}

// Read --
func (m *MC) Read() {
	m.Lock()
	defer m.Unlock()


	if m.n == 0 {
		f, err := os.OpenFile(m.file, os.O_RDONLY, 0600)
		if err != nil {
			log.Println("error opening file",err)
		}
		m.f = f
	}

	b := make([]byte, MaxFileSize)


	//m.f.Seek(m.ret,io.SeekStart)
	n, err := m.f.Read(b)
	if err != nil {
		log.Println("Error in MC Read",err)
		if err == io.EOF {
			return
		}
	}

	//m.ret, err =  m.f.Seek(0,io.SeekCurrent)
	if err != nil {
		fmt.Println("Err in m.f.Seek",err)
	}
	m.n = n
	m.b = b[0:n]
}



func (m *MC) logTest(event string) string {
	m.Read()
	log.Println("(MC)YES!!", event, m.Inc(),string(m.GetB()))
	return event
}
