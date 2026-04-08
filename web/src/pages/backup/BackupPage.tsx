import { Typography, Card } from 'antd';

const { Title, Text } = Typography;

export default function BackupPage() {
  return (
    <div>
      <Title level={3}>Backup & Restore</Title>
      <Card>
        <Text>Configuration backup, restore, and cloud sync</Text>
        <br /><br />
        <Text type="secondary">This module will be implemented in the corresponding development phase.</Text>
      </Card>
    </div>
  );
}
