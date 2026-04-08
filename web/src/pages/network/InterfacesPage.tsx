import { useEffect, useState } from 'react';
import {
  Typography, Table, Button, Space, Tag, Card, Row, Col, Statistic, Modal, Descriptions,
} from 'antd';
import {
  ReloadOutlined, GlobalOutlined, ArrowUpOutlined, ArrowDownOutlined, InfoCircleOutlined,
} from '@ant-design/icons';
import { apiClient } from '@api/client';

const { Title } = Typography;

interface NetIface {
  id: string; name: string; alias: string; type: string; enabled: boolean;
  mtu: number; mac: string;
  ipv4_config?: { mode: string; address: string; gateway: string };
  status: {
    oper_state: string; speed: string; duplex: string;
    addresses: string[]; rx_bytes: number; tx_bytes: number;
    rx_packets: number; tx_packets: number; rx_errors: number; tx_errors: number;
  };
}

function fmtBytes(b: number): string {
  if (!b) return '0 B';
  const k = 1024, s = ['B','KB','MB','GB','TB'];
  const i = Math.floor(Math.log(b) / Math.log(k));
  return parseFloat((b / Math.pow(k, i)).toFixed(1)) + ' ' + s[i];
}

export default function InterfacesPage() {
  const [ifaces, setIfaces] = useState<NetIface[]>([]);
  const [loading, setLoading] = useState(false);
  const [detail, setDetail] = useState<NetIface | null>(null);

  const load = async () => {
    setLoading(true);
    try { const r = await apiClient.get('/network/interfaces'); setIfaces(r.data || []); } catch {}
    setLoading(false);
  };
  useEffect(() => { load(); }, []);

  const columns = [
    { title: 'Interface', key: 'name', render: (_: any, r: NetIface) => (
      <Space><GlobalOutlined /><b>{r.name}</b>{r.alias && <Tag>{r.alias}</Tag>}</Space>
    )},
    { title: 'Status', key: 'status', width: 80, render: (_: any, r: NetIface) => {
      const up = r.status.oper_state?.toLowerCase() === 'up';
      return <Tag color={up ? 'green' : 'red'}>{up ? 'UP' : 'DOWN'}</Tag>;
    }},
    { title: 'Type', dataIndex: 'type', width: 90, render: (t: string) => <Tag>{t}</Tag> },
    { title: 'IP Address', key: 'ip', render: (_: any, r: NetIface) => {
      const a = [...(r.status.addresses || [])];
      if (r.ipv4_config?.address && !a.includes(r.ipv4_config.address)) a.unshift(r.ipv4_config.address);
      return a.length ? a.map((x, i) => <div key={i}>{x}</div>) : <Tag>No IP</Tag>;
    }},
    { title: 'MAC', dataIndex: 'mac', width: 150, render: (m: string) => <code style={{ fontSize: 11 }}>{m || '-'}</code> },
    { title: 'Speed', key: 'speed', width: 90, render: (_: any, r: NetIface) => r.status.speed || '-' },
    { title: 'Traffic', key: 'traffic', width: 170, render: (_: any, r: NetIface) => (
      <Space direction="vertical" size={0} style={{ fontSize: 12 }}>
        <span><ArrowDownOutlined style={{ color: '#1677ff' }} /> {fmtBytes(r.status.rx_bytes)}</span>
        <span><ArrowUpOutlined style={{ color: '#52c41a' }} /> {fmtBytes(r.status.tx_bytes)}</span>
      </Space>
    )},
    { title: '', width: 40, render: (_: any, r: NetIface) => <Button type="link" size="small" icon={<InfoCircleOutlined />} onClick={() => setDetail(r)} /> },
  ];

  return (
    <div>
      <Row justify="space-between" align="middle" style={{ marginBottom: 16 }}>
        <Col><Title level={3} style={{ margin: 0 }}>Network Interfaces</Title></Col>
        <Col><Button icon={<ReloadOutlined />} onClick={load}>Refresh</Button></Col>
      </Row>
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={6}><Card><Statistic title="Total" value={ifaces.length} /></Card></Col>
        <Col span={6}><Card><Statistic title="Active" value={ifaces.filter(i => i.status.oper_state?.toLowerCase() === 'up').length} valueStyle={{ color: '#3f8600' }} /></Card></Col>
        <Col span={6}><Card><Statistic title="Total RX" value={fmtBytes(ifaces.reduce((s, i) => s + (i.status.rx_bytes||0), 0))} /></Card></Col>
        <Col span={6}><Card><Statistic title="Total TX" value={fmtBytes(ifaces.reduce((s, i) => s + (i.status.tx_bytes||0), 0))} /></Card></Col>
      </Row>
      <Card bodyStyle={{ padding: 0 }}>
        <Table dataSource={ifaces} columns={columns} rowKey="id" loading={loading} size="small" pagination={false} />
      </Card>
      <Modal title={`Interface: ${detail?.name}`} open={!!detail} onCancel={() => setDetail(null)} footer={null} width={550}>
        {detail && (
          <Descriptions column={2} bordered size="small">
            <Descriptions.Item label="Name">{detail.name}</Descriptions.Item>
            <Descriptions.Item label="Alias">{detail.alias || '-'}</Descriptions.Item>
            <Descriptions.Item label="Type">{detail.type}</Descriptions.Item>
            <Descriptions.Item label="MTU">{detail.mtu}</Descriptions.Item>
            <Descriptions.Item label="MAC">{detail.mac || '-'}</Descriptions.Item>
            <Descriptions.Item label="Speed">{detail.status.speed || '-'}</Descriptions.Item>
            <Descriptions.Item label="RX">{fmtBytes(detail.status.rx_bytes)} ({detail.status.rx_packets?.toLocaleString()} pkts)</Descriptions.Item>
            <Descriptions.Item label="TX">{fmtBytes(detail.status.tx_bytes)} ({detail.status.tx_packets?.toLocaleString()} pkts)</Descriptions.Item>
            <Descriptions.Item label="Errors">{detail.status.rx_errors + detail.status.tx_errors}</Descriptions.Item>
            <Descriptions.Item label="State">{detail.status.oper_state}</Descriptions.Item>
          </Descriptions>
        )}
      </Modal>
    </div>
  );
}
