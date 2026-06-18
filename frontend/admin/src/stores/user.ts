import { create } from 'zustand';
import { UserInfo, getUserInfo, setUserInfo, removeToken, removeUserInfo } from '../utils/auth';
import { userService } from '../services/user';

interface UserState {
  userInfo: UserInfo | null;
  token: string | null;
  loading: boolean;
  setUserInfo: (info: UserInfo | null) => void;
  setToken: (token: string | null) => void;
  fetchUserInfo: () => Promise<void>;
  logout: () => void;
}

export const useUserStore = create<UserState>((set, get) => ({
  userInfo: getUserInfo(),
  token: localStorage.getItem('DISPUTE_ADMIN_TOKEN'),
  loading: false,

  setUserInfo: (info) => {
    set({ userInfo: info });
    if (info) {
      setUserInfo(info);
    } else {
      removeUserInfo();
    }
  },

  setToken: (token) => {
    set({ token });
    if (token) {
      localStorage.setItem('DISPUTE_ADMIN_TOKEN', token);
    } else {
      localStorage.removeItem('DISPUTE_ADMIN_TOKEN');
    }
  },

  fetchUserInfo: async () => {
    const { userInfo: existingInfo } = get();
    if (existingInfo) {
      return;
    }
    try {
      set({ loading: true });
      const res = await userService.getCurrentUser();
      const data = res.data || res;
      if (data) {
        set({ userInfo: data });
        setUserInfo(data);
      }
    } catch (error) {
      console.error('Fetch user info error:', error);
    } finally {
      set({ loading: false });
    }
  },

  logout: () => {
    set({ userInfo: null, token: null });
    removeToken();
    removeUserInfo();
  },
}));
