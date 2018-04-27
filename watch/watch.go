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

// FileExist -- general file checker
func FileExist(file string) bool {

	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}

// CMD -- format for config
type CMD struct {
	done         chan struct{}
	CmdWrite     func(string)
	CmdAllEvents func(string)
	TickCmd      func(string)
	File         string
	tickTime     time.Duration
	singleTime   time.Duration
	sync.Mutex
	restartCount int64
}

// OpenWatcher --
func OpenWatcher(cmdWrite func(string),
	cmdAllEvents func(string),
	tickcmd func(string), file string) *CMD {

	return &CMD{make(chan struct{}, 1),
		cmdWrite,
		cmdAllEvents,
		tickcmd, file, 600,
		1000, sync.Mutex{}, 0}
}

// Stop --
func (cmd *CMD) Stop() {
	cmd.done <- struct{}{}
}

// RestartCount --
func (cmd *CMD) RestartCount() int64 {
	cmd.Lock()
	defer cmd.Unlock()
	return cmd.restartCount
}

// BackOffFileCheck --
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
				//log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					//log.Println("modified file:", event.Name)
					go cmd.CmdWrite(event.Name)
				}
				if event.Op&fsnotify.Rename == fsnotify.Rename {
					go cmd.CmdAllEvents("RENAME")

					if BackOffFileCheck(cmd.File) {
						cmd.Lock()
						cmd.restartCount++
						watcher.Add(cmd.File)
						cmd.Unlock()

					}
				}
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					go cmd.CmdAllEvents("REMOVE")

					if BackOffFileCheck(cmd.File) {
						cmd.Lock()
						cmd.restartCount++
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
				if event.Op&fsnotify.Write == fsnotify.Write {
					go cmd.CmdAllEvents("WRITE")
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

type WatchFuncExec interface {
	WriteEvent(event string)
	AllEvents(event string)
	Tick(event string)
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
	writeLast      time.Time
	b              []byte
	removeOrRename bool
	events         map[string]time.Time
	allEvents      []EventMap
	Slave          WatchFuncExec
}

// EventMap --
type EventMap struct {
	event string
	time  time.Time
}

// MCError --
type MCError struct {
	When time.Time
	What string
}

// NewMC --
func NewMC(file string, slave WatchFuncExec) *MC {
	return &MC{file: file, events: map[string]time.Time{},
		allEvents: []EventMap{}, Slave: slave}
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

func (m *MC) FireSlaveWriteEvent(event string) {
	m.Lock()
	defer m.Unlock()
	go m.Slave.WriteEvent(event)
}

func (m *MC) FireSlaveTickEvent(event string) {
	m.Lock()
	defer m.Unlock()
	go m.Slave.Tick(event)
}

// StatusRemoveRename --
func (m *MC) StatusRemoveRename() bool {
	m.Lock()
	defer m.Unlock()
	return m.removeOrRename
}

// AddEvent --
func (m *MC) AddEvent(e string) {
	m.Lock()
	defer m.Unlock()
	if len(m.allEvents) > 9000 {
		m.allEvents = []EventMap{}
		m.events = map[string]time.Time{}
		m.events["EVENTS CLEARED"] = time.Now()
	}
	t := time.Now()
	m.events[e] = t
	lastEvent := EventMap{e, t}
	m.allEvents = append(m.allEvents, lastEvent)
}

func (e MCError) Error() string {
	return fmt.Sprintf("%v: %v", e.When, e.What)
}

func (m *MC) LastEvent() (EventMap, error) {
	m.Lock()
	defer m.Unlock()
	length := len(m.allEvents)
	if length > 0 {

		return m.allEvents[length-1], nil

	}
	err := MCError{
		When: time.Now(),
		What: "No events",
	}
	return EventMap{}, err
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
		// TODO: You get a lot of hits here
		// log.Println("Error in MC Read", err)
		if err == io.EOF {
			return
		}
	} else {
		m.removeOrRename = false
	}

	if n > 0 {
		m.writeLast = time.Now()
	}
	m.n = n
	m.b = b[0:n]
}

// No locks on functions below...

// WriteEvent -- hold no locks on this one
func (m *MC) WriteEvent(event string) {

	m.Read()
	m.Inc()
	m.FireSlaveWriteEvent(event)
	//log.Println("(MC)YES!!", event, string(m.GetB()))

}

// AllEvents --
func (m *MC) AllEvents(event string) {
	m.AddEvent(event)
	m.RemoveRename(event)
	//log.Println("All Events", event)

}

// Tick --
func (m *MC) Tick(event string) {
	m.TickUpdate()
	m.FireSlaveTickEvent(event)
	//log.Println("Tick", event, m.GetTick())

}
