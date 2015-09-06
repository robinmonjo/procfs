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

type Process struct {
	Pid int
}

//store information about a process, found in /proc/pid/status
type ProcessStatus struct {
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

//return ProcessStatus of the process
func (p *Process) Status() (*ProcessStatus, error) {
	f, err := os.Open(fmt.Sprintf("%s/%d/status", Mountpoint, p.Pid))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	s := &ProcessStatus{}

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		record := strings.SplitN(line, ":", 2)
		switch record[0] {
		case "Name":
			s.Name = strings.TrimSpace(record[1])
		case "PPid":
			s.PPid, err = strconv.Atoi(strings.TrimSpace(record[1]))
		case "State":
			s.State = strings.TrimSpace(record[1])
		case "Uid":
			s.Uid = strings.Fields(record[1])[0]
		case "SigBlk":
			s.SigBlk, err = decodeSigMask(record[1])
		case "SigIgn":
			s.SigIgn, err = decodeSigMask(record[1])
		case "SigCgt":
			s.SigCgt, err = decodeSigMask(record[1])
		}
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

//return all process's direct children
func (p *Process) Children() ([]*Process, error) {
	children := []*Process{}
	err := WalkProcesses(func(process *Process) (bool, error) {
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
func (p *Process) Descendants() ([]*Process, error) {
	descendants := []*Process{p}
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

func (p *Process) Fds() ([]*Fd, error) {
	d, err := os.Open(fmt.Sprintf("%s/%d/fd", Mountpoint, p.Pid))
	if err != nil {
		return nil, err
	}
	defer d.Close()
	fis, err := d.Readdir(-1)
	if err != nil {
		return nil, err
	}
	fds := []*Fd{}
	for _, fi := range fis {
		targ, err := os.Readlink(filepath.Join(d.Name(), fi.Name()))
		if err != nil {
			return nil, err
		}

		fds = append(fds, &Fd{
			Source: fi.Name(),
			Target: targ,
		})
	}
	return fds, nil
}

func (status *ProcessStatus) User() (*user.User, error) {
	return user.LookupId(status.Uid)
}

func (fd *Fd) SocketInode() string {
	re := regexp.MustCompile(`socket:\[(\d+)\]`)
	matches := re.FindStringSubmatch(filepath.Base(fd.Target))
	if matches == nil {
		return ""
	}
	return matches[1]
}

//implementation of signal mask decoding
//ref: http://jeff66ruan.github.io/blog/2014/03/31/sigpnd-sigblk-sigign-sigcgt-in-proc-status-file/
func decodeSigMask(maskStr string) ([]syscall.Signal, error) {
	b, err := hex.DecodeString(strings.TrimSpace(maskStr))
	if err != nil {
		return nil, err
	}
	//interested in the 32 right bits of the mask
	mask := int32(b[4])<<24 | int32(b[5])<<16 | int32(b[6])<<8 | int32(b[7])

	var signals []syscall.Signal

	for i := 0; i < 32; i++ {
		submask := int32(1 << uint(i))
		if mask&submask > 0 {
			signals = append(signals, syscall.Signal(i+1))
		}
	}

	return signals, nil
}
