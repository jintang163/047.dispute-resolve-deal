import React, { useRef, useState, useEffect } from 'react';
import { Button, Tag, Space, Switch, App, Modal, Form, Input, Select, Avatar } from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  ReloadOutlined,
  UserOutlined,
  EnvironmentOutlined,
  PhoneOutlined,
} from '@ant-design/icons';
import {
  ProTable,
  ProFormSelect,
  ProFormText,
  ModalForm,
  ProFormRadio,
} from '@ant-design/pro-components';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { patrolService, GridMember, CreateGridMemberParams } from '../../services/patrol';
import dayjs from 'dayjs';

const genderColorMap: Record<string, string> = {
  male: 'blue',
  female: 'pink',
  unknown: 'default',
};

const genderTextMap: Record<string, string> = {
  male: '男',
  female: '女',
  unknown: '未知',
};

const statusTextMap: Record<string, string> = {
  active: '启用',
  disabled: '禁用',
};

const GridMemberList: React.FC = () => {
  const { message, modal } = App.useApp();
  const actionRef = useRef<ActionType>();
  const [modalOpen, setModalOpen] = useState(false);
  const [editMode, setEditMode] = useState<'create' | 'edit'>('create');
  const [currentRecord, setCurrentRecord] = useState<GridMember | null>(null);
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [areaOptions, setAreaOptions] = useState<{ id: string; name: string }[]>([]);

  useEffect(() => {
    loadAreaOptions();
  }, []);

  const loadAreaOptions = async () => {
    try {
      const res = await patrolService.getAreaOptions();
      const data: any = (res as any)?.data ?? res;
      if (Array.isArray(data)) {
        setAreaOptions(data);
      }
    } catch (error) {
      console.error('加载区域选项失败:', error);
    }
  };

  const handleModalOpen = (mode: 'create' | 'edit', record?: GridMember) => {
    setEditMode(mode);
    setCurrentRecord(record || null);
    if (mode === 'edit' && record) {
      form.setFieldsValue({
        name: record.name,
        gender: record.gender,
        phone: record.phone,
        idCard: record.idCard,
        area: record.area,
        areaId: record.areaId,
        address: record.address,
        email: record.email,
        status: record.status,
      });
    } else {
      form.resetFields();
    }
    setModalOpen(true);
  };

  const handleModalFinish = async (values: CreateGridMemberParams) => {
    try {
      setLoading(true);
      if (editMode === 'create') {
        await patrolService.createMember(values);
        message.success('网格员创建成功');
      } else if (editMode === 'edit' && currentRecord) {
        await patrolService.updateMember(currentRecord.id, values);
        message.success('网格员更新成功');
      }
      actionRef.current?.reload();
      setModalOpen(false);
      return true;
    } catch (error: any) {
      message.error(error.message || '操作失败');
      return false;
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = (record: GridMember) => {
    modal.confirm({
      title: '确认删除该网格员?',
      icon: <DeleteOutlined />,
      content: `网格员: ${record.name} (${record.memberNo})`,
      okText: '确认删除',
      cancelText: '取消',
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          await patrolService.deleteMember(record.id);
          message.success('删除成功');
          actionRef.current?.reload();
        } catch (error: any) {
          message.error(error.message || '删除失败');
          return Promise.reject();
        }
      },
    });
  };

  const handleToggleStatus = async (record: GridMember, checked: boolean) => {
    try {
      await patrolService.updateMember(record.id, { status: checked ? 'active' : 'disabled' });
      message.success(`网格员已${checked ? '启用' : '禁用'}`);
      actionRef.current?.reload();
    } catch (error: any) {
      message.error(error.message || '操作失败');
      actionRef.current?.reload();
    }
  };

  const columns: ProColumns<GridMember>[] = [
    {
      title: '网格员信息',
      dataIndex: 'name',
      width: 220,
      fixed: 'left',
      render: (_, record) => (
        <Space>
          <Avatar
            size={40}
            src={record.avatar}
            style={{ fontSize: 18, background: '#1677ff' }}
            icon={<UserOutlined />}
          />
          <Space direction="vertical" size={0}>
            <Space>
              <span style={{ fontWeight: 500 }}>{record.name}</span>
              <Tag color={genderColorMap[record.gender] || 'default'} size="small">
                {record.genderName || genderTextMap[record.gender] || record.gender}
              </Tag>
            </Space>
            <span style={{ fontSize: 12, color: '#999' }}>{record.memberNo}</span>
          </Space>
        </Space>
      ),
    },
    {
      title: '联系电话',
      dataIndex: 'phone',
      width: 140,
      render: (_, record) => (
        <Space>
          <PhoneOutlined style={{ color: '#1677ff' }} />
          <span>{record.phone}</span>
        </Space>
      ),
    },
    {
      title: '所属区域',
      dataIndex: 'area',
      width: 160,
      ellipsis: true,
      render: (_, record) => (
        <Space>
          <EnvironmentOutlined style={{ color: '#52c41a' }} />
          <span>{record.area || '-'}</span>
        </Space>
      ),
    },
    {
      title: '当前积分',
      dataIndex: 'points',
      width: 100,
      render: (_, record) => (
        <span style={{ color: '#faad14', fontWeight: 500 }}>{record.points || 0}</span>
      ),
    },
    {
      title: '累计积分',
      dataIndex: 'totalPoints',
      width: 100,
      render: (_, record) => <span>{record.totalPoints || 0}</span>,
    },
    {
      title: '完成任务',
      dataIndex: 'completedTaskCount',
      width: 100,
      render: (_, record) => (
        <Tag color="green">{record.completedTaskCount || 0} 个</Tag>
      ),
    },
    {
      title: '加入时间',
      dataIndex: 'joinDate',
      width: 120,
      render: (_, record) =>
        record.joinDate ? dayjs(record.joinDate).format('YYYY-MM-DD') : '-',
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
      width: 160,
      sorter: true,
      render: (_, record) =>
        record.createTime ? dayjs(record.createTime).format('YYYY-MM-DD HH:mm:ss') : '-',
    },
    {
      title: '操作',
      key: 'option',
      valueType: 'option',
      width: 160,
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
      <ProTable<GridMember>
        columns={columns}
        actionRef={actionRef}
        cardBordered
        rowKey="id"
        search={{
          labelWidth: 'auto',
          defaultCollapsed: false,
        }}
        dateFormatter="string"
        headerTitle="网格员管理"
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
            新增网格员
          </Button>,
        ]}
        request={async (params, sort, filter) => {
          try {
            const res = await patrolService.getMemberList({
              pageNum: params.current,
              pageSize: params.pageSize,
              keyword: params.keyword as string,
              status: (params.status as string) || undefined,
              areaId: (params.areaId as string) || undefined,
            });
            const data: any = (res as any)?.data ?? res;
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
          persistenceKey: 'grid-member-list-columns',
          persistenceType: 'localStorage',
        }}
        pagination={{
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total) => `共 ${total} 条记录`,
        }}
        scroll={{ x: 1500 }}
      />

      <ModalForm
        title={editMode === 'create' ? '新增网格员' : '编辑网格员'}
        open={modalOpen}
        onOpenChange={setModalOpen}
        form={form}
        layout="vertical"
        autoFocusFirstInput
        modalProps={{
          destroyOnClose: true,
          maskClosable: false,
          width: 500,
        }}
        onFinish={handleModalFinish}
        submitter={{
          submitButtonProps: {
            loading,
          },
        }}
      >
        <ProFormText
          name="name"
          label="姓名"
          placeholder="请输入姓名"
          rules={[
            { required: true, message: '请输入姓名' },
            { max: 20, message: '姓名长度不能超过20个字符' },
          ]}
        />
        <ProFormRadio.Group
          name="gender"
          label="性别"
          rules={[{ required: true, message: '请选择性别' }]}
          options={Object.entries(genderTextMap).map(([value, label]) => ({ label, value }))}
        />
        <ProFormText
          name="phone"
          label="联系电话"
          placeholder="请输入联系电话"
          rules={[
            { required: true, message: '请输入联系电话' },
            {
              pattern: /^1[3-9]\d{9}$/,
              message: '请输入正确的手机号码',
            },
          ]}
        />
        <ProFormText
          name="idCard"
          label="身份证号"
          placeholder="请输入身份证号"
          rules={[
            {
              pattern: /(^\d{15}$)|(^\d{18}$)|(^\d{17}(\d|X|x)$)/,
              message: '请输入正确的身份证号',
            },
          ]}
        />
        <ProFormSelect
          name="areaId"
          label="所属区域"
          placeholder="请选择所属区域"
          rules={[{ required: true, message: '请选择所属区域' }]}
          options={areaOptions.map((a) => ({ label: a.name, value: a.id }))}
          fieldProps={{
            showSearch: true,
            optionFilterProp: 'label',
          }}
        />
        <ProFormText
          name="address"
          label="居住地址"
          placeholder="请输入居住地址"
        />
        <ProFormText
          name="email"
          label="邮箱"
          placeholder="请输入邮箱"
          rules={[
            {
              type: 'email',
              message: '请输入正确的邮箱地址',
            },
          ]}
        />
        {editMode === 'edit' && (
          <ProFormSelect
            name="status"
            label="状态"
            placeholder="请选择状态"
            initialValue="active"
            options={Object.entries(statusTextMap).map(([value, label]) => ({ label, value }))}
          />
        )}
      </ModalForm>
    </>
  );
};

export default GridMemberList;
