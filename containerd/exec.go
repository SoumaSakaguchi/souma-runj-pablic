package containerd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"sync"

	"github.com/containerd/console"
	"github.com/containerd/containerd/log"
	"github.com/containerd/containerd/sys/reaper"
	runc "github.com/containerd/go-runc"
	"github.com/sirupsen/logrus"
)

// execCreate runs the "create" subcommand for runj
func execCreate(ctx context.Context, id, bundle string, stdin io.Reader, stdout io.Writer, stderr io.Writer, terminal bool) (console.Console, error) {
	args := []string{"create", id, bundle, "--netns-compat"}
	var socket *runc.Socket
	if terminal {
		log.G(ctx).WithField("id", id).Warn("Creating terminal!")
		var err error
		socket, err = runc.NewTempConsoleSocket()
		if err != nil {
			return nil, fmt.Errorf("create: failed to create runj console socket: %w", err)
		}
		defer socket.Close()
		args = append(args, "--console-socket", socket.Path())
	}

	cmd := exec.CommandContext(ctx, "runj", args...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if terminal && cmd.Stderr == nil {
		cmd.Stderr = log.G(ctx).WithField("cmd", "runj create").WriterLevel(logrus.WarnLevel)
	}
	log.G(ctx).WithField("id", id).WithField("args", args).Warn("Starting runj create")
	ec, err := reaper.Default.Start(cmd)
	if err != nil {
		return nil, err
	}

	ready := make(chan struct {
		con console.Console
		err error
	})
	if socket != nil {
		go func() {
			con, err := func() (console.Console, error) {
				con, err := socket.ReceiveMaster()
				if err != nil {
					return nil, fmt.Errorf("failed to retrieve console master: %w", err)
				}
				log.G(ctx).WithField("id", id).Warn("Received console master!")
				err = copyConsole(ctx, con, stdin, stdout, stderr)
				if err != nil {
					return nil, fmt.Errorf("failed to start console copy: %w", err)
				}
				log.G(ctx).WithField("id", id).Warn("Copying console!")
				return con, nil
			}()
			ready <- struct {
				con console.Console
				err error
			}{con, err}
			close(ready)
		}()
	}

	ret, err := WaitNoFlush(cmd, ec)
	if err != nil {
		log.G(ctx).WithError(err).WithField("id", id).Error("runj create failed")
		return nil, err
	}
	if ret != 0 {
		log.G(ctx).WithField("exit", ret).Error("runj create failed")
		return nil, errors.New("runj create failed")
	}
	if socket != nil {
		ret := <-ready
		return ret.con, ret.err
	}
	return nil, nil
}

func copyConsole(ctx context.Context, console console.Console, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	var cwg sync.WaitGroup
	if stdin != nil {
		cwg.Add(1)
		go func() {
			cwg.Done()
			io.Copy(console, stdin)
		}()
	}
	if stdout != nil {
		cwg.Add(1)
		go func() {
			cwg.Done()
			io.Copy(stdout, console)
		}()
	}
	cwg.Wait()
	return nil
}

// WaitNoFlush waits for a process to exit but does not flush IO with cmd.Wait
func WaitNoFlush(c *exec.Cmd, ec chan runc.Exit) (int, error) {
	for e := range ec {
		if e.Pid == c.Process.Pid {
			reaper.Default.Unsubscribe(ec)
			return e.Status, nil
		}
	}
	// return no such process if the ec channel is closed and no more exit
	// events will be sent
	return -1, reaper.ErrNoSuchProcess
}

type ociState struct {
	OCIVersion  string            `json:"ociVersion"`
	ID          string            `json:"id"`
	Status      string            `json:"status"`
	PID         int               `json:"pid,omitempty"`
	Bundle      string            `json:"bundle"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// execState runs the "state" subcommand for runj
func execState(ctx context.Context, id string) (*ociState, error) {
	cmd := exec.CommandContext(ctx, "runj", "state", id)
	b, err := combinedOutput(cmd)
	if err != nil {
		log.G(ctx).
			WithError(err).
			WithField("output", string(b)).
			WithField("id", id).Error("runj state failed")
		return nil, err
	}
	s := &ociState{}
	err = json.Unmarshal(b, s)
	return s, err
}

// execDelete runs the "delete" subcommand for runj
func execDelete(ctx context.Context, id string) error {
	cmd := exec.CommandContext(ctx, "runj", "delete", id)
	b, err := combinedOutput(cmd)
	if err != nil {
		log.G(ctx).WithError(err).WithField("output", string(b)).WithField("id", id).Error("runj delete failed")
		return err
	}
	return nil
}

// execKill runs the "kill" subcommand for runj
func execKill(ctx context.Context, id string, signal string, all bool, pid int) error {
	args := []string{"kill", id, signal}
	if all {
		args = append(args, "--all")
	}
	if pid != 0 {
		args = append(args, "--pid", strconv.Itoa(pid))
	}
	cmd := exec.CommandContext(ctx, "runj", args...)
	b, err := combinedOutput(cmd)
	if err != nil {
		log.G(ctx).WithError(err).WithField("output", string(b)).WithField("id", id).Error("runj kill failed")
		return err
	}
	return nil
}

// execStart runs the "start" subcommand for runj
func execStart(ctx context.Context, id string) error {
	cmd := exec.CommandContext(ctx, "runj", "start", id)
	b, err := combinedOutput(cmd)
	if err != nil {
		log.G(ctx).WithError(err).WithField("output", string(b)).WithField("id", id).Error("runj start failed")
		return err
	}
	return nil
}

// execExec runs the "extension exec" subcommand for runj
func execExec(ctx context.Context, id, processJSONFilename string, stdin io.Reader, stdout io.Writer, stderr io.Writer, terminal bool) (int, console.Console, error) {
	args := []string{"extension", "exec", id, "--process", processJSONFilename}
	var socket *runc.Socket
	if terminal {
		log.G(ctx).WithField("id", id).Warn("Creating terminal!")
		var err error
		socket, err = runc.NewTempConsoleSocket()
		if err != nil {
			return -1, nil, fmt.Errorf("create: failed to create runj console socket: %w", err)
		}
		defer socket.Close()
		args = append(args, "--console-socket", socket.Path())
	}

	cmd := exec.CommandContext(ctx, "runj", args...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if cmd.Stderr == nil {
		cmd.Stderr = log.G(ctx).WithField("cmd", "runj ext exec").WriterLevel(logrus.WarnLevel)
	}
	log.G(ctx).WithField("id", id).Warn("Starting runj extension exec")
	err := cmd.Start()
	if err != nil {
		return 0, nil, err
	}
	pid := cmd.Process.Pid

	var con console.Console
	if socket != nil {
		con, err = socket.ReceiveMaster()
		if err != nil {
			return pid, nil, fmt.Errorf("failed to retrieve console master: %w", err)
		}
		log.G(ctx).WithField("id", id).Warn("Received exec console master!")
		err = copyConsole(ctx, con, stdin, stdout, stderr)
		if err != nil {
			return pid, nil, fmt.Errorf("failed to start console copy: %w", err)
		}
		log.G(ctx).WithField("id", id).Warn("Copying exec console!")
	}

	return pid, con, nil
}

func combinedOutput(cmd *exec.Cmd) ([]byte, error) {
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout
	ec, err := reaper.Default.Start(cmd)
	if err != nil {
		return nil, err
	}
	_, err = reaper.Default.Wait(cmd, ec)
	b := stdout.Bytes()
	return b, err
}
