import { Typography, Card } from 'antd';

const { Title, Text } = Typography;

export default function ProxyPage() {
  return (
    <div>
      <Title level={3}>Web Proxy</Title>
      <Card>
        <Text>Squid proxy, SSL bump, caching, and URL filtering</Text>
        <br /><br />
        <Text type="secondary">This module will be implemented in the corresponding development phase.</Text>
      </Card>
    </div>
  );
}
