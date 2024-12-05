package cmd

// $0 ao get-root --ao-uri|-a          // finds the root archival object for the ao-uri argument, returning itself if there are no ancestors

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(aoCmd)
	aoCmd.AddCommand(aoGetRootCmd)
}

// --------------------------------------------------------------------------------
// aoCmd represents the ao command
var aoCmd = &cobra.Command{
	Use:   "ao",
	Short: "Archival Object (ao) operations",
	Long: `The ao noun allows you to perform
certain operations on Archival Object (ao) resources`,
	Run: aoRoot,
}

func aoRoot(cmd *cobra.Command, args []string) { printNeedSubcommandHelp(cmd) }

// --------------------------------------------------------------------------------
var aoGetRootCmd = &cobra.Command{
	Use:   "get-root",
	Short: "Get the root archival object for the ao-uri argument",
	Long: `The get-root subcommand finds the root archival object for the ao-uri argument,
returning itself if there are no ancestors.`,
	RunE: aoGetRoot,
}

var aoURI string
var aoFlags = struct {
	URI      string
	URIShort string
}{
	URI:      "ao-uri",
	URIShort: "a",
}

func init() {
	aoGetRootCmd.PersistentFlags().StringVarP(&config, "config", "c", "", "path to the configuration file")
	aoGetRootCmd.PersistentFlags().StringVarP(&env, "environment", "e", "", "environment to use")
	aoGetRootCmd.PersistentFlags().BoolVar(&test, "test", false, "")

	aoGetRootCmd.Flags().StringVarP(&aoURI, aoFlags.URI, aoFlags.URIShort, "", "uri of the archival object")
	aoGetRootCmd.MarkFlagRequired(aoFlags.URI)
}

func aoGetRoot(cmd *cobra.Command, args []string) (err error) {
	setClient()
	for {
		ao, err := client.GetArchivalObjectFromURI(aoURI)
		if err != nil {
			panic(err)
		}
		if ao.Parent["ref"] == "" {
			fmt.Println(aoURI)
			break
		}
		aoURI = ao.Parent["ref"]
	}

	return nil
}

//--------------------------------------------------------------------------------
