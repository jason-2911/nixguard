import { Typography, Card } from 'antd';

const { Title, Text } = Typography;

export default function LoadBalancerPage() {
  return (
    <div>
      <Title level={3}>Load Balancer</Title>
      <Card>
        <Text>HAProxy frontends, backends, and health checks</Text>
        <br /><br />
        <Text type="secondary">This module will be implemented in the corresponding development phase.</Text>
      </Card>
    </div>
  );
}
