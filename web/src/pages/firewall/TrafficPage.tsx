import { useState } from 'react';
import {
  Button,
  Card,
  Col,
  Form,
  Input,
  InputNumber,
  Row,
  Select,
  Space,
  Table,
  Tag,
  Typography,
  message,
} from 'antd';
import { DownloadOutlined, ReloadOutlined } from '@ant-design/icons';
import { apiClient } from '@api/client';
import type { CapturedPacket, PCAPExport } from '@typedefs/firewall';

const { Title } = Typography;

export default function TrafficPage() {
  const [packets, setPackets] = useState<CapturedPacket[]>([]);
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();

  const capture = async () => {
    const values = await form.validateFields();
    setLoading(true);
    try {
      const response = await apiClient.get('/firewall/traffic', {
        params: {
          interface: values.interface ?? '',
          src: values.source_ip ?? '',
          dst: values.dest_ip ?? '',
          protocol: values.protocol ?? '',
          count: values.count ?? 50,
          snap_len: values.snap_len ?? 160,
        },
      });
      setPackets(response.data ?? []);
    } catch {
      message.error('Failed to capture traffic');
    } finally {
      setLoading(false);
    }
  };

  const exportPCAP = async () => {
    const values = await form.validateFields();
    setLoading(true);
    try {
      const response = await apiClient.post<PCAPExport>('/firewall/traffic/export', {
        interface: values.interface ?? '',
        source_ip: values.source_ip ?? '',
        dest_ip: values.dest_ip ?? '',
        protocol: values.protocol ?? '',
        count: values.count ?? 100,
        snap_len: values.snap_len ?? 256,
      });
      if (response.data?.download_url) {
        window.open(response.data.download_url, '_blank', 'noopener,noreferrer');
        message.success(`PCAP exported: ${response.data.name}`);
      }
    } catch {
      message.error('Failed to export traffic');
    } finally {
      setLoading(false);
    }
  };

  const columns = [
    { title: 'Time', dataIndex: 'timestamp', width: 140 },
    { title: 'Interface', dataIndex: 'interface', width: 120, render: (iface: string) => iface || 'any' },
    { title: 'Protocol', dataIndex: 'protocol', width: 100, render: (protocol: string) => <Tag color="blue">{protocol || 'unknown'}</Tag> },
    { title: 'Source', dataIndex: 'source' },
    { title: 'Destination', dataIndex: 'destination' },
    { title: 'Length', dataIndex: 'length', width: 90 },
    { title: 'Verdict', dataIndex: 'verdict', width: 110, render: (verdict: string) => <Tag color="geekblue">{verdict}</Tag> },
    { title: 'Summary', dataIndex: 'summary' },
  ];

  return (
    <div>
      <Row justify="space-between" align="middle" style={{ marginBottom: 16 }}>
        <Col><Title level={3} style={{ margin: 0 }}>Live Traffic</Title></Col>
      </Row>

      <Card style={{ marginBottom: 16 }}>
        <Form
          form={form}
          layout="vertical"
          initialValues={{ interface: 'any', protocol: 'any', count: 50, snap_len: 160 }}
        >
          <Row gutter={16}>
            <Col span={6}>
              <Form.Item name="interface" label="Interface">
                <Input placeholder="any / wan0 / lan0" />
              </Form.Item>
            </Col>
            <Col span={5}>
              <Form.Item name="protocol" label="Protocol">
                <Select options={[{ value: 'any', label: 'Any' }, { value: 'tcp', label: 'TCP' }, { value: 'udp', label: 'UDP' }, { value: 'icmp', label: 'ICMP' }, { value: 'icmp6', label: 'ICMPv6' }]} />
              </Form.Item>
            </Col>
            <Col span={5}>
              <Form.Item name="source_ip" label="Source">
                <Input placeholder="192.168.1.10" />
              </Form.Item>
            </Col>
            <Col span={5}>
              <Form.Item name="dest_ip" label="Destination">
                <Input placeholder="8.8.8.8" />
              </Form.Item>
            </Col>
            <Col span={3}>
              <Form.Item name="count" label="Packets">
                <InputNumber min={1} max={500} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={4}>
              <Form.Item name="snap_len" label="SnapLen">
                <InputNumber min={64} max={4096} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={20}>
              <Space style={{ marginTop: 30 }}>
                <Button type="primary" icon={<ReloadOutlined />} loading={loading} onClick={() => void capture()}>Capture Snapshot</Button>
                <Button icon={<DownloadOutlined />} loading={loading} onClick={() => void exportPCAP()}>Export PCAP</Button>
              </Space>
            </Col>
          </Row>
        </Form>
      </Card>

      <Card bodyStyle={{ padding: 0 }}>
        <Table
          dataSource={packets}
          columns={columns}
          rowKey={(packet, idx) => `${packet.timestamp}-${packet.source}-${packet.destination}-${idx}`}
          loading={loading}
          size="small"
          pagination={{ pageSize: 25 }}
          expandable={{
            expandedRowRender: (packet) => <pre style={{ margin: 0, whiteSpace: 'pre-wrap' }}>{packet.detail}</pre>,
          }}
        />
      </Card>
    </div>
  );
}
