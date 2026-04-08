import { Typography, Card } from 'antd';

const { Title, Text } = Typography;

export default function HAPage() {
  return (
    <div>
      <Title level={3}>High Availability</Title>
      <Card>
        <Text>CARP VIPs, config sync, state sync, and failover</Text>
        <br /><br />
        <Text type="secondary">This module will be implemented in the corresponding development phase.</Text>
      </Card>
    </div>
  );
}
