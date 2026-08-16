package utils

import (
	"os"
	"os/exec"
)

type IOStreams struct {
	STDIN  *os.File
	STDOUT *os.File
	STDERR *os.File
}

func LinkIOStreamsToCurrentShell(cmd *exec.Cmd, ioStreams IOStreams) {
	cmd.Stdin = ioStreams.STDIN
	cmd.Stderr = ioStreams.STDERR
	cmd.Stdout = ioStreams.STDOUT
}
