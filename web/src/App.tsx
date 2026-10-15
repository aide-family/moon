import { Navigate, Route, Routes } from "react-router-dom"

import { Shell } from "@/components/shell"
import { AlertHistoryPage } from "@/pages/alert/history"
import { AlertRealtimePage } from "@/pages/alert"
import { AlertLevelPage } from "@/pages/alert-level"
import { AlertRulePage } from "@/pages/alert-rule"
import { DatasourcePage } from "@/pages/datasource"
import { LoginPage } from "@/pages/login"
import { MemberPage } from "@/pages/member"
import { NamespacePage } from "@/pages/namespace"
import { ProfilePage } from "@/pages/profile"
import { RegisterPage } from "@/pages/register"
import { RolePage } from "@/pages/role"
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
        <Route path="/alerts" element={<AlertRealtimePage />} />
        <Route path="/alert-history" element={<AlertHistoryPage />} />
      </Route>
    </Routes>
  )
}

export default App
