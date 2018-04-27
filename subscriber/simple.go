package subscriber

import (
	"context"
	"fmt"
	. "github.com/mchirico/firewall/include"
	"github.com/mchirico/firewall/set"
	"log"
	"os/exec"
	"sync"
	"time"
)

// Cmd to execute
func Cmd(cmd string) {

	ctx, cancel := context.WithTimeout(context.Background(), 10000*time.Millisecond)
	defer cancel()

	if err := exec.CommandContext(ctx, "sh", "-c", cmd).Run(); err != nil {
		log.Printf("Command failed")
	}

}

// Example of Command Structure (or maybe broker)
type CmdS struct {
	sync.Mutex
	cmd        string
	status     map[int]bool
	createS    *set.Set
	whiteListS *set.Set
	stageCmd   string
	logSet     string
	exeEndS    string
	tickMsg    string
}

func CreateCmdS(stageCmd string) *CmdS {
	c := &CmdS{}
	c.createS = set.CreateS()
	c.whiteListS = set.CreateS()
	c.status = map[int]bool{}
	c.stageCmd = stageCmd

	return c
}

// SetWhiteListSet -- this makes a copy
func (cmdS *CmdS) SetWhiteListSet(wS *set.Set) *CmdS {
	cmdS.Lock()
	defer cmdS.Unlock()

	cmdS.whiteListS = wS.Copy()
	cmdS.createS = cmdS.createS.Difference(cmdS.whiteListS)

	return cmdS

}

// CopySet -- this makes a copy
func (cmdS *CmdS) SetSet(sS *set.Set) *CmdS {
	cmdS.Lock()
	defer cmdS.Unlock()

	cmdS.createS = sS.Copy()
	return cmdS

}

// CopySet -- this makes a copy
func (cmdS *CmdS) GopySet() *set.Set {
	cmdS.Lock()
	defer cmdS.Unlock()

	return cmdS.createS.Copy()

}

func (cmdS *CmdS) Build(i int, iprec IpRec) {
	cmdS.Lock()
	defer cmdS.Unlock()

	tmpSet := set.CreateS()
	tmpSet.Add(set.CreateIpRec(iprec.IP, iprec.Ports))
	tmpSet = tmpSet.Difference(cmdS.createS)

	s := ""
	for k, v := range tmpSet.Values() {
		s += fmt.Sprintf(cmdS.stageCmd,
			k, v)
	}
	cmdS.createS = cmdS.createS.Union(tmpSet)

	cmdS.cmd = s
	cmdS.status[i] = false

}

func (cmdS *CmdS) Exe(i int) {
	cmdS.Lock()
	defer cmdS.Unlock()
	Cmd(cmdS.cmd)

	cmdS.status[i] = true
}

func (cmdS *CmdS) SetWriteLog(file string) {
	cmdS.Lock()
	defer cmdS.Unlock()
	cmdS.logSet = file

}

func (cmdS *CmdS) WriteLog() {
	cmdS.Lock()
	defer cmdS.Unlock()
	if cmdS.logSet != "" {
		cmdS.createS.WriteToFile(cmdS.logSet)
	}
}

func (cmdS *CmdS) LoadFromFile() {
	cmdS.Lock()
	defer cmdS.Unlock()
	if cmdS.logSet != "" {
		cmdS.createS.LoadFromFile(cmdS.logSet)
	}
}

func (cmdS *CmdS) Values() map[string][]int {
	cmdS.Lock()
	defer cmdS.Unlock()

	return cmdS.createS.Values()

}

func (cmdS *CmdS) ExeEnd(s string) {
	cmdS.Lock()
	defer cmdS.Unlock()
	cmdS.exeEndS = s

	// TODO: need better way
	if cmdS.logSet != "" {
		cmdS.createS.WriteToFile(cmdS.logSet)
	}
}

func (cmdS *CmdS) Tick(s string) {
	cmdS.Lock()
	defer cmdS.Unlock()
	cmdS.tickMsg = s

	log.Printf("subscriber Tick: %v\n", s)

}
