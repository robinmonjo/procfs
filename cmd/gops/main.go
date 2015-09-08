package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/robinmonjo/procfs"
)

const version = "0.1"

func main() {
	app := cli.NewApp()
	app.Name = "gops"
	app.Version = fmt.Sprintf("v%s", version)
	app.Author = "Robin Monjo"
	app.Usage = "simple ps like tool"
	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "all, a", Usage: "display processes from all users"},
		cli.BoolFlag{Name: "socket, s", Usage: "display sockets (TCP, UDP) opened by processes"},
	}

	app.Action = func(c *cli.Context) {
		err := start(c)
		if err != nil {
			log.Fatal(err)
		}
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func start(c *cli.Context) error {

	uid := os.Getuid()

	var (
		sockets []*procfs.Socket
		err     error
	)

	if c.Bool("socket") {
		sockets, err = procfs.ReadNet()
		if err != nil {
			return err
		}
		sort.Sort(procfs.Sockets(sockets)) //sort output by inode for faster search
	}

	return procfs.WalkProcs(func(p *procfs.Proc) (bool, error) {

		st, err := p.Status()
		if err != nil {
			if os.IsNotExist(err) {
				return true, nil
			}
			return false, err
		}

		if !c.Bool("all") && st.Uid != uid {
			return true, nil
		}

		//get back user
		user, err := st.User()
		if err != nil {
			return false, err
		}

		//get back process name
		n := st.Name
		args, err := p.CmdLine()
		if err != nil {
			return false, err
		}
		if args != nil {
			n = strings.Join(args, " ")
		}

		//print basic infos
		fmt.Printf("%s %d %d %v", user.Username, p.Pid, st.PPid, n)

		if c.Bool("socket") {
			//get back port bound by the process
			if err := printSockets(p, sockets); err != nil {
				return false, err
			}
		}

		fmt.Printf("\n")
		return true, nil
	})
}

func printSockets(p *procfs.Proc, sockets []*procfs.Socket) error {
	fds, err := p.Fds()
	if err != nil {
		if !os.IsPermission(err) {
			return err
		}
	}

	inodes := []string{}

	for _, fd := range fds {
		inode := fd.SocketInode()
		if inode != "" {
			inodes = append(inodes, inode)
		}
	}

	str := []string{}
	for _, inode := range inodes {
		if s := procfs.Sockets(sockets).Find(inode); s != nil {
			str = append(str, fmt.Sprintf("%s %v %s", s.Protocol, s.LocalIP, s.LocalPort))
		}
	}
	fmt.Printf(" %s", strings.Join(str, ", "))
	return nil
}
