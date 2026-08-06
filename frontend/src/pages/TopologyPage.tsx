import React, { useState, useCallback, useMemo, useEffect } from 'react';
import { Card, Button, Space, Input, Select, Progress, Drawer, Descriptions, Tag, Empty, Typography, message, Alert } from 'antd';
import { PlayCircleOutlined, StopOutlined, ReloadOutlined, ExportOutlined, ThunderboltOutlined } from '@ant-design/icons';
import { useScanStore } from '../stores/scanStore';
import { useThemeStore } from '../stores/themeStore';
import TopologyMap from '../components/topology/TopologyMap';
import { startScan, stopScan, exportDrawIO, exportPDF, onEvent, getAvailableSubnets, getDevices } from '../hooks/api';
import type { Device, Link } from '../types';

const { Text } = Typography;

function formatDeviceTooltip(d: Device): string {
  const parts: string[] = [];
  if (d.hostname) parts.push(`<b>${d.hostname}</b>`);
  if ((d.ips || []).length > 0) parts.push(`IP: ${(d.ips || []).join(', ')}`);
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
  const { activeScanId, scanStatus, scanProgress, devices,
          setScanId, setProgress, setStatus, addDevice, setDevices, addFinding, reset } = useScanStore();
  const { isDark } = useThemeStore();
  const [selectedDevice, setSelectedDevice] = useState<Device | null>(null);
  const [selectedSubnet, setSelectedSubnet] = useState('192.168.1.0/24');
  const [availableSubnets, setAvailableSubnets] = useState<string[]>([]);
  const [longMinutes, setLongMinutes] = useState<number | null>(null);
  const [showLongInput, setShowLongInput] = useState(false);
  const [links, setLinks] = useState<Link[]>([]);
  const [loadError, setLoadError] = useState<string | null>(null);
  const activeScanRef = React.useRef<string | null>(activeScanId);

  useEffect(() => { activeScanRef.current = activeScanId; }, [activeScanId]);

  const isRunning = scanStatus === 'running';

  useEffect(() => {
    getAvailableSubnets().then(subs => {
      const subnets = (subs || []).length > 0 ? subs : ['192.168.1.0/24'];
      setAvailableSubnets(subnets);
      setSelectedSubnet(subnets[0]);
    }).catch((err: any) => setLoadError(`Unable to load subnets: ${err?.message || err}`));
  }, []);

  useEffect(() => {
    const unsubs = [
      onEvent('scan:progress', (scanId: string, progress: number) => {
        if (!scanId || scanId !== activeScanRef.current) return;
        setProgress(progress);
      }),
      onEvent('scan:device-found', (payload: Device | { scanId?: string; device: Device }) => {
        const event = payload as any;
        const scanId = event?.scanId;
        const device = event?.device || event;
        if (scanStatus !== 'running' || (scanId && scanId !== activeScanRef.current)) return;
        addDevice({ ...device, ips: device.ips || [], vlans: device.vlans || [] });
        if (Array.isArray(event?.links)) setLinks(event.links);
      }),
      onEvent('scan:complete', async (scanId: string, status: string) => {
        if (!scanId || scanId !== activeScanRef.current) return;
        const finalStatus = status === 'cancelled' || status === 'canceled' ? 'cancelled' : status;
        if (finalStatus === 'failed') {
          setStatus('failed');
          message.error(`Scan ${scanId.slice(0, 8)} failed`);
          return;
        }
        if (finalStatus === 'cancelled') {
          setStatus('cancelled');
          message.info(`Scan ${scanId.slice(0, 8)} cancelled`);
          return;
        }
        setStatus('completed');
        try {
          const response: any = await getDevices();
          const devs = Array.isArray(response) ? response : (response?.devices || []);
          const normalizedDevices = devs.map((device: Device) => ({ ...device, ips: device.ips || [], vlans: device.vlans || [] }));
          setDevices(normalizedDevices);
          if (Array.isArray(response?.links)) setLinks(response.links);
          else setLinks(normalizedDevices.flatMap((device: Device) => device.links || []));
          message.success(`Scan ${scanId.slice(0, 8)} complete — ${devs.length} devices found`);
        } catch {
          message.success(`Scan ${scanId.slice(0, 8)} complete`);
        }
      }),
      onEvent('survey:finding', (finding: any, scanId?: string) => {
        if (scanId && scanId !== activeScanRef.current) return;
        if (scanStatus !== 'running' && scanStatus !== 'completed') return;
        addFinding(finding);
      }),
    ];
    return () => unsubs.forEach(fn => fn());
  }, [setProgress, addDevice, setStatus, setDevices, addFinding, scanStatus]);

  useEffect(() => {
    setSelectedDevice((current) => current ? devices.find((device) => device.id === current.id) || null : null);
  }, [devices]);

  const graphNodes = useMemo(() => devices.map(d => ({
    id: d.id,
    label: d.hostname || ((d.ips || []).length > 0 ? d.ips[0] : d.mac),
    title: formatDeviceTooltip(d),
    group: roleToGroup(d.role),
    color: roleToColor(d.role),
    shape: 'dot' as const,
    size: 25,
  })), [devices]);

  const handleNodeClick = useCallback((nodeId: string) => {
    const device = devices.find(d => d.id === nodeId);
    setSelectedDevice(device || null);
  }, [devices]);

  const runScan = async (subnet: string, preset: string) => {
    try {
      reset();
      setLinks([]);
      const id = await startScan(subnet, preset);
      setScanId(id);
      setStatus('running');
      setProgress(0);
    } catch (err: any) {
      setStatus('failed');
      message.error(`Scan failed: ${err?.message || err}`);
    }
  };

  const handleQuickSurvey = () => runScan(selectedSubnet, 'quick');
  const handleShortSurvey = () => runScan(selectedSubnet, 'short');
  const handleLongSurvey = () => {
    if (longMinutes && longMinutes > 0) {
      runScan(selectedSubnet, 'long');
      setShowLongInput(false);
      setLongMinutes(null);
    } else {
      setShowLongInput(!showLongInput);
    }
  };

  const handleScan = () => runScan(selectedSubnet, 'manual');

  const handleStop = async () => {
    if (!activeScanId) return;
    try {
      await stopScan(activeScanId);
      activeScanRef.current = null;
      setScanId(null);
      setStatus('cancelled');
    } catch (err: any) {
      message.error(`Stop failed: ${err?.message || err}`);
    }
  };

  const handleExportDrawIO = async () => {
    if (!activeScanId) return;
    try {
      const path = await exportDrawIO(activeScanId);
      message.success(`Exported to ${path}`);
    } catch (err: any) {
      message.error(`Export failed: ${err?.message || err}`);
    }
  };

  const handleExportPDF = async () => {
    if (!activeScanId) return;
    try {
      const path = await exportPDF(activeScanId);
      message.success(`PDF exported to ${path}`);
    } catch (err: any) {
      message.error(`PDF export failed: ${err?.message || err}`);
    }
  };

  const topologyLinks = [...links, ...devices.flatMap(device => device.links || [])]
    .filter((link, index, all) => all.findIndex(item => item.sourceId === link.sourceId && item.targetId === link.targetId && item.type === link.type) === index);
  const topologyEdges = topologyLinks.filter((link) => devices.some(d => d.id === link.sourceId) && devices.some(d => d.id === link.targetId)).map((link, index) => ({
    id: `${link.sourceId}-${link.targetId}-${index}`,
    from: link.sourceId,
    to: link.targetId,
    label: [link.type, link.srcPort && `${link.srcPort}->${link.dstPort}`].filter(Boolean).join(' '),
    dashes: link.type === 'ARP',
    color: link.type === 'LLDP' ? '#52c41a' : '#8c8c8c',
    width: 2,
  }));

  return (
    <div className="h-full flex flex-col p-2">
      <Card size="small" className="mb-2 flex-shrink-0">
        {loadError && <Alert className="mb-2" type="error" showIcon message={loadError} />}
        <Space wrap>
          <Button icon={<ThunderboltOutlined />} onClick={handleQuickSurvey} disabled={isRunning}>
            Quick Survey (3m)
          </Button>
          <Button icon={<ThunderboltOutlined />} onClick={handleShortSurvey} disabled={isRunning}>
            Short Survey (10m)
          </Button>
          {showLongInput ? (
            <Space.Compact>
              <Input
                type="number"
                placeholder="Minutes"
                style={{ width: 100 }}
                value={longMinutes ?? ''}
                onChange={e => setLongMinutes(Number(e.target.value) || null)}
                disabled={isRunning}
              />
              <Button icon={<ThunderboltOutlined />} onClick={handleLongSurvey} disabled={isRunning}>
                Start
              </Button>
            </Space.Compact>
          ) : (
            <Button icon={<ThunderboltOutlined />} onClick={handleLongSurvey} disabled={isRunning}>
              Long Survey
            </Button>
          )}
        </Space>

        <div className="mt-2">
          <Text type="secondary" className="text-xs mr-2">Subnet Scan:</Text>
          <Space wrap>
            <Select
              value={selectedSubnet}
              onChange={setSelectedSubnet}
              style={{ width: 220 }}
              disabled={isRunning}
              options={availableSubnets.map(s => ({ value: s, label: s }))}
              notFoundContent={<Text type="secondary">No subnets detected</Text>}
            />
            <Button type="primary" icon={<PlayCircleOutlined />} onClick={handleScan} disabled={isRunning}>
              Scan
            </Button>
            <Button danger icon={<StopOutlined />} onClick={handleStop} disabled={!isRunning}>
              Stop
            </Button>
            <Button icon={<ReloadOutlined />} onClick={reset} disabled={isRunning}>
              Clear
            </Button>
            <Button icon={<ExportOutlined />} onClick={handleExportDrawIO} disabled={devices.length === 0}>
              Export Draw.io
            </Button>
            <Button icon={<ExportOutlined />} onClick={handleExportPDF} disabled={devices.length === 0 || !activeScanId}>
              Export PDF
            </Button>
          </Space>
        </div>

        {isRunning && (
          <div className="mt-2">
            <Progress percent={scanProgress} size="small" status="active" />
            <Text type="secondary" className="text-xs ml-2">
              Scanning... {devices.length} device(s) found
            </Text>
          </div>
        )}
      </Card>

      <Card className="flex-1 overflow-hidden" bodyStyle={{ height: '100%', padding: 0 }}>
        {devices.length === 0 && !isRunning ? (
          <div className="flex items-center justify-center h-full" style={{ background: isDark ? '#0d1117' : '#f5f5f5' }}>
            <Empty description={
              <span className={isDark ? 'text-gray-400' : 'text-gray-500'}>
                Enter a subnet and click Scan to discover devices
              </span>
            } />
          </div>
        ) : (
           <TopologyMap nodes={graphNodes} edges={topologyEdges} onNodeClick={handleNodeClick} darkMode={isDark} />
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
            {selectedDevice.hostname && <Descriptions.Item label="Hostname">{selectedDevice.hostname}</Descriptions.Item>}
            <Descriptions.Item label="IP Addresses">{(selectedDevice.ips || []).length > 0 ? (selectedDevice.ips || []).join(', ') : 'N/A'}</Descriptions.Item>
            <Descriptions.Item label="MAC Address"><Tag>{selectedDevice.mac || 'Unknown'}</Tag></Descriptions.Item>
            <Descriptions.Item label="Vendor">{selectedDevice.vendor || 'Unknown'}</Descriptions.Item>
            <Descriptions.Item label="Model">{selectedDevice.model || 'Unknown'}</Descriptions.Item>
            <Descriptions.Item label="OS">{selectedDevice.os || 'Unknown'}</Descriptions.Item>
            <Descriptions.Item label="Role"><Tag color={roleToColor(selectedDevice.role)}>{selectedDevice.role}</Tag></Descriptions.Item>
            <Descriptions.Item label="VLANs">{(selectedDevice.vlans || []).length > 0 ? (selectedDevice.vlans || []).map(v => <Tag key={v}>{v}</Tag>) : 'None'}</Descriptions.Item>
            <Descriptions.Item label="First Seen">{selectedDevice.firstSeen ? new Date(selectedDevice.firstSeen).toLocaleString() : 'N/A'}</Descriptions.Item>
            <Descriptions.Item label="Last Seen">{selectedDevice.lastSeen ? new Date(selectedDevice.lastSeen).toLocaleString() : 'N/A'}</Descriptions.Item>
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
