package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sync"
	"syscall"
	"testing"
	"time"
)

func TestServersStart(t *testing.T) {
	os.Args = []string{Application, "-control-addr", ":0", "-mock-addr", ":0"}
	logsReader, logWriter := io.Pipe()
	defer logWriter.Close()

	mockAddrRegexp := regexp.MustCompile(`\[mock\] Server is listening on ([a-f0-9\.:\[\]]+:\d+)`)
	mockAddrListen := ""
	controlAddrRegexp := regexp.MustCompile(`\[control\] Server is listening on ([a-f0-9\.:\[\]]+:\d+)`)
	controlAddrListen := ""
	go func() {
		reader := bufio.NewReader(logsReader)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					return
				}
				t.Errorf("reader.ReadString: %v", err)
			}
			t.Logf("# %v", line)
			matches := mockAddrRegexp.FindSubmatch([]byte(line))
			if len(matches) > 0 {
				mockAddrListen = string(matches[len(matches)-1])
			}
			matches = controlAddrRegexp.FindSubmatch([]byte(line))
			if len(matches) > 0 {
				controlAddrListen = string(matches[len(matches)-1])
			}
		}
	}()
	logOut = logWriter
	quitSignal = syscall.SIGUSR1
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		main()
	}()

	for {
		if mockAddrListen != "" && controlAddrListen != "" {
			break
		}
		time.Sleep(time.Millisecond)
	}

	for {
		if res, err := http.Get(fmt.Sprintf("http://%s/path", mockAddrListen)); err == nil {
			t.Log(res)
			break
		}
		time.Sleep(time.Millisecond)
	}

	for {
		if res, err := http.Get(fmt.Sprintf("http://%s/path", controlAddrListen)); err == nil {
			t.Log(res)
			break
		}
		time.Sleep(time.Millisecond)
	}

	if err := syscall.Kill(syscall.Getpid(), quitSignal.(syscall.Signal)); err != nil {
		t.Fatalf("syscall.Kill: %v", err)
	}
	wg.Wait()
}
