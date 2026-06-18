import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse, InternalAxiosRequestConfig } from 'axios';
import { message } from 'antd';
import { getToken, removeToken } from './auth';

const baseURL = '/api';

const service: AxiosInstance = axios.create({
  baseURL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

service.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = getToken();
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    console.error('Request error:', error);
    return Promise.reject(error);
  },
);

service.interceptors.response.use(
  (response: AxiosResponse) => {
    const res = response.data;
    if (res.code !== undefined && res.code !== 0 && res.code !== 200) {
      message.error(res.message || res.msg || '请求失败');
      if (res.code === 401 || res.code === 403) {
        removeToken();
        window.location.href = '/login';
      }
      return Promise.reject(new Error(res.message || res.msg || '请求失败'));
    }
    return res;
  },
  (error) => {
    const status = error.response?.status;
    const messageText = error.response?.data?.message || error.message;

    switch (status) {
      case 400:
        message.error(messageText || '请求参数错误');
        break;
      case 401:
        message.error('未授权，请重新登录');
        removeToken();
        window.location.href = '/login';
        break;
      case 403:
        message.error('没有权限访问');
        break;
      case 404:
        message.error('请求的资源不存在');
        break;
      case 500:
        message.error('服务器内部错误');
        break;
      case 502:
        message.error('网关错误');
        break;
      case 503:
        message.error('服务不可用');
        break;
      case 504:
        message.error('网关超时');
        break;
      default:
        message.error(messageText || '网络请求失败');
    }
    return Promise.reject(error);
  },
);

export const request = {
  get<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
    return service.get(url, config);
  },
  post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    return service.post(url, data, config);
  },
  put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    return service.put(url, data, config);
  },
  delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
    return service.delete(url, config);
  },
  patch<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    return service.patch(url, data, config);
  },
};

export default service;
