package geoip

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/nixguard/nixguard/internal/domain/firewall"
)

const (
	geoIPBaseURL = "https://download.maxmind.com/app/geoip_download"
	dbEdition    = "GeoLite2-Country-CSV"
)

// MaxMindProvider implements firewall.GeoIPProvider using MaxMind GeoLite2 CSV data.
type MaxMindProvider struct {
	licenseKey string
	dataDir    string
	log        *slog.Logger

	mu          sync.RWMutex
	countryMap  map[string][]string // country code → []CIDR
	lastUpdated time.Time
}

func NewMaxMindProvider(licenseKey, dataDir string, log *slog.Logger) *MaxMindProvider {
	return &MaxMindProvider{
		licenseKey: licenseKey,
		dataDir:    dataDir,
		log:        log,
		countryMap: make(map[string][]string),
	}
}

func (p *MaxMindProvider) Resolve(ctx context.Context, countryCode string) ([]string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	code := strings.ToUpper(countryCode)
	ranges, ok := p.countryMap[code]
	if !ok {
		return nil, nil
	}
	return ranges, nil
}

func (p *MaxMindProvider) Update(ctx context.Context) error {
	if p.licenseKey == "" {
		return fmt.Errorf("GeoIP license key not configured")
	}

	if err := os.MkdirAll(p.dataDir, 0750); err != nil {
		return fmt.Errorf("create geoip dir: %w", err)
	}

	archivePath := filepath.Join(p.dataDir, "GeoLite2-Country-CSV.tar.gz")
	if err := p.downloadDatabase(ctx, archivePath); err != nil {
		return fmt.Errorf("download geoip: %w", err)
	}

	countryMap, err := p.parseDatabase(archivePath)
	if err != nil {
		return fmt.Errorf("parse geoip: %w", err)
	}

	p.mu.Lock()
	p.countryMap = countryMap
	p.lastUpdated = time.Now().UTC()
	p.mu.Unlock()

	p.log.Info("GeoIP database updated",
		slog.Int("countries", len(countryMap)),
		slog.Time("updated_at", p.lastUpdated),
	)
	return nil
}

func (p *MaxMindProvider) LastUpdated(ctx context.Context) (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.lastUpdated.IsZero() {
		return "never", nil
	}
	return p.lastUpdated.Format(time.RFC3339), nil
}

func (p *MaxMindProvider) downloadDatabase(ctx context.Context, destPath string) error {
	url := fmt.Sprintf("%s?edition_id=%s&license_key=%s&suffix=tar.gz",
		geoIPBaseURL, dbEdition, p.licenseKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("MaxMind API returned %d", resp.StatusCode)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func (p *MaxMindProvider) parseDatabase(archivePath string) (map[string][]string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	// Step 1: Parse country locations to build geoname_id → country_code map
	geonameToCountry := make(map[string]string)
	countryMap := make(map[string][]string)

	tr := tar.NewReader(gz)
	var blocksV4Data, blocksV6Data []byte

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		name := filepath.Base(header.Name)
		switch name {
		case "GeoLite2-Country-Locations-en.csv":
			data, _ := io.ReadAll(tr)
			reader := csv.NewReader(strings.NewReader(string(data)))
			records, _ := reader.ReadAll()
			for i, rec := range records {
				if i == 0 || len(rec) < 5 {
					continue
				}
				geonameID := rec[0]
				countryCode := rec[4]
				if countryCode != "" {
					geonameToCountry[geonameID] = countryCode
				}
			}
		case "GeoLite2-Country-Blocks-IPv4.csv":
			blocksV4Data, _ = io.ReadAll(tr)
		case "GeoLite2-Country-Blocks-IPv6.csv":
			blocksV6Data, _ = io.ReadAll(tr)
		}
	}

	// Step 2: Parse blocks files
	for _, data := range [][]byte{blocksV4Data, blocksV6Data} {
		if len(data) == 0 {
			continue
		}
		reader := csv.NewReader(strings.NewReader(string(data)))
		records, _ := reader.ReadAll()
		for i, rec := range records {
			if i == 0 || len(rec) < 2 {
				continue
			}
			cidr := rec[0]
			geonameID := rec[1]
			if geonameID == "" && len(rec) > 2 {
				geonameID = rec[2] // registered_country_geoname_id
			}
			if code, ok := geonameToCountry[geonameID]; ok {
				countryMap[code] = append(countryMap[code], cidr)
			}
		}
	}

	return countryMap, nil
}

// StartAutoUpdater runs periodic GeoIP database updates.
func (p *MaxMindProvider) StartAutoUpdater(ctx context.Context, interval time.Duration) {
	// Try initial load
	if err := p.Update(ctx); err != nil {
		p.log.Warn("initial GeoIP update failed", slog.String("error", err.Error()))
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := p.Update(ctx); err != nil {
					p.log.Warn("GeoIP update failed", slog.String("error", err.Error()))
				}
			}
		}
	}()
}

// Compile-time interface check
var _ firewall.GeoIPProvider = (*MaxMindProvider)(nil)
