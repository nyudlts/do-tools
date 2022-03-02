package cmd

import (
	"bufio"
	"fmt"
	"github.com/nyudlts/go-aspace"
	"log"
	"os"
	"time"
)

var (
	client  *aspace.ASClient
	dos     = []DigitalObjectIDs{}
	workers = 12
	err     error
	config  string
)

type DigitalObjectIDs struct {
	RepoID int
	DOID   int
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func GetRoles(chunk []DigitalObjectIDs, resultsChannel chan map[string]int, worker int) {
	results := map[string]int{}
	fmt.Printf("Worker %d processing %d records\n", worker, len(chunk))
	for i, doid := range chunk {
		if i > 0 && i%100 == 0 {
			fmt.Printf("Worker %d has completed %d records\n", worker, i)
		}
		do, err := client.GetDigitalObject(doid.RepoID, doid.DOID)
		if err != nil {
			log.Println(err.Error())
		}
		fileVersions := do.FileVersions
		if len(fileVersions) > 0 {
			for _, fileVersion := range fileVersions {
				role := fileVersion.UseStatement
				if role == "" {
					role = "undefined"
				}
				if HasRole(results, role) == true {
					results[role] = results[role] + 1
				} else {
					results[role] = 1
				}
			}
		}
	}
	resultsChannel <- results
}

func PrintRoleMap(roles map[string]int) {
	for k, v := range roles {
		fmt.Printf("%s\t%d\n", k, v)
	}
}

func GenerateRoleReport(roles map[string]int) {
	t := time.Now()
	tf := t.Format("20060102-15-04")
	outfile, err := os.Create("roles-report-" + tf + ".tsv")
	if err != nil {
		panic(err)
	}
	defer outfile.Close()

	writer := bufio.NewWriter(outfile)
	for k, v := range roles {
		writer.WriteString(fmt.Sprintf("%s\t%d\n", k, v))
	}
	writer.Flush()
}

func HasRole(roles map[string]int, role string) bool {
	for k, _ := range roles {
		if k == role {
			return true
		}
	}
	return false
}

func getChunks(doids []DigitalObjectIDs) [][]DigitalObjectIDs {
	var divided [][]DigitalObjectIDs

	chunkSize := (len(doids) + workers - 1) / workers

	for i := 0; i < len(doids); i += chunkSize {
		end := i + chunkSize

		if end > len(doids) {
			end = len(doids)
		}

		divided = append(divided, doids[i:end])
	}
	return divided
}

func GetDOIDs() {
	for _, repoId := range []int{2, 3, 6} {
		doIDs, err := client.GetDigitalObjectIDs(repoId)
		if err != nil {
			panic(err)
		}
		for _, digitalObjectID := range doIDs {
			dos = append(dos, DigitalObjectIDs{repoId, digitalObjectID})
		}
	}
}
