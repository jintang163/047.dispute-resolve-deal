const TOKEN_KEY = 'DISPUTE_ADMIN_TOKEN';
const USER_INFO_KEY = 'DISPUTE_ADMIN_USER_INFO';

export interface UserInfo {
  id: string;
  username: string;
  realName: string;
  role: string;
  roleName: string;
  avatar?: string;
  orgId?: string;
  orgName?: string;
  phone?: string;
}

export const getToken = (): string | null => {
  return localStorage.getItem(TOKEN_KEY);
};

export const setToken = (token: string): void => {
  localStorage.setItem(TOKEN_KEY, token);
};

export const removeToken = (): void => {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(USER_INFO_KEY);
};

export const getUserInfo = (): UserInfo | null => {
  const info = localStorage.getItem(USER_INFO_KEY);
  if (info) {
    try {
      return JSON.parse(info);
    } catch {
      return null;
    }
  }
  return null;
};

export const setUserInfo = (userInfo: UserInfo): void => {
  localStorage.setItem(USER_INFO_KEY, JSON.stringify(userInfo));
};

export const removeUserInfo = (): void => {
  localStorage.removeItem(USER_INFO_KEY);
};

export const login = async (
  loginFn: (params: { username: string; password: string; role: string }) => Promise<any>,
  params: { username: string; password: string; role: string },
): Promise<UserInfo> => {
  const res = await loginFn(params);
  const data = res.data || res;
  if (data.token) {
    setToken(data.token);
  }
  if (data.userInfo) {
    setUserInfo(data.userInfo);
  }
  return data.userInfo || data;
};

export const logout = (): void => {
  removeToken();
  removeUserInfo();
  window.location.href = '/login';
};

export const isAuthenticated = (): boolean => {
  return !!getToken();
};
