import React from 'react';
import { Card, Input, Button, Select, Space, Row, Col } from 'antd';
import { SendOutlined, NodeIndexOutlined, SearchOutlined } from '@ant-design/icons';

const ToolsPage: React.FC = () => {
  return (
    <div className="h-full overflow-y-auto p-2">
      <Row gutter={16}>
        <Col span={12}>
          <Card title="Ping" size="small" className="mb-4">
            <Space direction="vertical" style={{ width: '100%' }}>
              <Input placeholder="Target IP or hostname" />
              <Space>
                <Select defaultValue={4} style={{ width: 80 }} options={[
                  { value: 4, label: '4 pings' },
                  { value: 10, label: '10 pings' },
                  { value: 0, label: 'Continuous' },
                ]} />
                <Button type="primary" icon={<SendOutlined />}>Ping</Button>
              </Space>
            </Space>
          </Card>
        </Col>

        <Col span={12}>
          <Card title="Traceroute" size="small" className="mb-4">
            <Space direction="vertical" style={{ width: '100%' }}>
              <Input placeholder="Target IP or hostname" />
              <Space>
                <Select defaultValue="icmp" style={{ width: 100 }} options={[
                  { value: 'icmp', label: 'ICMP' },
                  { value: 'tcp', label: 'TCP' },
                  { value: 'udp', label: 'UDP' },
                ]} />
                <Button type="primary" icon={<NodeIndexOutlined />}>Trace</Button>
              </Space>
            </Space>
          </Card>
        </Col>

        <Col span={12}>
          <Card title="NSLookup" size="small" className="mb-4">
            <Space direction="vertical" style={{ width: '100%' }}>
              <Input placeholder="Hostname or IP" />
              <Select mode="multiple" defaultValue={['A']} style={{ width: '100%' }} options={[
                { value: 'A', label: 'A (IPv4)' },
                { value: 'AAAA', label: 'AAAA (IPv6)' },
                { value: 'MX', label: 'MX' },
                { value: 'NS', label: 'NS' },
                { value: 'TXT', label: 'TXT' },
                { value: 'PTR', label: 'PTR' },
                { value: 'SOA', label: 'SOA' },
              ]} />
              <Button type="primary" icon={<SearchOutlined />}>Lookup</Button>
            </Space>
          </Card>
        </Col>

        <Col span={12}>
          <Card title="Wake-on-LAN" size="small" className="mb-4">
            <Space direction="vertical" style={{ width: '100%' }}>
              <Input placeholder="MAC Address (e.g. AA:BB:CC:DD:EE:FF)" />
              <Button type="primary">Send Magic Packet</Button>
            </Space>
          </Card>
        </Col>

        <Col span={12}>
          <Card title="iPerf" size="small" className="mb-4">
            <Space direction="vertical" style={{ width: '100%' }}>
              <Input placeholder="Target IP" />
              <Space>
                <Select defaultValue="client" style={{ width: 100 }} options={[
                  { value: 'client', label: 'Client' },
                  { value: 'server', label: 'Server' },
                ]} />
                <Input placeholder="Duration (s)" defaultValue="10" style={{ width: 100 }} />
                <Button type="primary">Run</Button>
              </Space>
            </Space>
          </Card>
        </Col>

        <Col span={12}>
          <Card title="Latency Monitor" size="small" className="mb-4">
            <Space direction="vertical" style={{ width: '100%' }}>
              <Input placeholder="Target IP" />
              <Button type="primary">Start Monitoring</Button>
              <div className="h-32 bg-gray-800 rounded flex items-center justify-center text-gray-500 text-sm">
                Latency graph area
              </div>
            </Space>
          </Card>
        </Col>

        <Col span={12}>
          <Card title="Subnet Calculator" size="small" className="mb-4">
            <Space direction="vertical" style={{ width: '100%' }}>
              <Input placeholder="e.g. 192.168.1.0/24" />
              <Button type="primary">Calculate</Button>
            </Space>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default ToolsPage;
