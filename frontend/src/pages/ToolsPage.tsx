import React, { useState, useEffect, useRef } from 'react';
import { Card, Input, Button, Select, Space, Row, Col, Table, Descriptions, Statistic, message, Typography, Tag, Alert } from 'antd';
import { SendOutlined, NodeIndexOutlined, SearchOutlined, ThunderboltOutlined } from '@ant-design/icons';
import { runPing, runTraceroute, runNSLookup, wakeOnLAN, runIPerf, getNetworkInfo, onEvent } from '../hooks/api';
import type { PingResult, Hop } from '../types';

const { Text } = Typography;

function maskToDotted(mask: number): string {
  const allOnes = 0xFFFFFFFF;
  const bitmask = mask === 0 ? 0 : ~(allOnes >>> mask) >>> 0;
  return [(bitmask >>> 24) & 255, (bitmask >>> 16) & 255, (bitmask >>> 8) & 255, bitmask & 255].join('.');
}

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

  // Subnet Planner
  const [subnetHosts, setSubnetHosts] = useState<number | null>(null);
  const [subnetBase, setSubnetBase] = useState('');
  const [subnetResult, setSubnetResult] = useState<any>(null);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const latencyIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  useEffect(() => {
    const unsubs = [
      onEvent('tool:ping-result', (result: PingResult) => {
        setPingResults(prev => [...prev.slice(-49), result]);
        if (latencyRunning && result?.target === latencyTarget && !result.timedOut) {
          setLatencyPoints(prev => [...prev.slice(-59), result.latencyMs]);
        }
      }),
      onEvent('tool:traceroute-hop', (hop: Hop) => {
        setTraceHops(prev => [...prev, hop]);
      }),
      onEvent('iperf:result', (result: any) => {
        setIperfResult(result);
      }),
    ];
    return () => unsubs.forEach(fn => fn());
  }, [latencyRunning, latencyTarget]);

  useEffect(() => () => {
    if (latencyIntervalRef.current) clearInterval(latencyIntervalRef.current);
  }, []);

  const setActionError = (key: string, err: any, fallback: string) => {
    setErrors(prev => ({ ...prev, [key]: err?.message || String(err) || fallback }));
  };

  const handlePing = async () => {
    if (!pingTarget.trim()) { setActionError('ping', null, 'Enter a target before starting ping.'); return; }
    setErrors(prev => ({ ...prev, ping: '' }));
    setPingResults([]);
    try {
      await runPing(pingTarget, pingCount);
    } catch (err: any) { setActionError('ping', err, 'Ping failed'); }
  };

  const handleTrace = async () => {
    if (!traceTarget.trim()) { setActionError('trace', null, 'Enter a target before starting traceroute.'); return; }
    setErrors(prev => ({ ...prev, trace: '' }));
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
      setErrors(prev => ({ ...prev, lookup: '' }));
    } catch (err: any) { setActionError('lookup', err, 'Lookup failed'); }
  };

  const handleWoL = async () => {
    if (!wolMac.trim()) { setActionError('wol', null, 'Enter a MAC address first.'); return; }
    try {
      await wakeOnLAN(wolMac);
      message.success(`Magic packet sent to ${wolMac}`);
    } catch (err: any) { setActionError('wol', err, 'WoL failed'); }
  };

  const handleIPerf = async () => {
    if (iperfMode === 'client' && !iperfTarget.trim()) {
      setActionError('iperf', null, 'Enter an iPerf server hostname or IP for client mode.');
      return;
    }
    if (!Number.isInteger(iperfDuration) || iperfDuration < 1 || iperfDuration > 3600) {
      setActionError('iperf', null, 'Duration must be a whole number between 1 and 3600 seconds.');
      return;
    }
    setErrors(prev => ({ ...prev, iperf: '' }));
    try {
      await runIPerf(iperfTarget, iperfMode === 'server', iperfDuration);
    } catch (err: any) { setActionError('iperf', err, 'iPerf failed'); }
  };

  const handleNetInfo = async () => {
    try {
      const info = await getNetworkInfo();
      setNetInfo(info);
      setErrors(prev => ({ ...prev, net: '' }));
    } catch (err: any) { setActionError('net', err, 'Unable to load network information.'); }
  };

  const handleLatencyToggle = () => {
    if (!latencyRunning && !latencyTarget.trim()) {
      setActionError('latency', null, 'Enter a target before starting latency monitoring.');
      return;
    }
    if (!latencyRunning) {
      setLatencyPoints([]);
      setLatencyRunning(true);
      latencyIntervalRef.current = setInterval(async () => {
        try {
          await runPing(latencyTarget, 1);
        } catch (err: any) { setActionError('latency', err, 'Latency probe failed.'); }
      }, 1000);
    } else {
      if (latencyIntervalRef.current) clearInterval(latencyIntervalRef.current);
      latencyIntervalRef.current = null;
      setLatencyRunning(false);
    }
  };

  const calculateSubnet = () => {
    const hosts = subnetHosts || 0;
    if (hosts <= 0) {
      message.warning('Enter a valid number of hosts');
      return;
    }
    const mask = 32 - Math.ceil(Math.log2(hosts + 2));
    const total = Math.pow(2, 32 - mask);
    const usable = Math.max(0, total - 2);
    setSubnetResult({
      cidr: mask,
      totalHosts: total,
      usableHosts: usable,
      netmask: maskToDotted(mask),
    });
  };

  return (
    <div className="h-full overflow-y-auto p-2">
      <Row gutter={16}>
        <Col span={12}>
           <Card title="Ping" size="small" className="mb-4">
            <Space direction="vertical" style={{ width: '100%' }}>
              {errors.ping && <Alert type="error" showIcon message={errors.ping} />}
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
              {errors.trace && <Alert type="error" showIcon message={errors.trace} />}
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
              {errors.lookup && <Alert type="error" showIcon message={errors.lookup} />}
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
              {errors.wol && <Alert type="error" showIcon message={errors.wol} />}
              <Input placeholder="MAC (AA:BB:CC:DD:EE:FF)" value={wolMac} onChange={e => setWolMac(e.target.value)} />
              <Button type="primary" onClick={handleWoL}>Send Magic Packet</Button>
            </Space>
          </Card>
        </Col>

        <Col span={12}>
           <Card title="iPerf" size="small" className="mb-4">
            <Space direction="vertical" style={{ width: '100%' }}>
              {errors.iperf && <Alert type="error" showIcon message={errors.iperf} />}
              <Space>
                <Input placeholder={iperfMode === 'server' ? 'Optional target' : 'Server IP or hostname'} value={iperfTarget} onChange={e => setIperfTarget(e.target.value)} />
                <Tag color="blue">iperf.he.net</Tag>
              </Space>
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
              {errors.latency && <Alert type="error" showIcon message={errors.latency} />}
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
            {errors.net && <Alert type="error" showIcon message={errors.net} />}
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
          <Card title="Subnet Planner" size="small" className="mb-4">
            <Space direction="vertical" style={{ width: '100%' }}>
              <Input
                type="number"
                placeholder="How many hosts?"
                value={subnetHosts ?? ''}
                onChange={e => setSubnetHosts(Number(e.target.value) || null)}
              />
              <Input
                placeholder="Base network (optional, e.g. 10.0.0.0)"
                value={subnetBase}
                onChange={e => setSubnetBase(e.target.value)}
              />
              <Button type="primary" onClick={calculateSubnet}>Calculate Subnet</Button>
              {subnetResult && (
                <Descriptions size="small" column={1}>
                  <Descriptions.Item label="CIDR">/{subnetResult.cidr}</Descriptions.Item>
                  <Descriptions.Item label="Total Hosts">{subnetResult.totalHosts}</Descriptions.Item>
                  <Descriptions.Item label="Usable Hosts">{subnetResult.usableHosts}</Descriptions.Item>
                  <Descriptions.Item label="Netmask">{subnetResult.netmask}</Descriptions.Item>
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
