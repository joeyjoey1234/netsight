package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"netsight/internal/model"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func RunIPerf(ctx context.Context, target string, serverMode bool, duration int, iperfBinary string, onResult func(*model.IPerfResult)) error {
	binary, err := resolveIPerfBinary(iperfBinary)
	if err != nil {
		return err
	}
	if duration <= 0 {
		duration = 10
	}

	if serverMode {
		return runIPerfServer(ctx, binary)
	}
	return runIPerfClient(ctx, target, duration, binary, onResult)
}

func resolveIPerfBinary(path string) (string, error) {
	if path == "" {
		path = "iperf3"
	}
	if filepath.IsAbs(path) {
		if _, err := os.Stat(path); err != nil {
			return "", fmt.Errorf("iperf3 unavailable: %w", err)
		}
		return path, nil
	}
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), path)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	if runtime.GOOS != "windows" {
		if found, err := exec.LookPath(path); err == nil {
			return found, nil
		}
	}
	return "", fmt.Errorf("iperf3 unavailable: binary %q was not found", path)
}

func runIPerfServer(ctx context.Context, binary string) error {
	cmd := exec.CommandContext(ctx, binary, "-s", "-p", "5201", "--json")
	return cmd.Run()
}

func runIPerfClient(ctx context.Context, target string, duration int, binary string, onResult func(*model.IPerfResult)) error {
	args := []string{
		"-c", target,
		"-p", "5201",
		"-t", fmt.Sprintf("%d", duration),
		"--json",
	}

	cmd := exec.CommandContext(ctx, binary, args...)
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("iperf3 error: %w (output: %s)", err, string(output))
	}

	if onResult != nil {
		result := parseIPerfJSON(output)
		if result == nil {
			return fmt.Errorf("iperf3 returned invalid JSON")
		}
		onResult(result)
	}

	return nil
}

func parseIPerfJSON(data []byte) *model.IPerfResult {
	result := &model.IPerfResult{}

	var raw struct {
		End struct {
			SumSent struct {
				Bytes         int64   `json:"bytes"`
				Seconds       float64 `json:"seconds"`
				BitsPerSecond float64 `json:"bits_per_second"`
			} `json:"sum_sent"`
			SumReceived struct {
				Bytes         int64   `json:"bytes"`
				Seconds       float64 `json:"seconds"`
				BitsPerSecond float64 `json:"bits_per_second"`
			} `json:"sum_received"`
			Sum struct {
				JitterMs    float64 `json:"jitter_ms"`
				LostPackets int     `json:"lost_packets"`
				Packets     int     `json:"packets"`
			} `json:"sum"`
		} `json:"end"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return nil
	}

	result.Interval = raw.End.SumSent.Seconds
	result.TransferBytes = raw.End.SumSent.Bytes + raw.End.SumReceived.Bytes
	result.BandwidthBps = int64(raw.End.SumSent.BitsPerSecond + raw.End.SumReceived.BitsPerSecond)
	result.JitterMs = raw.End.Sum.JitterMs
	result.LostPackets = raw.End.Sum.LostPackets
	result.TotalPackets = raw.End.Sum.Packets

	return result
}
