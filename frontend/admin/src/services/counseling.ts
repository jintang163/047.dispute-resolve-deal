import { request } from '../utils/request';

export interface Counselor {
  id: string | number;
  counselorNo: string;
  userId?: string | number;
  realName: string;
  gender?: number;
  phone?: string;
  email?: string;
  avatar?: string;
  title?: string;
  licenseNo?: string;
  specialty?: string;
  specialtyTags?: string;
  specialtyList?: string[];
  specialtyTagList?: string[];
  yearsOfExperience?: number;
  education?: string;
  introduction?: string;
  consultationTypes?: string;
  consultationTypeList?: Array<{ type: number; name: string }>;
  workDays?: string;
  workDayList?: string[];
  workStartTime?: string;
  workEndTime?: string;
  sessionDuration?: number;
  price?: number;
  organizationId?: string | number;
  organizationName?: string;
  ratingAvg?: number;
  ratingCount?: number;
  appointmentCount?: number;
  completedCount?: number;
  isEmergencyAvailable?: number;
  status?: number;
  statusName?: string;
  sortOrder?: number;
  recentRatings?: CounselorRating[];
}

export interface CounselorListParams {
  page?: number;
  pageSize?: number;
  keyword?: string;
  status?: number;
  specialty?: string;
  isEmergencyAvailable?: number;
  consultationType?: number;
  organizationId?: string | number;
  sortBy?: string;
  sortOrder?: string;
}

export interface CounselorCreateParams {
  realName: string;
  gender?: number;
  phone?: string;
  email?: string;
  avatar?: string;
  title?: string;
  licenseNo?: string;
  specialty?: string;
  specialtyTags?: string;
  yearsOfExperience?: number;
  education?: string;
  introduction?: string;
  consultationTypes?: string;
  workDays?: string;
  workStartTime?: string;
  workEndTime?: string;
  sessionDuration?: number;
  price?: number;
  organizationId?: string | number;
  organizationName?: string;
  isEmergencyAvailable?: number;
  status?: number;
  sortOrder?: number;
}

export interface CounselorUpdateParams extends Partial<CounselorCreateParams> {}

