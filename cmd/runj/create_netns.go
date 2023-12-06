package main

import (
	"fmt"

	"go.sbk.wtf/runj/hook"
	"go.sbk.wtf/runj/runj"
	"go.sbk.wtf/runj/oci"
	"go.sbk.wtf/runj/runtimespec"
	"go.sbk.wtf/runj/state"

	"github.com/spf13/cobra"
)

func create_netnsCommand() *cobra.Command{
	create_netns := &cobra.Command{
		Use: "crate_netns",
		Short: "Create a new vnet jail like Network Namesapce.",
		Long: "Atode kakuyo!!!"
	}
	/*
	flags := create_netns.Flags()
	flags.StringVerP{
		&bundle,
		"bundle",
		"b",
		"",
		"path to the root of the bundle directory")
	*/
	create_netns.RunE = func(cmd *cobra.command, args []string) (err error) {
		disableUsage(cmd)
		var s *state.State
		s, err = state.Create(id, bundle) /* idとbundleを取る必要あり */
		if err != nil {
			return err
		}
		defer func() {
			if err == nil {
				s.Status = state.StatusCreated
				err = s.Save()
			} else {
				state.Remove(id)
			}
		}()
		err = oci.StoreConfig(id, bundle)
		if err != nil {
			return err
		}
		var ociConfig *runtimespec.Spec
		ociConfig, err = oci.LoadConfig(id)
		if err != nil {
			return err
		}
		if ociConfig == nil {
			return errors.New("OCI config is required")
		}
		if ociConfig.Process == nil{
			return errors.New("OCI config Process is required")
		}
		rootPath := filepath.Join(bundle, "root")
		if ociConfig.Root != nil && ociConfig.Root.Path != "" {
			rootPath = ociConfig.Root.Path
			if rootPath[0] != filepath.Separator {
				rootPath = filepath.Join(bundle, rootPath)
			}
			ociConfig.Root.Path = rootPath
		} else {
			ociConfig.Root = &runtimespec.Root{Path: rootPath}
		}

		if ociConfig.Process.Terminal {
			if consoleSocket == "" {
				return errors.New("console-socket provided but Process. Terminal is true")
			}
			if socketStat, err := os.Stat(consoleSocket); err != nil {
				return fmt.Errorf("faild to stat console socket %q: %w", consoleSocket, err)
			} else if socketStat.Mode()&os.ModeSocket != os.ModeSocket {
				return fmt.Errorf("console-socket %q is not a socket", consoleSocket)
			}
		}else if consoleSocket != "" {
			return errors.New("console-socket provided but Process. Terminal is false")
		}
	}
}
