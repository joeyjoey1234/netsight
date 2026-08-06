package storage

import "netsight/internal/model"

type Store interface {
	CreateProject(project *model.Project) error
	GetProject(id string) (*model.Project, error)
	ListProjects() ([]*model.Project, error)
	DeleteProject(id string) error

	SaveScan(projectID string, scan *model.Scan) error
	UpdateScan(scan *model.Scan) error
	GetScan(id string) (*model.Scan, error)
	ListScans(projectID string) ([]*model.Scan, error)
	GetLatestScan(projectID string) (*model.Scan, error)

	SaveDevice(projectID string, device *model.Device) error
	SaveDevices(projectID string, devices []*model.Device) error
	SaveScanDevice(scanID, deviceID string) error
	GetDevice(id string) (*model.Device, error)
	ListDevices(projectID string) ([]*model.Device, error)
	GetDevicesByScan(scanID string) ([]*model.Device, error)

	SavePorts(deviceID, scanID string, ports []*model.Port) error
	GetPorts(deviceID string) ([]*model.Port, error)

	SaveLinks(links []*model.Link) error
	ListLinks() ([]*model.Link, error)

	SaveFinding(scanID string, finding *model.Finding) error

	GetNewDevices(projectID, scanA, scanB string) ([]*model.Device, error)
	GetMissingDevices(projectID, scanA, scanB string) ([]*model.Device, error)
	GetChangedDevices(projectID, scanA, scanB string) ([]*model.Device, error)

	Close() error
}
