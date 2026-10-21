import { Navigate, Route, Routes } from "react-router-dom"

import { Shell } from "@/components/shell"
import { AgentPage } from "@/pages/agent"
import { AgentShellPage } from "@/pages/agent/shell"
import { AlertHistoryPage } from "@/pages/alert/history"
import { AlertRealtimePage } from "@/pages/alert"
import { AlertDictionaryPage } from "@/pages/alert-dictionary"
import { AlertPageManagePage } from "@/pages/alert-page"
import { AlertRulePage } from "@/pages/alert-rule"
import { DatasourcePage } from "@/pages/datasource"
import { ForbiddenPage } from "@/pages/forbidden"
import { InvitationPage } from "@/pages/invitation"
import { InvitePage } from "@/pages/invite"
import { LoginPage } from "@/pages/login"
import { MemberPage } from "@/pages/member"
import { MessagePage } from "@/pages/message"
import { MessageLogPage } from "@/pages/message-log"
import { NamespacePage } from "@/pages/namespace"
import { NotificationChannelPage } from "@/pages/notification-channel"
import { NotificationGroupPage } from "@/pages/notification-group"
import { NotificationSubscriptionPage } from "@/pages/notification-subscription"
import { NotificationTemplatePage } from "@/pages/notification-template"
import { ProfilePage } from "@/pages/profile"
import { RegisterPage } from "@/pages/register"
import { RolePage } from "@/pages/role"
import { K8sPage } from "@/pages/k8s"
import { SFTPPage } from "@/pages/sftp"
import { SSHCommandPage } from "@/pages/ssh-command"
import { SSHCommandResultsPage } from "@/pages/ssh-command/results"
import { UserPage } from "@/pages/user"
import { WorkerStatusPage } from "@/pages/workers"

export function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route path="/invite/:token" element={<InvitePage />} />
      <Route element={<Shell />}>
        <Route path="/" element={<Navigate to="/profile" replace />} />
        <Route path="/profile" element={<ProfilePage />} />
        <Route path="/forbidden" element={<ForbiddenPage />} />
        <Route path="/namespaces" element={<NamespacePage />} />
        <Route path="/users" element={<UserPage />} />
        <Route path="/workers" element={<WorkerStatusPage />} />
        <Route path="/invitations" element={<InvitationPage />} />
        <Route path="/messages" element={<MessagePage />} />
        <Route path="/members" element={<MemberPage />} />
        <Route path="/roles" element={<RolePage />} />
        <Route path="/datasources" element={<DatasourcePage />} />
        <Route path="/alert-rules" element={<AlertRulePage />} />
        <Route path="/alert-dictionaries" element={<AlertDictionaryPage />} />
        <Route path="/alert-pages" element={<AlertPageManagePage />} />
        <Route path="/alerts" element={<AlertRealtimePage />} />
        <Route path="/alert-history" element={<AlertHistoryPage />} />
        <Route path="/notification-channels" element={<NotificationChannelPage />} />
        <Route path="/notification-templates" element={<NotificationTemplatePage />} />
        <Route path="/notification-groups" element={<NotificationGroupPage />} />
        <Route path="/notification-subscriptions" element={<NotificationSubscriptionPage />} />
        <Route path="/message-logs" element={<MessageLogPage />} />
        <Route path="/agents" element={<AgentPage />} />
        <Route path="/agents/:uid/shell" element={<AgentShellPage />} />
        <Route path="/ssh-commands" element={<SSHCommandPage />} />
        <Route path="/ssh-commands/:uid/results" element={<SSHCommandResultsPage />} />
        <Route path="/sftp" element={<SFTPPage />} />
        <Route path="/k8s" element={<K8sPage />} />
      </Route>
    </Routes>
  )
}

export default App
