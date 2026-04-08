import { Typography, Card } from 'antd';

const { Title, Text } = Typography;

export default function OpenVPNPage() {
  return (
    <div>
      <Title level={3}>OpenVPN</Title>
      <Card>
        <Text>OpenVPN server/client management and client export</Text>
        <br /><br />
        <Text type="secondary">This module will be implemented in the corresponding development phase.</Text>
      </Card>
    </div>
  );
}
