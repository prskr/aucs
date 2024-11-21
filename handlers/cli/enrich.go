package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/CycloneDX/cyclonedx-go"
	"github.com/gojek/heimdall/v7"
	"github.com/gojek/heimdall/v7/hystrix"
	"github.com/prskr/aucs/core/ports"
	"github.com/prskr/aucs/infrastructure/checker"
	"github.com/prskr/aucs/infrastructure/checker/npm"
	"github.com/prskr/aucs/infrastructure/checker/nuget"
	"github.com/prskr/aucs/infrastructure/checker/pypi"
)

type EnrichCLiHandler struct {
	SBOMFile *os.File `arg:"" help:"SBOM file to enrich"`

	BOMFormat       BOMFileFormatFlag `name:"bom-format" help:"BOM file format" default:"json"`
	DB              DBFlag            `embed:"" prefix:"db."`
	Parallelism     uint8             `name:"parallelism" help:"Number of parallel requests" default:"20"`
	WriteBackToFile bool              `name:"write" help:"If aucs should write the SBOM to the source file - if not will be written to STDOUT" default:"false"`

	HttpClient struct {
		Timeout               time.Duration `name:"timeout" help:"HTTP client timeout" default:"30s"`
		HystrixTimeout        time.Duration `name:"hystrix-timeout" help:"Hystrix timeout" default:"30s"`
		MaxConcurrentRequests int           `name:"max-concurrent-requests" help:"Maximum concurrent requests" default:"100"`
		Retry                 struct {
			InitialTimeout time.Duration `name:"initial-timeout" help:"Initial retry timeout" default:"1s"`
			MaxTimeout     time.Duration `name:"max-timeout" help:"Maximum retry timeout" default:"10s"`
			ExponentFactor float64       `name:"exponent-factor" help:"Exponential backoff factor" default:"2"`
			MaximumJitter  time.Duration `name:"maximum-jitter" help:"Maximum retry jitter" default:"200ms"`
		} `embed:"" prefix:"retry."`
	} `embed:"" prefix:"http-client."`

	KV       ports.KeyValueStore `kong:"-"`
	Checkers *checker.Registry   `kong:"-"`
}

func (h *EnrichCLiHandler) Run(ctx context.Context, stdout ports.STDOUT) (err error) {
	defer func() {
		err = errors.Join(err, h.SBOMFile.Close(), h.KV.Close())
	}()

	decoder := cyclonedx.NewBOMDecoder(h.SBOMFile, h.BOMFormat.Format)
	bom := cyclonedx.NewBOM()

	if err := decoder.Decode(bom); err != nil {
		return fmt.Errorf("failed to decode BOM: %w", err)
	}

	var wg sync.WaitGroup
	wg.Add(len(*bom.Components))

	scanInput := make(chan *cyclonedx.Component, h.Parallelism)
	processCtx, cancel := context.WithCancel(ctx)

	for range h.Parallelism {
		go func() {
			for {
				select {
				case in := <-scanInput:
					h.processComponent(processCtx, in)
					wg.Done()
				case <-processCtx.Done():
					return
				}
			}
		}()
	}

	for _, c := range *bom.Components {
		scanInput <- &c
	}

	wg.Wait()
	cancel()

	out := stdout
	if h.WriteBackToFile {
		if _, err := h.SBOMFile.Seek(0, 0); err != nil {
			return fmt.Errorf("failed to seek to the beginning of the SBOM file: %w", err)
		}

		out = h.SBOMFile
	}

	encoder := cyclonedx.NewBOMEncoder(out, h.BOMFormat.Format)

	return encoder.Encode(bom)
}

func (h *EnrichCLiHandler) processComponent(ctx context.Context, in *cyclonedx.Component) {
	if info, err := h.Checkers.LatestVersionFor(ctx, in.PackageURL); err != nil {
		slog.WarnContext(ctx, "Failed to determine latest version for package", slog.String("package_url", in.PackageURL), slog.String("err", err.Error()))
	} else {
		slog.DebugContext(ctx, "Found latest package version",
			slog.String("package_url", in.PackageURL),
			slog.String("latest_version", info.LatestVersion),
			slog.String("current_version", in.Version),
		)

		if in.Properties == nil {
			*in.Properties = make([]cyclonedx.Property, 0)
		}
		*in.Properties = append(*in.Properties, cyclonedx.Property{Name: "aucs:package:latest_version", Value: info.LatestVersion})
	}
}

func (h *EnrichCLiHandler) AfterApply() error {
	if h.SBOMFile == nil {
		return errors.New("missing SBOM file")
	}

	if kv, err := h.DB.Open(); err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	} else {
		h.KV = kv
	}

	backoff := heimdall.NewExponentialBackoff(
		h.HttpClient.Retry.InitialTimeout,
		h.HttpClient.Retry.MaxTimeout,
		h.HttpClient.Retry.ExponentFactor,
		h.HttpClient.Retry.MaximumJitter,
	)

	retrier := heimdall.NewRetrier(backoff)

	h.Checkers = checker.NewRegistry(h.KV)
	h.Checkers.Register(
		nuget.NewChecker(h.heimdallClient("CheckLatestNugetVersion", retrier)),
		npm.NewChecker(h.heimdallClient("CheckLatestNPMVersion", retrier)),
		pypi.NewChecker(h.heimdallClient("CheckLatestPyPiVersion", retrier)),
	)

	return nil
}

func (h *EnrichCLiHandler) heimdallClient(commandName string, retier heimdall.Retriable) *http.Client {
	hystrixClient := hystrix.NewClient(
		hystrix.WithCommandName(commandName),
		hystrix.WithHTTPTimeout(h.HttpClient.Timeout),
		hystrix.WithCommandName("MyCommand"),
		hystrix.WithHystrixTimeout(h.HttpClient.HystrixTimeout),
		hystrix.WithMaxConcurrentRequests(h.HttpClient.MaxConcurrentRequests),
		hystrix.WithRetrier(retier),
	)

	return &http.Client{Transport: hystrixRoundtrip{client: hystrixClient}}
}

var _ http.RoundTripper = (*hystrixRoundtrip)(nil)

type hystrixRoundtrip struct {
	client *hystrix.Client
}

// RoundTrip implements http.RoundTripper.
func (h hystrixRoundtrip) RoundTrip(req *http.Request) (*http.Response, error) {
	return h.client.Do(req)
}
