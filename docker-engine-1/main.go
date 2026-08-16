package main

import (
	"log"
	"os"
	"syscall"
)

// this starts a go container within the docker container where we run this code.
// there is no isolation at this point.
// there is a parent-child relationship b/w the docker container and this 'so-called' go container.
func main() {
	// arg[0] is name of the binary to fork.
	// the args[1] will have the command to be executed with exec.
	// we bind the process's(go code) i/o streams with the host(docker container).
	proc, err := os.StartProcess("/bin/bash", []string{"bash", "-c", "ps ax"}, &os.ProcAttr{
		Files: []*os.File{
			os.Stdin,
			os.Stdout,
			os.Stderr,
		},
	})
	if err != nil {
		log.Println("error starting the process, err=", err)
	}
	log.Println("Process started with pid=", proc.Pid)
	err = proc.Signal(syscall.SIGTERM)
	if err != nil {
		log.Println("error sending signal to process, err=", err)
	}
	log.Println("process shut down")
}
