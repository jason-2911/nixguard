import { Typography, Card } from 'antd';

const { Title, Text } = Typography;

export default function FilteringPage() {
  return (
    <div>
      <Title level={3}>DNS Filtering</Title>
      <Card>
        <Text>DNS-based ad blocking and domain filtering</Text>
        <br /><br />
        <Text type="secondary">This module will be implemented in the corresponding development phase.</Text>
      </Card>
    </div>
  );
}
