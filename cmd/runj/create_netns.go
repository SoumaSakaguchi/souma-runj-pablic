package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func create_netnsCommand() *cobra.Command{
	create_netns := &cobra.Command{
		Use: "crate_netns",
		Short: "Create a new vnet jail like Network Namesapce.",
		PreRunE: func() {
			fmt.Printf("Test create_netns.go");
		},
	}
}
