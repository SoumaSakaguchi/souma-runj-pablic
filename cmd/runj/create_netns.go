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
	// flagの追加
	flags := create_netns.Flags()
	flags.StringVerP{
		&bundle,
		"bundle",
		"b",
		"",
		"path to the root of the bundle directory")
	*/
	create_netns.RunE = func(cmd *cobra.command, args []string) (err error) { /* 実行部 */
		disableUsage(cmd) /* usage出力の無効化 */
		var s *state.State /* id, jid, Status, Bundle, pid */
		s, err = state.Create(id, bundle) /* /var/run/runj/jails/<id>にステータスファイルを作成 */
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
		err = oci.StoreConfig(id, bundle) /* コンテナディレクトリにconfig.jsonをコピー */
		if err != nil {
			return err
		}
		var ociConfig *runtimespec.Spec /* spec情報構造体 */
		ociConfig, err = oci.LoadConfig(id) /* config情報のロード */
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

		jailcfg := &jail.Config{
			Name:		id,
			Root:		rootPath,
			Hostname:	ociConfig.Hostname,
		}
		if ociConfig.FreeBSD != nil && ociConfig.FreeBSD.Network != nil {
			if ociConfig.FreeBSD.Network.IPv4 != nil {
				jailcfg.IP4 = string(ociConfig.FreeBSD.Network.IPv4.Mode)
				jailcfg.IP4Addr = ociConfig.FreeBSD.Network.IPv4.Addr
			}
			if ociConfig.FreeBSD.Network.VNet != nil {
				jailcfg.VNet = string(ociConfig.FreeBSD.Network.VNet.Mode)
				jailcfg.VNetInterface = ociConfig.FreeBSD.Network.VNet.Interfaces
			}
		}

		var confPath string
		confPath, err = jail.CreateConfig(jailcfg) // ここらへんを変える
		if err != nil{
			return err
		}
		if err := jail.CreateJail(cmd.Context(), confPath); err != nil {
			return err
		}
		err = jail.Mount(ociConfig) // ociConfig構造体の中身見たい
		if err != nil {
			return err
		}
		defer func() {
			if err == nil {
				return
			}
			jail.Unmount(ociConfig)
		}()

		var entrypoint *exec.Cmd
		entrypoint, err = jail.SetupEntrypoint(id, true, ociConfig.Process.Args, ociConfig.Process.Env, consoleSocket)
		if err != nil{
			return err
		}

		s.PID = entrypoint.Process.Pid // entrypoint is なに
		if pidFile != "" {
			pidValue := strconv.Itoa(s.PID)
			err = os.WriteFile(pidFile, []byte(pidValue), 0o666)
			if err != nil {
				return err
			}
		}

		if ociConfig.Hooks != nil {
			for _, h := range ociCofig.Hooks.CreateRuntime {
				output := s.Output()
				output.Annotations = ociConfig.Annotations
				err = hook.Run(&output, &h)
				if err != nil {
					return err
				}
			}
		}

		return nil
	}
	return create_netns
}
