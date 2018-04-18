package watch

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

// MaxFileSize to open
var MaxFileSize = 20000000

// FileExist --
func FileExist(file string) bool {

	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}

// CMD --
type CMD struct {
	done         chan struct{}
	CmdWrite     func(string)
	CmdAllEvents func(string)
	TickCmd      func(string)
	File         string
	tickTime     time.Duration
	singleTime   time.Duration
	sync.Mutex
}

// OpenWatcher --
func OpenWatcher(cmdWrite func(string),
	cmdAllEvents func(string),
	tickcmd func(string), file string) *CMD {

	return &CMD{make(chan struct{}, 1),
		cmdWrite,
		cmdAllEvents,
		tickcmd, file, 600,
		1000, sync.Mutex{}}
}

// Stop --
func (cmd *CMD) Stop() {
	cmd.done <- struct{}{}
}

func BackOffFileCheck(file string) bool {
	for i := time.Duration(1); i < 30; i += 3 {
		if !FileExist(file) {
			time.Sleep(3 * i * time.Second)
		} else {
			return true
		}
	}
	return false
}

// Watcher --
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
					go cmd.CmdWrite(event.Name)
				}
				if event.Op&fsnotify.Rename == fsnotify.Rename {
					go cmd.CmdAllEvents("RENAME")

					if BackOffFileCheck(cmd.File) {
						cmd.Lock()
						watcher.Add(cmd.File)
						cmd.Unlock()

					}
				}
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					go cmd.CmdAllEvents("REMOVE")

					if BackOffFileCheck(cmd.File) {
						cmd.Lock()
						watcher.Add(cmd.File)
						cmd.Unlock()

					}
				}
				if event.Op&fsnotify.Chmod == fsnotify.Chmod {
					go cmd.CmdAllEvents("CHMOD")
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					go cmd.CmdAllEvents("CREATE")
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
	count          int
	file           string
	n              int
	ret            int64
	f              *os.File
	tickLast       time.Time
	b              []byte
	removeOrRename bool
}

// NewMC --
func NewMC(file string) *MC {
	return &MC{file: file}
}

// Inc --
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

// Count --
func (m *MC) TickUpdate() {
	m.Lock()
	defer m.Unlock()
	m.tickLast = time.Now()
}

func (m *MC) GetTick() time.Time {
	m.Lock()
	defer m.Unlock()
	return m.tickLast
}

// File --
func (m *MC) File() string {
	m.Lock()
	defer m.Unlock()
	return m.file
}

// SetFile --
func (m *MC) SetFile(file string) {
	m.Lock()
	defer m.Unlock()
	m.file = file
}

// GetB --
func (m *MC) GetB() []byte {
	m.Lock()
	defer m.Unlock()
	return m.b

}

// StatusRemoveRename --
func (m *MC) StatusRemoveRename() bool {
	m.Lock()
	defer m.Unlock()
	return m.removeOrRename
}

// RemoveRename --
func (m *MC) RemoveRename(event string) {
	m.Lock()
	defer m.Unlock()

	if event == "REMOVE" {
		m.removeOrRename = true
	}
	if event == "RENAME" {
		m.removeOrRename = true
	}
}

// ResetRemoveRename --
func (m *MC) ResetRemoveRename() {
	m.Lock()
	defer m.Unlock()
	m.removeOrRename = false
}

// Read --
func (m *MC) Read() {
	m.Lock()
	defer m.Unlock()

	if m.n == 0 {
		f, err := os.OpenFile(m.file, os.O_RDONLY, 0600)
		if err != nil {
			log.Println("error opening file", err)
		}
		m.f = f
	}

	b := make([]byte, MaxFileSize)

	//m.f.Seek(m.ret,io.SeekStart)
	n, err := m.f.Read(b)
	if err != nil {
		log.Println("Error in MC Read", err)
		if err == io.EOF {
			return
		}
	} else {
		m.removeOrRename = false
	}

	//m.ret, err =  m.f.Seek(0,io.SeekCurrent)
	if err != nil {
		fmt.Println("Err in m.f.Seek", err)
	}
	m.n = n
	m.b = b[0:n]
}

// No locks on functions below...

// LogTest -- hold no locks on this one
func (m *MC) WriteEvent(event string) {
	m.Read()
	log.Println("(MC)YES!!", event, m.Inc(), string(m.GetB()))

}

func (m *MC) AllEvents(event string) {

	m.RemoveRename(event)
	log.Println("All Events", event)

}

func (m *MC) Tick(event string) {
	m.TickUpdate()
	//log.Println("Tick", event, m.GetTick())

}