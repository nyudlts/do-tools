package cmd

import (
	"fmt"
	"github.com/nyudlts/go-aspace"
	"github.com/spf13/cobra"
)

func init() {
	thumbnailCmd.PersistentFlags().StringVarP(&config, "config", "c", "", "")
	rootCmd.AddCommand(thumbnailCmd)
}

var thumbnailCmd = &cobra.Command{
	Use: "remove-thumbnails",
	Run: func(cmd *cobra.Command, args []string) {
		client, err = aspace.NewClient(config, "fade", 20)
		if err != nil {
			panic(err)
		}
		GetDOIDs()
		removeThumbs()
	},
}

type Result struct {
	Code string
	Msg  string
}

func removeThumbs() {

	chunks := getChunks(dos)
	resultChannel := make(chan []Result)

	for i, chunk := range chunks {
		go removeThumbnails(chunk, resultChannel, i+1)
	}

	for range chunks {
		for _, result := range <-resultChannel {
			fmt.Printf("%v\n", result)
		}
	}

}

func removeThumbnails(chunk []DigitalObjectIDs, resultChannel chan []Result, worker int) {
	results := []Result{}
	fmt.Printf("Worker %d started, processing %d records\n", worker, len(chunk))
	for i, doid := range chunk {
		if i > 0 && i%250 == 0 {
			fmt.Printf("Worker %d has completed %d digital objects\n", worker, i)
		}
		do, err := client.GetDigitalObject(doid.RepoID, doid.DOID)
		if err != nil {
			results = append(results, Result{"ERROR", err.Error()})
			continue
		}

		if len(do.FileVersions) > 0 {
			if checkForRole("image-thumbnail", do.FileVersions) == true {
				//update the do
				do.FileVersions = updateFileVersions(do)
				response, err := client.UpdateDigitalObject(doid.RepoID, doid.DOID, do)
				if err != nil {
					results = append(results, Result{"ERROR", err.Error()})
					continue
				}
				results = append(results, Result{"SUCCESS", fmt.Sprintf("%v", response)})
				continue
			} else {
				results = append(results, Result{"SKIPPED", fmt.Sprintf("%v", do.URI)})
				continue
			}
		}
		results = append(results, Result{"SKIPPED", fmt.Sprintf("%v", do.URI)})
	}
	fmt.Printf("Worker %d finished\n", worker)
	resultChannel <- results
}

func checkForRole(role string, versions []aspace.FileVersion) bool {
	for _, version := range versions {
		if version.UseStatement == role {
			return true
		}
	}
	return false
}

func updateFileVersions(do aspace.DigitalObject) []aspace.FileVersion {
	newFileVersions := []aspace.FileVersion{}
	for _, fileVersion := range do.FileVersions {
		if fileVersion.UseStatement != "image-thumbnail" {
			newFileVersions = append(newFileVersions, fileVersion)
		}
	}
	return newFileVersions
}
