import { request } from '../utils/request';

const BASE = '/v1/dispute';

export const videoApi = {
  createRoom: (caseId: number, data: any) =>
    request.post(`${BASE}/${caseId}/video/create`, data),

  getRoomList: (caseId: number, params?: any) =>
    request.get(`${BASE}/${caseId}/video`, { params }),

  getRoomDetail: (caseId: number, roomId: string) =>
    request.get(`${BASE}/${caseId}/video/${roomId}`),

  joinRoom: (caseId: number, roomId: string, data: any) =>
    request.post(`${BASE}/${caseId}/video/${roomId}/join`, data),

  endRoom: (caseId: number, roomId: string) =>
    request.post(`${BASE}/${caseId}/video/${roomId}/end`),

  cancelRoom: (caseId: number, roomId: string, data: any) =>
    request.post(`${BASE}/${caseId}/video/${roomId}/cancel`, data),

  getTRTCUserSig: (caseId: number, data: { roomId: number; userId: string }) =>
    request.post(`${BASE}/${caseId}/video/trtc/usersig`, data),

  startRecord: (caseId: number, data: { roomId: number; userId?: string }) =>
    request.post(`${BASE}/${caseId}/video/record/start`, data),

  stopRecord: (caseId: number, data: { roomId: number }) =>
    request.post(`${BASE}/${caseId}/video/record/stop`, data),

  getRecordSegments: (caseId: number, roomId: string) =>
    request.get(`${BASE}/${caseId}/video/${roomId}/segments`),

  updateScreenShare: (caseId: number, data: { roomId: number; userId: number }) =>
    request.post(`${BASE}/${caseId}/video/screen-share`, data),

  updateVirtualBg: (caseId: number, data: { roomId: number; enabled: boolean }) =>
    request.post(`${BASE}/${caseId}/video/virtual-bg`, data),

  updateBeauty: (caseId: number, data: { roomId: number; enabled: boolean }) =>
    request.post(`${BASE}/${caseId}/video/beauty`, data),

  generateMinutes: (caseId: number, data: { roomId: number; caseId: number; transcript: string; durationMinutes?: number }) =>
    request.post(`${BASE}/${caseId}/video/minutes/generate`, data),

  getMinutes: (caseId: number, roomId: string) =>
    request.get(`${BASE}/${caseId}/video/${roomId}/minutes`),

  approveMinutes: (caseId: number, minutesId: string) =>
    request.post(`${BASE}/${caseId}/video/minutes/${minutesId}/approve`),
};

export const queueApi = {
  enqueue: (data: any) =>
    request.post(`${BASE}/video-queue/enqueue`, data),

  getPosition: (params: { caseId: number; userId: number }) =>
    request.get(`${BASE}/video-queue/position`, { params }),

  getList: () =>
    request.get(`${BASE}/video-queue/list`),

  leave: (data: { caseId: number; userId: number }) =>
    request.post(`${BASE}/video-queue/leave`, data),

  getStatus: () =>
    request.get(`${BASE}/video-queue/status`),
};
