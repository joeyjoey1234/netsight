import React from 'react';
import { Card, Switch, Row, Col, Tag, Input, Space } from 'antd';
import { useServerStore } from '../stores/serverStore';

const serverDefs = [
  { type: 'tftp', label: 'TFTP Server', port: 69, color: 'blue' },
  { type: 'http', label: 'HTTP Server', port: 8080, color: 'green' },
  { type: 'ftp', label: 'FTP Server', port: 21, color: 'orange' },
  { type: 'syslog', label: 'Syslog Server', port: 514, color: 'purple' },
  { type: 'netcat', label: 'Netcat Listener', port: 4444, color: 'cyan' },
  { type: 'dhcp', label: 'DHCP Server', port: 67, color: 'red' },
  { type: 'ntp', label: 'NTP Server', port: 123, color: 'gold' },
  { type: 'dns', label: 'DNS Server', port: 53, color: 'magenta' },
];

const ServersPage: React.FC = () => {
  const { servers } = useServerStore();

  return (
    <div className="h-full overflow-y-auto p-2">
      <Row gutter={[16, 16]}>
        {serverDefs.map((srv) => {
          const state = servers[srv.type];
          const isRunning = state?.status === 'running';

          return (
            <Col span={12} key={srv.type}>
              <Card
                size="small"
                title={
                  <Space>
                    <Tag color={srv.color}>{srv.type.toUpperCase()}</Tag>
                    <span>{srv.label}</span>
                  </Space>
                }
                extra={
                  <Space>
                    <Tag color={isRunning ? 'success' : 'default'}>
                      {isRunning ? 'Running' : 'Stopped'}
                    </Tag>
                    <Switch checked={isRunning} />
                  </Space>
                }
              >
                <Space direction="vertical" style={{ width: '100%' }}>
                  <Input addonBefore="Port" defaultValue={srv.port} style={{ width: '100%' }} />
                  <Input addonBefore="Interface" placeholder="0.0.0.0" />
                  {srv.type === 'tftp' || srv.type === 'http' || srv.type === 'ftp' ? (
                    <Input addonBefore="Root Dir" placeholder="/tmp" />
                  ) : null}
                </Space>
              </Card>
            </Col>
          );
        })}
      </Row>
    </div>
  );
};

export default ServersPage;
