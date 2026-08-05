package tools

import (
	"context"
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

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("iperf3 failed to start: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		cmd.Process.Kill()
		return ctx.Err()
	case err := <-done:
		if err != nil {
			return fmt.Errorf("iperf3 error: %w", err)
		}
	}

	if onResult != nil {
		onResult(&model.IPerfResult{
			Interval:      float64(duration),
			TransferBytes: 0,
			BandwidthBps:  0,
		})
	}

	return nil
}
