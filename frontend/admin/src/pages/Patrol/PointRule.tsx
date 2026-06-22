import React, { useRef, useState } from 'react';
import { Button, Tag, Space, Switch, App, Modal, Form, Input, InputNumber } from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  ReloadOutlined,
  ThunderboltOutlined,
  GiftOutlined,
} from '@ant-design/icons';
import {
  ProTable,
  ProFormSelect,
  ProFormText,
  ModalForm,
  ProFormRadio,
  ProFormTextArea,
  ProFormDigit,
} from '@ant-design/pro-components';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { patrolService, PointRule } from '../../services/patrol';
import dayjs from 'dayjs';

const typeColorMap: Record<string, string> = {
  earn: 'green',
  spend: 'red',
};

const typeTextMap: Record<string, string> = {
  earn: '获取积分',
  spend: '消耗积分',
};

const PointRule: React.FC = () => {
  const { message, modal } = App.useApp();
  const actionRef = useRef<ActionType>();
  const [modalOpen, setModalOpen] = useState(false);
  const [editMode, setEditMode] = useState<'create' | 'edit'>('create');
  const [currentRecord, setCurrentRecord] = useState<PointRule | null>(null);
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);

  const handleModalOpen = (mode: 'create' | 'edit', record?: PointRule) => {
    setEditMode(mode);
    setCurrentRecord(record || null);
    if (mode === 'edit' && record) {
      form.setFieldsValue({
        name: record.name,
        code: record.code,
        type: record.type,
        points: record.points,
        description: record.description,
        status: record.status,
        sort: record.sort,
      });
    } else {
      form.resetFields();
    }
    setModalOpen(true);
  };

  const handleModalFinish = async (values: Partial<PointRule>) => {
    try {
      setLoading(true);
      if (editMode === 'create') {
        await patrolService.createPointRule(values);
        message.success('积分规则创建成功');
      } else if (editMode === 'edit' && currentRecord) {
        await patrolService.updatePointRule(currentRecord.id, values);
        message.success('积分规则更新成功');
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

  const handleDelete = (record: PointRule) => {
    modal.confirm({
      title: '确认删除该积分规则?',
      icon: <DeleteOutlined />,
      content: `规则: ${record.name} (${record.code})`,
      okText: '确认删除',
      cancelText: '取消',
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          await patrolService.deletePointRule(record.id);
          message.success('删除成功');
          actionRef.current?.reload();
        } catch (error: any) {
          message.error(error.message || '删除失败');
          return Promise.reject();
        }
      },
    });
  };

  const handleToggleStatus = async (record: PointRule, checked: boolean) => {
    try {
      await patrolService.updatePointRule(record.id, { status: checked ? 'active' : 'inactive' });
      message.success(`规则已${checked ? '启用' : '禁用'}`);
      actionRef.current?.reload();
    } catch (error: any) {
      message.error(error.message || '操作失败');
      actionRef.current?.reload();
    }
  };

  const columns: ProColumns<PointRule>[] = [
    {
      title: '规则名称',
      dataIndex: 'name',
      width: 200,
      ellipsis: true,
      render: (_, record) => (
        <Space>
          {record.type === 'earn' ? (
            <ThunderboltOutlined style={{ color: '#52c41a' }} />
          ) : (
            <GiftOutlined style={{ color: '#ff4d4f' }} />
          )}
          <span style={{ fontWeight: 500 }}>{record.name}</span>
        </Space>
      ),
    },
    {
      title: '规则编码',
      dataIndex: 'code',
      width: 160,
      copyable: true,
    },
    {
      title: '类型',
      dataIndex: 'type',
      width: 120,
      render: (_, record) => {
        const typeName = record.typeName || typeTextMap[record.type] || record.type;
        return (
          <Tag color={typeColorMap[record.type] || 'default'}>
            {typeName}
          </Tag>
        );
      },
    },
    {
      title: '积分值',
      dataIndex: 'points',
      width: 100,
      render: (_, record) => (
        <span
          style={{
            color: record.type === 'earn' ? '#52c41a' : '#ff4d4f',
            fontWeight: 500,
            fontSize: 16,
          }}
        >
          {record.type === 'earn' ? '+' : '-'}{record.points}
        </span>
      ),
    },
    {
      title: '描述',
      dataIndex: 'description',
      width: 280,
      ellipsis: true,
      render: (_, record) => record.description || '-',
    },
    {
      title: '排序',
      dataIndex: 'sort',
      width: 80,
      sorter: true,
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
      <ProTable<PointRule>
        columns={columns}
        actionRef={actionRef}
        cardBordered
        rowKey="id"
        search={{
          labelWidth: 'auto',
          defaultCollapsed: false,
        }}
        dateFormatter="string"
        headerTitle="积分规则配置"
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
            新增规则
          </Button>,
        ]}
        request={async (params, sort, filter) => {
          try {
            const res = await patrolService.getPointRuleList({
              pageNum: params.current,
              pageSize: params.pageSize,
              keyword: params.keyword as string,
              type: (params.type as string) || undefined,
              status: (params.status as string) || undefined,
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
          persistenceKey: 'point-rule-list-columns',
          persistenceType: 'localStorage',
        }}
        pagination={{
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total) => `共 ${total} 条记录`,
        }}
        scroll={{ x: 1300 }}
      />

      <ModalForm
        title={editMode === 'create' ? '新增积分规则' : '编辑积分规则'}
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
          label="规则名称"
          placeholder="请输入规则名称"
          rules={[
            { required: true, message: '请输入规则名称' },
            { max: 50, message: '规则名称长度不能超过50个字符' },
          ]}
        />
        <ProFormText
          name="code"
          label="规则编码"
          placeholder="请输入规则编码，如：PATROL_COMPLETE"
          rules={[
            { required: true, message: '请输入规则编码' },
            { max: 50, message: '规则编码长度不能超过50个字符' },
          ]}
        />
        <ProFormRadio.Group
          name="type"
          label="类型"
          rules={[{ required: true, message: '请选择类型' }]}
          options={Object.entries(typeTextMap).map(([value, label]) => ({ label, value }))}
        />
        <ProFormDigit
          name="points"
          label="积分值"
          placeholder="请输入积分值"
          min={1}
          max={10000}
          rules={[{ required: true, message: '请输入积分值' }]}
          fieldProps={{
            precision: 0,
            addonBefore: 'P',
          }}
        />
        <ProFormDigit
          name="sort"
          label="排序"
          placeholder="请输入排序号，数字越小越靠前"
          min={0}
          max={999}
          initialValue={0}
        />
        <ProFormTextArea
          name="description"
          label="描述"
          placeholder="请输入规则描述"
          rows={3}
        />
        {editMode === 'edit' && (
          <ProFormSelect
            name="status"
            label="状态"
            placeholder="请选择状态"
            initialValue="active"
            options={[
              { label: '启用', value: 'active' },
              { label: '禁用', value: 'inactive' },
            ]}
          />
        )}
      </ModalForm>
    </>
  );
};

export default PointRule;
