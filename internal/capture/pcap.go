package capture

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcapgo"
	"os"
	"time"
)

func ExportPCAP(filename string, packets []byte) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create pcap file: %w", err)
	}
	defer f.Close()
	w := pcapgo.NewWriter(f)
	if err := w.WriteFileHeader(65536, 1); err != nil {
		return err
	}
	if len(packets) == 0 {
		return nil
	}
	return w.WritePacket(gopacket.CaptureInfo{Timestamp: time.Now(), CaptureLength: len(packets), Length: len(packets)}, packets)
}
