package procfs

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const (
	localAddrColumn = 1
	inodeColumn     = 9
)

var (
	protocols = []string{"tcp", "tcp6", "udp", "udp6"}
)

type Socket struct {
	Protocol string
	BindPort string
	Inode    string
}

func Net() ([]*Socket, error) {
	var (
		sockets = []*Socket{}
		err     error
		mutex   = &sync.Mutex{}
		wg      sync.WaitGroup
	)

	wg.Add(len(protocols))
	for _, proto := range protocols {
		go func(p string) {
			s, e := parseNetfile(p)
			mutex.Lock()
			if e != nil {
				err = e
			} else {
				sockets = append(sockets, s...)
			}
			mutex.Unlock()
			wg.Done()
		}(proto)
	}

	wg.Wait()
	return sockets, err
}

func parseNetfile(protocol string) ([]*Socket, error) {
	f, err := os.Open(filepath.Join(Mountpoint, "net", protocol))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sockets := []*Socket{}

	scanner := bufio.NewScanner(f)

	flushedHeader := false
	for scanner.Scan() {
		line := scanner.Text()
		if !flushedHeader {
			flushedHeader = true
			continue
		}
		sock := processLine(line, protocol)
		sockets = append(sockets, sock)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return sockets, nil
}

func processLine(line, protocol string) *Socket {
	columns := strings.Fields(line)

	s := &Socket{
		Protocol: protocol,
	}

	for i, c := range columns {
		switch i {
		case localAddrColumn: // "0.0.0.0:9999"
			hexPort := strings.Split(c, ":")[1]
			p, _ := strconv.ParseInt(hexPort, 16, 32)
			s.BindPort = strconv.Itoa(int(p))
		case inodeColumn:
			s.Inode = c
		}
	}

	return s
}

//sort warppers
type Sockets []*Socket

func (sockets Sockets) Len() int           { return len(sockets) }
func (sockets Sockets) Swap(i, j int)      { sockets[i], sockets[j] = sockets[j], sockets[i] }
func (sockets Sockets) Less(i, j int) bool { return sockets[i].Inode < sockets[j].Inode }

func (sockets Sockets) Find(inode string) *Socket {
	i := sort.Search(len(sockets), func(i int) bool {
		return sockets[i].Inode >= inode
	})
	if i < len(sockets) && sockets[i].Inode == inode {
		return sockets[i]
	}
	return nil
}