// Package procfs provides primitives for interacting with the linux proc
// virtual file system
package procfs

import (
	"os"
	"strconv"
)

// DefaultMountpoint define the default mount point of the proc file system
const DefaultMountpoint = "/proc"

// MountPoint path of the proc file system mount point.
// Default to DefaultMountPoint
var Mountpoint = DefaultMountpoint

// CountRunningProcesses return the number of running processes or an error
func CountRunningProcesses() (int, error) {
	cpt := 0
	err := WalkProcesses(func(process *Process) (bool, error) {
		cpt++
		return true, nil
	})
	return cpt, err
}

// WalkFunc WalkFunc is the type of the function called for each process visited by WalkProcesses.
// The process argument contains the current process. If the function return false or an error,
// the WalkProcesses func stop, and reurn the eventual error
type WalkFunc func(process *Process) (bool, error)

// WalkProcesses walks all the processes and call walk on each process
func WalkProcesses(walk WalkFunc) error {
	d, err := os.Open(Mountpoint)
	if err != nil {
		return err
	}
	defer d.Close()
	fis, err := d.Readdir(-1)
	if err != nil {
		return err
	}

	for _, fi := range fis {
		if pid, err := strconv.Atoi(fi.Name()); err == nil {
			loop, err := walk(&Process{
				Pid: pid,
			})
			if err != nil {
				return err
			}
			if !loop {
				return nil
			}
		}
	}
	return nil
}
