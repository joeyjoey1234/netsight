import React, { useState, useEffect } from 'react';
import { Card, Input, Button, Select, Space, Row, Col, Table, Descriptions, Statistic, message, Typography } from 'antd';
import { SendOutlined, NodeIndexOutlined, SearchOutlined, ThunderboltOutlined } from '@ant-design/icons';
import { runPing, runTraceroute, runNSLookup, wakeOnLAN, runIPerf, getNetworkInfo, onEvent } from '../hooks/api';
import type { PingResult, Hop } from '../types';

const { Text } = Typography;

const ToolsPage: React.FC = () => {
  // Ping
  const [pingTarget, setPingTarget] = useState('');
  const [pingCount, setPingCount] = useState(4);
  const [pingResults, setPingResults] = useState<PingResult[]>([]);

  // Traceroute
  const [traceTarget, setTraceTarget] = useState('');
  const [traceMode, setTraceMode] = useState('icmp');
  const [traceHops, setTraceHops] = useState<Hop[]>([]);

  // NSLookup
  const [lookupQuery, setLookupQuery] = useState('');
  const [lookupTypes, setLookupTypes] = useState<string[]>(['A']);
  const [lookupResults, setLookupResults] = useState<any[]>([]);

  // WoL
  const [wolMac, setWolMac] = useState('');

  // iPerf
  const [iperfTarget, setIperfTarget] = useState('');
  const [iperfMode, setIperfMode] = useState('client');
  const [iperfDuration, setIperfDuration] = useState(10);
  const [iperfResult, setIperfResult] = useState<any>(null);

  // Network info
  const [netInfo, setNetInfo] = useState<any>(null);

  // Latency
  const [latencyTarget, setLatencyTarget] = useState('');
  const [latencyRunning, setLatencyRunning] = useState(false);
  const [latencyPoints, setLatencyPoints] = useState<number[]>([]);

  // Subnet
  const [subnetCidr, setSubnetCidr] = useState('');
  const [subnetResult, setSubnetResult] = useState<any>(null);

  useEffect(() => {
    const unsubs = [
      onEvent('tool:ping-result', (result: PingResult) => {
        setPingResults(prev => [...prev.slice(-49), result]);
      }),
      onEvent('tool:traceroute-hop', (hop: Hop) => {
        setTraceHops(prev => [...prev, hop]);
      }),
      onEvent('iperf:result', (result: any) => {
        setIperfResult(result);
      }),
    ];
    return () => unsubs.forEach(fn => fn());
  }, []);

  const handlePing = async () => {
    if (!pingTarget) return;
    setPingResults([]);
    try {
      await runPing(pingTarget, pingCount);
    } catch (err: any) { message.error(err?.message || 'Ping failed'); }
  };

  const handleTrace = async () => {
    if (!traceTarget) return;
    setTraceHops([]);
    try {
      await runTraceroute(traceTarget, traceMode);
    } catch (err: any) { message.error(err?.message || 'Trace failed'); }
  };

  const handleLookup = async () => {
    if (!lookupQuery) return;
    try {
      const results = await runNSLookup(lookupQuery, lookupTypes);
      setLookupResults(results || []);
    } catch (err: any) { message.error(err?.message || 'Lookup failed'); }
  };

  const handleWoL = async () => {
    if (!wolMac) return;
    try {
      await wakeOnLAN(wolMac);
      message.success(`Magic packet sent to ${wolMac}`);
    } catch (err: any) { message.error(err?.message || 'WoL failed'); }
  };

  const handleIPerf = async () => {
    if (!iperfTarget) return;
    try {
      await runIPerf(iperfTarget, iperfMode === 'server', iperfDuration);
    } catch (err: any) { message.error(err?.message || 'iPerf failed'); }
  };

  const handleNetInfo = async () => {
    try {
      const info = await getNetworkInfo();
      setNetInfo(info);
    } catch (err: any) { message.error(err?.message || 'Failed'); }
  };

  const handleLatencyToggle = () => {
    setLatencyRunning(!latencyRunning);
    // Latency monitoring uses periodic RunPing in a loop
    if (!latencyRunning && latencyTarget) {
      const interval = setInterval(async () => {
        try {
          await runPing(latencyTarget, 1);
        } catch {}
      }, 1000);
      (window as any).__latencyInterval = interval;
    } else {
      clearInterval((window as any).__latencyInterval);
    }
  };

  return (
    <div className="h-full overflow-y-auto p-2">
      <Row gutter={16}>
        <Col span={12}>
          <Card title="Ping" size="small" className="mb-4">
            <Space direction="vertical" style={{ width: '100%' }}>
              <Input placeholder="Target IP or hostname" value={pingTarget} onChange={e => setPingTarget(e.target.value)} />
              <Space>
                <Select value={pingCount} onChange={setPingCount} style={{ width: 100 }} options={[
                  { value: 4, label: '4 pings' }, { value: 10, label: '10 pings' }, { value: 0, label: 'Continuous' },
                ]} />
                <Button type="primary" icon={<SendOutlined />} onClick={handlePing}>Ping</Button>
              </Space>
              {pingResults.map((r, i) => (
                <Text key={i} className="text-xs font-mono" type={r.timedOut ? 'danger' : 'secondary'}>
                  {r.timedOut ? 'Request timed out' : `Reply from ${r.target}: bytes=${r.bytes} time=${r.latencyMs}ms TTL=${r.ttl}`}
                </Text>
              ))}
            </Space>
          </Card>
        </Col>

        <Col span={12}>
          <Card title="Traceroute" size="small" className="mb-4">
            <Space direction="vertical" style={{ width: '100%' }}>
              <Input placeholder="Target IP or hostname" value={traceTarget} onChange={e => setTraceTarget(e.target.value)} />
              <Space>
                <Select value={traceMode} onChange={setTraceMode} style={{ width: 100 }} options={[
                  { value: 'icmp', label: 'ICMP' }, { value: 'tcp', label: 'TCP' }, { value: 'udp', label: 'UDP' },
                ]} />
                <Button type="primary" icon={<NodeIndexOutlined />} onClick={handleTrace}>Trace</Button>
              </Space>
              {traceHops.map((h, i) => (
                <Text key={i} className="text-xs font-mono" type="secondary">
                  {h.number}  {h.ip} ({h.hostname || '???'})  {h.latencyMs.toFixed(1)}ms
                </Text>
              ))}
            </Space>
          </Card>
        </Col>

        <Col span={12}>
          <Card title="NSLookup" size="small" className="mb-4">
            <Space direction="vertical" style={{ width: '100%' }}>
              <Input placeholder="Hostname or IP" value={lookupQuery} onChange={e => setLookupQuery(e.target.value)} />
              <Select mode="multiple" value={lookupTypes} onChange={setLookupTypes} style={{ width: '100%' }} options={[
                { value: 'A', label: 'A (IPv4)' }, { value: 'AAAA', label: 'AAAA (IPv6)' }, { value: 'MX', label: 'MX' },
                { value: 'NS', label: 'NS' }, { value: 'TXT', label: 'TXT' }, { value: 'PTR', label: 'PTR' },
              ]} />
              <Button type="primary" icon={<SearchOutlined />} onClick={handleLookup}>Lookup</Button>
              {lookupResults.map((r: any, i: number) => (
                <Text key={i} className="text-xs" type="secondary">{r?.results?.join(', ') || r?.error || ''}</Text>
              ))}
            </Space>
          </Card>
        </Col>

        <Col span={12}>
          <Card title="Wake-on-LAN" size="small" className="mb-4">
            <Space direction="vertical" style={{ width: '100%' }}>
              <Input placeholder="MAC (AA:BB:CC:DD:EE:FF)" value={wolMac} onChange={e => setWolMac(e.target.value)} />
              <Button type="primary" onClick={handleWoL}>Send Magic Packet</Button>
            </Space>
          </Card>
        </Col>

        <Col span={12}>
          <Card title="iPerf" size="small" className="mb-4">
            <Space direction="vertical" style={{ width: '100%' }}>
              <Input placeholder="Target IP" value={iperfTarget} onChange={e => setIperfTarget(e.target.value)} />
              <Space>
                <Select value={iperfMode} onChange={setIperfMode} style={{ width: 100 }} options={[
                  { value: 'client', label: 'Client' }, { value: 'server', label: 'Server' },
                ]} />
                <Input type="number" value={iperfDuration} onChange={e => setIperfDuration(Number(e.target.value))} style={{ width: 100 }} />
                <Button type="primary" onClick={handleIPerf}>Run</Button>
              </Space>
              {iperfResult && <Text className="text-xs" type="secondary">{iperfResult.bandwidthBps} bps</Text>}
            </Space>
          </Card>
        </Col>

        <Col span={12}>
          <Card title="Latency Monitor" size="small" className="mb-4">
            <Space direction="vertical" style={{ width: '100%' }}>
              <Input placeholder="Target IP" value={latencyTarget} onChange={e => setLatencyTarget(e.target.value)} />
              <Button type="primary" onClick={handleLatencyToggle}>
                {latencyRunning ? 'Stop Monitoring' : 'Start Monitoring'}
              </Button>
              <div className="h-24 bg-gray-800 rounded p-2 text-xs text-gray-400 font-mono">
                {latencyPoints.slice(-20).map((v, i) => <span key={i}>{v}ms </span>)}
              </div>
            </Space>
          </Card>
        </Col>

        <Col span={12}>
          <Card title="Network Info" size="small" className="mb-4">
            <Button type="primary" icon={<ThunderboltOutlined />} onClick={handleNetInfo}>Get Network Info</Button>
            {netInfo && (
              <Descriptions size="small" column={1} className="mt-2">
                {Object.entries(netInfo).map(([k, v]) => (
                  <Descriptions.Item key={k} label={k}>{String(v)}</Descriptions.Item>
                ))}
              </Descriptions>
            )}
          </Card>
        </Col>

        <Col span={12}>
          <Card title="Subnet Calculator" size="small" className="mb-4">
            <Space direction="vertical" style={{ width: '100%' }}>
              <Input placeholder="e.g. 192.168.1.0/24" value={subnetCidr} onChange={e => setSubnetCidr(e.target.value)} />
              <Button type="primary" onClick={() => {
                const parts = subnetCidr.split('/');
                const ipOctets = parts[0].split('.').map(Number);
                const mask = parseInt(parts[1]);
                const total = Math.pow(2, 32 - mask);
                setSubnetResult({ cidr: subnetCidr, totalHosts: total, usableHosts: Math.max(0, total - 2) });
              }}>Calculate</Button>
              {subnetResult && (
                <Descriptions size="small" column={1}>
                  <Descriptions.Item label="CIDR">{subnetResult.cidr}</Descriptions.Item>
                  <Descriptions.Item label="Total Hosts">{subnetResult.totalHosts}</Descriptions.Item>
                  <Descriptions.Item label="Usable Hosts">{subnetResult.usableHosts}</Descriptions.Item>
                </Descriptions>
              )}
            </Space>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default ToolsPage;
