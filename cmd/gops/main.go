package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/robinmonjo/procfs"
)

func main() {
	sockets, err := procfs.Net()
	if err != nil {
		panic(err)
	}
	sort.Sort(procfs.Sockets(sockets)) //sort output by inode for faster search

	err = procfs.WalkProcs(func(p *procfs.Proc) (bool, error) {
		status, err := p.Status()
		if err != nil {
			if os.IsNotExist(err) {
				return true, nil
			}
			return false, err
		}

		fds, err := p.Fds()
		if err != nil {
			if !os.IsPermission(err) {
				return false, err
			}
		}

		inodes := []string{}

		for _, fd := range fds {
			inode := fd.SocketInode()
			if inode != "" {
				inodes = append(inodes, inode)
			}
		}

		ports := []string{}
		for _, inode := range inodes {
			if s := procfs.Sockets(sockets).Find(inode); s != nil {
				ports = append(ports, s.BindPort)
			}
		}

		user, err := status.User()
		if err != nil {
			panic(err)
		}

		fmt.Printf("%s %d %d %s %v\n", user.Username, p.Pid, status.PPid, status.Name, ports)

		return true, nil
	})

	if err != nil {
		panic(err)
	}
}
