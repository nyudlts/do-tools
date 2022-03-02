package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "do-tools",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("**DO TOOLS**")
	},
}
