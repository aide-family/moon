package agent

import (
	"io"

	"golang.org/x/crypto/ssh"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
)

type liveShell struct {
	uid     string
	stdin   io.WriteCloser
	session *ssh.Session
	client  *ssh.Client
}

func (c *Connector) handleShell(stream v1.AgentService_ConnectClient, frame *v1.AgentFrame) {
	if open := frame.GetShellOpen(); open != nil {
		c.startShell(stream, open)
		return
	}
	c.shellMu.Lock()
	current := c.shell
	c.shellMu.Unlock()
	if current == nil {
		return
	}
	if data := frame.GetShellData(); data != nil && data.GetSessionUid() == current.uid {
		_, _ = current.stdin.Write(data.GetData())
		return
	}
	if resize := frame.GetShellResize(); resize != nil && resize.GetSessionUid() == current.uid {
		_ = current.session.WindowChange(int(resize.GetRows()), int(resize.GetCols()))
		return
	}
	if closed := frame.GetShellClose(); closed != nil && closed.GetSessionUid() == current.uid {
		c.closeShell()
	}
}

func (c *Connector) startShell(stream v1.AgentService_ConnectClient, open *v1.AgentShellOpen) {
	c.closeShell()
	client, err := sshDial(open.GetSshHost(), open.GetSshPort(), open.GetSshUser(), open.GetSshPassword(), open.GetSshPrivateKey())
	if err != nil {
		_ = c.send(stream, &v1.AgentFrame{ShellClose: &v1.AgentShellClose{SessionUid: open.GetSessionUid(), Reason: err.Error()}})
		return
	}
	session, err := client.NewSession()
	if err != nil {
		_ = client.Close()
		_ = c.send(stream, &v1.AgentFrame{ShellClose: &v1.AgentShellClose{SessionUid: open.GetSessionUid(), Reason: err.Error()}})
		return
	}
	cols, rows := int(open.GetCols()), int(open.GetRows())
	if cols <= 0 {
		cols = 80
	}
	if rows <= 0 {
		rows = 24
	}
	if err := session.RequestPty("xterm-256color", rows, cols, ssh.TerminalModes{ssh.ECHO: 1, ssh.TTY_OP_ISPEED: 14400, ssh.TTY_OP_OSPEED: 14400}); err != nil {
		_ = session.Close()
		_ = client.Close()
		_ = c.send(stream, &v1.AgentFrame{ShellClose: &v1.AgentShellClose{SessionUid: open.GetSessionUid(), Reason: err.Error()}})
		return
	}
	stdin, err := session.StdinPipe()
	if err != nil {
		_ = session.Close()
		_ = client.Close()
		_ = c.send(stream, &v1.AgentFrame{ShellClose: &v1.AgentShellClose{SessionUid: open.GetSessionUid(), Reason: err.Error()}})
		return
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		_ = session.Close()
		_ = client.Close()
		_ = c.send(stream, &v1.AgentFrame{ShellClose: &v1.AgentShellClose{SessionUid: open.GetSessionUid(), Reason: err.Error()}})
		return
	}
	if err := session.Shell(); err != nil {
		_ = session.Close()
		_ = client.Close()
		_ = c.send(stream, &v1.AgentFrame{ShellClose: &v1.AgentShellClose{SessionUid: open.GetSessionUid(), Reason: err.Error()}})
		return
	}
	live := &liveShell{uid: open.GetSessionUid(), stdin: stdin, session: session, client: client}
	c.shellMu.Lock()
	c.shell = live
	c.shellMu.Unlock()
	go func() {
		buf := make([]byte, 4096)
		for {
			n, readErr := stdout.Read(buf)
			if n > 0 {
				chunk := append([]byte(nil), buf[:n]...)
				if err := c.send(stream, &v1.AgentFrame{ShellData: &v1.AgentShellData{SessionUid: live.uid, Data: chunk}}); err != nil {
					c.closeShell()
					return
				}
			}
			if readErr != nil {
				_ = c.send(stream, &v1.AgentFrame{ShellClose: &v1.AgentShellClose{SessionUid: live.uid, Reason: "closed"}})
				c.closeShell()
				return
			}
		}
	}()
}

func (c *Connector) closeShell() {
	c.shellMu.Lock()
	live := c.shell
	c.shell = nil
	c.shellMu.Unlock()
	if live == nil {
		return
	}
	_ = live.stdin.Close()
	_ = live.session.Close()
	_ = live.client.Close()
}
