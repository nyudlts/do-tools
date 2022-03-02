package cmd

import (
	"bufio"
	"fmt"
	"github.com/nyudlts/go-aspace"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"time"
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
	URI  string
	Msg  string
	Time time.Time
}

func removeThumbs() {

	chunks := getChunks(dos)
	resultChannel := make(chan []Result)

	for i, chunk := range chunks {
		go removeThumbnails(chunk, resultChannel, i+1)
	}

	t := time.Now()
	tf := t.Format("20060102-15-04")
	outfile, err := os.Create("thumbnail-updated" + tf + ".tsv")
	if err != nil {
		panic(err)
	}
	defer outfile.Close()

	writer := bufio.NewWriter(outfile)
	for range chunks {
		for _, result := range <-resultChannel {
			writer.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\n", result.Time.Format("20060102T15:04:05"), result.Code, result.URI, result.Msg))
			writer.Flush()
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

		//request the digital object
		do, err := client.GetDigitalObject(doid.RepoID, doid.DOID)
		if err != nil {
			results = append(results, Result{"ERROR", do.URI, err.Error(), time.Now()})
			continue
		}

		//check that there are file versions
		if len(do.FileVersions) > 0 {

			//check for thumnbnails
			if do.ContainsUseStatement("image-thumbnail") == true {
				//delete any dos that only have a thumbnail
				if len(do.FileVersions) == 1 {
					response, err := client.DeleteDigitalObject(doid.RepoID, doid.DOID)
					if err != nil {
						results = append(results, Result{"ERROR", do.URI, err.Error(), time.Now()})
						continue
					}
					results = append(results, Result{"DELETED", do.URI, fmt.Sprintf("%s", strings.ReplaceAll(response, "\n", "")), time.Now()})
					continue
				}

				//update dos with more than one file versions
				do.FileVersions = updateFileVersions(do)
				response, err := client.UpdateDigitalObject(doid.RepoID, doid.DOID, do)
				if err != nil {
					results = append(results, Result{"ERROR", do.URI, err.Error(), time.Now()})
					continue
				}
				results = append(results, Result{"UPDATED", do.URI, fmt.Sprintf("%s", strings.ReplaceAll(response, "\n", "")), time.Now()})
				continue
			} else {
				results = append(results, Result{"SKIPPED", do.URI, fmt.Sprintf("No image-thumbnails in file versions"), time.Now()})
				continue
			}
		}
		results = append(results, Result{"SKIPPED", do.URI, "No file versions", time.Now()})
	}
	fmt.Printf("Worker %d finished\n", worker)
	resultChannel <- results
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
