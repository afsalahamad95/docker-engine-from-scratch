package main

import (
	"log"
	"os"
	"syscall"
)

func main() {
	proc, err := os.StartProcess("/bin/bash", []string{"bash", "-c", "ps ax"}, &os.ProcAttr{
		Files: []*os.File{
			os.Stdin,
			os.Stdout,
			os.Stderr,
		},
		Sys: &syscall.SysProcAttr{
			// process-level(NEWPID) and mount-level isolation(NEWNS)
			Cloneflags: syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		},
	})
	if err != nil {
		log.Println("error starting the process, err=", err)
	}
	log.Println("process started with pid:", proc.Pid)
}
