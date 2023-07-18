package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/nyudlts/go-aspace"
	"github.com/spf13/cobra"
	"os"
	"regexp"
	"strings"
	"time"
)

var handlePtn = regexp.MustCompile("^https://hdl.handle.net")
var aoPtn = regexp.MustCompile("/repositories/7/archival_objects/[0-9]+")

type DORef struct {
	URI   string
	Index int
}

func init() {
	vlpCmd.PersistentFlags().StringVar(&config, "config", "", "")
	vlpCmd.PersistentFlags().StringVar(&env, "environment", "", "")
	vlpCmd.PersistentFlags().BoolVar(&test, "test", false, "")
	rootCmd.AddCommand(vlpCmd)
}

var vlpCmd = &cobra.Command{
	Use: "remove-vlp-dos",
	Run: func(cmd *cobra.Command, args []string) {
		setClient()
		if err = removeVLP(); err != nil {
			panic(err)
		}
	},
}

func removeVLP() error {
	aos, err := getAOs()
	if err != nil {
		return err
	}
	aoChunks := getChunks(aos)
	resultChannel := make(chan []Result)

	for i, chunk := range aoChunks {
		go removeAOChunk(chunk, resultChannel, i+1)
	}

	t := time.Now()
	tf := t.Format("20060102-030405")
	var outfile *os.File
	if test {
		outfile, _ = os.Create("remove-vlp-dos" + env + "-TEST-" + tf + ".tsv")

	} else {
		outfile, _ = os.Create("remove-vlp-dos" + env + "-" + tf + ".tsv")
	}
	defer outfile.Close()

	writer := bufio.NewWriter(outfile)
	writer.WriteString("timestamp\tworkerID\tresult\taspaceURL\tmessage\n")
	writer.Flush()

	for range aoChunks {
		for _, result := range <-resultChannel {
			writer.WriteString(fmt.Sprintf("%s\t%d\t%s\t%s\t%s\n", result.Time.Format(time.RFC3339), result.Worker, result.Code, result.URI, result.Msg))
			writer.Flush()
		}
	}

	return nil
}

func removeAOChunk(aoChunk []ObjectID, resultChannel chan []Result, workerID int) {
	results := []Result{}
	fmt.Printf("Worker %d is processing %d records\n", workerID, len(aoChunk))
	for i, obj := range aoChunk {
		if i > 0 && i%50 == 0 {
			fmt.Printf("Worker %d completed %d of %d records\n", workerID, i, len(aoChunk))
		}
		ao, err := client.GetArchivalObject(obj.RepoID, obj.ObjectID)
		if err != nil {
			results = append(results, Result{
				Code:   "ERROR",
				URI:    fmt.Sprintf("/repositories/%d/archival_objects/%d", obj.RepoID, obj.ObjectID),
				Msg:    strings.ReplaceAll(err.Error(), "\n", ""),
				Time:   time.Time{},
				Worker: workerID,
			})
			continue
		}

		msg, err := removeAO(ao, obj.ObjectID)
		if err != nil {
			results = append(results, Result{
				Code:   "ERROR",
				URI:    fmt.Sprintf("/repositories/%d/archival_objects/%d", obj.RepoID, obj.ObjectID),
				Msg:    strings.ReplaceAll(err.Error(), "\n", ""),
				Time:   time.Time{},
				Worker: workerID,
			})
			continue
		}

		results = append(results, Result{
			Code:   "SUCCESS",
			URI:    ao.URI,
			Msg:    msg,
			Time:   time.Time{},
			Worker: workerID,
		})
	}
	resultChannel <- results
}

func removeAO(ao aspace.ArchivalObject, aoID int) (string, error) {
	DORefs := []DORef{}

	//iterate through the instances
	for i, instance := range ao.Instances {
		//check if the instance is a digital object
		if instance.InstanceType == "digital_object" {
			//iterate through the digital object map
			for _, doURI := range instance.DigitalObject {
				//check for aeon link objects
				res, err := hasHandleLinks(doURI)
				if err != nil {
					return "", err
				}
				if res {
					DORefs = append(DORefs, DORef{URI: doURI, Index: i})
				}
			}
		}
	}

	if len(DORefs) < 1 {
		return fmt.Sprintf("No handles found in found in %s", ao.URI), nil
	}

	if len(DORefs) > 1 {
		return fmt.Sprintf("Multiple handles found in found in %s, SKIPPING", ao.URI), nil
	}

	msg, err := unlinkDO(7, aoID, ao, 0)
	if err != nil {
		return msg, err
	}

	return msg, nil

}

func unlinkDO(repoID int, aoID int, ao aspace.ArchivalObject, i int) (string, error) {
	//remove the instance from instance slice
	oLength := len(ao.Instances)
	ao.Instances = append(ao.Instances[:i], ao.Instances[i+1:]...)
	nLength := len(ao.Instances)

	//check that the instance was removed
	if nLength != oLength-1 {
		return "", fmt.Errorf("%d is not equal to %d -1", nLength, oLength)
	}

	if !test {
		msg, err := client.UpdateArchivalObject(repoID, aoID, ao)
		if err != nil {
			return "", err
		}
		return msg, nil
	}

	return fmt.Sprintf("TEST MODE SKIPPING %s", ao.URI), nil

}

func hasHandleLinks(doURI string) (bool, error) {
	repoID, doID, err := aspace.URISplit(doURI)
	if err != nil {
		return false, err
	}

	do, err := client.GetDigitalObject(repoID, doID)
	if err != nil {
		return false, err
	}

	uri := do.FileVersions[0].FileURI
	if len(do.FileVersions) == 1 && handlePtn.MatchString(uri) {
		return true, nil
	}

	return false, nil
}

func getAOs() ([]ObjectID, error) {
	var objects = []ObjectID{}

	tree, err := client.GetResourceTree(7, 3379)
	if err != nil {
		return objects, err
	}

	jBytes, err := json.Marshal(tree)
	if err != nil {
		return objects, err
	}

	aos := aoPtn.FindAll(jBytes, -1)

	for _, ao := range aos {
		repoID, aoID, _ := aspace.URISplit(string(ao))
		objects = append(objects, ObjectID{repoID, aoID})
	}

	return objects, nil
}
