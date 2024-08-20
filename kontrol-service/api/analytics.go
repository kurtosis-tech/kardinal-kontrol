package api

import (
	"github.com/segmentio/analytics-go/v3"
	"github.com/sirupsen/logrus"
)

// AnalyticsWrapper wraps the Segment analytics client
type AnalyticsWrapper struct {
	client *analytics.Client
}

type AnalyticsEvent string

const (
	EVENT_FLOW_CREATE AnalyticsEvent = "FLOW_CREATE"
	EVENT_FLOW_DELETE AnalyticsEvent = "FLOW_DELETE"
	EVENT_DEPLOY      AnalyticsEvent = "DEPLOY"
)

// NewAnalyticsWrapper creates a new AnalyticsWrapper
func NewAnalyticsWrapper(isDevMode bool) *AnalyticsWrapper {
	if !isDevMode {
		// This is the Segment write key for the "kontrol-service" project. It is not
		// a sensitive value, but it could be extracted to an env var in the future
		// to separate dev and prod traffic if desired.
		client := analytics.New("1TeZVRY3ta9VYaNknKTCCKBZtcBllE6U")
		logrus.Info("Segment analytics client initialized")
		return &AnalyticsWrapper{client: &client}
	}
	logrus.Info("Dev mode: Segment analytics client not initialized")
	return &AnalyticsWrapper{client: nil}
}

// TrackEvent sends an analytics event to Segment
func (aw *AnalyticsWrapper) TrackEvent(event AnalyticsEvent, tenantUuid string) {
	if aw.client == nil {
		logrus.Infof(
			"Analytics client not initialized, skipping track event '%s' for uuid '%s'",
			event,
			tenantUuid,
		)
		return
	}
	logrus.Infof("Track event '%s' for uuid '%s'", event, tenantUuid)
	err := (*aw.client).Enqueue(analytics.Track{
		Event:      string(event),
		UserId:     tenantUuid,
		Properties: analytics.NewProperties(),
	})
	if err != nil {
		logrus.WithError(err).Error("Failed to enqueue analytics event")
	}
}

// Close closes the analytics client if it was initialized
func (aw *AnalyticsWrapper) Close() {
	if aw.client != nil {
		(*aw.client).Close()
	}
}