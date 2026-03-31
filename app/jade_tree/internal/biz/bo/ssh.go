// Package bo contains business objects used across layers.
package bo

import "time"

// SSHExecRequest defines connection and command options for remote execution.
type SSHExecRequest struct {
	Host       string
	Port       int
	Username   string
	Password   string
	PrivateKey string
	Timeout    time.Duration
	Command    string
	WorkDir    string
	Env        map[string]string
}

// SSHExecReply is the result of remote command execution.
type SSHExecReply struct {
	Stdout   string
	Stderr   string
	ExitCode int
}
