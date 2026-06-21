import React, { useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate, useLocation, useParams, useNavigate } from 'react-router-dom';
import BasicLayout from './layouts/BasicLayout';
import Login from './pages/Login';
import DisputeList from './pages/Dispute/List';
import DisputeDetail from './pages/Dispute/Detail';
import DisputeCreate from './pages/Dispute/Create';
import MediationList from './pages/Mediation/List';
import VideoMediation from './pages/Video';
import VideoRoom from './pages/Video/VideoRoom';
import TodoApproval from './pages/Approval/Todo';
import DoneApproval from './pages/Approval/Done';
import Dashboard from './pages/Stats/Dashboard';
import HeatMap from './pages/HeatMap';
import UserList from './pages/System/UserList';
import OrgList from './pages/System/OrgList';
import Performance from './pages/Performance/Index';
import JudicialList from './pages/Judicial/List';
import JudicialDetail from './pages/Judicial/Detail';
import CourtConfig from './pages/Judicial/CourtConfig';
import EsignList from './pages/ESign/List';
import EsignDetail from './pages/ESign/Detail';
import CertificateList from './pages/ESign/Certificate';
import PublicVerify from './pages/ESign/Verify';
import CallbackList from './pages/Callback/List';
import SatisfactionAnalysis from './pages/Satisfaction/Analysis';
import ImprovementList from './pages/Satisfaction/ImprovementList';
import CaseLibraryList from './pages/CaseLibrary/List';
import CaseLibraryDetail from './pages/CaseLibrary/Detail';
import CaseLibraryCreate from './pages/CaseLibrary/Create';
import CounselorList from './pages/Counseling/CounselorList';
import AppointmentList from './pages/Counseling/AppointmentList';
import ExportLogList from './pages/ExportLog/List';
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

const VideoRoomPage: React.FC = () => {
  const { caseId, roomId } = useParams();
  const navigate = useNavigate();
  const caseIdNum = parseInt(caseId || '0');
  const roomIdNum = parseInt(roomId || '0');
  const userInfo = useUserStore((state) => state.userInfo);

  return (
    <VideoRoom
      caseId={caseIdNum}
      roomId={roomId || ''}
      trtcRoomId={roomIdNum}
      userId={String(userInfo?.id || 0)}
      userSig=""
      sdkAppId={0}
      isHost={true}
      onClose={() => navigate('/video')}
    />
  );
};

const App: React.FC = () => {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/verify/:certNo" element={<PublicVerify />} />
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
          <Route path="heatmap" element={<HeatMap />} />
          <Route path="dispute">
            <Route index element={<DisputeList />} />
            <Route path="create" element={<DisputeCreate />} />
            <Route path=":id" element={<DisputeDetail />} />
          </Route>
          <Route path="mediation" element={<MediationList />} />
          <Route path="video" element={<VideoMediation />} />
          <Route path="video/room/:caseId/:roomId" element={<VideoRoomPage />} />
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
          <Route path="judicial">
            <Route index element={<JudicialList />} />
            <Route path=":id" element={<JudicialDetail />} />
          </Route>
          <Route path="court">
            <Route path="config" element={<CourtConfig />} />
          </Route>
          <Route path="esign">
            <Route index element={<EsignList />} />
            <Route path=":caseId/:flowId" element={<EsignDetail />} />
            <Route path="certificate" element={<CertificateList />} />
          </Route>
          <Route path="callback" element={<CallbackList />} />
          <Route path="satisfaction">
            <Route index element={<SatisfactionAnalysis />} />
            <Route path="improvement" element={<ImprovementList />} />
          </Route>
          <Route path="case-library">
            <Route index element={<CaseLibraryList />} />
            <Route path="create" element={<CaseLibraryCreate />} />
            <Route path=":id" element={<CaseLibraryDetail />} />
          </Route>
          <Route path="counseling">
            <Route path="counselor" element={<CounselorList />} />
            <Route path="appointment" element={<AppointmentList />} />
          </Route>
          <Route path="export">
            <Route path="log" element={<ExportLogList />} />
          </Route>
        </Route>
        <Route path="*" element={<Navigate to="/dashboard" replace />} />
      </Routes>
    </BrowserRouter>
  );
};

export default App;
