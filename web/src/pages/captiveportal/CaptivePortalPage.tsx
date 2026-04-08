import { Typography, Card } from 'antd';

const { Title, Text } = Typography;

export default function CaptivePortalPage() {
  return (
    <div>
      <Title level={3}>Captive Portal</Title>
      <Card>
        <Text>Guest network portal, vouchers, and sessions</Text>
        <br /><br />
        <Text type="secondary">This module will be implemented in the corresponding development phase.</Text>
      </Card>
    </div>
  );
}
