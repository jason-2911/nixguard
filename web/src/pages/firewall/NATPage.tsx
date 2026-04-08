import { useCallback, useEffect, useState } from 'react';
import {
  Button,
  Card,
  Col,
  Form,
  Input,
  Modal,
  Popconfirm,
  Row,
  Select,
  Space,
  Switch,
  Table,
  Tag,
  Typography,
  message,
} from 'antd';
import { DeleteOutlined, EditOutlined, PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import { useAutoRefresh } from '@hooks/useAutoRefresh';
import { apiClient } from '@api/client';
import type { Address, NATRule } from '@typedefs/firewall';

const { Title } = Typography;

function formatAddress(address: Address): string {
  if (address.type === 'any') {
    return '*';
  }
  return address.port ? `${address.value}:${address.port}` : address.value;
}

export default function NATPage() {
  const [rules, setRules] = useState<NATRule[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingRule, setEditingRule] = useState<NATRule | null>(null);
  const [form] = Form.useForm();

  const loadRules = async () => {
    setLoading(true);
    try {
      const response = await apiClient.get('/firewall/nat');
      setRules(response.data ?? []);
    } catch {
      message.error('Failed to load NAT rules');
    } finally {
      setLoading(false);
    }
  };

  const refresh = useCallback(() => { void loadRules(); }, []);
  useEffect(() => { refresh(); }, [refresh]);
  useAutoRefresh(refresh, 5000);

  const openCreate = () => {
    setEditingRule(null);
    form.resetFields();
    form.setFieldsValue({
      type: 'port_forward',
      protocol: 'tcp',
      source_type: 'any',
      dest_type: 'single',
      nat_reflection: false,
      enabled: true,
    });
    setModalOpen(true);
  };

  const openEdit = (rule: NATRule) => {
    setEditingRule(rule);
    form.setFieldsValue({
      type: rule.type,
      interface: rule.interface,
      protocol: rule.protocol,
      source_type: rule.source.type,
      source_value: rule.source.value,
      source_port: rule.source.port,
      dest_type: rule.destination.type,
      dest_value: rule.destination.value,
      dest_port: rule.destination.port,
      redirect_target: rule.redirect_target,
      redirect_port: rule.redirect_port,
      nat_reflection: rule.nat_reflection,
      enabled: rule.enabled,
      description: rule.description,
    });
    setModalOpen(true);
  };

  const handleSubmit = async () => {
    const values = await form.validateFields();
    const payload = {
      type: values.type,
      interface: values.interface,
      protocol: values.protocol,
      source: {
        type: values.source_type ?? 'any',
        value: values.source_value ?? '',
        port: values.source_port ?? '',
        not: false,
      },
      destination: {
        type: values.dest_type ?? 'any',
        value: values.dest_value ?? '',
        port: values.dest_port ?? '',
        not: false,
      },
      redirect_target: values.redirect_target,
      redirect_port: values.redirect_port ?? '',
      nat_reflection: values.nat_reflection ?? false,
      enabled: values.enabled ?? true,
      description: values.description ?? '',
    };

    try {
      if (editingRule) {
        await apiClient.put(`/firewall/nat/${editingRule.id}`, payload);
        message.success('NAT rule updated');
      } else {
        await apiClient.post('/firewall/nat', payload);
        message.success('NAT rule created');
      }
      setModalOpen(false);
      void loadRules();
    } catch {
      message.error('Failed to save NAT rule');
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await apiClient.delete(`/firewall/nat/${id}`);
      message.success('NAT rule deleted');
      void loadRules();
    } catch {
      message.error('Failed to delete NAT rule');
    }
  };

  const columns = [
    {
      title: 'Type',
      dataIndex: 'type',
      width: 140,
      render: (type: NATRule['type']) => <Tag color="purple">{type.toUpperCase()}</Tag>,
    },
    {
      title: 'Interface',
      dataIndex: 'interface',
      width: 120,
    },
    {
      title: 'Protocol',
      dataIndex: 'protocol',
      width: 100,
      render: (protocol: string) => <Tag>{protocol.toUpperCase()}</Tag>,
    },
    {
      title: 'Match',
      key: 'match',
      render: (_: unknown, rule: NATRule) => `${formatAddress(rule.source)} -> ${formatAddress(rule.destination)}`,
    },
    {
      title: 'Target',
      key: 'target',
      render: (_: unknown, rule: NATRule) => `${rule.redirect_target}${rule.redirect_port ? `:${rule.redirect_port}` : ''}`,
    },
    {
      title: 'Reflection',
      dataIndex: 'nat_reflection',
      width: 110,
      render: (enabled: boolean) => enabled ? <Tag color="blue">ON</Tag> : '-',
    },
    {
      title: 'Description',
      dataIndex: 'description',
    },
    {
      title: '',
      key: 'actions',
      width: 96,
      render: (_: unknown, rule: NATRule) => (
        <Space>
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => openEdit(rule)} />
          <Popconfirm title="Delete NAT rule?" onConfirm={() => void handleDelete(rule.id)}>
            <Button type="link" size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Row justify="space-between" align="middle" style={{ marginBottom: 16 }}>
        <Col><Title level={3} style={{ margin: 0 }}>NAT Rules</Title></Col>
        <Col>
          <Space>
            <Button icon={<ReloadOutlined />} onClick={() => void loadRules()}>Refresh</Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>Add NAT Rule</Button>
          </Space>
        </Col>
      </Row>

      <Card bodyStyle={{ padding: 0 }}>
        <Table dataSource={rules} columns={columns} rowKey="id" loading={loading} size="small" pagination={{ pageSize: 25 }} />
      </Card>

      <Modal
        title={editingRule ? 'Edit NAT Rule' : 'Create NAT Rule'}
        open={modalOpen}
        onOk={() => void handleSubmit()}
        onCancel={() => setModalOpen(false)}
        width={800}
      >
        <Form form={form} layout="vertical">
          <Row gutter={16}>
            <Col span={8}>
              <Form.Item name="type" label="NAT Type" rules={[{ required: true }]}>
                <Select
                  options={[
                    { value: 'port_forward', label: 'Port Forward' },
                    { value: 'one_to_one', label: '1:1 NAT' },
                    { value: 'outbound', label: 'Outbound NAT' },
                  ]}
                />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="interface" label="Interface" rules={[{ required: true }]}>
                <Input placeholder="wan0" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="protocol" label="Protocol" rules={[{ required: true }]}>
                <Select
                  options={[
                    { value: 'any', label: 'Any' },
                    { value: 'tcp', label: 'TCP' },
                    { value: 'udp', label: 'UDP' },
                  ]}
                />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={8}><Form.Item name="source_type" label="Source Type"><Select options={[{ value: 'any', label: 'Any' }, { value: 'single', label: 'IP' }, { value: 'network', label: 'Network' }, { value: 'alias', label: 'Alias' }]} /></Form.Item></Col>
            <Col span={8}><Form.Item name="source_value" label="Source Value"><Input placeholder="192.168.1.0/24" /></Form.Item></Col>
            <Col span={8}><Form.Item name="source_port" label="Source Port"><Input placeholder="1024-65535" /></Form.Item></Col>
          </Row>

          <Row gutter={16}>
            <Col span={8}><Form.Item name="dest_type" label="Destination Type"><Select options={[{ value: 'any', label: 'Any' }, { value: 'single', label: 'IP' }, { value: 'network', label: 'Network' }, { value: 'alias', label: 'Alias' }]} /></Form.Item></Col>
            <Col span={8}><Form.Item name="dest_value" label="Destination Value"><Input placeholder="203.0.113.10" /></Form.Item></Col>
            <Col span={8}><Form.Item name="dest_port" label="Destination Port"><Input placeholder="443" /></Form.Item></Col>
          </Row>

          <Row gutter={16}>
            <Col span={8}>
              <Form.Item name="redirect_target" label="Redirect Target" rules={[{ required: true }]}>
                <Input placeholder="192.168.1.10 or fd00::10" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="redirect_port" label="Redirect Port">
                <Input placeholder="8443" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="description" label="Description">
                <Input placeholder="HTTPS reverse proxy" />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={6}><Form.Item name="nat_reflection" label="NAT Reflection" valuePropName="checked"><Switch /></Form.Item></Col>
            <Col span={6}><Form.Item name="enabled" label="Enabled" valuePropName="checked"><Switch /></Form.Item></Col>
          </Row>
        </Form>
      </Modal>
    </div>
  );
}
