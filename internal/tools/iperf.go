package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"netsight/internal/model"
	"os/exec"
)

func RunIPerf(ctx context.Context, target string, serverMode bool, duration int, iperfBinary string, onResult func(*model.IPerfResult)) error {
	if duration <= 0 {
		duration = 10
	}

	if serverMode {
		return runIPerfServer(ctx, iperfBinary)
	}
	return runIPerfClient(ctx, target, duration, iperfBinary, onResult)
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
		onResult(result)
	}

	return nil
}

func parseIPerfJSON(data []byte) *model.IPerfResult {
	result := &model.IPerfResult{}

	var raw struct {
		End struct {
			SumSent struct {
				Bytes   int64   `json:"bytes"`
				Seconds float64 `json:"seconds"`
				BitsPerSecond float64 `json:"bits_per_second"`
			} `json:"sum_sent"`
			SumReceived struct {
				Bytes   int64   `json:"bytes"`
				Seconds float64 `json:"seconds"`
				BitsPerSecond float64 `json:"bits_per_second"`
			} `json:"sum_received"`
			Sum struct {
				JitterMs     float64 `json:"jitter_ms"`
				LostPackets  int     `json:"lost_packets"`
				Packets      int     `json:"packets"`
			} `json:"sum"`
		} `json:"end"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return result
	}

	result.Interval = raw.End.SumSent.Seconds
	result.TransferBytes = raw.End.SumSent.Bytes + raw.End.SumReceived.Bytes
	result.BandwidthBps = int64(raw.End.SumSent.BitsPerSecond + raw.End.SumReceived.BitsPerSecond)
	result.JitterMs = raw.End.Sum.JitterMs
	result.LostPackets = raw.End.Sum.LostPackets
	result.TotalPackets = raw.End.Sum.Packets

	return result
}
