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
	cmd     string
	status  map[int]bool
	createS *set.Set
}

func CreateCmdS() *CmdS {
	c := &CmdS{}
	c.createS = set.CreateS()
	c.status = map[int]bool{}

	return c
}

func (cmdS *CmdS) Build(i int, iprec IpRec) {
	cmdS.Lock()
	defer cmdS.Unlock()

	tmpSet := set.CreateS()
	tmpSet.Add(set.CreateIpRec(iprec.IP, iprec.Ports))
	tmpSet = tmpSet.Difference(cmdS.createS)

	s := ""
	for k, v := range tmpSet.Values() {
		s += fmt.Sprintf("echo  %v:  %v >>/tmp/firewall.cmd\n",
			k, v)
	}
	cmdS.createS.Union(tmpSet)

	cmdS.cmd = s
	cmdS.status[i] = false
}

func (cmdS *CmdS) Exe(i int) {
	cmdS.Lock()
	defer cmdS.Unlock()
	Cmd(cmdS.cmd)
	cmdS.status[i] = true
}
