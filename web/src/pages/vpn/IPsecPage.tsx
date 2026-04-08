import { Typography, Card } from 'antd';

const { Title, Text } = Typography;

export default function IPsecPage() {
  return (
    <div>
      <Title level={3}>IPsec VPN</Title>
      <Card>
        <Text>Site-to-site and road warrior IPsec tunnels (StrongSwan)</Text>
        <br /><br />
        <Text type="secondary">This module will be implemented in the corresponding development phase.</Text>
      </Card>
    </div>
  );
}
