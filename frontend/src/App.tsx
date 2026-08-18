import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import { Layout } from "./components/Layout";
import { AppDetailPage } from "./pages/AppDetailPage";
import { AppsPage, NewAppPage } from "./pages/AppsPage";
import { Dashboard } from "./pages/Dashboard";
import { LoginPage } from "./pages/LoginPage";
import { useAuthStore } from "./store/auth";

function Guard({ children }: { children: React.ReactNode }) {
  const token = useAuthStore((s) => s.token);
  if (!token) return <Navigate to="/login" replace />;
  return <>{children}</>;
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route
          element={
            <Guard>
              <Layout />
            </Guard>
          }
        >
          <Route path="/" element={<Dashboard />} />
          <Route path="/apps" element={<AppsPage />} />
          <Route path="/apps/new" element={<NewAppPage />} />
          <Route path="/apps/:id" element={<AppDetailPage />} />
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
}
