package cmd

import (
	"fmt"
	"testing"

	"github.com/nyudlts/go-aspace"
)

func TestDORefresh(t *testing.T) {
	var args []string
	aoURI := "/repositories/3/archival_objects/512700"

	setClient()

	ao, err := client.GetArchivalObjectFromURI(aoURI)
	if err != nil {
		t.Errorf("before test: could not get Archival Object")
		t.FailNow()
	}

	doURIs, err := client.GetDigitalObjectIDsForArchivalObjectFromURI(aoURI)
	if err != nil {
		t.Errorf("before test: could not get Digital Object URIs")
		t.FailNow()
	}

	want := ""
	for _, doURI := range doURIs {
		do, err := client.GetDigitalObjectFromURI(doURI)
		if err != nil {
			t.Errorf("before test: could not get Digital Object %s", doURI)
		}

		if do.Title != ao.Title {
			t.Errorf("before test: Digital Object title does not match Archival Object title: %s != %s", do.Title, ao.Title)
			t.FailNow()
		}
		want += fmt.Sprintf("updated ao: %s do: %s\n", aoURI, doURI)
	}

	saveAOTitle := ao.Title

	// change the title of the Archival Object
	ao.Title = "waffles are delicious"
	repoID, aoObjectID, err := aspace.URISplit(aoURI)
	if err != nil {
		t.Errorf("before test: could not split AO URI: %s", aoURI)
		t.FailNow()
	}

	body, err := client.UpdateArchivalObject(repoID, aoObjectID, ao)
	if err != nil {
		t.Errorf("%v, %s", err, body)
	}

	// RUN THE COMMAND
	// refresh the Digital Objects
	setCmdFlag(doRefreshCmd, aoFlags.URI, aoURI)
	got, errOut, err := CaptureCmdStdoutStderrE(doRefresh, doRefreshCmd, args)
	if err != nil {
		t.Errorf("doRefresh: %v, %v", err, errOut)
	}

	if want != got {
		t.Errorf("wanted: %s\n got: %s\n", want, got)
	}

	// reset the title of the Archival Object
	// need to fetch the AO again or else we get a stale object error
	ao, err = client.GetArchivalObjectFromURI(aoURI)
	if err != nil {
		t.Errorf("post test: could not get Archival Object")
	}

	ao.Title = saveAOTitle
	body, err = client.UpdateArchivalObject(repoID, aoObjectID, ao)
	if err != nil {
		t.Errorf("during AO restore: %v, %s", err, body)
	}

	// reset the title of the Digital Object(s)
	for _, doURI := range doURIs {
		do, err := client.GetDigitalObjectFromURI(doURI)
		if err != nil {
			t.Errorf("post test: could not get Digital Object %s", doURI)
		}

		do.Title = ao.Title
		repoID, doObjectID, err := aspace.URISplit(doURI)
		if err != nil {
			t.Errorf("post test: could not split DO URI: %s", doURI)
			t.FailNow()
		}

		body, err := client.UpdateDigitalObject(repoID, doObjectID, do)

		if err != nil {
			t.Errorf("post test: %v : %s", err, body)
		}
	}
}
