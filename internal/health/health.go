package health

import (
	"netsight/internal/model"
)

func RunAllHealthChecks(devices []*model.Device) []*model.Finding {
	var findings []*model.Finding

	duplex := NewDuplexCheck()
	ifaces := duplex.LocalInterfaces()
	for _, iface := range ifaces {
		duplex.RegisterInterface(iface)
	}
	findings = append(findings, duplex.CheckInterfaces()...)

	_ = devices

	return findings
}
