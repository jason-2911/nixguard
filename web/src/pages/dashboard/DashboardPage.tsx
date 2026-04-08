import { useCallback, useEffect, useState } from 'react';
import { Card, Col, Row, Statistic, Table, Tag, Typography, message } from 'antd';
import {
  SafetyOutlined,
  GlobalOutlined,
  ApartmentOutlined,
  ClusterOutlined,
} from '@ant-design/icons';
import { useAutoRefresh } from '@hooks/useAutoRefresh';
import { apiClient } from '@api/client';
import type { FirewallRule } from '@typedefs/firewall';
import type { Gateway } from '@typedefs/network';

const { Title } = Typography;

interface DashboardInterface {
  id: string;
  status?: { oper_state?: string; rx_bytes?: number; tx_bytes?: number };
}

function formatBytes(bytes: number): string {
  if (!bytes) {
    return '0 B';
  }
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const exponent = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
  return `${(bytes / 1024 ** exponent).toFixed(1)} ${units[exponent]}`;
}

export default function DashboardPage() {
  const [rules, setRules] = useState<FirewallRule[]>([]);
  const [interfaces, setInterfaces] = useState<DashboardInterface[]>([]);
  const [gateways, setGateways] = useState<Gateway[]>([]);
  const [loading, setLoading] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [ruleResp, ifaceResp, gatewayResp] = await Promise.all([
        apiClient.get('/firewall/rules'),
        apiClient.get('/network/interfaces'),
        apiClient.get('/network/gateways'),
      ]);
      setRules(ruleResp.data ?? []);
      setInterfaces(ifaceResp.data ?? []);
      setGateways(gatewayResp.data ?? []);
    } catch {
      message.error('Failed to load dashboard data');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { void load(); }, [load]);
  useAutoRefresh(() => { void load(); }, 10000);

  const activeInterfaces = interfaces.filter((iface) => iface.status?.oper_state?.toLowerCase() === 'up').length;
  const totalRx = interfaces.reduce((sum, iface) => sum + (iface.status?.rx_bytes ?? 0), 0);
  const totalTx = interfaces.reduce((sum, iface) => sum + (iface.status?.tx_bytes ?? 0), 0);

  return (
    <div>
      <Title level={3}>Dashboard</Title>

      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading}>
            <Statistic title="Firewall Rules" value={rules.length} prefix={<SafetyOutlined />} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading}>
            <Statistic title="Active Interfaces" value={activeInterfaces} prefix={<GlobalOutlined />} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading}>
            <Statistic title="Configured Gateways" value={gateways.length} prefix={<ApartmentOutlined />} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading}>
            <Statistic title="Floating Rules" value={rules.filter((rule) => rule.is_floating).length} prefix={<ClusterOutlined />} />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading}>
            <Statistic title="Total RX" value={formatBytes(totalRx)} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading}>
            <Statistic title="Total TX" value={formatBytes(totalTx)} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading}>
            <Statistic title="Logged Rules" value={rules.filter((rule) => rule.log).length} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading}>
            <Statistic title="Online Gateways" value={gateways.filter((gateway) => gateway.status?.state === 'online').length} />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} lg={12}>
          <Card title="Gateway Status" loading={loading}>
            <Table
              dataSource={gateways}
              rowKey="id"
              size="small"
              pagination={false}
              columns={[
                { title: 'Gateway', dataIndex: 'name' },
                { title: 'Address', dataIndex: 'address' },
                { title: 'Priority', dataIndex: 'priority', width: 90 },
                {
                  title: 'Status',
                  key: 'status',
                  render: (_: unknown, gateway: Gateway) => <Tag color={gateway.status?.state === 'online' ? 'green' : 'red'}>{(gateway.status?.state || 'unknown').toUpperCase()}</Tag>,
                },
                {
                  title: 'Latency',
                  key: 'latency',
                  render: (_: unknown, gateway: Gateway) => gateway.status?.latency_ms ? `${gateway.status.latency_ms.toFixed(1)} ms` : '-',
                },
              ]}
            />
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card title="Top Rules by Hits" loading={loading}>
            <Table
              dataSource={[...rules].sort((a, b) => (b.stats?.packets ?? 0) - (a.stats?.packets ?? 0)).slice(0, 8)}
              rowKey="id"
              size="small"
              pagination={false}
              columns={[
                { title: 'Rule', dataIndex: 'description', render: (value: string) => value || 'Untitled rule' },
                { title: 'Interface', dataIndex: 'interface', width: 110, render: (value: string) => value || 'any' },
                { title: 'Action', dataIndex: 'action', width: 90, render: (action: string) => <Tag>{action.toUpperCase()}</Tag> },
                { title: 'Packets', key: 'packets', width: 110, render: (_: unknown, rule: FirewallRule) => (rule.stats?.packets ?? 0).toLocaleString() },
              ]}
            />
          </Card>
        </Col>
      </Row>
    </div>
  );
}
