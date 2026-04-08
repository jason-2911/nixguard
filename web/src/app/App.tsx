import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { Provider } from 'react-redux';
import { ConfigProvider, theme } from 'antd';
import { store } from '@store/store';
import { useAppSelector } from '@hooks/useStore';
import MainLayout from '@components/layout/MainLayout';

// ── Page imports (lazy loaded) ────────────────────────────────
import { lazy, Suspense } from 'react';
import LoadingPage from '@components/common/LoadingPage';

const Dashboard       = lazy(() => import('@pages/dashboard/DashboardPage'));
const FirewallRules   = lazy(() => import('@pages/firewall/RulesPage'));
const FirewallNAT     = lazy(() => import('@pages/firewall/NATPage'));
const FirewallAliases = lazy(() => import('@pages/firewall/AliasesPage'));
const FirewallTraffic = lazy(() => import('@pages/firewall/TrafficPage'));
const Interfaces      = lazy(() => import('@pages/network/InterfacesPage'));
const Routing         = lazy(() => import('@pages/network/RoutingPage'));
const Gateways        = lazy(() => import('@pages/network/GatewaysPage'));
const IPsec           = lazy(() => import('@pages/vpn/IPsecPage'));
const OpenVPN         = lazy(() => import('@pages/vpn/OpenVPNPage'));
const WireGuard       = lazy(() => import('@pages/vpn/WireGuardPage'));
const DNSResolver     = lazy(() => import('@pages/dns/ResolverPage'));
const DNSFiltering    = lazy(() => import('@pages/dns/FilteringPage'));
const DHCPServer      = lazy(() => import('@pages/dhcp/ServerPage'));
const DHCPLeases      = lazy(() => import('@pages/dhcp/LeasesPage'));
const IDSAlerts       = lazy(() => import('@pages/ids/AlertsPage'));
const IDSRules        = lazy(() => import('@pages/ids/RulesPage'));
const Proxy           = lazy(() => import('@pages/proxy/ProxyPage'));
const LoadBalancer    = lazy(() => import('@pages/loadbalancer/LoadBalancerPage'));
const HA              = lazy(() => import('@pages/ha/HAPage'));
const TrafficShaper   = lazy(() => import('@pages/traffic_shaper/TrafficShaperPage'));
const CaptivePortal   = lazy(() => import('@pages/captiveportal/CaptivePortalPage'));
const Monitoring      = lazy(() => import('@pages/monitor/MonitoringPage'));
const Logs            = lazy(() => import('@pages/monitor/LogsPage'));
const Users           = lazy(() => import('@pages/auth/UsersPage'));
const Certificates    = lazy(() => import('@pages/auth/CertificatesPage'));
const SystemSettings  = lazy(() => import('@pages/system/SettingsPage'));
const Backup          = lazy(() => import('@pages/backup/BackupPage'));
const Diagnostics     = lazy(() => import('@pages/diagnostics/DiagnosticsPage'));
const Login           = lazy(() => import('@pages/auth/LoginPage'));

function AppRoutes() {
  const isDarkMode = useAppSelector((state) => state.ui.darkMode);

  return (
    <ConfigProvider
      theme={{
        algorithm: isDarkMode ? theme.darkAlgorithm : theme.defaultAlgorithm,
        token: {
          colorPrimary: '#1677ff',
          borderRadius: 6,
        },
      }}
    >
      <BrowserRouter>
        <Suspense fallback={<LoadingPage />}>
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route element={<MainLayout />}>
              {/* Dashboard */}
              <Route path="/" element={<Navigate to="/dashboard" replace />} />
              <Route path="/dashboard" element={<Dashboard />} />

              {/* Firewall */}
              <Route path="/firewall/rules" element={<FirewallRules />} />
              <Route path="/firewall/nat" element={<FirewallNAT />} />
              <Route path="/firewall/aliases" element={<FirewallAliases />} />
              <Route path="/firewall/traffic" element={<FirewallTraffic />} />

              {/* Network */}
              <Route path="/network/interfaces" element={<Interfaces />} />
              <Route path="/network/routing" element={<Routing />} />
              <Route path="/network/gateways" element={<Gateways />} />

              {/* VPN */}
              <Route path="/vpn/ipsec" element={<IPsec />} />
              <Route path="/vpn/openvpn" element={<OpenVPN />} />
              <Route path="/vpn/wireguard" element={<WireGuard />} />

              {/* Services */}
              <Route path="/dns/resolver" element={<DNSResolver />} />
              <Route path="/dns/filtering" element={<DNSFiltering />} />
              <Route path="/dhcp/server" element={<DHCPServer />} />
              <Route path="/dhcp/leases" element={<DHCPLeases />} />
              <Route path="/ids/alerts" element={<IDSAlerts />} />
              <Route path="/ids/rules" element={<IDSRules />} />
              <Route path="/proxy" element={<Proxy />} />
              <Route path="/loadbalancer" element={<LoadBalancer />} />
              <Route path="/ha" element={<HA />} />
              <Route path="/traffic-shaper" element={<TrafficShaper />} />
              <Route path="/captive-portal" element={<CaptivePortal />} />

              {/* Monitoring */}
              <Route path="/monitor" element={<Monitoring />} />
              <Route path="/monitor/logs" element={<Logs />} />

              {/* System */}
              <Route path="/system/users" element={<Users />} />
              <Route path="/system/certificates" element={<Certificates />} />
              <Route path="/system/settings" element={<SystemSettings />} />
              <Route path="/system/backup" element={<Backup />} />
              <Route path="/system/diagnostics" element={<Diagnostics />} />
            </Route>
          </Routes>
        </Suspense>
      </BrowserRouter>
    </ConfigProvider>
  );
}

export default function App() {
  return (
    <Provider store={store}>
      <AppRoutes />
    </Provider>
  );
}
