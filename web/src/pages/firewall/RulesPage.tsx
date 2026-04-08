import { useCallback, useEffect, useState } from 'react';
import {
  Typography, Table, Button, Space, Tag, Switch, Modal, Form, Select, Input,
  Tooltip, Popconfirm, Card, Row, Col, message,
} from 'antd';
import {
  PlusOutlined, DeleteOutlined, EditOutlined, ReloadOutlined,
  ArrowUpOutlined, ArrowDownOutlined, SearchOutlined,
} from '@ant-design/icons';
import { useAppDispatch, useAppSelector } from '@hooks/useStore';
import { fetchRules, createRule, updateRule, deleteRule } from '@store/slices/firewallSlice';
import { useAutoRefresh } from '@hooks/useAutoRefresh';
import { apiClient } from '@api/client';
import type { FirewallRule, Address } from '@typedefs/firewall';

const { Title } = Typography;
const { Option } = Select;

const actionColors: Record<string, string> = { pass: 'green', block: 'red', reject: 'orange' };
const protocolColors: Record<string, string> = { tcp: 'blue', udp: 'purple', icmp: 'cyan', any: 'default' };

export default function RulesPage() {
  const dispatch = useAppDispatch();
  const { rules, loading } = useAppSelector((state) => state.firewall);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingRule, setEditingRule] = useState<FirewallRule | null>(null);
  const [form] = Form.useForm();
  const [searchText, setSearchText] = useState('');

  const refresh = useCallback(() => { dispatch(fetchRules()); }, [dispatch]);
  useEffect(() => { refresh(); }, [refresh]);
  useAutoRefresh(refresh, 5000);

  const filteredRules = rules.filter((r) =>
    r.description.toLowerCase().includes(searchText.toLowerCase()) ||
    r.interface.toLowerCase().includes(searchText.toLowerCase()),
  );

  const handleCreate = () => {
    setEditingRule(null);
    form.resetFields();
    form.setFieldsValue({
      direction: 'in',
      action: 'pass',
      protocol: 'any',
      source_type: 'any',
      dest_type: 'any',
      log: false,
      is_floating: false,
      state_type: 'keep',
    });
    setModalOpen(true);
  };

  const handleEdit = (rule: FirewallRule) => {
    setEditingRule(rule);
    form.setFieldsValue({
      interface: rule.interface, direction: rule.direction, action: rule.action, protocol: rule.protocol,
      source_type: rule.source.type, source_value: rule.source.value, source_port: rule.source.port,
      dest_type: rule.destination.type, dest_value: rule.destination.value, dest_port: rule.destination.port,
      log: rule.log, description: rule.description,
      category: rule.category,
      is_floating: rule.is_floating,
      floating_interfaces: (rule.interfaces || []).join(', '),
      gateway: rule.gateway,
      state_type: rule.state_type,
      max_states: rule.max_states,
      tag: rule.tag,
      tagged: rule.tagged,
      schedule_name: rule.schedule?.name,
      schedule_start: rule.schedule?.start_time,
      schedule_end: rule.schedule?.end_time,
      schedule_days: rule.schedule?.weekdays,
    });
    setModalOpen(true);
  };

  const handleSubmit = async () => {
    const values = await form.validateFields();
    const floatingInterfaces = String(values.floating_interfaces || '')
      .split(',')
      .map((value: string) => value.trim())
      .filter(Boolean);

    const ruleData = {
      interface: values.interface || '', direction: values.direction, action: values.action, protocol: values.protocol,
      source: { type: values.source_type, value: values.source_value || '', port: values.source_port || '', not: false } as Address,
      destination: { type: values.dest_type, value: values.dest_value || '', port: values.dest_port || '', not: false } as Address,
      log: values.log || false,
      description: values.description || '',
      category: values.category || '',
      is_floating: values.is_floating || false,
      interfaces: floatingInterfaces,
      gateway: values.gateway || '',
      state_type: values.state_type || 'keep',
      max_states: Number(values.max_states || 0),
      tag: values.tag || '',
      tagged: values.tagged || '',
      schedule: values.schedule_name || values.schedule_start || values.schedule_end || (values.schedule_days || []).length
        ? {
            name: values.schedule_name || '',
            start_time: values.schedule_start || '',
            end_time: values.schedule_end || '',
            weekdays: values.schedule_days || [],
          }
        : undefined,
    };
    if (editingRule) {
      await dispatch(updateRule({ id: editingRule.id, ...ruleData }));
      message.success('Rule updated');
    } else {
      await dispatch(createRule(ruleData));
      message.success('Rule created');
    }
    setModalOpen(false);
    dispatch(fetchRules());
  };

  const handleDelete = async (id: string) => { await dispatch(deleteRule(id)); message.success('Rule deleted'); };
  const handleToggle = async (rule: FirewallRule) => {
    await dispatch(updateRule({ id: rule.id, enabled: !rule.enabled } as any));
    dispatch(fetchRules());
  };
  const handleApply = async () => {
    await apiClient.post('/firewall/apply');
    message.success('Rules applied');
    dispatch(fetchRules());
  };
  const handleMove = async (rule: FirewallRule, direction: -1 | 1) => {
    const ordered = [...rules];
    const index = ordered.findIndex((item) => item.id === rule.id);
    const nextIndex = index + direction;
    if (index < 0 || nextIndex < 0 || nextIndex >= ordered.length) {
      return;
    }
    [ordered[index], ordered[nextIndex]] = [ordered[nextIndex], ordered[index]];
    await apiClient.post('/firewall/rules/reorder', { rule_ids: ordered.map((item) => item.id) });
    dispatch(fetchRules());
  };

  const formatAddr = (addr: Address) => {
    if (addr.type === 'any') return '*';
    let s = addr.value || addr.type;
    if (addr.port) s += ':' + addr.port;
    return s;
  };

  const columns = [
    { title: '#', key: 'order', width: 50, render: (_: any, __: any, i: number) => i + 1 },
    { title: '', key: 'enabled', width: 50, render: (_: any, r: FirewallRule) => <Switch size="small" checked={r.enabled} onChange={() => handleToggle(r)} /> },
    { title: 'Action', dataIndex: 'action', width: 80, render: (a: string) => <Tag color={actionColors[a]}>{a.toUpperCase()}</Tag> },
    { title: 'Iface', dataIndex: 'interface', width: 90, render: (v: string) => v || 'any' },
    { title: 'Dir', dataIndex: 'direction', width: 50, render: (d: string) => d === 'in' ? <ArrowDownOutlined style={{ color: '#1677ff' }} /> : <ArrowUpOutlined style={{ color: '#52c41a' }} /> },
    { title: 'Proto', dataIndex: 'protocol', width: 70, render: (p: string) => <Tag color={protocolColors[p] || 'default'}>{p.toUpperCase()}</Tag> },
    { title: 'Source', key: 'source', width: 210, ellipsis: true, render: (_: any, r: FirewallRule) => formatAddr(r.source) },
    { title: 'Destination', key: 'dest', width: 210, ellipsis: true, render: (_: any, r: FirewallRule) => formatAddr(r.destination) },
    { title: 'Scope', key: 'scope', width: 100, render: (_: any, r: FirewallRule) => r.is_floating ? <Tag color="cyan">FLOATING</Tag> : <Tag>{r.interface || 'any'}</Tag> },
    { title: 'Log', dataIndex: 'log', width: 50, render: (v: boolean) => v ? <Tag color="blue">LOG</Tag> : null },
    { title: 'Description', dataIndex: 'description', width: 240, ellipsis: true },
    { title: 'Packets', key: 'stats', width: 90, render: (_: any, r: FirewallRule) => <Tooltip title={`${r.stats.bytes} bytes`}>{r.stats.packets.toLocaleString()}</Tooltip> },
    { title: '', key: 'actions', width: 144, fixed: 'right' as const, render: (_: any, r: FirewallRule) => (
      <Space size={4} wrap={false}>
        <Button type="link" size="small" icon={<ArrowUpOutlined />} onClick={() => void handleMove(r, -1)} />
        <Button type="link" size="small" icon={<ArrowDownOutlined />} onClick={() => void handleMove(r, 1)} />
        <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEdit(r)} />
        <Popconfirm title="Delete?" onConfirm={() => handleDelete(r.id)}><Button type="link" size="small" danger icon={<DeleteOutlined />} /></Popconfirm>
      </Space>
    )},
  ];

  return (
    <div>
      <Row justify="space-between" align="middle" style={{ marginBottom: 16 }}>
        <Col><Title level={3} style={{ margin: 0 }}>Firewall Rules</Title></Col>
        <Col>
          <Space>
            <Input placeholder="Search..." prefix={<SearchOutlined />} value={searchText} onChange={(e) => setSearchText(e.target.value)} style={{ width: 200 }} />
            <Button icon={<ReloadOutlined />} onClick={() => dispatch(fetchRules())}>Refresh</Button>
            <Button onClick={() => void handleApply()}>Apply</Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>Add Rule</Button>
          </Space>
        </Col>
      </Row>

      <Card bodyStyle={{ padding: 0 }}>
        <Table
          dataSource={filteredRules}
          columns={columns}
          rowKey="id"
          loading={loading}
          size="small"
          scroll={{ x: 1320 }}
          pagination={{ pageSize: 50, showSizeChanger: true }}
        />
      </Card>

      <Modal title={editingRule ? 'Edit Rule' : 'Add Firewall Rule'} open={modalOpen} onOk={handleSubmit} onCancel={() => setModalOpen(false)} width={700} okText={editingRule ? 'Update' : 'Create'}>
        <Form form={form} layout="vertical">
          <Row gutter={16}>
            <Col span={8}><Form.Item name="action" label="Action" rules={[{ required: true }]}><Select><Option value="pass">Pass</Option><Option value="block">Block</Option><Option value="reject">Reject</Option></Select></Form.Item></Col>
            <Col span={8}><Form.Item name="direction" label="Direction" rules={[{ required: true }]}><Select><Option value="in">In</Option><Option value="out">Out</Option></Select></Form.Item></Col>
            <Col span={8}><Form.Item name="interface" label="Interface"><Input placeholder="e.g. eth0" /></Form.Item></Col>
          </Row>
          <Row gutter={16}>
            <Col span={8}><Form.Item name="protocol" label="Protocol" rules={[{ required: true }]}><Select><Option value="any">Any</Option><Option value="tcp">TCP</Option><Option value="udp">UDP</Option><Option value="icmp">ICMP</Option><Option value="esp">ESP</Option><Option value="gre">GRE</Option></Select></Form.Item></Col>
            <Col span={8}><Form.Item name="category" label="Category"><Input placeholder="wan-inbound" /></Form.Item></Col>
            <Col span={8}><Form.Item name="gateway" label="Policy Gateway"><Input placeholder="wan-primary" /></Form.Item></Col>
          </Row>
          <Title level={5}>Source</Title>
          <Row gutter={16}>
            <Col span={8}><Form.Item name="source_type" label="Type"><Select><Option value="any">Any</Option><Option value="single">IP</Option><Option value="network">Network</Option><Option value="alias">Alias</Option></Select></Form.Item></Col>
            <Col span={8}><Form.Item name="source_value" label="Address"><Input placeholder="192.168.1.0/24" /></Form.Item></Col>
            <Col span={8}><Form.Item name="source_port" label="Port"><Input placeholder="80, 80-443" /></Form.Item></Col>
          </Row>
          <Title level={5}>Destination</Title>
          <Row gutter={16}>
            <Col span={8}><Form.Item name="dest_type" label="Type"><Select><Option value="any">Any</Option><Option value="single">IP</Option><Option value="network">Network</Option><Option value="alias">Alias</Option></Select></Form.Item></Col>
            <Col span={8}><Form.Item name="dest_value" label="Address"><Input placeholder="10.0.0.1" /></Form.Item></Col>
            <Col span={8}><Form.Item name="dest_port" label="Port"><Input placeholder="443" /></Form.Item></Col>
          </Row>
          <Row gutter={16}>
            <Col span={8}><Form.Item name="log" label="Log" valuePropName="checked"><Switch /></Form.Item></Col>
            <Col span={16}><Form.Item name="description" label="Description"><Input placeholder="Rule description" /></Form.Item></Col>
          </Row>
          <Row gutter={16}>
            <Col span={8}><Form.Item name="is_floating" label="Floating Rule" valuePropName="checked"><Switch /></Form.Item></Col>
            <Col span={16}><Form.Item name="floating_interfaces" label="Floating Interfaces"><Input placeholder="wan0, lan0" /></Form.Item></Col>
          </Row>
          <Row gutter={16}>
            <Col span={8}><Form.Item name="state_type" label="State Type"><Select><Option value="keep">Keep</Option><Option value="sloppy">Sloppy</Option><Option value="synproxy">Synproxy</Option></Select></Form.Item></Col>
            <Col span={8}><Form.Item name="max_states" label="Max States"><Input placeholder="0" /></Form.Item></Col>
            <Col span={8}><Form.Item name="tag" label="Tag"><Input placeholder="web" /></Form.Item></Col>
          </Row>
          <Row gutter={16}>
            <Col span={8}><Form.Item name="tagged" label="Tagged"><Input placeholder="web" /></Form.Item></Col>
            <Col span={8}><Form.Item name="schedule_name" label="Schedule Name"><Input placeholder="work-hours" /></Form.Item></Col>
            <Col span={8}><Form.Item name="schedule_days" label="Weekdays"><Select mode="multiple" options={[{ value: 1, label: 'Mon' }, { value: 2, label: 'Tue' }, { value: 3, label: 'Wed' }, { value: 4, label: 'Thu' }, { value: 5, label: 'Fri' }, { value: 6, label: 'Sat' }, { value: 0, label: 'Sun' }]} /></Form.Item></Col>
          </Row>
          <Row gutter={16}>
            <Col span={12}><Form.Item name="schedule_start" label="Schedule Start"><Input placeholder="09:00" /></Form.Item></Col>
            <Col span={12}><Form.Item name="schedule_end" label="Schedule End"><Input placeholder="18:00" /></Form.Item></Col>
          </Row>
        </Form>
      </Modal>
    </div>
  );
}
