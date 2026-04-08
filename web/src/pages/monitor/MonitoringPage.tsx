import { Typography, Card } from 'antd';

const { Title, Text } = Typography;

export default function MonitoringPage() {
  return (
    <div>
      <Title level={3}>Monitoring</Title>
      <Card>
        <Text>System metrics, traffic graphs, and top talkers</Text>
        <br /><br />
        <Text type="secondary">This module will be implemented in the corresponding development phase.</Text>
      </Card>
    </div>
  );
}
