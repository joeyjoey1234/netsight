import React, { useState } from 'react';
import { Card, Button, Input, Space, Table, Tag, Statistic, Row, Col, Empty } from 'antd';
import { PlayCircleOutlined, PauseCircleOutlined, DownloadOutlined } from '@ant-design/icons';
import { useCaptureStore } from '../stores/captureStore';
import type { PacketSummary } from '../types';

const CapturePage: React.FC = () => {
  const { isCapturing, packets, packetsPerSec, bytesPerSec, filter, setFilter, setCapturing } = useCaptureStore();
  const [iface, setIface] = useState('');

  const columns = [
    { title: '#', dataIndex: 'number', width: 60 },
    { title: 'Time', dataIndex: 'timestamp', width: 100 },
    { title: 'Source', key: 'src', width: 180, render: (_: any, r: PacketSummary) => `${r.srcIp}:${r.srcPort || ''}` },
    { title: 'Destination', key: 'dst', width: 180, render: (_: any, r: PacketSummary) => `${r.dstIp}:${r.dstPort || ''}` },
    { title: 'Protocol', dataIndex: 'protocol', width: 80, render: (p: string) => <Tag color="blue">{p}</Tag> },
    { title: 'Length', dataIndex: 'length', width: 70 },
    { title: 'Info', dataIndex: 'info', ellipsis: true },
  ];

  return (
    <div className="h-full flex flex-col p-2">
      <Card size="small" className="mb-4">
        <Space>
          <Input placeholder="Interface (e.g. eth0)" value={iface} onChange={(e) => setIface(e.target.value)} style={{ width: 180 }} />
          <Input placeholder="BPF filter (e.g. tcp port 80)" value={filter} onChange={(e) => setFilter(e.target.value)} style={{ width: 240 }} />
          <Button type="primary" icon={<PlayCircleOutlined />} onClick={() => setCapturing(true)} disabled={isCapturing}>
            Start
          </Button>
          <Button icon={<PauseCircleOutlined />} onClick={() => setCapturing(false)} disabled={!isCapturing}>
            Stop
          </Button>
          <Button icon={<DownloadOutlined />} disabled={packets.length === 0}>
            Export PCAP
          </Button>
        </Space>
      </Card>

      <Row gutter={16} className="mb-4">
        <Col span={8}>
          <Card size="small"><Statistic title="Packets/sec" value={packetsPerSec} /></Card>
        </Col>
        <Col span={8}>
          <Card size="small"><Statistic title="Total Packets" value={packets.length} /></Card>
        </Col>
        <Col span={8}>
          <Card size="small"><Statistic title="Throughput" value={`${(bytesPerSec / 1024).toFixed(1)} KB/s`} /></Card>
        </Col>
      </Row>

      <Card className="flex-1 overflow-hidden" bodyStyle={{ padding: 0, height: '100%' }}>
        {packets.length === 0 ? (
          <div className="flex items-center justify-center h-full">
            <Empty description="Start a capture to see packets" />
          </div>
        ) : (
          <Table
            dataSource={packets.slice(0, 500)}
            columns={columns}
            rowKey="number"
            size="small"
            scroll={{ y: 'calc(100vh - 400px)' }}
            pagination={false}
          />
        )}
      </Card>
    </div>
  );
};

export default CapturePage;
