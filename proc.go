package procfs

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

const (
	statusName   = "Name"
	statusPPid   = "PPid"
	statusState  = "State"
	statusUid    = "Uid"
	statusSigBlk = "SigBlk"
	statusSigIgn = "SigIgn"
	statusSigCgt = "SigCgt"

	socketLinkRegex = `socket:\[(\d+)\]`
)

type Proc struct {
	Pid int
}

//store information about a process, found in /proc/pid/status
type ProcStatus struct {
	Name   string
	PPid   int
	State  string
	Uid    string
	SigBlk []syscall.Signal
	SigIgn []syscall.Signal
	SigCgt []syscall.Signal
}

//file descriptor are symlinks
type Fd struct {
	Source string
	Target string
}

//return ProcStatus of the process
func (p *Proc) Status() (*ProcStatus, error) {
	f, err := os.Open(fmt.Sprintf("%s/%d/status", Mountpoint, p.Pid))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	s := &ProcStatus{}

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		records := strings.SplitN(scanner.Text(), ":", 2)
		key, value := records[0], strings.TrimSpace(records[1])

		switch key {
		case statusName:
			s.Name = value
		case statusPPid:
			s.PPid, _ = strconv.Atoi(value)
		case statusState:
			s.State = value
		case statusUid:
			s.Uid = strings.Fields(value)[0]
		case statusSigBlk:
			s.SigBlk = decodeSigMask(value)
		case statusSigIgn:
			s.SigIgn = decodeSigMask(value)
		case statusSigCgt:
			s.SigCgt = decodeSigMask(value)
		}
	}

	return s, scanner.Err()
}

//return all process's direct children
func (p *Proc) Children() ([]*Proc, error) {
	children := []*Proc{}
	err := WalkProcs(func(process *Proc) (bool, error) {
		if process.Pid == p.Pid { //myself
			return true, nil
		}
		status, err := process.Status()
		if err != nil {
			return false, err
		}

		if status.PPid == p.Pid {
			children = append(children, process)
		}
		return true, nil
	})
	if err != nil {
		return nil, err
	}
	return children, nil
}

// return process's descendants (children, grand children ...)
func (p *Proc) Descendants() ([]*Proc, error) {
	descendants := []*Proc{p}
	cursor := 0

	for {
		if cursor >= len(descendants) {
			break
		}

		cp := descendants[cursor]
		children, err := cp.Children()
		if err != nil {
			return nil, err
		}
		descendants = append(descendants, children...)
		cursor++
	}

	return descendants[1:], nil //remove self from the descendants
}

func (p *Proc) Fds() ([]*Fd, error) {
	d, err := os.Open(fmt.Sprintf("%s/%d/fd", Mountpoint, p.Pid))
	if err != nil {
		return nil, err
	}
	defer d.Close()

	names, err := d.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	fds := []*Fd{}
	for _, name := range names {
		targ, err := os.Readlink(filepath.Join(d.Name(), name))
		if err != nil {
			return nil, err
		}

		fds = append(fds, &Fd{
			Source: name,
			Target: targ,
		})
	}
	return fds, nil
}

func (status *ProcStatus) User() (*user.User, error) {
	return user.LookupId(status.Uid)
}

func (fd *Fd) SocketInode() string {
	matches := regexp.MustCompile(socketLinkRegex).FindStringSubmatch(filepath.Base(fd.Target))
	if matches == nil {
		return ""
	}
	return matches[1]
}

//implementation of signal mask decoding
//ref: http://jeff66ruan.github.io/blog/2014/03/31/sigpnd-sigblk-sigign-sigcgt-in-proc-status-file/
func decodeSigMask(maskStr string) []syscall.Signal {
	b, _ := hex.DecodeString(maskStr)
	//interested in the 32 right bits of the mask
	mask := int32(b[4])<<24 | int32(b[5])<<16 | int32(b[6])<<8 | int32(b[7])

	var signals []syscall.Signal

	for i := 0; i < 32; i++ {
		submask := int32(1 << uint(i))
		if mask&submask > 0 {
			signals = append(signals, syscall.Signal(i+1))
		}
	}

	return signals
}
