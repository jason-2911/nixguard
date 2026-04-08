import { Typography, Card } from 'antd';

const { Title, Text } = Typography;

export default function LeasesPage() {
  return (
    <div>
      <Title level={3}>DHCP Leases</Title>
      <Card>
        <Text>View and manage active DHCP leases</Text>
        <br /><br />
        <Text type="secondary">This module will be implemented in the corresponding development phase.</Text>
      </Card>
    </div>
  );
}
