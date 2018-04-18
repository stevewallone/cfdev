package cfanalytics

import (
	"runtime"

	analytics "gopkg.in/segmentio/analytics-go.v3"
)

type ClientInterface interface {
	Enqueue(message analytics.Message) error
}

type AnalyticsEvent struct {
	SegClient ClientInterface
	Event     string
	UserId    string
	Data      interface{}
	OS        string
	Version   string
}

func (s *AnalyticsEvent) SendAnalytics() error {
	properties := analytics.NewProperties()
	properties.Set("os", runtime.GOOS)
	properties.Set("version", s.Version)
	properties.Set("data", s.Data)

	analyticsTrack := analytics.Track{
		UserId:     s.UserId,
		Event:      s.Event,
		Properties: properties,
	}

	s.SegClient.Enqueue(analyticsTrack)
	return nil
}
