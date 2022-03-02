package cmd

import (
	"fmt"
	"github.com/nyudlts/go-aspace"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(reportCmd)
}

var reportCmd = &cobra.Command{
	Use: "report",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running Role Report")
		client, err = aspace.NewClient("/home/menneric/.config/go-aspace", "fade", 20)
		ReportDOs()
	},
}

func ReportDOs() {
	GetDOIDs()
	doChunks := getChunks(dos)

	resultsChannel := make(chan map[string]int)

	//get the dos
	for i, chunk := range doChunks {
		go GetRoles(chunk, resultsChannel, i+1)
	}

	results := map[string]int{}
	for range doChunks {
		chunk := <-resultsChannel
		for k, v := range chunk {
			if HasRole(results, k) == true {
				results[k] = results[k] + v
			} else {
				results[k] = v
			}
		}
	}
	GenerateRoleReport(results)
	PrintRoleMap(results)
}