export interface CounselorAppointment {
  id: string | number;
  appointmentNo: string;
  counselorId: string | number;
  counselorName?: string;
  counselorRealName?: string;
  counselorTitle?: string;
  counselorAvatar?: string;
  counselorPhone?: string;
  caseId?: string | number;
  partyId?: string | number;
  partyName?: string;
  partyPhone?: string;
  partyIdCard?: string;
  partyNameDisplay?: string;
  partyPhoneDisplay?: string;
  partyIdCardDisplay?: string;
  isAnonymous?: number;
  anonymousCode?: string;
  appointmentDate: string;
  startTime: string;
  endTime: string;
  consultationType?: number;
  consultationTypeName?: string;
  appointmentSource?: number;
  isEmergency?: number;
  emergencyTriggerWords?: string;
  emergencyLevel?: number;
  emergencyLevelName?: string;
  concernType?: string;
  concernDescription?: string;
  status?: number;
  statusName?: string;
  cancelReason?: string;
  cancelledBy?: string | number;
  cancelledAt?: string;
  confirmedBy?: string | number;
  confirmedAt?: string;
  startedAt?: string;
  completedAt?: string;
  consultationSummary?: string;
  followUpSuggestion?: string;
  nextAppointmentSuggestion?: string;
  roomId?: string;
  roomUrl?: string;
  location?: string;
  reminderSent?: number;
  ratingSubmitted?: number;
  createdBy?: string | number;
  createdByName?: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface AppointmentListParams {
  page?: number;
  pageSize?: number;
  keyword?: string;
  status?: number;
  counselorId?: string | number;
  caseId?: string | number;
  partyName?: string;
  isEmergency?: number;
  isAnonymous?: number;
  startDate?: string;
  endDate?: string;
  appointmentDate?: string;
  sortBy?: string;
  sortOrder?: string;
}

export interface AppointmentCreateParams {
  counselorId: string | number;
  caseId?: string | number;
  partyId?: string | number;
  partyName?: string;
  partyPhone?: string;
  partyIdCard?: string;
  isAnonymous?: number;
  appointmentDate: string;
  startTime: string;
  endTime: string;
  consultationType?: number;
  concernType?: string;
  concernDescription?: string;
  isEmergency?: number;
}

export interface AppointmentUpdateParams {
  status?: number;
  counselorId?: string | number;
  appointmentDate?: string;
  startTime?: string;
  endTime?: string;
  consultationType?: number;
  concernType?: string;
  concernDescription?: string;
  consultationSummary?: string;
  followUpSuggestion?: string;
  location?: string;
  cancelReason?: string;
}

export interface CounselorRating {
  id: string | number;
  appointmentId: string | number;
  counselorId: string | number;
  raterId?: string | number;
  raterName?: string;
  isAnonymousRating?: number;
  overallScore: number;
  professionalScore?: number;
  attitudeScore?: number;
  empathyScore?: number;
  helpfulScore?: number;
  content?: string;
  tags?: string;
  tagList?: string[];
  counselorReply?: string;
  counselorReplyAt?: string;
  isHelpful?: number;
  status?: number;
  createdAt?: string;
}

export interface RatingCreateParams {
  appointmentId: string | number;
  counselorId: string | number;
  isAnonymousRating?: number;
  overallScore: number;
  professionalScore?: number;
  attitudeScore?: number;
  empathyScore?: number;
  helpfulScore?: number;
  content?: string;
  tags?: string;
}

export interface CounselorSchedule {
  id: string | number;
  counselorId: string | number;
  scheduleDate: string;
  startTime: string;
  endTime: string;
  scheduleType?: number;
  scheduleTypeName?: string;
  title?: string;
  remark?: string;
  appointmentId?: string | number;
  createdBy?: string | number;
  createdAt?: string;
}

export interface AvailableSlot {
  startTime: string;
  endTime: string;
  available: boolean;
}

export interface CounselorRecommendParams {
  caseId?: string | number;
  disputeType?: string;
  keywords?: string;
  description?: string;
  isEmergency?: number;
}

export interface CounselorStats {
  totalAppointments: number;
  completedAppointments: number;
  avgRating: number;
  totalRatings: number;
  ratingDistribution: Array<{ score: number; count: number }>;
}

export const counselingService = {
  getCounselorList: (params?: CounselorListParams) => {
    return request.get<{ list: Counselor[]; total: number; page: number; pageSize: number }>(
      '/counseling/counselor',
      { params },
    );
  },

  getCounselorDetail: (id: string | number) => {
    return request.get<Counselor>(`/counseling/counselor/${id}`);
  },

  createCounselor: (params: CounselorCreateParams) => {
    return request.post<{ id: string | number; counselorNo: string }>(
      '/counseling/counselor',
      params,
    );
  },

  updateCounselor: (id: string | number, params: CounselorUpdateParams) => {
    return request.put(`/counseling/counselor/${id}`, params);
  },

  deleteCounselor: (id: string | number) => {
    return request.delete(`/counseling/counselor/${id}`);
  },

  getCounselorAvailableSlots: (id: string | number, date?: string) => {
    return request.get<{ date: string; slots: AvailableSlot[] }>(
      `/counseling/counselor/${id}/available-slots`,
      { params: { date } },
    );
  },

  getCounselorRatings: (id: string | number, page = 1, pageSize = 10) => {
    return request.get<{ list: CounselorRating[]; total: number; page: number; pageSize: number }>(
      `/counseling/counselor/${id}/ratings`,
      { params: { page, pageSize } },
    );
  },

  getCounselorStats: (id: string | number) => {
    return request.get<CounselorStats>(`/counseling/counselor/${id}/stats`);
  },

  recommendCounselors: (params: CounselorRecommendParams) => {
    return request.post<Counselor[]>('/counseling/counselor/recommend', params);
  },

  getAppointmentList: (params?: AppointmentListParams) => {
    return request.get<{ list: CounselorAppointment[]; total: number; page: number; pageSize: number }>(
      '/counseling/appointment',
      { params },
    );
  },

  getAppointmentDetail: (id: string | number) => {
    return request.get<CounselorAppointment>(`/counseling/appointment/${id}`);
  },

  createAppointment: (params: AppointmentCreateParams) => {
    return request.post<{
      id: string | number;
      appointmentNo: string;
      isEmergency?: number;
      emergencyLevel?: number;
      warning?: string;
      anonymousCode?: string;
    }>('/counseling/appointment', params);
  },

  updateAppointment: (id: string | number, params: AppointmentUpdateParams) => {
    return request.put(`/counseling/appointment/${id}`, params);
  },

  cancelAppointment: (id: string | number, reason?: string) => {
    return request.post(`/counseling/appointment/${id}/cancel`, { reason });
  },

  createRating: (params: RatingCreateParams) => {
    return request.post<{ id: string | number }>('/counseling/rating', params);
  },

  getScheduleList: (params?: {
    page?: number;
    pageSize?: number;
    counselorId?: string | number;
    scheduleDate?: string;
    startDate?: string;
    endDate?: string;
  }) => {
    return request.get<{ list: CounselorSchedule[]; total: number; page: number; pageSize: number }>(
      '/counseling/schedule',
      { params },
    );
  },

  createSchedule: (params: {
    counselorId: string | number;
    scheduleDate: string;
    startTime: string;
    endTime: string;
    scheduleType?: number;
    title?: string;
    remark?: string;
  }) => {
    return request.post<{ id: string | number }>('/counseling/schedule', params);
  },

  deleteSchedule: (id: string | number) => {
    return request.delete(`/counseling/schedule/${id}`);
  },
};
