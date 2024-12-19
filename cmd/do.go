package cmd

// implemented:
// $0 do refresh  --ao-uri|-a <ao URI> // updates the metadata of all DOs attached to the AO

// pending:
// $0 do create   --ao-uri|-a <ao URI> --file-version|-f <file URI> --use-statement|-u <use statement>
// $0 do update   --ao-uri|-a <ao URI> --old-file-version|-o <file URI to replace> --file-version|-f <new file URI value> --use-statement|-u <new FV use statement>
import (
	"fmt"

	"github.com/nyudlts/go-aspace"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(doCmd)
	doCmd.AddCommand(doRefreshCmd)
}

// --------------------------------------------------------------------------------
// Flags and Parameters
var aoURI string
var aoFlags = struct {
	URI      string
	URIShort string
}{
	URI:      "ao-uri",
	URIShort: "a",
}

// var fv string // file version
// var fvFlags = struct {
// 	FileVersion      string
// 	FileVersionShort string
// }{
// 	FileVersion:      "file-version",
// 	FileVersionShort: "f",
// }

// var oldFV string // old file version
// var oldFVFlags = struct {
// 	OldFileVersion      string
// 	OldFileVersionShort string
// }{
// 	OldFileVersion:      "old-file-version",
// 	OldFileVersionShort: "o",
// }

// --------------------------------------------------------------------------------
// doCmd represents the do command
var doCmd = &cobra.Command{
	Use:   "do",
	Short: "Digital Object (do) operations",
	Long: `The 'do' noun allows you to perform
certain operations on Digital Object (do) resources`,
	Run: doRoot,
}

func doRoot(cmd *cobra.Command, args []string) { printNeedSubcommandHelp(cmd) }

// --------------------------------------------------------------------------------
var doRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Update the titles of all Digital Objects",
	Long: `The refresh subcommand updates the titles of all digital objects ('do's) attached to
the specified archival object (ao) by copying the ao title to each do title`,
	RunE: doRefresh,
}

func init() {
	doRefreshCmd.PersistentFlags().StringVarP(&config, "config", "c", "", "path to the configuration file")
	doRefreshCmd.PersistentFlags().StringVarP(&env, "environment", "e", "", "environment to use")
	doRefreshCmd.PersistentFlags().BoolVar(&test, "test", false, "")

	doRefreshCmd.Flags().StringVarP(&aoURI, aoFlags.URI, aoFlags.URIShort, "", "uri of the archival object")
	doRefreshCmd.MarkFlagRequired(aoFlags.URI)

	// build command hierarchy
	rootCmd.AddCommand(doCmd)
	doCmd.AddCommand(doRefreshCmd)
}

func doRefresh(cmd *cobra.Command, args []string) (err error) {
	setClient()

	ao, err := client.GetArchivalObjectFromURI(aoURI)
	if err != nil {
		return err
	}

	doURIs, err := client.GetDigitalObjectIDsForArchivalObjectFromURI(aoURI)
	if err != nil {
		return err
	}

	// refresh the titles of all the DOs
	for _, doURI := range doURIs {
		do, err := client.GetDigitalObjectFromURI(doURI)
		if err != nil {
			return err
		}

		do.Title = ao.Title

		repoID, objectID, err := aspace.URISplit(doURI)
		if err != nil {
			return err
		}

		// update the do
		body, err := client.UpdateDigitalObject(repoID, objectID, do)
		if err != nil {
			return fmt.Errorf("%v : %s", err, body)
		}
		fmt.Printf("updated ao: %s do: %s\n", aoURI, doURI)
	}
	return nil
}
