package oci

import (
	"go.sbk.wtf/runj/runtimespec"
)

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
