package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/nyudlts/go-aspace"
	"github.com/spf13/cobra"
	"log"
	"regexp"
	"strings"
)

func init() {
	convertHandlesCmd.PersistentFlags().StringVar(&config, "config", "", "")
	convertHandlesCmd.PersistentFlags().StringVar(&env, "environment", "", "")
	convertHandlesCmd.PersistentFlags().BoolVar(&test, "test", false, "")
	rootCmd.AddCommand(convertHandlesCmd)
}

var (
	httpMatch = regexp.MustCompile("^http://hdl.handle")
)

var convertHandlesCmd = &cobra.Command{
	Use: "convert-handles",
	Run: func(cmd *cobra.Command, args []string) {
		setClient()

		err := convertHandles()
		if err != nil {
			panic(err)
		}
	},
}

func convertHandles() error {

	for _, repoID := range []int{2, 3, 6} {
		fmt.Println(repoID)
		initialResult, err := client.Search(repoID, "digital_object", "http:", 1)
		if err != nil {
			panic(err)
		}

		for pageID := 1; pageID < initialResult.LastPage; pageID++ {
			results, err := client.Search(repoID, "digital_object", "hdl.handle.net", pageID)
			if err != nil {
				panic(err)
			}

			for _, hit := range results.Results {
				hitJson := hit["json"].(string)
				do := aspace.DigitalObject{}
				err := json.Unmarshal([]byte(hitJson), &do)
				if err != nil {
					log.Println(err.Error())
					continue
				}
				if len(do.FileVersions) == 1 {
					if strings.Contains(do.FileVersions[0].FileURI, "http://") {
						fmt.Println(do.URI, do.FileVersions[0].FileURI)
					}
				}
			}
		}
	}

	return nil
}
