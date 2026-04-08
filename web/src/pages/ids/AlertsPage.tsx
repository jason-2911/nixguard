import { Typography, Card } from 'antd';

const { Title, Text } = Typography;

export default function AlertsPage() {
  return (
    <div>
      <Title level={3}>IDS Alerts</Title>
      <Card>
        <Text>Suricata intrusion detection alerts and events</Text>
        <br /><br />
        <Text type="secondary">This module will be implemented in the corresponding development phase.</Text>
      </Card>
    </div>
  );
}
