import React, { useState, useCallback, useMemo } from 'react';
import { Card, Button, Space, Input, Select, Progress, Drawer, Descriptions, Tag, Empty, Typography } from 'antd';
import { PlayCircleOutlined, StopOutlined, ReloadOutlined, ExportOutlined } from '@ant-design/icons';
import { useScanStore } from '../stores/scanStore';
import { useThemeStore } from '../stores/themeStore';
import TopologyMap from '../components/topology/TopologyMap';
import type { Device } from '../types';

function devicesToTopologyNodes(devices: Device[]) {
  return devices.map(d => ({
    id: d.id,
    label: d.hostname || (d.ips.length > 0 ? d.ips[0] : d.mac),
    title: formatDeviceTooltip(d),
    group: roleToGroup(d.role),
    color: roleToColor(d.role),
    shape: 'dot' as const,
    size: 25,
  }));
}

const placeholderEdges: any[] = [];

function formatDeviceTooltip(d: Device): string {
  const parts: string[] = [];
  if (d.hostname) parts.push(`<b>${d.hostname}</b>`);
  if (d.ips.length > 0) parts.push(`IP: ${d.ips.join(', ')}`);
  if (d.mac) parts.push(`MAC: ${d.mac}`);
  if (d.vendor) parts.push(`Vendor: ${d.vendor}`);
  if (d.model) parts.push(`Model: ${d.model}`);
  if (d.os) parts.push(`OS: ${d.os}`);
  if (d.role && d.role !== 'unknown') parts.push(`Role: ${d.role}`);
  return parts.join('<br>');
}

function roleToGroup(role: string): string {
  switch (role) {
    case 'switch': case 'router': case 'L3 switch': return 'network';
    case 'server': return 'server';
    case 'workstation': return 'endpoint';
    default: return 'unknown';
  }
}

function roleToColor(role: string): string {
  switch (role) {
    case 'switch': return '#52c41a';
    case 'router': return '#1677ff';
    case 'L3 switch': return '#722ed1';
    case 'server': return '#fa8c16';
    case 'workstation': return '#13c2c2';
    default: return '#8c8c8c';
  }
}

const TopologyPage: React.FC = () => {
  const { scanStatus, scanProgress, devices, reset } = useScanStore();
  const { isDark } = useThemeStore();
  const [selectedDevice, setSelectedDevice] = useState<Device | null>(null);
  const [subnetInput, setSubnetInput] = useState('192.168.1.0/24');
  const [preset, setPreset] = useState('quick');

  const isRunning = scanStatus === 'running';
  const graphNodes = useMemo(() => devicesToTopologyNodes(devices), [devices]);

  const handleNodeClick = useCallback((nodeId: string) => {
    const device = devices.find(d => d.id === nodeId);
    setSelectedDevice(device || null);
  }, [devices]);

  const handleScan = async () => {
    console.log('Starting scan:', subnetInput, preset);
  };

  const handleStop = async () => {
    console.log('Stopping scan');
  };

  return (
    <div className="h-full flex flex-col p-2">
      <Card size="small" className="mb-2 flex-shrink-0">
        <Space wrap>
          <Input
            placeholder="Subnet (e.g. 192.168.1.0/24)"
            value={subnetInput}
            onChange={e => setSubnetInput(e.target.value)}
            style={{ width: 200 }}
            disabled={isRunning}
          />
          <Select
            value={preset}
            onChange={setPreset}
            style={{ width: 130 }}
            disabled={isRunning}
            options={[
              { value: 'quick', label: 'Quick (3m)' },
              { value: 'short', label: 'Short (10m)' },
              { value: 'long', label: 'Long' },
            ]}
          />
          <Button
            type="primary"
            icon={<PlayCircleOutlined />}
            onClick={handleScan}
            disabled={isRunning}
          >
            Scan
          </Button>
          <Button
            danger
            icon={<StopOutlined />}
            onClick={handleStop}
            disabled={!isRunning}
          >
            Stop
          </Button>
          <Button icon={<ReloadOutlined />} onClick={reset} disabled={isRunning}>
            Clear
          </Button>
          <Button icon={<ExportOutlined />} disabled={devices.length === 0}>
            Export Draw.io
          </Button>
        </Space>

        {isRunning && (
          <div className="mt-2">
            <Progress percent={scanProgress} size="small" status="active" />
            <Typography.Text type="secondary" className="text-xs ml-2">
              Scanning... {devices.length} device(s) found
            </Text>
          </div>
        )}
      </Card>

      <Card className="flex-1 overflow-hidden" bodyStyle={{ height: '100%', padding: 0 }}>
        {devices.length === 0 && !isRunning ? (
          <div className="flex items-center justify-center h-full" style={{ background: isDark ? '#0d1117' : '#f5f5f5' }}>
            <Empty
              description={
                <span className={isDark ? 'text-gray-400' : 'text-gray-500'}>
                  Enter a subnet and click Scan to discover devices
                </span>
              }
            />
          </div>
        ) : (
          <TopologyMap
            nodes={graphNodes}
            edges={placeholderEdges}
            onNodeClick={handleNodeClick}
            darkMode={isDark}
          />
        )}
      </Card>

      <Drawer
        title={selectedDevice?.hostname || selectedDevice?.ips?.[0] || 'Device Details'}
        placement="right"
        width={400}
        onClose={() => setSelectedDevice(null)}
        open={!!selectedDevice}
      >
        {selectedDevice && (
          <Descriptions column={1} size="small" bordered>
            {selectedDevice.hostname && (
              <Descriptions.Item label="Hostname">{selectedDevice.hostname}</Descriptions.Item>
            )}
            <Descriptions.Item label="IP Addresses">
              {selectedDevice.ips.length > 0 ? selectedDevice.ips.join(', ') : 'N/A'}
            </Descriptions.Item>
            <Descriptions.Item label="MAC Address">
              <Tag>{selectedDevice.mac || 'Unknown'}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="Vendor">
              {selectedDevice.vendor || 'Unknown'}
            </Descriptions.Item>
            <Descriptions.Item label="Model">
              {selectedDevice.model || 'Unknown'}
            </Descriptions.Item>
            <Descriptions.Item label="OS">
              {selectedDevice.os || 'Unknown'}
            </Descriptions.Item>
            <Descriptions.Item label="Role">
              <Tag color={roleToColor(selectedDevice.role)}>{selectedDevice.role}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="VLANs">
              {selectedDevice.vlans.length > 0
                ? selectedDevice.vlans.map(v => <Tag key={v}>{v}</Tag>)
                : 'None'}
            </Descriptions.Item>
            <Descriptions.Item label="First Seen">
              {selectedDevice.firstSeen ? new Date(selectedDevice.firstSeen).toLocaleString() : 'N/A'}
            </Descriptions.Item>
            <Descriptions.Item label="Last Seen">
              {selectedDevice.lastSeen ? new Date(selectedDevice.lastSeen).toLocaleString() : 'N/A'}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Drawer>

      <div className="text-xs text-gray-500 mt-1 flex-shrink-0">
        {devices.length} devices | Status: {scanStatus}
        {scanStatus === 'completed' && ' — Scan complete'}
      </div>
    </div>
  );
};

export default TopologyPage;
