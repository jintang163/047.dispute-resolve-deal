import React, { useRef, useState } from 'react';
import { Button, Tag, Space, Switch, App, Modal, Form, Input, Select } from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  ReloadOutlined,
  KeyOutlined,
  UserOutlined,
} from '@ant-design/icons';
import {
  ProTable,
  ProFormSelect,
  ProFormText,
  ModalForm,
} from '@ant-design/pro-components';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { userService, User } from '../../services/user';
import dayjs from 'dayjs';

const roleColorMap: Record<string, string> = {
  admin: 'purple',
  mediator: 'blue',
  approver: 'orange',
  leader: 'red',
  operator: 'green',
};

const roleTextMap: Record<string, string> = {
  admin: '系统管理员',
  mediator: '调解员',
  approver: '审批员',
  leader: '综治领导',
  operator: '办事员',
};

const UserList: React.FC = () => {
  const { message, modal } = App.useApp();
  const actionRef = useRef<ActionType>();
  const [modalOpen, setModalOpen] = useState(false);
  const [editMode, setEditMode] = useState<'create' | 'edit'>('create');
  const [currentRecord, setCurrentRecord] = useState<User | null>(null);
  const [form] = Form.useForm();
  const [resetModalOpen, setResetModalOpen] = useState(false);
  const [resetUserId, setResetUserId] = useState<string | null>(null);
  const [resetForm] = Form.useForm();
  const [loading, setLoading] = useState(false);

  const handleModalOpen = (mode: 'create' | 'edit', record?: User) => {
    setEditMode(mode);
    setCurrentRecord(record || null);
    if (mode === 'edit' && record) {
      form.setFieldsValue(record);
    } else {
      form.resetFields();
    }
    setModalOpen(true);
  };

  const handleModalFinish = async (values: any) => {
    try {
      setLoading(true);
      if (editMode === 'create') {
        await userService.create(values);
        message.success('用户创建成功');
      } else if (editMode === 'edit' && currentRecord) {
        await userService.update(currentRecord.id, values);
        message.success('用户更新成功');
      }
      actionRef.current?.reload();
      setModalOpen(false);
    } catch (error: any) {
      message.error(error.message || '操作失败');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = (record: User) => {
    modal.confirm({
      title: '确认删除该用户?',
      icon: <DeleteOutlined />,
      content: `用户: ${record.realName} (${record.username})`,
      okText: '确认删除',
      cancelText: '取消',
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          await userService.delete(record.id);
          message.success('删除成功');
          actionRef.current?.reload();
        } catch (error: any) {
          message.error(error.message || '删除失败');
          return Promise.reject();
        }
      },
    });
  };

  const handleToggleStatus = async (record: User, checked: boolean) => {
    try {
      await userService.update(record.id, { status: checked ? 'active' : 'disabled' });
      message.success(`用户已${checked ? '启用' : '禁用'}`);
      actionRef.current?.reload();
    } catch (error: any) {
      message.error(error.message || '操作失败');
      actionRef.current?.reload();
    }
  };

  const handleResetPassword = (record: User) => {
    setResetUserId(record.id);
    resetForm.resetFields();
    setResetModalOpen(true);
  };

  const handleResetPasswordFinish = async (values: any) => {
    if (!resetUserId) return;
    try {
      setLoading(true);
      await userService.resetPassword(resetUserId, values.newPassword);
      message.success('密码重置成功');
      setResetModalOpen(false);
      setResetUserId(null);
    } catch (error: any) {
      message.error(error.message || '密码重置失败');
    } finally {
      setLoading(false);
    }
  };

  const columns: ProColumns<User>[] = [
    {
      title: '用户信息',
      dataIndex: 'username',
      width: 180,
      fixed: 'left',
      render: (_, record) => (
        <Space>
          <div
            style={{
              width: 36,
              height: 36,
              borderRadius: '50%',
              background: '#1677ff15',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: '#1677ff',
              fontSize: 16,
            }}
          >
            {record.realName?.charAt(0) || <UserOutlined />}
          </div>
          <Space direction="vertical" size={0}>
            <span style={{ fontWeight: 500 }}>{record.realName}</span>
            <span style={{ fontSize: 12, color: '#999' }}>{record.username}</span>
          </Space>
        </Space>
      ),
    },
    {
      title: '角色',
      dataIndex: 'role',
      width: 120,
      valueEnum: roleTextMap,
      render: (_, record) => (
        <Tag color={roleColorMap[record.role] || 'default'}>
          {record.roleName || roleTextMap[record.role] || record.role}
        </Tag>
      ),
    },
    {
      title: '所属组织',
      dataIndex: 'orgName',
      width: 180,
      ellipsis: true,
    },
    {
      title: '联系电话',
      dataIndex: 'phone',
      width: 140,
    },
    {
      title: '邮箱',
      dataIndex: 'email',
      width: 200,
      ellipsis: true,
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (_, record) => (
        <Switch
          checked={record.status === 'active'}
          checkedChildren="启用"
          unCheckedChildren="禁用"
          onChange={(checked) => handleToggleStatus(record, checked)}
        />
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'createTime',
      width: 180,
      sorter: true,
      render: (_, record) =>
        record.createTime ? dayjs(record.createTime).format('YYYY-MM-DD HH:mm:ss') : '-',
    },
    {
      title: '操作',
      key: 'option',
      valueType: 'option',
      width: 220,
      fixed: 'right',
      render: (_, record) => [
        <Button
          type="link"
          key="edit"
          icon={<EditOutlined />}
          onClick={() => handleModalOpen('edit', record)}
        >
          编辑
        </Button>,
        <Button
          type="link"
          key="reset"
          icon={<KeyOutlined />}
          onClick={() => handleResetPassword(record)}
        >
          重置密码
        </Button>,
        <Button
          type="link"
          key="delete"
          danger
          icon={<DeleteOutlined />}
          onClick={() => handleDelete(record)}
        >
          删除
        </Button>,
      ],
    },
  ];

  return (
    <>
      <ProTable<User>
        columns={columns}
        actionRef={actionRef}
        cardBordered
        rowKey="id"
        search={{
          labelWidth: 'auto',
          defaultCollapsed: false,
        }}
        dateFormatter="string"
        headerTitle="用户管理"
        toolBarRender={() => [
          <Button
            key="reload"
            icon={<ReloadOutlined />}
            onClick={() => actionRef.current?.reload()}
          >
            刷新
          </Button>,
          <Button
            key="create"
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => handleModalOpen('create')}
          >
            新增用户
          </Button>,
        ]}
        request={async (params, sort, filter) => {
          try {
            const res = await userService.getList({
              pageNum: params.current,
              pageSize: params.pageSize,
              keyword: params.keyword,
              role: params.role,
              status: params.status,
            });
            const data = res.data || res;
            return {
              data: data.list || [],
              success: true,
              total: data.total || 0,
            };
          } catch (error) {
            return {
              data: [],
              success: false,
              total: 0,
            };
          }
        }}
        columnsState={{
          persistenceKey: 'user-list-columns',
          persistenceType: 'localStorage',
        }}
        pagination={{
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total) => `共 ${total} 条记录`,
        }}
        scroll={{ x: 1400 }}
      />

      <ModalForm
        title={editMode === 'create' ? '新增用户' : '编辑用户'}
        open={modalOpen}
        onOpenChange={setModalOpen}
        form={form}
        layout="vertical"
        autoFocusFirstInput
        modalProps={{
          destroyOnClose: true,
          maskClosable: false,
        }}
        submitTimeout={2000}
        onFinish={handleModalFinish}
        submitter={{
          submitButtonProps: {
            loading,
          },
          resetButtonProps: {
            onClick: () => {
              form.resetFields();
            },
          },
        }}
      >
        <Form.Item
          label="用户名"
          name="username"
          rules={[
            { required: true, message: '请输入用户名' },
            { min: 2, max: 32, message: '用户名长度为 2-32 个字符' },
          ]}
        >
          <Input placeholder="请输入用户名" disabled={editMode === 'edit'} />
        </Form.Item>
        {editMode === 'create' && (
          <Form.Item
            label="初始密码"
            name="password"
            rules={[
              { required: true, message: '请输入初始密码' },
              { min: 6, max: 32, message: '密码长度为 6-32 个字符' },
            ]}
          >
            <Input.Password placeholder="请输入初始密码" />
          </Form.Item>
        )}
        <Form.Item
          label="真实姓名"
          name="realName"
          rules={[{ required: true, message: '请输入真实姓名' }]}
        >
          <Input placeholder="请输入真实姓名" />
        </Form.Item>
        <Form.Item
          label="角色"
          name="role"
          rules={[{ required: true, message: '请选择角色' }]}
        >
          <Select
            placeholder="请选择角色"
            options={Object.entries(roleTextMap).map(([value, label]) => ({ value, label }))}
          />
        </Form.Item>
        <Form.Item
          label="联系电话"
          name="phone"
          rules={[
            {
              pattern: /^1[3-9]\d{9}$/,
              message: '请输入正确的手机号码',
            },
          ]}
        >
          <Input placeholder="请输入联系电话" />
        </Form.Item>
        <Form.Item
          label="邮箱"
          name="email"
          rules={[
            {
              type: 'email',
              message: '请输入正确的邮箱地址',
            },
          ]}
        >
          <Input placeholder="请输入邮箱" />
        </Form.Item>
        <Form.Item
          label="所属组织"
          name="orgId"
        >
          <Select
            placeholder="请选择所属组织"
            allowClear
            options={[
              { value: 'org_001', label: '综治中心' },
              { value: 'org_002', label: '东街街道办事处' },
              { value: 'org_003', label: '西街街道办事处' },
              { value: 'org_004', label: '南区调解委员会' },
              { value: 'org_005', label: '北区调解委员会' },
            ]}
          />
        </Form.Item>
        {editMode === 'edit' && (
          <Form.Item
            label="状态"
            name="status"
            initialValue="active"
          >
            <Select
              placeholder="请选择状态"
              options={[
                { value: 'active', label: '启用' },
                { value: 'disabled', label: '禁用' },
              ]}
            />
          </Form.Item>
        )}
      </ModalForm>

      <ModalForm
        title="重置密码"
        open={resetModalOpen}
        onOpenChange={setResetModalOpen}
        form={resetForm}
        layout="vertical"
        modalProps={{
          destroyOnClose: true,
          maskClosable: false,
        }}
        onFinish={handleResetPasswordFinish}
        submitter={{
          submitButtonProps: {
            loading,
          },
        }}
      >
        <Form.Item
          label="新密码"
          name="newPassword"
          rules={[
            { required: true, message: '请输入新密码' },
            { min: 6, max: 32, message: '密码长度为 6-32 个字符' },
          ]}
        >
          <Input.Password placeholder="请输入新密码" />
        </Form.Item>
        <Form.Item
          label="确认密码"
          name="confirmPassword"
          dependencies={['newPassword']}
          rules={[
            { required: true, message: '请再次输入新密码' },
            ({ getFieldValue }) => ({
              validator(_, value) {
                if (!value || getFieldValue('newPassword') === value) {
                  return Promise.resolve();
                }
                return Promise.reject(new Error('两次输入的密码不一致'));
              },
            }),
          ]}
        >
          <Input.Password placeholder="请再次输入新密码" />
        </Form.Item>
      </ModalForm>
    </>
  );
};

export default UserList;
