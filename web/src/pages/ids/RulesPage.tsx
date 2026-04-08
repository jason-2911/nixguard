import { Typography, Card } from 'antd';

const { Title, Text } = Typography;

export default function RulesPage() {
  return (
    <div>
      <Title level={3}>IDS Rules</Title>
      <Card>
        <Text>Suricata ruleset management and overrides</Text>
        <br /><br />
        <Text type="secondary">This module will be implemented in the corresponding development phase.</Text>
      </Card>
    </div>
  );
}
