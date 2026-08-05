import React, { useState, useEffect } from 'react';
import { Card, Button, Input, Space, Table, Tag, Statistic, Row, Col, Select, Empty, message } from 'antd';
import { PlayCircleOutlined, PauseCircleOutlined, DownloadOutlined } from '@ant-design/icons';
import { useCaptureStore } from '../stores/captureStore';
import { startPacketCapture, stopPacketCapture, onEvent, getAllNetworkInfo } from '../hooks/api';
import type { PacketSummary } from '../types';

interface InterfaceOption {
  value: string;
  label: string;
}

const CapturePage: React.FC = () => {
  const { isCapturing, packets, packetsPerSec, bytesPerSec, filter, setFilter, setCapturing, addPacket, setStats, clear } = useCaptureStore();
  const [iface, setIface] = useState('');
  const [interfaces, setInterfaces] = useState<InterfaceOption[]>([]);

  useEffect(() => {
    getAllNetworkInfo().then((infos: any[]) => {
      const opts: InterfaceOption[] = (infos || []).map((info: any) => {
        const firstIpv4 = (info.ips || []).find((ip: string) => !ip.includes(':'));
        return {
          value: info.name || '',
          label: `${info.name} (${firstIpv4 || 'no IP'})`,
        };
      });
      setInterfaces(opts);
      if (opts.length > 0 && !iface) {
        setIface(opts[0].value);
      }
    });
  }, []);

  useEffect(() => {
    const unsubs = [
      onEvent('capture:packet', (packet: PacketSummary) => {
        addPacket(packet);
      }),
      onEvent('capture:stats', (pps: number, bps: number) => {
        setStats(pps, bps);
      }),
    ];
    return () => unsubs.forEach(fn => fn());
  }, [addPacket, setStats]);

  const handleStart = async () => {
    try {
      await startPacketCapture(iface, filter);
      setCapturing(true);
      clear();
    } catch (err: any) {
      message.error(`Capture start failed: ${err?.message || err}`);
    }
  };

  const handleStop = async () => {
    try {
      await stopPacketCapture();
      setCapturing(false);
    } catch (err: any) {
      message.error(`Capture stop failed: ${err?.message || err}`);
    }
  };

  const columns = [
    { title: '#', dataIndex: 'number', width: 60 },
    { title: 'Time', dataIndex: 'timestamp', width: 100 },
    { title: 'Source', key: 'src', width: 200, render: (_: any, r: PacketSummary) => `${r.srcIp}:${r.srcPort || ''}` },
    { title: 'Destination', key: 'dst', width: 200, render: (_: any, r: PacketSummary) => `${r.dstIp}:${r.dstPort || ''}` },
    { title: 'Protocol', dataIndex: 'protocol', width: 80, render: (p: string) => <Tag color="blue">{p}</Tag> },
    { title: 'Length', dataIndex: 'length', width: 70 },
    { title: 'Info', dataIndex: 'info', ellipsis: true },
  ];

  return (
    <div className="h-full flex flex-col p-2">
      <Card size="small" className="mb-2 flex-shrink-0">
        <Space>
          <Select
            placeholder="Interface"
            value={iface || undefined}
            onChange={setIface}
            style={{ width: 220 }}
            options={interfaces}
            notFoundContent="No interfaces found"
          />
          {/* TODO: dynamic protocol filter hints */}
          <Input placeholder="BPF filter (e.g. tcp port 80)" value={filter} onChange={e => setFilter(e.target.value)} style={{ width: 240 }} />
          <Button type="primary" icon={<PlayCircleOutlined />} onClick={handleStart} disabled={isCapturing}>Start</Button>
          <Button icon={<PauseCircleOutlined />} onClick={handleStop} disabled={!isCapturing}>Stop</Button>
        </Space>
      </Card>

      <Row gutter={16} className="mb-2 flex-shrink-0">
        <Col span={8}><Card size="small"><Statistic title="Packets/sec" value={packetsPerSec} /></Card></Col>
        <Col span={8}><Card size="small"><Statistic title="Total Packets" value={packets.length} /></Card></Col>
        <Col span={8}><Card size="small"><Statistic title="Throughput" value={`${(bytesPerSec / 1024).toFixed(1)} KB/s`} /></Card></Col>
      </Row>

      <Card className="flex-1 overflow-hidden" bodyStyle={{ padding: 0, height: '100%' }}>
        {packets.length === 0 ? (
          <div className="flex items-center justify-center h-full"><Empty description="Start a capture to see packets" /></div>
        ) : (
          <Table dataSource={packets.slice(-500)} columns={columns} rowKey="number" size="small" scroll={{ y: 'calc(100vh - 400px)' }} pagination={false} />
        )}
      </Card>
    </div>
  );
};

export default CapturePage;
