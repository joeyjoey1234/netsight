import React, { useEffect, useState } from 'react';
import { Card, Switch, Row, Col, Tag, Input, Space, message } from 'antd';
import { useServerStore } from '../stores/serverStore';
import { startServer, stopServer, onEvent } from '../hooks/api';
import type { ServerState } from '../types';

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
  const { servers, updateServer } = useServerStore();
  const [configs, setConfigs] = useState<Record<string, { port: number; iface: string; rootDir: string }>>({});

  useEffect(() => {
    const unsub = onEvent('server:status', (state: ServerState) => {
      console.error('[ServerPage] server:status received:', state);
      updateServer(state);
    });
    return unsub;
  }, [updateServer]);

  const getConfigValue = (type: string, key: string, fallback: any) => {
    return configs[type]?.[key as keyof typeof configs[string]] ?? fallback;
  };

  const updateConfig = (type: string, key: string, value: any) => {
    setConfigs(prev => ({
      ...prev,
      [type]: { ...prev[type], [key]: value },
    }));
  };

  const handleToggle = async (type: string, checked: boolean) => {
    try {
      if (checked) {
        const cfg = {
          port: getConfigValue(type, 'port', serverDefs.find(s => s.type === type)?.port || 0),
          interface: getConfigValue(type, 'iface', '0.0.0.0'),
          rootDir: getConfigValue(type, 'rootDir', ''),
        };
        await startServer(type, cfg);
        message.success(`${type.toUpperCase()} server started`);
      } else {
        await stopServer(type);
        message.info(`${type.toUpperCase()} server stopped`);
      }
    } catch (err: any) {
      message.error(`${type}: ${err?.message || err}`);
    }
  };

  return (
    <div className="h-full overflow-y-auto p-2">
      <Row gutter={[16, 16]}>
        {serverDefs.map(srv => {
          const state = servers[srv.type];
          const isRunning = state?.status === 'running';

          return (
            <Col span={12} key={srv.type}>
              <Card
                size="small"
                title={<Space><Tag color={srv.color}>{srv.type.toUpperCase()}</Tag><span>{srv.label}</span></Space>}
                extra={
                  <Space>
                    <Tag color={isRunning ? 'success' : 'default'}>{isRunning ? 'Running' : 'Stopped'}</Tag>
                    <Switch checked={isRunning} onChange={(v) => handleToggle(srv.type, v)} />
                  </Space>
                }
              >
                <Space direction="vertical" style={{ width: '100%' }}>
                  <Input
                    addonBefore="Port"
                    defaultValue={srv.port}
                    onChange={e => updateConfig(srv.type, 'port', parseInt(e.target.value) || srv.port)}
                    style={{ width: '100%' }}
                  />
                  <Input
                    addonBefore="Interface"
                    placeholder="0.0.0.0"
                    onChange={e => updateConfig(srv.type, 'iface', e.target.value)}
                  />
                  {(srv.type === 'tftp' || srv.type === 'http' || srv.type === 'ftp') && (
                    <Input
                      addonBefore="Root Dir"
                      placeholder="/tmp"
                      onChange={e => updateConfig(srv.type, 'rootDir', e.target.value)}
                    />
                  )}
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
