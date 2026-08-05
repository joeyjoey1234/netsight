package export

import (
	"encoding/json"
	"fmt"
	"netsight/internal/model"
	"os"
	"time"
)

func ExportJSON(scanID string, devices []*model.Device, links []*model.Link, findings []*model.Finding) (string, error) {
	outputPath := fmt.Sprintf("netsight-scan-%s-%s.json", scanID[:8], time.Now().Format("150405"))

	type exportData struct {
		ScanID    string           `json:"scanId"`
		Timestamp time.Time        `json:"timestamp"`
		Devices   []*model.Device  `json:"devices"`
		Links     []*model.Link    `json:"links"`
		Findings  []*model.Finding `json:"findings"`
	}

	data := exportData{
		ScanID:    scanID,
		Timestamp: time.Now(),
		Devices:   devices,
		Links:     links,
		Findings:  findings,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON marshal failed: %w", err)
	}

	if err := os.WriteFile(outputPath, jsonData, 0644); err != nil {
		return "", fmt.Errorf("failed to write JSON: %w", err)
	}

	return outputPath, nil
}
