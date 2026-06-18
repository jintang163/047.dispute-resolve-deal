import React, { useState, useEffect } from 'react';
import { Outlet, useNavigate, useLocation, Link } from 'react-router-dom';
import { Dropdown, Avatar, Badge, theme } from 'antd';
import type { MenuProps } from 'antd';
import {
  ProLayout,
  DefaultFooter,
  PageContainer,
} from '@ant-design/pro-components';
import {
  LogoutOutlined,
  UserOutlined,
  SettingOutlined,
  BellOutlined,
  SafetyCertificateOutlined,
} from '@ant-design/icons';
import { useUserStore } from '../stores/user';
import { menuConfig, findMenuByPath } from '../config/menu';
import { approvalService } from '../services/approval';
import dayjs from 'dayjs';

const { defaultAlgorithm, darkAlgorithm } = theme;

const BasicLayout: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const userInfo = useUserStore((state) => state.userInfo);
  const logoutStore = useUserStore((state) => state.logout);
  const [isDark, setIsDark] = useState(false);
  const [todoCount, setTodoCount] = useState(0);

  useEffect(() => {
    fetchTodoCount();
  }, []);

  const fetchTodoCount = async () => {
    try {
      const res = await approvalService.getTodoCount();
      const data = res.data || res;
      setTodoCount(data.count || 0);
    } catch {}
  };

  const currentMenu = findMenuByPath(location.pathname);

  const userMenuItems: MenuProps['items'] = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人中心',
      onClick: () => {
        navigate('/profile');
      },
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: '账户设置',
      onClick: () => {
        navigate('/settings');
      },
    },
    {
      type: 'divider',
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      danger: true,
      onClick: () => {
        logoutStore();
        navigate('/login', { replace: true });
      },
    },
  ];

  const notificationItems: MenuProps['items'] = [
    {
      key: 'approvals',
      icon: <SafetyCertificateOutlined />,
      label: (
        <Link to="/approval/todo">
          待审批事项 <Badge count={todoCount} size="small" style={{ marginLeft: 8 }} />
        </Link>
      ),
    },
  ];

  return (
    <div
      id="test-pro-layout"
      style={{
        height: '100vh',
      }}
    >
      <ProLayout
        title="矛盾纠纷管理系统"
        logo="https://gw.alipayobjects.com/zos/rmsportal/KDpgvguMpGfqaHPjicRK.svg"
        theme={isDark ? darkAlgorithm : defaultAlgorithm}
        layout="mix"
        navTheme={isDark ? 'realDark' : 'light'}
        splitMenus={false}
        fixSiderbar
        fixedHeader
        location={{ pathname: location.pathname }}
        menu={{
          type: 'submenu',
        }}
        route={{
          path: '/',
          routes: menuConfig.map((item) => ({
            path: item.path,
            name: item.label,
            icon: item.icon as React.ReactNode,
            children: item.children?.map((child) => ({
              path: child.path,
              name: child.label,
              icon: child.icon as React.ReactNode,
            })),
          })),
        }}
        menuDataRender={() => menuConfig as any}
        itemRender={(route, params, routes, paths) => {
          const first = routes.indexOf(route) === 0;
          return first ? (
            <Link to={paths.join('/')}>{route.breadcrumbName}</Link>
          ) : (
            <span>{route.breadcrumbName}</span>
          );
        }}
        breadcrumbRender={(routers) => [
          {
            path: '/',
            breadcrumbName: '首页',
          },
          ...routers,
        ]}
        headerContentRender={() => {
          return (
            <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
              {currentMenu && <span style={{ fontSize: 16, fontWeight: 600 }}>{currentMenu.label}</span>}
            </div>
          );
        }}
        actionsRender={() => [
          <Dropdown key="notification" menu={{ items: notificationItems }} placement="bottomRight">
            <Badge count={todoCount} size="small" offset={[-2, 2]}>
              <span
                style={{
                  cursor: 'pointer',
                  padding: 4,
                  borderRadius: 4,
                  display: 'inline-flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                }}
              >
                <BellOutlined style={{ fontSize: 18 }} />
              </span>
            </Badge>
          </Dropdown>,
          <div
            key="theme"
            style={{
              cursor: 'pointer',
              padding: 4,
              borderRadius: 4,
              display: 'inline-flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
            onClick={() => setIsDark(!isDark)}
            title={isDark ? '切换为浅色模式' : '切换为深色模式'}
          >
            {isDark ? (
              <span role="img" aria-label="sun" style={{ fontSize: 18 }}>
                ☀️
              </span>
            ) : (
              <span role="img" aria-label="moon" style={{ fontSize: 18 }}>
                🌙
              </span>
            )}
          </div>,
        ]}
        avatarProps={{
          src: userInfo?.avatar,
          title: userInfo?.realName || userInfo?.username,
          icon: <UserOutlined />,
          size: 'small',
          render: (props, dom) => {
            return (
              <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
                <div style={{ cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 8 }}>
                  {dom}
                  <span style={{ fontSize: 14 }}>{userInfo?.realName || userInfo?.username}</span>
                </div>
              </Dropdown>
            );
          },
        }}
        subMenuItemRender={(itemDom, props) => {
          return itemDom;
        }}
        footerRender={() => (
          <DefaultFooter
            copyright={`${dayjs().year()} 综治中心矛盾纠纷管理系统`}
            style={{ background: 'transparent' }}
            links={[
              {
                key: '1',
                title: '技术支持',
                href: '#',
                blankTarget: true,
              },
              {
                key: '2',
                title: '帮助文档',
                href: '#',
                blankTarget: true,
              },
            ]}
          />
        )}
        onPageChange={(location) => {
          navigate(location.pathname);
        }}
      >
        <PageContainer
          header={{
            title: '',
            breadcrumb: {},
          }}
        >
          <Outlet />
        </PageContainer>
      </ProLayout>
    </div>
  );
};

export default BasicLayout;
