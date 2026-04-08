import { Typography, Card } from 'antd';

const { Title, Text } = Typography;

export default function UsersPage() {
  return (
    <div>
      <Title level={3}>Users & Groups</Title>
      <Card>
        <Text>User management, RBAC, LDAP, RADIUS, and MFA</Text>
        <br /><br />
        <Text type="secondary">This module will be implemented in the corresponding development phase.</Text>
      </Card>
    </div>
  );
}
