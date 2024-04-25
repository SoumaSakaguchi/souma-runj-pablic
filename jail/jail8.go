package jail

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// CreateJail wraps the jail(8) command to create a jail
func CreateJail(ctx context.Context, confPath string) error {
	cmd := exec.CommandContext(ctx, "jail", "-cf", confPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintln(os.Stderr, string(out))
	}
	return err
}

// CreateNestJail wraps the jail(8) and jexec(8) command to crate a jail in existing jail
func CreateNestedJail(ctx context.Context, confPath, jid string) error {
	cmd := exec.CommandContext(ctx, "jexec", jid, "jail", "-cf", confPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintln(os.Stderr, string(out))
	}
	return err
}

// DestroyJail wraps the jail(8) command to destroy a jail
func DestroyJail(ctx context.Context, confPath, jail string) error {
	cmd := exec.CommandContext(ctx, "jail", "-f", confPath, "-r", jail)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintln(os.Stderr, string(out))
	}
	return err
}

func DestriyNestedJail(ctx context.Context, confPath, jail, parent_jail string) error {
	cmd := exec.CommandContext(ctx, "jexec", parent_jail, "jail", "-f", confPath, "-r", jail)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintln(os.Stderr, string(out))
	}
	return err
}
