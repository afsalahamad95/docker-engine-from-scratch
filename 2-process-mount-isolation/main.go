package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	// note that we mount a separate /proc directory, as ps will query this directory instead of asking the kernel for
	// process states.
	getParentProcesses()
	go func() {
		// isolated mount for /proc to trick ps commands to think that this process is the PID 1.
		// host will maintain a separate namespace map - host can modify/terminate the child process.
		// child cannot do anything to the host process whatsoever as it is isolated from one-side.

		// mount isolation - rprivate is used as the mount will use 'shared' mounts which will apply the mount operation to the host as well.
		// to avoid the corruption of host mount, we will make this private - so only the child is modified.
		proc, err := os.StartProcess("/bin/bash", []string{"bash", "-c", "mount --make-rprivate / && mount -t proc proc /proc && ps ax && sleep 5"}, &os.ProcAttr{
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
		proc.Wait() // clean up the resources clogged by the above process
	}()
	getParentProcesses()
}

// get running processes on the host system
func getParentProcesses() {
	cmd := exec.Command("ps", "ax")

	var out bytes.Buffer
	cmd.Stdout = &out // set the stdout stream for the above command

	err := cmd.Run()
	if err != nil {
		log.Println("error running process, err:", err)
	}

	log.Println("output:", string(out.String()))

}
