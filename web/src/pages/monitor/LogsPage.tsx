import { Typography, Card } from 'antd';

const { Title, Text } = Typography;

export default function LogsPage() {
  return (
    <div>
      <Title level={3}>System Logs</Title>
      <Card>
        <Text>Firewall, system, service, and audit logs</Text>
        <br /><br />
        <Text type="secondary">This module will be implemented in the corresponding development phase.</Text>
      </Card>
    </div>
  );
}
