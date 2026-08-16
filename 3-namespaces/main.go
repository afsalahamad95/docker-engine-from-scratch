package main

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func run() {
	// run will simply clone itself after requesting new namespaces from the kernel
	// like how docker engine runs, it will simply spin up containers(which are just more containers of the same type)

	// spawn a child process of the same instance - but under the new namespace
	// since a running process cannot change the namespace, spawn a child process under the new namespace.
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)

	// link current shell to process
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWNS | syscall.CLONE_NEWPID,
	}

	err := cmd.Run()
	if err != nil {
		log.Println("error running command, err=", err)
	}
}

func child() {
	hostname := "isolated-hostname"
	err := syscall.Sethostname([]byte(hostname))
	if err != nil {
		log.Println("error setting hostname, err=", err)
	}
	err = syscall.Chroot("alpine-rootfs")
	if err != nil {
		log.Println("error setting root dir, err=", err)
	}
	err = os.Chdir("/")
	if err != nil {
		log.Println("error moving to root dir, err=", err)
	}
	command := exec.Command("/bin/sh", "-c", "ls -a && (mount --make-rprivate / || true); (mount -t proc proc /proc || true)")
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	_ = command.Run() // ignore errors, these may fail in container environments

	targetCmd := os.Args[2]
	targetCmdArgs := os.Args[3:]
	shellCmd := targetCmd
	if len(targetCmdArgs) > 0 {
		shellCmd = targetCmd + " " + strings.Join(targetCmdArgs, " ")
	}
	// since we are using alpine linux, /bin/bash is unavailable.
	command = exec.Command("/bin/sh", "-c", shellCmd)
	// TODO: write a helper to do the below wiring
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	err = command.Run()
	if err != nil {
		log.Println("error running command, err=", err)
	}

	_ = syscall.Unmount("/proc", 0) // ignore error, may not be mounted
}

func main() {
	// "assume go run main.go is alias for docker -> this is just docker run <command> <args>"
	if len(os.Args) < 2 {
		log.Println("Usage: go run main.go <command> <args...>")
		return
	}

	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		log.Println("invalid command:", os.Args[1])
		return
	}
}
