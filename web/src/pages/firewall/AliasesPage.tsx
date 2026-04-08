import { useCallback, useEffect, useState } from 'react';
import {
  Button,
  Card,
  Form,
  Input,
  Modal,
  Popconfirm,
  Row,
  Col,
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
import type { FirewallAlias } from '@typedefs/firewall';

const { Title } = Typography;
const { TextArea } = Input;

const aliasTypes: FirewallAlias['type'][] = ['host', 'network', 'port', 'url', 'url_table', 'geoip', 'nested'];

function parseEntries(raw: string): string[] {
  return raw
    .split(/\r?\n|,/)
    .map((entry) => entry.trim())
    .filter(Boolean);
}

export default function AliasesPage() {
  const [aliases, setAliases] = useState<FirewallAlias[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingAlias, setEditingAlias] = useState<FirewallAlias | null>(null);
  const [form] = Form.useForm();

  const loadAliases = async () => {
    setLoading(true);
    try {
      const response = await apiClient.get('/firewall/aliases');
      setAliases(response.data ?? []);
    } catch {
      message.error('Failed to load aliases');
    } finally {
      setLoading(false);
    }
  };

  const refresh = useCallback(() => { void loadAliases(); }, []);
  useEffect(() => { refresh(); }, [refresh]);
  useAutoRefresh(refresh, 5000);

  const openCreate = () => {
    setEditingAlias(null);
    form.resetFields();
    form.setFieldsValue({ type: 'host', enabled: true });
    setModalOpen(true);
  };

  const openEdit = (alias: FirewallAlias) => {
    setEditingAlias(alias);
    form.setFieldsValue({
      name: alias.name,
      type: alias.type,
      description: alias.description,
      update_freq: alias.update_freq,
      enabled: alias.enabled,
      entries_text: alias.entries.join('\n'),
    });
    setModalOpen(true);
  };

  const handleSubmit = async () => {
    const values = await form.validateFields();
    const payload = {
      name: values.name,
      type: values.type,
      description: values.description ?? '',
      update_freq: values.update_freq ?? '',
      enabled: values.enabled ?? true,
      entries: parseEntries(values.entries_text ?? ''),
    };

    try {
      if (editingAlias) {
        await apiClient.put(`/firewall/aliases/${editingAlias.id}`, payload);
        message.success('Alias updated');
      } else {
        await apiClient.post('/firewall/aliases', payload);
        message.success('Alias created');
      }
      setModalOpen(false);
      void loadAliases();
    } catch {
      message.error('Failed to save alias');
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await apiClient.delete(`/firewall/aliases/${id}`);
      message.success('Alias deleted');
      void loadAliases();
    } catch {
      message.error('Failed to delete alias');
    }
  };

  const columns = [
    {
      title: 'Name',
      dataIndex: 'name',
      render: (name: string, alias: FirewallAlias) => (
        <Space>
          <strong>{name}</strong>
          {!alias.enabled && <Tag color="default">DISABLED</Tag>}
        </Space>
      ),
    },
    {
      title: 'Type',
      dataIndex: 'type',
      width: 120,
      render: (type: FirewallAlias['type']) => <Tag color="blue">{type.toUpperCase()}</Tag>,
    },
    {
      title: 'Entries',
      key: 'entries',
      render: (_: unknown, alias: FirewallAlias) => `${alias.entries.length} item(s)`,
    },
    {
      title: 'Update',
      dataIndex: 'update_freq',
      width: 120,
      render: (freq: string) => freq || '-',
    },
    {
      title: 'Description',
      dataIndex: 'description',
    },
    {
      title: '',
      key: 'actions',
      width: 96,
      render: (_: unknown, alias: FirewallAlias) => (
        <Space>
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => openEdit(alias)} />
          <Popconfirm title="Delete alias?" onConfirm={() => void handleDelete(alias.id)}>
            <Button type="link" size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Row justify="space-between" align="middle" style={{ marginBottom: 16 }}>
        <Col><Title level={3} style={{ margin: 0 }}>Firewall Aliases</Title></Col>
        <Col>
          <Space>
            <Button icon={<ReloadOutlined />} onClick={() => void loadAliases()}>Refresh</Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>Add Alias</Button>
          </Space>
        </Col>
      </Row>

      <Card bodyStyle={{ padding: 0 }}>
        <Table dataSource={aliases} columns={columns} rowKey="id" loading={loading} size="small" pagination={{ pageSize: 25 }} />
      </Card>

      <Modal
        title={editingAlias ? 'Edit Alias' : 'Create Alias'}
        open={modalOpen}
        onOk={() => void handleSubmit()}
        onCancel={() => setModalOpen(false)}
        width={760}
      >
        <Form form={form} layout="vertical">
          <Row gutter={16}>
            <Col span={10}>
              <Form.Item name="name" label="Name" rules={[{ required: true, message: 'Alias name is required' }]}>
                <Input placeholder="trusted_hosts" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="type" label="Type" rules={[{ required: true, message: 'Alias type is required' }]}>
                <Select options={aliasTypes.map((type) => ({ value: type, label: type }))} />
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item name="enabled" label="Enabled" valuePropName="checked">
                <Switch />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item name="description" label="Description">
            <Input placeholder="Trusted hosts, GeoIP country list, URL table..." />
          </Form.Item>
          <Form.Item name="update_freq" label="Update Frequency">
            <Input placeholder="24h" />
          </Form.Item>
          <Form.Item
            name="entries_text"
            label="Entries"
            rules={[{ required: true, message: 'Provide at least one entry' }]}
            extra="One entry per line. Supports IP/CIDR, ports, URLs, GeoIP country codes, or alias names for nested aliases."
          >
            <TextArea rows={10} placeholder={'192.168.1.10\n192.168.1.11\n2001:db8::10'} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
