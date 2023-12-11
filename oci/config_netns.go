package oci

import (
	"encoding/json"
	"os"

	"go.sbk.wtf/runj/runtimespec"
)

func setNetnsConf() (*runtimespec.Spec, error) {
	config := *runtimespec.Spec{}

	config.Version = runtimespec.Version
	config.Process = {
		false,
		["/usr/bin/sh"],
		["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"]
	}
	cofig.Root = {
		"/"
	}
	config.Hostname = "netns"
	config.FreeBSD = {
		{nil,{"new",nil}}
	}

	return config, nil
}
