package main

import (
	"docker-engine/utils"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

func Run() {
	// run will simply clone itself after requesting new namespaces from the kernel
	// like how docker engine runs, it will simply spin up containers(which are just more containers of the same type)

	// spawn a child process of the same instance - but under the new namespace
	// since a running process cannot change the namespace, spawn a child process under the new namespace.
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)

	// link current shell to process
	utils.LinkIOStreamsToCurrentShell(cmd, utils.IOStreams{
		STDIN:  os.Stdin,
		STDOUT: os.Stdout,
		STDERR: os.Stderr,
	})

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWNS | syscall.CLONE_NEWPID,
	}

	// use start as it is non-blocking, unlike Run()
	err := cmd.Start()
	if err != nil {
		log.Println("error running command, err=", err)
	}

	if cmd.Process == nil {
		log.Println("process is nil")
		return
	}

	err = configureCgroups(cmd.Process.Pid)
	if err != nil {
		return
	}
}

func configureCgroups(pid int) error {
	err := os.WriteFile("/sys/fs/cgroup/cgroup.subtree_control", []byte("+memory +pids"), 0700)
	if err != nil {
		log.Println("error enabling cgroup controllers, err=", err)
		return err
	}
	cGroupPath := filepath.Join("/sys/fs/cgroup", "test-container")
	err = os.MkdirAll(cGroupPath, 0755)
	if err != nil {
		log.Println("error creating cgroup directory, err=", err)
		return err
	}
	memoryLimit := filepath.Join(cGroupPath, "memory.max")
	err = os.WriteFile(memoryLimit, []byte("100M"), 0700)
	if err != nil {
		log.Println("error writing memory limit, err=", err)
		return err
	}
	processLimit := filepath.Join(cGroupPath, "pids.max")
	err = os.WriteFile(processLimit, []byte("20"), 0700)
	if err != nil {
		log.Println("error writing process limit, err=", err)
		return err
	}
	processFile := filepath.Join(cGroupPath, "cgroup.procs")
	err = os.WriteFile(processFile, []byte(strconv.Itoa(pid)), 0700)
	if err != nil {
		log.Println("error writing to processes file, err=", err)
		return err
	}
	return nil
}

func Child() {
	hostname := "isolated-hostname"
	err := syscall.Sethostname([]byte(hostname))
	if err != nil {
		log.Println("error setting hostname, err=", err)
	}
	err = syscall.Mount("/dev", "../alpine-rootfs/dev", "", syscall.MS_BIND, "")
	if err != nil {
		log.Println("error bind mounting /dev, err=", err)
	}
	err = syscall.Chroot("../alpine-rootfs")
	if err != nil {
		log.Println("error setting root dir, err=", err)
	}
	err = os.Chdir("/")
	if err != nil {
		log.Println("error moving to root dir, err=", err)
	}
	installCmd := exec.Command("/bin/busybox", "--install", "-s", "/bin")
	utils.LinkIOStreamsToCurrentShell(installCmd, utils.IOStreams{
		STDIN:  os.Stdin,
		STDOUT: os.Stdout,
		STDERR: os.Stderr,
	})
	err = installCmd.Run()
	if err != nil {
		log.Println("error installing busybox, err", err)
		return
	}
	command := exec.Command("/bin/sh", "-c", "ls -a && (mount --make-rprivate / || true); (mount -t proc proc /proc || true)")
	utils.LinkIOStreamsToCurrentShell(command, utils.IOStreams{
		STDIN:  os.Stdin,
		STDOUT: os.Stdout,
		STDERR: os.Stderr,
	})
	_ = command.Run() // ignore errors, these may fail in container environments

	targetCmd := os.Args[2]
	targetCmdArgs := os.Args[3:]
	shellCmd := targetCmd
	if len(targetCmdArgs) > 0 {
		shellCmd = targetCmd + " " + strings.Join(targetCmdArgs, " ")
	}
	// since we are using alpine linux, /bin/bash is unavailable.
	command = exec.Command("/bin/sh", "-c", shellCmd)
	utils.LinkIOStreamsToCurrentShell(command, utils.IOStreams{
		STDIN:  os.Stdin,
		STDOUT: os.Stdout,
		STDERR: os.Stderr,
	})
	err = command.Run()
	if err != nil {
		log.Println("error running command, err=", err)
	}

	_ = syscall.Unmount("/proc", 0) // ignore error, may not be mounted

	// limit number of processes
	for i := 0; i < 40; i++ {
		command = exec.Command("/bin/sh", "-c", "sleep 10")
		if err := command.Start(); err != nil {
			log.Println("error running command, err=", err)
			return
		}
	}
}

func main() {
	// "assume go run main.go is alias for docker -> this is just docker run <command> <args>"
	if len(os.Args) < 2 {
		log.Println("Usage: go run main.go <command> <args...>")
		return
	}

	switch os.Args[1] {
	case "run":
		Run()
	case "child":
		Child()
	default:
		log.Println("invalid command:", os.Args[1])
		return
	}
}
