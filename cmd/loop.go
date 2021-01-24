package cmd

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

const (
	// loop flags
	LoFlagsPartscan = 8

	LoNameSize = 64
	LoKeySize  = 32

	// Loop Filter types
	LoCryptNone      = 0
	LoCryptXor       = 1
	LoCryptDes       = 2
	LoCryptFish2     = 3
	LoCryptBlow      = 4
	LoCryptCast128   = 5
	LoCryptIdea      = 6
	LoCryptDummy     = 9
	LoCryptSkipjack  = 10
	LoCryptCryptoApi = 18
	MaxLoCrypt       = 20

	// IOCTL commands
	LoopSetFd        = 0x4C00
	LoopClrFd        = 0x4C01
	LoopSetStatus    = 0x4C02
	LoopGetStatus    = 0x4C03
	LoopSetStatus64  = 0x4C04
	LoopGetStatus64  = 0x4C05
	LoopChangeFd     = 0x4C06
	LoopSetCapacity  = 0x4C07
	LoopSetDirectIo  = 0x4C08
	LoopSetBlockSize = 0x4C09
	LoopConfigure    = 0x4C0A

	// Loop control commands
	LoopCtlAdd     = 0x4C80
	LoopCtlRemove  = 0x4C81
	LoopCtlGetFree = 0x4C82
)

type loopInfo struct {
	LoDevice         uint64
	LoInode          uint64
	LoRdevice        uint64
	LoOffset         uint64
	LoSizeLimit      uint64
	LoNumber         uint32
	LoEncryptType    uint32
	LoEncryptKeySize uint32
	LoFlags          uint32
	LoFileName       [LoNameSize]uint8
	LoCryptName      [LoNameSize]uint8
	LoCryptKey       [LoKeySize]uint8
	LoInit           [2]uint8
}

type loopDevice struct {
	*os.File
	info loopInfo
}

// Return next available loopback device's FD
func getFreeLoopbackDev() (uint64, error) {
	loopControl, err := os.Open("/dev/loop-control")
	if err != nil {
		return 0, err
	}
	fd, _, err := syscall.Syscall(syscall.SYS_IOCTL, loopControl.Fd(), LoopCtlGetFree, 0)
	if fd < 0 || err != nil {
		return uint64(fd), err
	}

	return uint64(fd), nil
}

// Attaches a file at path to the next available loopback device
func Attach(path string) (loopDevice, error) {
	backingFile, err := os.OpenFile(path, os.O_RDWR, 0660)
	if err != nil {
		return loopDevice{}, err
	}
	defer backingFile.Close()

	freeFd, err := getFreeLoopbackDev()
	if err != nil {
		return loopDevice{}, err
	}

	var dev loopDevice
	dev.File, err = os.Open(fmt.Sprintf("/dev/loop%d", freeFd))
	if err != nil {
		return loopDevice{}, err
	}

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, dev.File.Fd(), LoopSetFd, backingFile.Fd())
	if errno != 0 {
		return loopDevice{}, errors.New(fmt.Sprintf("Recieved errno %d", errno))
	}

	info := loopInfo{LoFlags: LoFlagsPartscan}

	_, _, errno = syscall.Syscall(syscall.SYS_IOCTL, dev.Fd(), LoopSetStatus64, uintptr(unsafe.Pointer(&info)))
	if errno != 0 {
		return loopDevice{}, errors.New(fmt.Sprintf("Recieved errno %d", errno))
	}

	dev.info = info

	return dev, nil
}
