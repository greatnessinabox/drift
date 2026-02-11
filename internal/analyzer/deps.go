package analyzer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/mod/modfile"
)

type DepStatus struct {
	Module         string
	CurrentVersion string
	LatestVersion  string
	StaleDays      int
	Status         string // "current", "stale", "outdated"
}

func analyzeDeps(root string) ([]DepStatus, error) {
	gomodPath := filepath.Join(root, "go.mod")
	data, err := os.ReadFile(gomodPath)
	if err != nil {
		return nil, fmt.Errorf("reading go.mod: %w", err)
	}

	f, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return nil, fmt.Errorf("parsing go.mod: %w", err)
	}

	var results []DepStatus

	for _, req := range f.Require {
		if req.Indirect {
			continue
		}

		dep := DepStatus{
			Module:         shortModuleName(req.Mod.Path),
			CurrentVersion: req.Mod.Version,
		}

		latest, latestTime, err := fetchLatestVersion(req.Mod.Path)
		if err != nil {
			dep.Status = "unknown"
			dep.LatestVersion = "?"
		} else {
			dep.LatestVersion = latest
			if dep.CurrentVersion == latest {
				dep.Status = "current"
				dep.StaleDays = 0
			} else {
				staleDays := int(time.Since(latestTime).Hours() / 24)
				dep.StaleDays = staleDays
				if staleDays > 90 {
					dep.Status = "outdated"
				} else {
					dep.Status = "stale"
				}
			}
		}

		results = append(results, dep)
	}

	return results, nil
}

func shortModuleName(mod string) string {
	parts := strings.Split(mod, "/")
	if len(parts) <= 1 {
		return mod
	}
	return parts[len(parts)-1]
}

type proxyInfo struct {
	Version string    `json:"Version"`
	Time    time.Time `json:"Time"`
}

func fetchLatestVersion(module string) (string, time.Time, error) {
	url := fmt.Sprintf("https://proxy.golang.org/%s/@latest", module)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", time.Time{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", time.Time{}, fmt.Errorf("proxy returned %d", resp.StatusCode)
	}

	var info proxyInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", time.Time{}, err
	}

	return info.Version, info.Time, nil
}
