import React, { useState } from 'react';
import { Form, Input, Button, Select, Card, App } from 'antd';
import { UserOutlined, LockOutlined, SafetyCertificateOutlined } from '@ant-design/icons';
import { useNavigate, useLocation } from 'react-router-dom';
import { ProFormText } from '@ant-design/pro-components';
import { userService } from '../../services/user';
import { useUserStore } from '../../stores/user';
import { setToken, setUserInfo } from '../../utils/auth';

interface LocationState {
  from?: { pathname: string };
}

const Login: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const state = location.state as LocationState;
  const { message } = App.useApp();
  const [loading, setLoading] = useState(false);
  const setUserInfoStore = useUserStore((state) => state.setUserInfo);
  const setTokenStore = useUserStore((state) => state.setToken);
  const [form] = Form.useForm();

  const roleOptions = [
    { value: 'admin', label: '系统管理员' },
    { value: 'mediator', label: '调解员' },
    { value: 'approver', label: '审批员' },
    { value: 'leader', label: '综治领导' },
    { value: 'operator', label: '办事员' },
  ];

  const onFinish = async (values: { username: string; password: string; role: string }) => {
    try {
      setLoading(true);
      const res = await userService.login({
        username: values.username,
        password: values.password,
        role: values.role,
      });
      const data = res.data || res;
      if (data.token) {
        setToken(data.token);
        setTokenStore(data.token);
      }
      if (data.userInfo) {
        setUserInfo(data.userInfo);
        setUserInfoStore(data.userInfo);
      }
      message.success('登录成功');
      const from = state?.from?.pathname || '/dashboard';
      navigate(from, { replace: true });
    } catch (error: any) {
      message.error(error.message || '登录失败，请检查用户名和密码');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="login-container">
      <div style={{ width: 420 }}>
        <h1 className="login-title">
          <SafetyCertificateOutlined style={{ marginRight: 12 }} />
          综治中心矛盾纠纷管理系统
        </h1>
        <p className="login-subtitle">构建和谐社会，化解矛盾纠纷</p>
        <Card className="login-card" bordered={false}>
          <Form
            form={form}
            layout="vertical"
            onFinish={onFinish}
            initialValues={{ role: 'admin' }}
            size="large"
          >
            <Form.Item
              name="username"
              label="用户名"
              rules={[
                { required: true, message: '请输入用户名' },
                { min: 2, max: 32, message: '用户名长度为 2-32 个字符' },
              ]}
            >
              <ProFormText
                prefix={<UserOutlined className="site-form-item-icon" />}
                placeholder="请输入用户名"
                fieldProps={{
                  autoComplete: 'username',
                }}
              />
            </Form.Item>

            <Form.Item
              name="password"
              label="密码"
              rules={[
                { required: true, message: '请输入密码' },
                { min: 6, max: 32, message: '密码长度为 6-32 个字符' },
              ]}
            >
              <Input.Password
                prefix={<LockOutlined className="site-form-item-icon" />}
                placeholder="请输入密码"
                autoComplete="current-password"
              />
            </Form.Item>

            <Form.Item
              name="role"
              label="角色选择"
              rules={[{ required: true, message: '请选择登录角色' }]}
            >
              <Select
                placeholder="请选择登录角色"
                options={roleOptions}
              />
            </Form.Item>

            <Form.Item style={{ marginBottom: 0, marginTop: 24 }}>
              <Button
                block
                size="large"
                type="primary"
                htmlType="submit"
                loading={loading}
              >
                登 录
              </Button>
            </Form.Item>

            <div style={{ marginTop: 16, display: 'flex', justifyContent: 'space-between' }}>
              <a href="#forgot" style={{ color: '#1677ff' }}>忘记密码?</a>
              <span style={{ color: '#999', fontSize: 12 }}>
                默认账号: admin / 123456
              </span>
            </div>
          </Form>
        </Card>
      </div>
    </div>
  );
};

export default Login;
