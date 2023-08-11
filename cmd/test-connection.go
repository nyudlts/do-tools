package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	testCmd.PersistentFlags().StringVarP(&config, "config", "c", "", "")
	testCmd.PersistentFlags().StringVarP(&env, "environment", "e", "", "")
	rootCmd.AddCommand(testCmd)
}

var testCmd = &cobra.Command{
	Use: "test-connection",
	Run: func(cmd *cobra.Command, args []string) {
		setClient()
		fmt.Println(client)
	},
}
