package daemon

//go:generate mockgen -package mocks -destination mocks/analytics.go gopkg.in/segmentio/analytics-go.v3 Client

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"

	"gopkg.in/segmentio/analytics-go.v3"
)

const ccTimeStampFormat = "2006-01-02T15:04:05Z"

type Daemon struct {
	ccHost          string
	httpClient      *http.Client
	UUID            string
	version         string
	analyticsClient analytics.Client
	ticker          *time.Ticker
	pollingInterval time.Duration
	logger          *log.Logger
	lastTime        *time.Time
	doneChan        chan bool
}

var buildpackWhitelist = map[string]string{
	"staticfile_buildpack":  "staticfile",
	"java_buildpack":        "java",
	"ruby_buildpack":        "ruby",
	"dotnet_core_buildpack": "dotnet_core",
	"nodejs_buildpack":      "nodejs",
	"go_buildpack":          "go",
	"python_buildpack":      "python",
	"php_buildpack":         "php",
	"binary_buildpack":      "binary",
	"": "unspecified",
}

func New(
	ccHost string,
	UUID string,
	version string,
	writer io.Writer,
	httpClient *http.Client,
	analyticsClient analytics.Client,
	pollingInterval time.Duration,
) *Daemon {
	return &Daemon{
		ccHost:          ccHost,
		UUID:            UUID,
		version:         version,
		httpClient:      httpClient,
		analyticsClient: analyticsClient,
		ticker:          time.NewTicker(pollingInterval),
		pollingInterval: pollingInterval,
		logger:          log.New(writer, "[ANALYTICSD] ", log.LstdFlags),
		doneChan:        make(chan bool, 1),
	}
}

type Request struct {
	Buildpack string
	ServicePlanGUID string `json:"service_plan_guid"`
}

type Metadata struct {
	Request Request
}

type Entity struct {
	Type      string
	Timestamp string
	Metadata  Metadata
}

type Resource struct {
	Entity Entity
}

type Response struct {
	NextURL *string `json:"next_url"`
	Resources []Resource
}

type ServicePlanResponse struct {
	ServicePlanEntity ServicePlanEntity `json:"entity"`
}

type ServicePlanEntity struct {
	ServicePlanGUID string `json:"service_guid"`
}

type ServiceResponse struct {
	ServiceEntity ServiceEntity `json:"entity"`
}

type ServiceEntity struct {
	ServiceLabel string `json:"label"`
}

var (
	eventTypes = map[string]string{
		"audit.app.create": "app created",
		"audit.service_instance.create": "service created",
	}
)

func (d *Daemon) Start() {
	err := d.do(false)
	if err != nil {
		d.logger.Println(err)
	}
	for {
		select {
		case <-d.doneChan:
			return
		case <-time.NewTicker(d.pollingInterval).C:
			isTimestampSet := d.lastTime != nil
			err := d.do(isTimestampSet)

			if err != nil {
				d.logger.Println(err)
			}
		}
	}
}

func (d *Daemon) Stop() {
	d.doneChan <- true
}

