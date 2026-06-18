import React, { useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate, useLocation } from 'react-router-dom';
import BasicLayout from './layouts/BasicLayout';
import Login from './pages/Login';
import DisputeList from './pages/Dispute/List';
import DisputeDetail from './pages/Dispute/Detail';
import DisputeCreate from './pages/Dispute/Create';
import MediationList from './pages/Mediation/List';
import TodoApproval from './pages/Approval/Todo';
import DoneApproval from './pages/Approval/Done';
import Dashboard from './pages/Stats/Dashboard';
import UserList from './pages/System/UserList';
import OrgList from './pages/System/OrgList';
import Performance from './pages/Performance/Index';
import { getToken } from './utils/auth';
import { useUserStore } from './stores/user';

const PrivateRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const token = getToken();
  const location = useLocation();
  const fetchUserInfo = useUserStore((state) => state.fetchUserInfo);

  useEffect(() => {
    if (token) {
      fetchUserInfo();
    }
  }, [token, fetchUserInfo]);

  if (!token) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }
  return <>{children}</>;
};

const App: React.FC = () => {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route
          path="/"
          element={
            <PrivateRoute>
              <BasicLayout />
            </PrivateRoute>
          }
        >
          <Route index element={<Navigate to="/dashboard" replace />} />
          <Route path="dashboard" element={<Dashboard />} />
          <Route path="dispute">
            <Route index element={<DisputeList />} />
            <Route path="create" element={<DisputeCreate />} />
            <Route path=":id" element={<DisputeDetail />} />
          </Route>
          <Route path="mediation" element={<MediationList />} />
          <Route path="approval">
            <Route index element={<Navigate to="/approval/todo" replace />} />
            <Route path="todo" element={<TodoApproval />} />
            <Route path="done" element={<DoneApproval />} />
          </Route>
          <Route path="system">
            <Route path="users" element={<UserList />} />
            <Route path="orgs" element={<OrgList />} />
          </Route>
          <Route path="performance" element={<Performance />} />
        </Route>
        <Route path="*" element={<Navigate to="/dashboard" replace />} />
      </Routes>
    </BrowserRouter>
  );
};

export default App;
