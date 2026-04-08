import { useEffect, useState } from 'react';
import {
  Button,
  Card,
  Col,
  Form,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Row,
  Select,
  Space,
  Switch,
  Table,
  Tabs,
  Tag,
  Typography,
  message,
} from 'antd';
import { DeleteOutlined, PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import { apiClient } from '@api/client';
import type { Gateway, GatewayGroup } from '@typedefs/network';

const { Title } = Typography;

function statusColor(state: string): string {
  switch (state) {
    case 'online':
      return 'green';
    case 'offline':
      return 'red';
    default:
      return 'default';
  }
}

export default function GatewaysPage() {
  const [gateways, setGateways] = useState<Gateway[]>([]);
  const [groups, setGroups] = useState<GatewayGroup[]>([]);
  const [loading, setLoading] = useState(false);
  const [gatewayModalOpen, setGatewayModalOpen] = useState(false);
  const [groupModalOpen, setGroupModalOpen] = useState(false);
  const [gatewayForm] = Form.useForm();
  const [groupForm] = Form.useForm();

  const load = async () => {
    setLoading(true);
    try {
      const [gatewayResp, groupResp] = await Promise.all([
        apiClient.get('/network/gateways'),
        apiClient.get('/network/gateway-groups'),
      ]);
      setGateways(gatewayResp.data ?? []);
      setGroups(groupResp.data ?? []);
    } catch {
      message.error('Failed to load gateways');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void load();
  }, []);

  const handleCreateGateway = async () => {
    const values = await gatewayForm.validateFields();
    try {
      await apiClient.post('/network/gateways', {
        name: values.name,
        interface: values.interface,
        address: values.address,
        protocol: values.protocol ?? 'inet',
        monitor_ip: values.monitor_ip ?? '',
        weight: values.weight ?? 1,
        priority: values.priority ?? 10,
        is_default: values.is_default ?? false,
        description: values.description ?? '',
        monitor_interval: values.monitor_interval ?? 5,
        loss_threshold: values.loss_threshold ?? 20,
        latency_threshold: values.latency_threshold ?? 500,
        down_count: values.down_count ?? 3,
        monitor_method: values.monitor_method ?? 'icmp',
      });
      message.success('Gateway created');
      setGatewayModalOpen(false);
      gatewayForm.resetFields();
      void load();
    } catch {
      message.error('Failed to create gateway');
    }
  };

  const handleDeleteGateway = async (id: string) => {
    try {
      await apiClient.delete(`/network/gateways/${id}`);
      message.success('Gateway deleted');
      void load();
    } catch {
      message.error('Failed to delete gateway');
    }
  };

  const handleCreateGroup = async () => {
    const values = await groupForm.validateFields();
    try {
      await apiClient.post('/network/gateway-groups', {
        name: values.name,
        trigger: values.trigger ?? 'member_down',
        description: values.description ?? '',
        members: (values.members ?? []).map((member: { gateway_id: string; tier: number; weight: number }) => ({
          gateway_id: member.gateway_id,
          tier: member.tier ?? 1,
          weight: member.weight ?? 1,
        })),
      });
      message.success('Gateway group created');
      setGroupModalOpen(false);
      groupForm.resetFields();
      void load();
    } catch {
      message.error('Failed to create gateway group');
    }
  };

  const handleDeleteGroup = async (id: string) => {
    try {
      await apiClient.delete(`/network/gateway-groups/${id}`);
      message.success('Gateway group deleted');
      void load();
    } catch {
      message.error('Failed to delete gateway group');
    }
  };

  const gatewayColumns = [
    { title: 'Name', dataIndex: 'name' },
    { title: 'Interface', dataIndex: 'interface', width: 120 },
    { title: 'Address', dataIndex: 'address' },
    { title: 'Monitor', dataIndex: 'monitor_ip', render: (value: string) => value || '-' },
    { title: 'Priority', dataIndex: 'priority', width: 90 },
    { title: 'Weight', dataIndex: 'weight', width: 90 },
    {
      title: 'Status',
      key: 'status',
      width: 140,
      render: (_: unknown, gateway: Gateway) => (
        <Space>
          <Tag color={statusColor(gateway.status?.state)}>{(gateway.status?.state || 'unknown').toUpperCase()}</Tag>
          {gateway.status?.latency_ms ? `${gateway.status.latency_ms.toFixed(1)} ms` : ''}
        </Space>
      ),
    },
    { title: 'Default', dataIndex: 'is_default', width: 100, render: (value: boolean) => value ? <Tag color="gold">DEFAULT</Tag> : '-' },
    {
      title: '',
      key: 'actions',
      width: 64,
      render: (_: unknown, gateway: Gateway) => (
        <Popconfirm title="Delete gateway?" onConfirm={() => void handleDeleteGateway(gateway.id)}>
          <Button type="link" size="small" danger icon={<DeleteOutlined />} />
        </Popconfirm>
      ),
    },
  ];

  const groupColumns = [
    { title: 'Name', dataIndex: 'name' },
    { title: 'Trigger', dataIndex: 'trigger', width: 140, render: (trigger: string) => <Tag color="blue">{trigger}</Tag> },
    {
      title: 'Members',
      key: 'members',
      render: (_: unknown, group: GatewayGroup) => group.members.map((member) => `${member.gateway_id} (tier ${member.tier}, weight ${member.weight})`).join(', '),
    },
    { title: 'Description', dataIndex: 'description' },
    {
      title: '',
      key: 'actions',
      width: 64,
      render: (_: unknown, group: GatewayGroup) => (
        <Popconfirm title="Delete gateway group?" onConfirm={() => void handleDeleteGroup(group.id)}>
          <Button type="link" size="small" danger icon={<DeleteOutlined />} />
        </Popconfirm>
      ),
    },
  ];

  return (
    <div>
      <Row justify="space-between" align="middle" style={{ marginBottom: 16 }}>
        <Col><Title level={3} style={{ margin: 0 }}>Gateways</Title></Col>
        <Col>
          <Space>
            <Button icon={<ReloadOutlined />} onClick={() => void load()}>Refresh</Button>
            <Button onClick={() => setGroupModalOpen(true)}>Add Group</Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => setGatewayModalOpen(true)}>Add Gateway</Button>
          </Space>
        </Col>
      </Row>

      <Card bodyStyle={{ paddingTop: 12 }}>
        <Tabs
          items={[
            {
              key: 'gateways',
              label: `Gateways (${gateways.length})`,
              children: <Table dataSource={gateways} columns={gatewayColumns} rowKey="id" loading={loading} size="small" pagination={{ pageSize: 25 }} />,
            },
            {
              key: 'groups',
              label: `Gateway Groups (${groups.length})`,
              children: <Table dataSource={groups} columns={groupColumns} rowKey="id" loading={loading} size="small" pagination={{ pageSize: 25 }} />,
            },
          ]}
        />
      </Card>

      <Modal title="Create Gateway" open={gatewayModalOpen} onOk={() => void handleCreateGateway()} onCancel={() => setGatewayModalOpen(false)} width={760}>
        <Form form={gatewayForm} layout="vertical">
          <Row gutter={16}>
            <Col span={8}><Form.Item name="name" label="Name" rules={[{ required: true }]}><Input placeholder="wan-primary" /></Form.Item></Col>
            <Col span={8}><Form.Item name="interface" label="Interface" rules={[{ required: true }]}><Input placeholder="wan0" /></Form.Item></Col>
            <Col span={8}><Form.Item name="address" label="Address" rules={[{ required: true }]}><Input placeholder="192.0.2.1" /></Form.Item></Col>
          </Row>
          <Row gutter={16}>
            <Col span={8}><Form.Item name="protocol" label="Protocol"><Select options={[{ value: 'inet', label: 'IPv4' }, { value: 'inet6', label: 'IPv6' }]} /></Form.Item></Col>
            <Col span={8}><Form.Item name="monitor_ip" label="Monitor IP"><Input placeholder="1.1.1.1" /></Form.Item></Col>
            <Col span={8}><Form.Item name="description" label="Description"><Input placeholder="Primary uplink" /></Form.Item></Col>
          </Row>
          <Row gutter={16}>
            <Col span={6}><Form.Item name="weight" label="Weight"><InputNumber min={1} style={{ width: '100%' }} /></Form.Item></Col>
            <Col span={6}><Form.Item name="priority" label="Priority"><InputNumber min={1} style={{ width: '100%' }} /></Form.Item></Col>
            <Col span={6}><Form.Item name="monitor_interval" label="Interval"><InputNumber min={1} style={{ width: '100%' }} /></Form.Item></Col>
            <Col span={6}><Form.Item name="down_count" label="Down Count"><InputNumber min={1} style={{ width: '100%' }} /></Form.Item></Col>
          </Row>
          <Row gutter={16}>
            <Col span={8}><Form.Item name="loss_threshold" label="Loss Threshold (%)"><InputNumber min={0} max={100} style={{ width: '100%' }} /></Form.Item></Col>
            <Col span={8}><Form.Item name="latency_threshold" label="Latency Threshold (ms)"><InputNumber min={1} style={{ width: '100%' }} /></Form.Item></Col>
            <Col span={8}><Form.Item name="monitor_method" label="Monitor Method"><Select options={[{ value: 'icmp', label: 'ICMP' }]} /></Form.Item></Col>
          </Row>
          <Form.Item name="is_default" label="Default Gateway" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>

      <Modal title="Create Gateway Group" open={groupModalOpen} onOk={() => void handleCreateGroup()} onCancel={() => setGroupModalOpen(false)} width={760}>
        <Form form={groupForm} layout="vertical" initialValues={{ trigger: 'member_down', members: [{ tier: 1, weight: 1 }] }}>
          <Row gutter={16}>
            <Col span={12}><Form.Item name="name" label="Name" rules={[{ required: true }]}><Input placeholder="wan-failover" /></Form.Item></Col>
            <Col span={12}><Form.Item name="trigger" label="Trigger"><Select options={[{ value: 'member_down', label: 'Member Down' }, { value: 'packet_loss', label: 'Packet Loss' }, { value: 'high_latency', label: 'High Latency' }]} /></Form.Item></Col>
          </Row>
          <Form.Item name="description" label="Description">
            <Input placeholder="Primary/failover uplinks" />
          </Form.Item>
          <Form.List name="members">
            {(fields, { add, remove }) => (
              <>
                {fields.map((field) => (
                  <Row gutter={12} key={field.key} align="middle">
                    <Col span={10}>
                      <Form.Item {...field} name={[field.name, 'gateway_id']} label="Gateway" rules={[{ required: true }]}>
                        <Select options={gateways.map((gateway) => ({ value: gateway.id, label: gateway.name }))} />
                      </Form.Item>
                    </Col>
                    <Col span={5}>
                      <Form.Item {...field} name={[field.name, 'tier']} label="Tier" rules={[{ required: true }]}>
                        <InputNumber min={1} style={{ width: '100%' }} />
                      </Form.Item>
                    </Col>
                    <Col span={5}>
                      <Form.Item {...field} name={[field.name, 'weight']} label="Weight" rules={[{ required: true }]}>
                        <InputNumber min={1} style={{ width: '100%' }} />
                      </Form.Item>
                    </Col>
                    <Col span={4}>
                      <Button danger style={{ marginTop: 30 }} onClick={() => remove(field.name)}>Remove</Button>
                    </Col>
                  </Row>
                ))}
                <Button onClick={() => add({ tier: 1, weight: 1 })}>Add Member</Button>
              </>
            )}
          </Form.List>
        </Form>
      </Modal>
    </div>
  );
}
