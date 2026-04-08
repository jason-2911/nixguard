import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { Layout, Menu, Typography, Avatar, Dropdown, Badge, Switch } from 'antd';
import {
  DashboardOutlined,
  SafetyOutlined,
  GlobalOutlined,
  LockOutlined,
  CloudServerOutlined,
  DesktopOutlined,
  AlertOutlined,
  FilterOutlined,
  ApartmentOutlined,
  ClusterOutlined,
  FundOutlined,
  WifiOutlined,
  MonitorOutlined,
  SettingOutlined,
  UserOutlined,
  SafetyCertificateOutlined,
  SaveOutlined,
  ToolOutlined,
  BellOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  BulbOutlined,
} from '@ant-design/icons';
import { useAppDispatch, useAppSelector } from '@hooks/useStore';
import { toggleDarkMode, toggleSidebar } from '@store/slices/uiSlice';
import type { MenuProps } from 'antd';

const { Header, Sider, Content } = Layout;

type MenuItem = Required<MenuProps>['items'][number];

const menuItems: MenuItem[] = [
  {
    key: '/dashboard',
    icon: <DashboardOutlined />,
    label: 'Dashboard',
  },
  {
    key: 'firewall',
    icon: <SafetyOutlined />,
    label: 'Firewall',
    children: [
      { key: '/firewall/rules', label: 'Rules' },
      { key: '/firewall/nat', label: 'NAT' },
      { key: '/firewall/aliases', label: 'Aliases' },
      { key: '/firewall/traffic', label: 'Live Traffic' },
    ],
  },
  {
    key: 'network',
    icon: <GlobalOutlined />,
    label: 'Network',
    children: [
      { key: '/network/interfaces', label: 'Interfaces' },
      { key: '/network/routing', label: 'Routing' },
      { key: '/network/gateways', label: 'Gateways' },
    ],
  },
  {
    key: 'vpn',
    icon: <LockOutlined />,
    label: 'VPN',
    children: [
      { key: '/vpn/ipsec', label: 'IPsec' },
      { key: '/vpn/openvpn', label: 'OpenVPN' },
      { key: '/vpn/wireguard', label: 'WireGuard' },
    ],
  },
  {
    key: 'services',
    icon: <CloudServerOutlined />,
    label: 'Services',
    children: [
      { key: '/dns/resolver', icon: <DesktopOutlined />, label: 'DNS Resolver' },
      { key: '/dns/filtering', icon: <FilterOutlined />, label: 'DNS Filtering' },
      { key: '/dhcp/server', icon: <ApartmentOutlined />, label: 'DHCP Server' },
      { key: '/dhcp/leases', label: 'DHCP Leases' },
      { key: '/ids/alerts', icon: <AlertOutlined />, label: 'IDS Alerts' },
      { key: '/ids/rules', label: 'IDS Rules' },
      { key: '/proxy', label: 'Web Proxy' },
      { key: '/loadbalancer', icon: <ClusterOutlined />, label: 'Load Balancer' },
      { key: '/ha', label: 'High Availability' },
      { key: '/traffic-shaper', icon: <FundOutlined />, label: 'Traffic Shaper' },
      { key: '/captive-portal', icon: <WifiOutlined />, label: 'Captive Portal' },
    ],
  },
  {
    key: 'monitoring',
    icon: <MonitorOutlined />,
    label: 'Monitoring',
    children: [
      { key: '/monitor', label: 'Dashboard' },
      { key: '/monitor/logs', label: 'Logs' },
    ],
  },
  {
    key: 'system',
    icon: <SettingOutlined />,
    label: 'System',
    children: [
      { key: '/system/users', icon: <UserOutlined />, label: 'Users & Groups' },
      { key: '/system/certificates', icon: <SafetyCertificateOutlined />, label: 'Certificates' },
      { key: '/system/settings', label: 'Settings' },
      { key: '/system/backup', icon: <SaveOutlined />, label: 'Backup' },
      { key: '/system/diagnostics', icon: <ToolOutlined />, label: 'Diagnostics' },
    ],
  },
];

export default function MainLayout() {
  const navigate = useNavigate();
  const location = useLocation();
  const dispatch = useAppDispatch();
  const { sidebarCollapsed, darkMode, notifications } = useAppSelector((state) => state.ui);
  const { user } = useAppSelector((state) => state.auth);
  const unreadCount = notifications.filter((n) => !n.read).length;

  const userMenuItems: MenuProps['items'] = [
    { key: 'profile', label: 'Profile' },
    { key: 'settings', label: 'Settings' },
    { type: 'divider' },
    { key: 'logout', label: 'Logout', danger: true },
  ];

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        collapsible
        collapsed={sidebarCollapsed}
        onCollapse={() => dispatch(toggleSidebar())}
        theme={darkMode ? 'dark' : 'light'}
        width={256}
        trigger={null}
      >
        <div style={{ padding: '16px', textAlign: 'center' }}>
          <Typography.Title level={4} style={{ margin: 0, color: '#1677ff' }}>
            {sidebarCollapsed ? 'NG' : 'NixGuard'}
          </Typography.Title>
        </div>
        <Menu
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
          theme={darkMode ? 'dark' : 'light'}
        />
      </Sider>

      <Layout>
        <Header
          style={{
            padding: '0 24px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            background: darkMode ? '#141414' : '#fff',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
            {sidebarCollapsed ? (
              <MenuUnfoldOutlined onClick={() => dispatch(toggleSidebar())} />
            ) : (
              <MenuFoldOutlined onClick={() => dispatch(toggleSidebar())} />
            )}
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
            <Switch
              checkedChildren={<BulbOutlined />}
              unCheckedChildren={<BulbOutlined />}
              checked={darkMode}
              onChange={() => dispatch(toggleDarkMode())}
            />
            <Badge count={unreadCount} size="small">
              <BellOutlined style={{ fontSize: 18, cursor: 'pointer' }} />
            </Badge>
            <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
              <div style={{ cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 8 }}>
                <Avatar icon={<UserOutlined />} size="small" />
                <span>{user?.username || 'admin'}</span>
              </div>
            </Dropdown>
          </div>
        </Header>

        <Content style={{ margin: 24 }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}
