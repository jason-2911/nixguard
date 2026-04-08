import { Typography, Card } from 'antd';

const { Title, Text } = Typography;

export default function CertificatesPage() {
  return (
    <div>
      <Title level={3}>Certificates</Title>
      <Card>
        <Text>Internal CA, server/client certs, and ACME/Let's Encrypt</Text>
        <br /><br />
        <Text type="secondary">This module will be implemented in the corresponding development phase.</Text>
      </Card>
    </div>
  );
}
