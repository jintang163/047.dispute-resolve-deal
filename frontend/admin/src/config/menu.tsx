import React from 'react';
import {
  DashboardOutlined,
  FileTextOutlined,
  AuditOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  UserOutlined,
  TeamOutlined,
  TrophyOutlined,
  SettingOutlined,
  SafetyCertificateOutlined,
  BankOutlined,
  VideoCameraOutlined,
  FileProtectOutlined,
  HeatMapOutlined,
  PhoneOutlined,
  BarChartOutlined,
  ToolOutlined,
} from '@ant-design/icons';

export interface MenuItem {
  key: string;
  label: string;
  icon?: React.ReactNode;
  path?: string;
  children?: MenuItem[];
}

export const menuConfig: MenuItem[] = [
  {
    key: 'dashboard',
    label: '数据概览',
    icon: React.createElement(DashboardOutlined),
    path: '/dashboard',
  },
  {
    key: 'heatmap',
    label: '纠纷热力图',
    icon: React.createElement(HeatMapOutlined),
    path: '/heatmap',
  },
  {
    key: 'dispute',
    label: '纠纷案件管理',
    icon: React.createElement(FileTextOutlined),
    children: [
      {
        key: 'dispute-list',
        label: '案件列表',
        path: '/dispute',
      },
      {
        key: 'dispute-create',
        label: '新增案件',
        path: '/dispute/create',
      },
    ],
  },
  {
    key: 'mediation',
    label: '调解记录',
    icon: React.createElement(AuditOutlined),
    path: '/mediation',
  },
  {
    key: 'video',
    label: '视频调解室',
    icon: React.createElement(VideoCameraOutlined),
    path: '/video',
  },
  {
    key: 'approval',
    label: '审批中心',
    icon: React.createElement(CheckCircleOutlined),
    children: [
      {
        key: 'approval-todo',
        label: '待审批',
        icon: React.createElement(ClockCircleOutlined),
        path: '/approval/todo',
      },
      {
        key: 'approval-done',
        label: '已审批',
        icon: React.createElement(CheckCircleOutlined),
        path: '/approval/done',
      },
    ],
  },
  {
    key: 'system',
    label: '系统管理',
    icon: React.createElement(SettingOutlined),
    children: [
      {
        key: 'system-users',
        label: '用户管理',
        icon: React.createElement(UserOutlined),
        path: '/system/users',
      },
      {
        key: 'system-orgs',
        label: '组织管理',
        icon: React.createElement(TeamOutlined),
        path: '/system/orgs',
      },
    ],
  },
  {
    key: 'performance',
    label: '绩效考核看板',
    icon: React.createElement(TrophyOutlined),
    path: '/performance',
  },
  {
    key: 'judicial',
    label: '司法确认',
    icon: React.createElement(SafetyCertificateOutlined),
    children: [
      {
        key: 'judicial-list',
        label: '确认申请管理',
        icon: React.createElement(FileTextOutlined),
        path: '/judicial',
      },
      {
        key: 'court-config',
        label: '法院配置管理',
        icon: React.createElement(BankOutlined),
        path: '/court/config',
      },
    ],
  },
  {
    key: 'esign',
    label: '电子签章与存证',
    icon: React.createElement(FileProtectOutlined),
    children: [
      {
        key: 'esign-list',
        label: '签章管理',
        icon: React.createElement(FileTextOutlined),
        path: '/esign',
      },
      {
        key: 'esign-certificate',
        label: '区块链存证',
        icon: React.createElement(SafetyCertificateOutlined),
        path: '/esign/certificate',
      },
    ],
  },
  {
    key: 'callback',
    label: '自动回访管理',
    icon: React.createElement(PhoneOutlined),
    path: '/callback',
  },
  {
    key: 'satisfaction',
    label: '满意度分析',
    icon: React.createElement(BarChartOutlined),
    children: [
      {
        key: 'satisfaction-analysis',
        label: '情感分析统计',
        icon: React.createElement(BarChartOutlined),
        path: '/satisfaction',
      },
      {
        key: 'improvement-list',
        label: '改进工单管理',
        icon: React.createElement(ToolOutlined),
        path: '/satisfaction/improvement',
      },
    ],
  },
];

export const getFlatMenuKeys = (items: MenuItem[] = menuConfig): string[] => {
  const keys: string[] = [];
  items.forEach((item) => {
    keys.push(item.key);
    if (item.children) {
      keys.push(...getFlatMenuKeys(item.children));
    }
  });
  return keys;
};

export const findMenuByPath = (
  path: string,
  items: MenuItem[] = menuConfig,
): MenuItem | undefined => {
  for (const item of items) {
    if (item.path === path) {
      return item;
    }
    if (item.children) {
      const found = findMenuByPath(path, item.children);
      if (found) return found;
    }
  }
  return undefined;
};
