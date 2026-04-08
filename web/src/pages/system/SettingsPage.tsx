import { Typography, Card } from 'antd';

const { Title, Text } = Typography;

export default function SettingsPage() {
  return (
    <div>
      <Title level={3}>System Settings</Title>
      <Card>
        <Text>Hostname, DNS, NTP, timezone, and tunables</Text>
        <br /><br />
        <Text type="secondary">This module will be implemented in the corresponding development phase.</Text>
      </Card>
    </div>
  );
}
