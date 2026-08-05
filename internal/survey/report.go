package survey

import (
	"fmt"
	"strings"
	"time"

	"netsight/internal/model"
)

type SurveyReport struct {
	ScanID       string
	Preset       string
	Subnet       string
	Duration     time.Duration
	StartTime    time.Time
	DevicesFound int
	Findings     []*model.Finding
	Sections     []ReportSection
}

type ReportSection struct {
	Title   string
	Content string
}

func GenerateReport(scanID string, preset string, subnet string, duration time.Duration, startTime time.Time, devices []*model.Device, findings []*model.Finding) *SurveyReport {
	report := &SurveyReport{
		ScanID:       scanID,
		Preset:       preset,
		Subnet:       subnet,
		Duration:     duration,
		StartTime:    startTime,
		DevicesFound: len(devices),
		Findings:     findings,
	}

	report.Sections = append(report.Sections, ReportSection{
		Title: "Survey Summary",
		Content: fmt.Sprintf(
			"Preset: %s\nSubnet: %s\nDuration: %s\nDevices Found: %d\nFindings: %d\n",
			preset, subnet, duration.Round(time.Second), len(devices), len(findings),
		),
	})

	var deviceList []string
	for _, d := range devices {
		deviceList = append(deviceList, fmt.Sprintf("- %s (%s, %s, %s)",
			d.Hostname, d.IPs[0], d.Vendor, d.Role))
	}
	if len(deviceList) > 0 {
		report.Sections = append(report.Sections, ReportSection{
			Title:   "Discovered Devices",
			Content: strings.Join(deviceList, "\n"),
		})
	}

	severityCount := make(map[string]int)
	for _, f := range findings {
		severityCount[f.Severity]++
	}
	severityOrder := []string{"critical", "high", "medium", "low", "info"}
	var findingSummary []string
	for _, sev := range severityOrder {
		if count, ok := severityCount[sev]; ok {
			findingSummary = append(findingSummary, fmt.Sprintf("  %s: %d", sev, count))
		}
	}
	if len(findingSummary) > 0 {
		report.Sections = append(report.Sections, ReportSection{
			Title:   "Findings by Severity",
			Content: strings.Join(findingSummary, "\n"),
		})
	}

	return report
}
