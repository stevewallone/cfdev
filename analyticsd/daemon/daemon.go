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
	"":                      "unspecified",
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
	Buildpack       string
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
	NextURL   *string `json:"next_url"`
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
		"audit.app.create":              "app created",
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
		nextURL   *string = nil
		resources []Resource
		fetch     = func(params url.Values) error {
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

		cmd := CreateResponseCommand(resource , isTimestampSet , d , eventType , t )
		err = cmd.HandleResponse()
		if err != nil {
			return err
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

type ResponseCommand interface {
	HandleResponse() error
}

type AppCreatedCmd struct {
	resource Resource
	isTimestampSet bool
	d *Daemon
	eventType string
	t time.Time
}

func(ac *AppCreatedCmd) HandleResponse() error {
	buildpack, ok := buildpackWhitelist[ac.resource.Entity.Metadata.Request.Buildpack]
	if !ok {
		buildpack = "custom"
	}
	var properties = analytics.Properties{
		"buildpack": buildpack,
		"os":        runtime.GOOS,
		"version":   ac.d.version,
	}

	var err error

	if ac.isTimestampSet {
		err = ac.d.analyticsClient.Enqueue(analytics.Track{
			UserId:     ac.d.UUID,
			Event:      ac.eventType,
			Timestamp:  ac.t,
			Properties: properties,
		})
	}

	if err != nil {
		return fmt.Errorf("failed to send analytics: %v", err)
	}

	return nil
}

type ServiceCreatedCmd struct {
	resource Resource
	isTimestampSet bool
	d *Daemon
	eventType string
	t time.Time
}

func(sc *ServiceCreatedCmd) HandleResponse() error {
	var servicePlanResponse ServicePlanResponse
	servicePlanEndpoint := "/v2/service_plans/" + sc.resource.Entity.Metadata.Request.ServicePlanGUID
	err := sc.d.fetch(servicePlanEndpoint, nil, &servicePlanResponse)
	if err != nil {
		return err
	}

	serviceGUID := servicePlanResponse.ServicePlanEntity.ServicePlanGUID

	var serviceResponse ServiceResponse
	serviceEndpoint := "/v2/services/" + serviceGUID
	err = sc.d.fetch(serviceEndpoint, nil, &serviceResponse)
	if err != nil {
		return err
	}
	serviceType := serviceResponse.ServiceEntity.ServiceLabel

	var properties = analytics.Properties{
		"service": serviceType,
		"os":      runtime.GOOS,
		"version": sc.d.version,
	}

	if sc.isTimestampSet {
		err = sc.d.analyticsClient.Enqueue(analytics.Track{
			UserId:     sc.d.UUID,
			Event:      sc.eventType,
			Timestamp:  sc.t,
			Properties: properties,
		})
	}

	if err != nil {
		return fmt.Errorf("failed to send analytics: %v", err)
	}

	return nil
}

func CreateResponseCommand(resource Resource, isTimestampSet bool, d *Daemon, eventType string, t time.Time) ResponseCommand {
	switch eventType {
	case "app created":
		return  &AppCreatedCmd {
			resource: resource,
			isTimestampSet: isTimestampSet,
			d: d,
			eventType: eventType,
			t: t,
		}
	case "service created":
		return &ServiceCreatedCmd {
			resource: resource,
			isTimestampSet: isTimestampSet,
			d: d,
			eventType: eventType,
			t: t,
		}
	}

	return nil
}
