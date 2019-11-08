package job

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"strings"
	"time"

	"github.com/pierreprinetti/go-junit"
)

const (
	baseURL = "https://gcsweb-ci.svc.ci.openshift.org/gcs/origin-ci-test/logs"
)

type Job struct {
	Name, ID   string
	StartedAt  time.Time
	FinishedAt time.Time
	Result     string

	client *http.Client
}

func (j Job) Tests(ctx context.Context) (junit.TestSuite, error) {
	var testSuite junit.TestSuite

	req, err := http.NewRequest(http.MethodGet, baseURL+"/"+j.Name+"/"+j.ID+"/artifacts/e2e-openstack-serial/junit/junit_e2e_20191108-114348.xml", nil)
	if err != nil {
		return testSuite, err
	}

	res, err := j.client.Do(req.WithContext(ctx))
	if err != nil {
		return testSuite, err
	}

	err = xml.NewDecoder(res.Body).Decode(&testSuite)
	if err != nil {
		return testSuite, err
	}

	return testSuite, err
}

func Fetch(jobName, jobID string) (Job, error) {
	j := Job{
		Name:   jobName,
		ID:     jobID,
		client: new(http.Client),
	}

	// Get start metadata
	{
		req, err := http.NewRequest(http.MethodGet, baseURL+"/"+j.Name+"/"+j.ID+"/started.json", nil)
		if err != nil {
			return j, err
		}

		res, err := j.client.Do(req)
		if err != nil {
			return j, err
		}

		var started metadata
		if err := json.NewDecoder(res.Body).Decode(&started); err != nil {
			return j, err
		}

		j.StartedAt = started.time
	}

	// Get finish metadata
	{
		req, err := http.NewRequest(http.MethodGet, baseURL+"/"+jobName+"/"+jobID+"/finished.json", nil)
		if err != nil {
			return j, err
		}

		res, err := j.client.Do(req)
		if err != nil {
			return j, err
		}

		var finished metadata
		if err := json.NewDecoder(res.Body).Decode(&finished); err != nil {
			return j, err
		}

		j.FinishedAt = finished.time
		j.Result = finished.result
	}

	return j, nil
}

// String implements fmt.Stringer
func (j Job) String() string {
	return strings.Join([]string{
		j.ID,                                   // ID
		j.StartedAt.String(),                   // Started
		j.FinishedAt.Sub(j.StartedAt).String(), // Duration
		j.Result,                               // Result
	}, "\t")

}
