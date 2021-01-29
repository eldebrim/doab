package cmd

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	testFileName = string("test.img")
)

// Make sure Attach() attaches a file to the next available loopback device. requires root permissions
func TestAttach(t *testing.T) {
	testFile, err := os.Create(testFileName)
	if err != nil {
		log.Fatal(err)
	}
	testFile.Close()
	defer os.Remove(testFileName)

	testPath, _ := filepath.Abs(testFile.Name())
	_, err = Attach(testPath)
	if err != nil {
		log.Fatal(err)
	}

	losetupOut, err := exec.Command("losetup", "--list").Output()
	if err != nil {
		log.Fatal(err)
	}
	defer exec.Command("losetup", "-D")

	require.Contains(t, string(losetupOut), testFileName)

}

// Attach file to next available loopback device with `losetup` then make sure Detach() can detach it
func TestDetach(t *testing.T) {
	testFile, err := os.Create(testFileName)
	if err != nil {
		log.Fatal(err)
	}
	testFile.Close()
	defer os.Remove(testFileName)

	losetupOut, err := exec.Command("losetup", "-f", testFileName, "--show").Output()
	if err != nil {
		log.Fatal(err)
	}
	loPath := strings.TrimRight(string(losetupOut), "\n")

	loDev := loopDevice{path: loPath}
	err = Detach(loDev)
	if err != nil {
		log.Fatal(err)
	}
	// wait for a bit so detach operation completes
	time.Sleep(4 * time.Second)

	losetupOut, err = exec.Command("losetup", "--list").Output()
	if err != nil {
		log.Fatal(err)
	}
	require.NotContains(t, string(losetupOut), loPath)

}
