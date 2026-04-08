import { Typography, Card } from 'antd';

const { Title, Text } = Typography;

export default function TrafficShaperPage() {
  return (
    <div>
      <Title level={3}>Traffic Shaper</Title>
      <Card>
        <Text>QoS pipes, queues, and traffic classification</Text>
        <br /><br />
        <Text type="secondary">This module will be implemented in the corresponding development phase.</Text>
      </Card>
    </div>
  );
}
