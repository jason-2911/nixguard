import { Typography, Card } from 'antd';

const { Title, Text } = Typography;

export default function DiagnosticsPage() {
  return (
    <div>
      <Title level={3}>Diagnostics</Title>
      <Card>
        <Text>Ping, traceroute, DNS lookup, packet capture</Text>
        <br /><br />
        <Text type="secondary">This module will be implemented in the corresponding development phase.</Text>
      </Card>
    </div>
  );
}
