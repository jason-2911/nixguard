import { Typography, Card } from 'antd';

const { Title, Text } = Typography;

export default function ResolverPage() {
  return (
    <div>
      <Title level={3}>DNS Resolver</Title>
      <Card>
        <Text>Unbound DNS configuration, overrides, and forwarding</Text>
        <br /><br />
        <Text type="secondary">This module will be implemented in the corresponding development phase.</Text>
      </Card>
    </div>
  );
}
