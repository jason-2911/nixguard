import { Typography, Card } from 'antd';

const { Title, Text } = Typography;

export default function LoginPage() {
  return (
    <div>
      <Title level={3}>Login</Title>
      <Card>
        <Text>NixGuard authentication</Text>
        <br /><br />
        <Text type="secondary">This module will be implemented in the corresponding development phase.</Text>
      </Card>
    </div>
  );
}
