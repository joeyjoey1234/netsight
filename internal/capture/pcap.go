package capture

import (
	"fmt"
	"os"
)

func ExportPCAP(filename string, packets []byte) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create pcap file: %w", err)
	}
	defer f.Close()

	_, _ = f.Write(make([]byte, 24))
	return nil
}
