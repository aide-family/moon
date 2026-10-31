import { Route, Routes } from "react-router-dom"

import { AppShell } from "@/components/app-shell"
import { LandingRedirect } from "@/components/landing-redirect"
import { AgentPage } from "@/pages/agent"
import { AgentShellPage } from "@/pages/agent/shell"
import { AlertHistoryPage } from "@/pages/alert/history"
import { AlertRealtimePage } from "@/pages/alert"
import { AlertDictionaryPage } from "@/pages/alert-dictionary"
import { AlertPageManagePage } from "@/pages/alert-page"
import { AlertRulePage } from "@/pages/alert-rule"
import { ClusterPage } from "@/pages/cluster"
import { ClusterDetailPage } from "@/pages/cluster/detail"
import { DatasourcePage } from "@/pages/datasource"
import { ForbiddenPage } from "@/pages/forbidden"
import { InvitationPage } from "@/pages/invitation"
import { InvitePage } from "@/pages/invite"
import { LoginPage } from "@/pages/login"
import { MemberPage } from "@/pages/member"
import { MessagePage } from "@/pages/message"
import { MessageLogPage } from "@/pages/message-log"
import { NamespacePage } from "@/pages/namespace"
import { NotFoundPage } from "@/pages/not-found"
import { NotificationChannelPage } from "@/pages/notification-channel"
import { NotificationGroupPage } from "@/pages/notification-group"
import { NotificationOnCallPage } from "@/pages/notification-oncall"
import { NotificationSubscriptionPage } from "@/pages/notification-subscription"
import { NotificationTemplatePage } from "@/pages/notification-template"
import { NotificationPresetPage } from "@/pages/notification-template/presets"
import { OAuthCallbackPage } from "@/pages/oauth-callback"
import { OrderPage } from "@/pages/order"
import { OverviewPage } from "@/pages/overview"
import { PackagePage } from "@/pages/package"
import { ProfilePage } from "@/pages/profile"
import { ProbeTargetPage } from "@/pages/probe-target"
import { RegisterPage } from "@/pages/register"
import { RolePage } from "@/pages/role"
import { SSHCommandPage } from "@/pages/ssh-command"
import { SSHCommandResultsPage } from "@/pages/ssh-command/results"
import { TenantPage } from "@/pages/tenant"
import { TenantOrdersPage } from "@/pages/tenant/orders"
import { UserPage } from "@/pages/user"
import { WorkerStatusPage } from "@/pages/workers"
import { UpgradePage } from "@/pages/upgrade"

export function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route path="/auth/callback" element={<OAuthCallbackPage />} />
      <Route path="/invite/:token" element={<InvitePage />} />
      <Route element={<AppShell />}>
        {/* 登录后的着陆页：没有租户去租户页，否则回放上次访问的页面。 */}
        <Route path="/" element={<LandingRedirect />} />
        <Route path="/overview" element={<OverviewPage />} />
        <Route path="/profile" element={<ProfilePage />} />
        <Route path="/forbidden" element={<ForbiddenPage />} />
        <Route path="/namespaces" element={<NamespacePage />} />
        <Route path="/orders" element={<OrderPage />} />
        <Route path="/tenants" element={<TenantPage />} />
        <Route path="/tenants/:uid/orders" element={<TenantOrdersPage />} />
        <Route path="/packages" element={<PackagePage />} />
        <Route path="/users" element={<UserPage />} />
        <Route path="/workers" element={<WorkerStatusPage />} />
        <Route path="/system/upgrade" element={<UpgradePage />} />
        <Route path="/invitations" element={<InvitationPage />} />
        <Route path="/messages" element={<MessagePage />} />
        <Route path="/members" element={<MemberPage />} />
        <Route path="/roles" element={<RolePage />} />
        <Route path="/datasources" element={<DatasourcePage />} />
        <Route path="/probe-targets" element={<ProbeTargetPage />} />
        <Route path="/alert-rules" element={<AlertRulePage />} />
        <Route path="/alert-dictionaries" element={<AlertDictionaryPage />} />
        <Route path="/alert-pages" element={<AlertPageManagePage />} />
        <Route path="/alerts" element={<AlertRealtimePage />} />
        <Route path="/alert-history" element={<AlertHistoryPage />} />
        <Route
          path="/notification-channels"
          element={<NotificationChannelPage />}
        />
        <Route
          path="/notification-templates"
          element={<NotificationTemplatePage />}
        />
        <Route
          path="/notification-templates/presets"
          element={<NotificationPresetPage />}
        />
        <Route
          path="/notification-groups"
          element={<NotificationGroupPage />}
        />
        <Route
          path="/notification-subscriptions"
          element={<NotificationSubscriptionPage />}
        />
        <Route
          path="/notification-oncalls"
          element={<NotificationOnCallPage />}
        />
        <Route path="/message-logs" element={<MessageLogPage />} />
        <Route path="/clusters" element={<ClusterPage />} />
        <Route path="/clusters/:uid" element={<ClusterDetailPage />} />
        <Route path="/agents" element={<AgentPage />} />
        <Route path="/agents/:uid/shell" element={<AgentShellPage />} />
        <Route path="/ssh-commands" element={<SSHCommandPage />} />
        <Route
          path="/ssh-commands/:uid/results"
          element={<SSHCommandResultsPage />}
        />
        <Route path="*" element={<NotFoundPage />} />
      </Route>
    </Routes>
  )
}

export default App
