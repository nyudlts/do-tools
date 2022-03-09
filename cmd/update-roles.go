package cmd

import (
	"bufio"
	"fmt"
	"github.com/nyudlts/go-aspace"
	"github.com/spf13/cobra"
	"os"
	"regexp"
	"strings"
	"time"
)

func init() {
	updateCmd.PersistentFlags().StringVarP(&config, "config", "c", "", "")
	updateCmd.PersistentFlags().StringVar(&env, "environment", "", "")
	rootCmd.AddCommand(updateCmd)
}

var (
	aeonMatcher    = regexp.MustCompile("aeon.library.nyu.edu")
	waybackMatcher = regexp.MustCompile("wayback.archive-it.org")
	cdlibMatcher   = regexp.MustCompile("webarchives.cdlib.org")
)

var updateCmd = &cobra.Command{
	Use: "update-roles",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("UPDATE ROLES")
		setClient()
		updateRoles()
	},
}

func updateRoles() {
	fmt.Println("Updating Roles")
	GetDOIDs()
	doChunks := getChunks(dos)
	resultChannel := make(chan []Result)

	for i, chunk := range doChunks {
		go updateRoleChunk(chunk, resultChannel, i+1)
	}

	t := time.Now()
	tf := t.Format("20060102T15:04")
	outfile, _ := os.Create("update-roles-" + tf + ".tsv")
	defer outfile.Close()
	writer := bufio.NewWriter(outfile)

	for range doChunks {
		for _, result := range <-resultChannel {
			writer.WriteString(result.String())
		}
	}
}

func updateRoleChunk(chunk []ObjectID, resultChannel chan []Result, workerID int) {
	results := []Result{}
	fmt.Printf("Worker %d is processing %d records\n", workerID, len(chunk))
	for i, oid := range chunk {
		if i > 2 && (i-1)%100 == 0 {
			fmt.Printf("Worker %d has completed %d records\n", workerID, i-1)
		}
		//get the do
		do, err := client.GetDigitalObject(oid.RepoID, oid.ObjectID)
		if err != nil {
			results = append(results, Result{Code: "ERROR", URI: oid.String(), Msg: err.Error(), Time: time.Time{}, Worker: workerID})
			continue
		}

		//check that there is at least one FV
		fileVersions := do.FileVersions
		if len(fileVersions) < 1 {
			results = append(results, Result{Code: "SKIPPED", URI: do.URI, Msg: "", Time: time.Time{}, Worker: workerID})
			continue
		}

		//check for blank roles
		if hasUndefinedFV(fileVersions) == true {
			do.FileVersions = updateFileVersionRoles(fileVersions)
			msg, err := client.UpdateDigitalObject(oid.RepoID, oid.ObjectID, do)
			if err != nil {
				results = append(results, Result{Code: "ERROR", URI: do.URI, Msg: err.Error(), Time: time.Time{}, Worker: workerID})
				continue
			}
			results = append(results, Result{Code: "UPDATED", URI: do.URI, Msg: strings.ReplaceAll(msg, "\n", ""), Time: time.Time{}, Worker: workerID})
			continue
		}
	}
	fmt.Printf("Worker %d completed %d records\n", workerID, len(chunk))
	resultChannel <- results
}

func updateFileVersionRoles(fvs []aspace.FileVersion) []aspace.FileVersion {
	newFvs := []aspace.FileVersion{}

	for _, fv := range fvs {
		if fv.UseStatement == "electronic-records-service" {
			fv.UseStatement = "electronic-records-reading-room"
			newFvs = append(newFvs, fv)
			continue
		}
		if fv.UseStatement == "" || fv.UseStatement == "service" {
			if aeonMatcher.MatchString(fv.FileURI) == true {
				fv.UseStatement = "electronic-records-reading-room"
				newFvs = append(newFvs, fv)
				continue
			}

			if waybackMatcher.MatchString(fv.FileURI) == true || cdlibMatcher.MatchString(fv.FileURI) == true {
				fv.UseStatement = "external-link"
				newFvs = append(newFvs, fv)
				continue
			}
		} else {
			newFvs = append(newFvs, fv)
		}
	}
	return newFvs
}

func hasUndefinedFV(fvs []aspace.FileVersion) bool {
	for _, fv := range fvs {
		if fv.UseStatement == "" || fv.UseStatement == "service" || fv.UseStatement == "electronic-records-service" {
			return true
		}
	}
	return false
}
