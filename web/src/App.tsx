import { Navigate, Route, Routes } from "react-router-dom"

import { Shell } from "@/components/shell"
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
      </Route>
    </Routes>
  )
}

export default App
