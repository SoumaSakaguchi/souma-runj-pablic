package main

import (
	"errors"

	"go.sbk.wtf/runj/jail"
	"go.sbk.wtf/runj/runtimespec"

	"github.com/spf13/cobra"
)

func create_netnsCommand() *cobra.Command{
	create_netns := &cobra.Command{
		Use: "create_netns <netns-id>",
		Short: "Create a new vnet jail like Network Namesapce.",
		Long: "Atode kakuyo!!!",
		Args: cobra.ExactArgs(1),
	}
	create_netns.RunE = func(cmd *cobra.Command, args []string) (err error) { /* 実行部 */
		disableUsage(cmd) /* usage出力の無効化 */
		id := args[0]
		var ociConfig *runtimespec.Spec /* spec情報構造体 */
		ociConfig = setNetnsConf(string(id)) /* config情報のロード */
		if ociConfig == nil {
			return errors.New("OCI config is required")
		}
		if ociConfig.Process == nil{
			return errors.New("OCI config Process is required")
		}
		rootPath := ociConfig.Root.Path

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
		confPath, err = jail.CreateConfig(jailcfg)
		if err != nil{
			return err
		}
		if err := jail.CreateJail(cmd.Context(), confPath); err != nil {
			return err
		}
		err = jail.Mount(ociConfig)
		if err != nil {
			return err
		}
		defer func() {
			if err == nil {
				return
			}
			jail.Unmount(ociConfig)
		}()

		return nil
	}
	return create_netns
}

func setNetnsConf(id string) (*runtimespec.Spec) {
	config := runtimespec.Spec{}

	config.Version = runtimespec.Version
	config.Process = &runtimespec.Process{
		Terminal: false,
		Args: []string{
			"/usr/bin/sh",
		},
		Env: []string{
			"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
			"TERM=xterm",
		},
	}
	config.Root = &runtimespec.Root{
		Path: "/",
	}
	config.Hostname = id
	config.FreeBSD = &runtimespec.FreeBSD{
		Network: &runtimespec.FreeBSDNetwork{
			VNet: &runtimespec.FreeBSDVNet{
				Mode: runtimespec.FreeBSDVNetModeNew,
			},
		},
	}

	return &config
}
