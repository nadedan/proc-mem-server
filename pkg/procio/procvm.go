package procio

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

type ProcVm struct {
	pid   int
	addrs procMapAddrs
	size  uint64
}

func NewProcVm(binName string) (*ProcVm, error) {
	p := &ProcVm{}

	output, err := exec.Command("pidof", binName).Output()
	if err != nil {
		return nil, fmt.Errorf("could not get the pid of %s, perhaps it is not running: %w", binName, err)
	}
	pidStr := string(output)
	pidStr = strings.TrimRight(pidStr, "\n")

	pid, err := strconv.ParseUint(pidStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("pidof gave us '%s', and that cannot be parsed as an int: %w", pidStr, err)
	}
	p.pid = int(pid)

	mapPath := fmt.Sprintf("/proc/%d/maps", p.pid)
	fProcMap, err := os.Open(mapPath)
	if err != nil {
		return nil, fmt.Errorf("could not open %s: %w", mapPath, err)
	}
	defer fProcMap.Close()

	p.addrs, err = readProcMap(fProcMap, binName)
	if err != nil {
		return nil, fmt.Errorf("could not read proc map: %w", err)
	}
	p.size = p.addrs.rwEnd - p.addrs.rwStart

	fmt.Printf("addrs. base: 0x%X rwStart: 0x%X rwEnd: 0x%0X\n", p.addrs.base, p.addrs.rwStart, p.addrs.rwEnd)

	return p, nil
}

func (p ProcVm) Read(addr uint64, buff []byte) (n int, err error) {

	addr += uint64(p.addrs.base)
	//addr -= uint64(p.addrs.rwStart)
	//if len(buff) > int(p.size) {
	//	panic(fmt.Sprintf("cannot get %d bytes from an mmap of size %d", len(buff), p.size))
	//}

	localIov := []unix.Iovec{
		{Base: &buff[0], Len: uint64(len(buff))},
	}
	remoteIov := []unix.RemoteIovec{
		{Base: uintptr(addr), Len: len(buff)},
	}

	n, err = unix.ProcessVMReadv(p.pid, localIov, remoteIov, 0)
	if err != nil {
		return 0, fmt.Errorf("%T.Read: could not ProcessVMReadv: %w", p, err)
	}
	if n != int(len(buff)) {
		return 0, fmt.Errorf("%T.Read: short read: got %d, want %d", p, n, len(buff))
	}

	return n, nil
}

func (p ProcVm) Write(addr uint64, buff []byte) (n int, err error) {

	localIov := []unix.Iovec{
		{Base: &buff[0], Len: uint64(len(buff))},
	}
	remoteIov := []unix.RemoteIovec{
		{Base: uintptr(addr), Len: len(buff)},
	}

	n, err = unix.ProcessVMWritev(p.pid, localIov, remoteIov, 0)
	if err != nil {
		return 0, fmt.Errorf("%T.Write: could not ProcessVMWritev: %w", p, err)
	}
	if n != int(len(buff)) {
		return 0, fmt.Errorf("%T.Write: short write: got %d, want %d", p, n, len(buff))
	}

	return n, nil
}
