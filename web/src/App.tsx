import { Navigate, Route, Routes } from "react-router-dom"

import { Shell } from "@/components/shell"
import { AgentPage } from "@/pages/agent"
import { AlertHistoryPage } from "@/pages/alert/history"
import { AlertRealtimePage } from "@/pages/alert"
import { AlertLevelPage } from "@/pages/alert-level"
import { AlertPageManagePage } from "@/pages/alert-page"
import { AlertRulePage } from "@/pages/alert-rule"
import { DatasourcePage } from "@/pages/datasource"
import { LoginPage } from "@/pages/login"
import { MemberPage } from "@/pages/member"
import { MessageLogPage } from "@/pages/message-log"
import { NamespacePage } from "@/pages/namespace"
import { NotificationChannelPage } from "@/pages/notification-channel"
import { NotificationGroupPage } from "@/pages/notification-group"
import { NotificationTemplatePage } from "@/pages/notification-template"
import { ProfilePage } from "@/pages/profile"
import { RegisterPage } from "@/pages/register"
import { RolePage } from "@/pages/role"
import { K8sPage } from "@/pages/k8s"
import { SFTPPage } from "@/pages/sftp"
import { SSHCommandPage } from "@/pages/ssh-command"
import { UserPage } from "@/pages/user"

export function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route element={<Shell />}>
        <Route path="/" element={<Navigate to="/profile" replace />} />
        <Route path="/profile" element={<ProfilePage />} />
        <Route path="/namespaces" element={<NamespacePage />} />
        <Route path="/users" element={<UserPage />} />
        <Route path="/members" element={<MemberPage />} />
        <Route path="/roles" element={<RolePage />} />
        <Route path="/datasources" element={<DatasourcePage />} />
        <Route path="/alert-rules" element={<AlertRulePage />} />
        <Route path="/alert-levels" element={<AlertLevelPage />} />
        <Route path="/alert-pages" element={<AlertPageManagePage />} />
        <Route path="/alerts" element={<AlertRealtimePage />} />
        <Route path="/alert-history" element={<AlertHistoryPage />} />
        <Route path="/notification-channels" element={<NotificationChannelPage />} />
        <Route path="/notification-templates" element={<NotificationTemplatePage />} />
        <Route path="/notification-groups" element={<NotificationGroupPage />} />
        <Route path="/message-logs" element={<MessageLogPage />} />
        <Route path="/agents" element={<AgentPage />} />
        <Route path="/ssh-commands" element={<SSHCommandPage />} />
        <Route path="/sftp" element={<SFTPPage />} />
        <Route path="/k8s" element={<K8sPage />} />
      </Route>
    </Routes>
  )
}

export default App
