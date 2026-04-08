import { Typography, Card } from 'antd';

const { Title, Text } = Typography;

export default function WireGuardPage() {
  return (
    <div>
      <Title level={3}>WireGuard VPN</Title>
      <Card>
        <Text>WireGuard tunnel and peer management</Text>
        <br /><br />
        <Text type="secondary">This module will be implemented in the corresponding development phase.</Text>
      </Card>
    </div>
  );
}
