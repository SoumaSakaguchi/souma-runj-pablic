// +build integration

package integration

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"go.sbk.wtf/runj/runtimespec"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDelete(t *testing.T) {
	dir, err := ioutil.TempDir("", "runj-integ-test-"+t.Name())
	require.NoError(t, err)
	defer func() {
		if !t.Failed() {
			os.RemoveAll(dir)
		} else {
			t.Log("preserving tempdir due to failure", dir)
		}
	}()

	tests := []runtimespec.Spec{
		// minimal
		{
			Process: &runtimespec.Process{},
		},
		// arguments
		{
			Process: &runtimespec.Process{
				Args: []string{"one", "two", "three"},
			},
		},
		// environment variables
		{
			Process: &runtimespec.Process{
				Env: []string{"one=two", "three=four", "five"},
			},
		},
	}

	for i, tc := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			bundleDir := filepath.Join(dir, strconv.Itoa(i))
			defer func() {
				if !t.Failed() {
					os.RemoveAll(bundleDir)
				} else {
					t.Log("preserving tempdir due to failure", bundleDir)
				}
			}()
			rootDir := filepath.Join(bundleDir, "root")
			err := os.MkdirAll(rootDir, 0755)
			require.NoError(t, err, "create bundle dir")
			t.Log("bundle", bundleDir)

			configJSON, err := json.Marshal(tc)
			require.NoError(t, err, "marshal config")
			err = ioutil.WriteFile(filepath.Join(bundleDir, "config.json"), configJSON, 0644)
			require.NoError(t, err, "write config")

			id := "test-create-delete-" + strconv.Itoa(i)
			cmd := exec.Command("runj", "create", id, bundleDir)
			cmd.Stdin = nil
			out, err := os.OpenFile(filepath.Join(bundleDir, "out"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
			require.NoError(t, err, "out file")
			cmd.Stdout = out
			cmd.Stderr = out
			err = cmd.Run()
			assert.NoError(t, err, "runj create")
			err = out.Close()
			assert.NoError(t, err, "out file close")
			outBytes, err := ioutil.ReadFile(filepath.Join(bundleDir, "out"))
			assert.NoError(t, err, "out file read")
			t.Log("runj create output:", string(outBytes))

			cmd = exec.Command("runj", "delete", id)
			cmd.Stdin = nil
			outBytes, err = cmd.CombinedOutput()
			assert.NoError(t, err, "runj delete")
			t.Log("runj delete output:", string(outBytes))
		})
	}
}

func TestJailHello(t *testing.T) {
	root, err := ioutil.TempDir("", "runj-integ-test-"+t.Name())
	require.NoError(t, err, "create root")
	defer os.RemoveAll(root)

	err = copyFile("bin/integ-inside", filepath.Join(root, "integ-inside"))
	require.NoError(t, err, "copy inside binary")

	spec := runtimespec.Spec{
		Process: &runtimespec.Process{
			Args: []string{"/integ-inside", "-test.v", "-test.run", "TestHello"},
		},
		Root: &runtimespec.Root{Path: root},
	}
	stdout, stderr, err := runSimpleExitingJail(t, "integ-test-hello", spec, 2*time.Second)
	assert.NoError(t, err)
	t.Log("STDOUT:", string(stdout))
	t.Log("STDERR:", string(stderr))
}

func copyFile(source, dest string) error {
	stat, err := os.Stat(source)
	if err != nil {
		return err
	}
	in, err := os.OpenFile(source, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, stat.Mode())
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// runSimpleExitingJail is a helper that takes a spec as input, sets up a bundle
// starts a jail, collects its output, and waits for the jail's entrypoint to
// exit.  It can be used in tests where the entrypoint embeds the test
// assertions.
// TODO: Build a better non-racy or less-racy end condition.
// The wait parameter is currently used as a simple sleep between `runj start`
// and `runj delete`.  A normal wait is not used as the jail's main process is
// not a direct child of this test; it's instead a child of the `runj create`
// process.
func runSimpleExitingJail(t *testing.T, id string, spec runtimespec.Spec, wait time.Duration) ([]byte, []byte, error) {
	t.Helper()
	bundleDir, err := ioutil.TempDir("", "runj-integ-test-"+t.Name()+"-"+id)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if err == nil {
			os.RemoveAll(bundleDir)
		} else {
			t.Log("preserving tempdir due to error", bundleDir, err)
		}
	}()
	rootDir := filepath.Join(bundleDir, "root")
	err = os.MkdirAll(rootDir, 0755)
	if err != nil {
		return nil, nil, fmt.Errorf("create bundle dir: %w", err)
	}
	t.Log("bundle", bundleDir)

	configJSON, err := json.Marshal(spec)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal config: %w", err)
	}
	err = ioutil.WriteFile(filepath.Join(bundleDir, "config.json"), configJSON, 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("write config: %w", err)
	}

	cmd := exec.Command("runj", "create", id, bundleDir)
	cmd.Stdin = nil
	stdout, err := os.OpenFile(filepath.Join(bundleDir, "stdout"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("create stdout file: %w", err)
	}
	cmd.Stdout = stdout
	stderr, err := os.OpenFile(filepath.Join(bundleDir, "stderr"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("create stderr file: %w", err)
	}
	cmd.Stderr = stderr

	err = cmd.Run()
	if err != nil {
		return nil, nil, fmt.Errorf("runj create: %w", err)
	}
	err = stdout.Close()
	if err != nil {
		return nil, nil, fmt.Errorf("close stdout file: %w", err)
	}
	err = stderr.Close()
	if err != nil {
		return nil, nil, fmt.Errorf("close stderr file: %w", err)
	}

	defer func() {
		cmd = exec.Command("runj", "delete", id)
		cmd.Stdin = nil
		outBytes, cleanupErr := cmd.CombinedOutput()
		if cleanupErr != nil && err == nil {
			err = fmt.Errorf("runj delete: %w", cleanupErr)
		}
		t.Log("runj delete output:", string(outBytes))
	}()

	// runj start
	cmd = exec.Command("runj", "start", id)
	err = cmd.Run()
	if err != nil {
		return nil, nil, fmt.Errorf("runj start: %w", err)
	}
	time.Sleep(wait)

	stdoutBytes, err := ioutil.ReadFile(filepath.Join(bundleDir, "stdout"))
	if err != nil {
		return nil, nil, fmt.Errorf("read stdout file: %w", err)
	}
	stderrBytes, err := ioutil.ReadFile(filepath.Join(bundleDir, "stderr"))
	if err != nil {
		return nil, nil, fmt.Errorf("read stderr file: %w", err)
	}
	return stdoutBytes, stderrBytes, nil
}