func (d *Daemon) do(isTimestampSet bool) error {
	var (
		nextURL *string = nil
		resources []Resource
		fetch = func(params url.Values) error {
			var appResponse Response
			err := d.fetch("/v2/events", params, &appResponse)
			if err != nil {
				return err
			}

			resources = append(resources, appResponse.Resources...)
			nextURL = appResponse.NextURL
			return nil
		}
	)

	params := url.Values{}
	params.Add("q", "type IN "+eventTypesFilter())
	if isTimestampSet {
		params.Add("q", "timestamp>"+d.lastTime.Format(ccTimeStampFormat))
	}
	err := fetch(params)
	if err != nil {
		return err
	}

	for nextURL != nil {
		t, err := url.Parse(*nextURL)
		if err != nil {
			return fmt.Errorf("failed to parse params out of %s: %s", nextURL, err)
		}

		err = fetch(t.Query())
		if err != nil {
			return err
		}
	}

	if len(resources) == 0 {
		d.saveLatestTime(time.Now())
	}

	for _, resource := range resources {
		eventType, ok := eventTypes[resource.Entity.Type]
		if !ok {
			continue
		}

		t, err := time.Parse(time.RFC3339, resource.Entity.Timestamp)
		if err != nil {
			return err
		}

		d.saveLatestTime(t)

		switch eventType {
		case "app created":
			buildpack, ok := buildpackWhitelist[resource.Entity.Metadata.Request.Buildpack]
			if !ok {
				buildpack = "custom"
			}
			var properties = analytics.Properties{
				"buildpack": buildpack,
				"os":        runtime.GOOS,
				"version":   d.version,
			}

			if isTimestampSet {
				err = d.analyticsClient.Enqueue(analytics.Track{
					UserId:     d.UUID,
					Event:      eventType,
					Timestamp:  t,
					Properties: properties,
				})
			}

			if err != nil {
				return fmt.Errorf("failed to send analytics: %v", err)
			}

		case "service created":
			var servicePlanResponse ServicePlanResponse
			servicePlanEndpoint := "/v2/service_plans/" + resource.Entity.Metadata.Request.ServicePlanGUID
			os.Mkdir("/tmp/ServicePlanEndpoint:" + servicePlanEndpoint, 0777)
			err := d.fetch(servicePlanEndpoint, nil, servicePlanResponse)
			if err != nil {
				os.Mkdir("/tmp/" + err.Error(), 0777)
				return err
			}
			serviceGUID := servicePlanResponse.ServicePlanEntity.ServicePlanGUID
			os.Mkdir("/tmp/second-api-call-successful", 0777)

			var serviceResponse ServiceResponse
			serviceEndpoint := "/v2/services/" + serviceGUID
			err = d.fetch(serviceEndpoint, nil, serviceResponse)
			if err != nil {
				return err
			}
			serviceType := serviceResponse.ServiceEntity.ServiceLabel
			os.Mkdir("/tmp/third-api-call-successful", 0777)


			var properties = analytics.Properties{
				"service":	serviceType,
				"os":        runtime.GOOS,
				"version":   d.version,
			}

			if isTimestampSet {
				os.Mkdir("/tmp/sending-service-create", 0777)
				err = d.analyticsClient.Enqueue(analytics.Track{
					UserId:     d.UUID,
					Event:      eventType,
					Timestamp:  t,
					Properties: properties,
				})
			}

			if err != nil {
				return fmt.Errorf("failed to send analytics: %v", err)
			}

		}
	}
	return nil
}
func (d *Daemon) fetch(apiEndpoint string, params url.Values, dest interface{}) error {
	req, err := http.NewRequest(http.MethodGet, d.ccHost+apiEndpoint, nil)
	if err != nil {
		return err
	}

	req.URL.RawQuery = params.Encode()

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to query cloud controller: %s", err)
	}

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		var properties = analytics.Properties{
			"message": fmt.Sprintf("failed to contact cc api: [%v] %s", resp.Status, contents),
			"os":      runtime.GOOS,
			"version": d.version,
		}

		err := d.analyticsClient.Enqueue(analytics.Track{
			UserId:     d.UUID,
			Event:      "analytics error",
			Timestamp:  time.Now().UTC(),
			Properties: properties,
		})

		if err != nil {
			return fmt.Errorf("failed to send analytics: %v", err)
		}

		//think about logging error anyway if failed to contact cc
		//instead of return nil
		return nil
	}

	return json.Unmarshal(contents, dest)
}

func eventTypesFilter() string {
	var coll []string
	for k, _ := range eventTypes {
		coll = append(coll, k)
	}
	return strings.Join(coll, ",")
}

func (d *Daemon) saveLatestTime(t time.Time) {
	t = t.UTC()
	if d.lastTime == nil || t.After(*d.lastTime) {
		d.lastTime = &t
	}
}
