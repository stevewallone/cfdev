package main

import (
	"code.cloudfoundry.org/analyticsd/daemon"
	"context"
	"crypto/tls"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"gopkg.in/segmentio/analytics-go.v3"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var (
		analyticsKey = os.Getenv("ANALYTICSD_API_KEY")
		userID       = os.Getenv("ANALYTICSD_USER_ID")
	)

	cfg := &clientcredentials.Config{
		ClientID:     "",
		ClientSecret: "",
		TokenURL:     "https://uaa.dev.cfdev.sh/oauth/token",
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)

	analyticsDaemon := daemon.New(
		"https://api.dev.cfdev.sh",
		userID,
		os.Stdout,
		cfg.Client(ctx),
		analytics.New(analyticsKey),
		10*time.Second, //<-- TODO decide on polling interval
	)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		analyticsDaemon.Stop()
	}()

	analyticsDaemon.Start()
}