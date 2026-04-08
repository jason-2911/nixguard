import { useEffect, useState } from 'react';
import {
  Button,
  Card,
  Col,
  Form,
  Input,
  Modal,
  Popconfirm,
  Row,
  Space,
  Table,
  Tabs,
  Tag,
  Typography,
  message,
} from 'antd';
import { DeleteOutlined, PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import { apiClient } from '@api/client';
import type { NetworkRoute } from '@typedefs/network';

const { Title } = Typography;

export default function RoutingPage() {
  const [routes, setRoutes] = useState<NetworkRoute[]>([]);
  const [systemRoutes, setSystemRoutes] = useState<NetworkRoute[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [form] = Form.useForm();

  const load = async () => {
    setLoading(true);
    try {
      const [saved, system] = await Promise.all([
        apiClient.get('/network/routes'),
        apiClient.get('/network/routes/system'),
      ]);
      setRoutes(saved.data ?? []);
      setSystemRoutes(system.data ?? []);
    } catch {
      message.error('Failed to load routes');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void load();
  }, []);

  const handleSubmit = async () => {
    const values = await form.validateFields();
    try {
      await apiClient.post('/network/routes', {
        destination: values.destination,
        gateway: values.gateway ?? '',
        interface: values.interface ?? '',
        metric: Number(values.metric ?? 0),
        table: Number(values.table ?? 254),
        description: values.description ?? '',
      });
      message.success('Route created');
      setModalOpen(false);
      form.resetFields();
      void load();
    } catch {
      message.error('Failed to create route');
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await apiClient.delete(`/network/routes/${id}`);
      message.success('Route deleted');
      void load();
    } catch {
      message.error('Failed to delete route');
    }
  };

  const savedColumns = [
    { title: 'Destination', dataIndex: 'destination' },
    { title: 'Gateway', dataIndex: 'gateway', render: (gateway: string) => gateway || '-' },
    { title: 'Interface', dataIndex: 'interface', render: (iface: string) => iface || '-' },
    { title: 'Metric', dataIndex: 'metric', width: 90 },
    { title: 'Table', dataIndex: 'table', width: 90 },
    { title: 'Type', dataIndex: 'type', width: 100, render: (type: string) => <Tag>{type.toUpperCase()}</Tag> },
    { title: 'Description', dataIndex: 'description' },
    {
      title: '',
      key: 'actions',
      width: 64,
      render: (_: unknown, route: NetworkRoute) => (
        <Popconfirm title="Delete route?" onConfirm={() => void handleDelete(route.id)}>
          <Button type="link" size="small" danger icon={<DeleteOutlined />} />
        </Popconfirm>
      ),
    },
  ];

  const systemColumns = [
    { title: 'Destination', dataIndex: 'destination' },
    { title: 'Gateway', dataIndex: 'gateway', render: (gateway: string) => gateway || '-' },
    { title: 'Interface', dataIndex: 'interface', render: (iface: string) => iface || '-' },
    { title: 'Metric', dataIndex: 'metric', width: 90 },
    { title: 'Type', dataIndex: 'type', width: 100, render: (type: string) => <Tag color="blue">{type.toUpperCase()}</Tag> },
  ];

  return (
    <div>
      <Row justify="space-between" align="middle" style={{ marginBottom: 16 }}>
        <Col><Title level={3} style={{ margin: 0 }}>Routing</Title></Col>
        <Col>
          <Space>
            <Button icon={<ReloadOutlined />} onClick={() => void load()}>Refresh</Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>Add Route</Button>
          </Space>
        </Col>
      </Row>

      <Card bodyStyle={{ paddingTop: 12 }}>
        <Tabs
          items={[
            {
              key: 'saved',
              label: `Configured Routes (${routes.length})`,
              children: <Table dataSource={routes} columns={savedColumns} rowKey="id" loading={loading} size="small" pagination={{ pageSize: 25 }} />,
            },
            {
              key: 'system',
              label: `Kernel Routes (${systemRoutes.length})`,
              children: <Table dataSource={systemRoutes} columns={systemColumns} rowKey={(route) => `${route.destination}-${route.gateway}-${route.interface}`} loading={loading} size="small" pagination={{ pageSize: 25 }} />,
            },
          ]}
        />
      </Card>

      <Modal title="Create Static Route" open={modalOpen} onOk={() => void handleSubmit()} onCancel={() => setModalOpen(false)} width={720}>
        <Form form={form} layout="vertical">
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="destination" label="Destination" rules={[{ required: true, message: 'Destination is required' }]}>
                <Input placeholder="10.10.0.0/16 or 2001:db8:100::/64" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="gateway" label="Gateway">
                <Input placeholder="192.0.2.1 or 2001:db8::1" />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={8}>
              <Form.Item name="interface" label="Interface">
                <Input placeholder="wan0" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="metric" label="Metric">
                <Input placeholder="10" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="table" label="Routing Table">
                <Input placeholder="254" />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item name="description" label="Description">
            <Input placeholder="Backup path to branch office" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
