package include

// IpRec -- ip's to block
type IpRec struct {
	IP    string
	Count int
	Ports []int
}

// CmdSlave -- fires cmd
type CmdSlave interface {
	Build(int, IpRec)
	Exe(int)
	ExeEnd(string)
	Tick(string)
}
