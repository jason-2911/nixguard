import { Typography, Card } from 'antd';

const { Title, Text } = Typography;

export default function ServerPage() {
  return (
    <div>
      <Title level={3}>DHCP Server</Title>
      <Card>
        <Text>DHCPv4/v6 server configuration and static mappings</Text>
        <br /><br />
        <Text type="secondary">This module will be implemented in the corresponding development phase.</Text>
      </Card>
    </div>
  );
}
