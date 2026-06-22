import React, { useState, useRef, useEffect } from 'react';
import { Button, Tag, Space, App, Modal, Form, Input, Select, DatePicker, InputNumber, Switch } from 'antd';
import {
  PlusOutlined,
  EyeOutlined,
  EditOutlined,
  DeleteOutlined,
  ExclamationCircleOutlined,
  SendOutlined,
  StopOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import {
  ProTable,
  ProFormSelect,
  ProFormDateRangePicker,
  ProFormText,
  ModalForm,
  ProForm,
  ProFormTextArea,
  ProFormRadio,
  ProFormDatePicker,
} from '@ant-design/pro-components';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { useNavigate } from 'react-router-dom';
import { patrolService, PatrolTask, GridMember, CreatePatrolTaskParams } from '../../services/patrol';
import dayjs from 'dayjs';

const { confirm } = Modal;
const { TextArea } = Input;
const { Option } = Select;

const STATUS_MAP: Record<string, string> = {
  pending: '待下发',
  assigned: '已下发',
  in_progress: '进行中',
  completed: '已完成',
  cancelled: '已取消',
};

const STATUS_COLOR: Record<string, string> = {
  pending: 'default',
  assigned: 'blue',
  in_progress: 'processing',
  completed: 'success',
  cancelled: 'default',
};

const PRIORITY_MAP: Record<string, string> = {
  low: '低',
  medium: '中',
  high: '高',
  urgent: '紧急',
};

const PRIORITY_COLOR: Record<string, string> = {
  low: 'default',
  medium: 'blue',
  high: 'orange',
  urgent: 'red',
};

const TYPE_MAP: Record<string, string> = {
  routine: '日常排查',
  key_area: '重点区域排查',
  special: '专项排查',
  emergency: '应急排查',
  complaint: '投诉核查',
};

const TaskList: React.FC = () => {
  const navigate = useNavigate();
  const { message, modal } = App.useApp();
  const actionRef = useRef<ActionType>();
  const [assignModalVisible, setAssignModalVisible] = useState(false);
  const [editModalVisible, setEditModalVisible] = useState(false);
  const [editMode, setEditMode] = useState<'create' | 'edit'>('create');
  const [currentTask, setCurrentTask] = useState<PatrolTask | null>(null);
  const [assignForm] = Form.useForm<{ gridMemberId: string; remark?: string }>();
  const [taskForm] = Form.useForm<CreatePatrolTaskParams>();
  const [loading, setLoading] = useState(false);
  const [taskTypes, setTaskTypes] = useState<{ code: string; name: string }[]>([]);
  const [memberOptions, setMemberOptions] = useState<GridMember[]>([]);
  const [areaOptions, setAreaOptions] = useState<{ id: string; name: string }[]>([]);

  useEffect(() => {
    loadOptions();
  }, []);

  const loadOptions = async () => {
    try {
      const [typesRes, membersRes, areasRes] = await Promise.all([
        patrolService.getTaskTypes(),
        patrolService.getMemberOptions(),
        patrolService.getAreaOptions(),
      ]);
      const typesData: any = (typesRes as any)?.data ?? typesRes;
      if (Array.isArray(typesData)) setTaskTypes(typesData);

      const membersData: any = (membersRes as any)?.data ?? membersRes;
      if (Array.isArray(membersData)) setMemberOptions(membersData);

      const areasData: any = (areasRes as any)?.data ?? areasRes;
      if (Array.isArray(areasData)) setAreaOptions(areasData);
    } catch (error) {
      console.error('加载选项失败:', error);
    }
  };

  const handleOpenAssign = (record: PatrolTask) => {
    setCurrentTask(record);
    assignForm.resetFields();
    setAssignModalVisible(true);
  };

  const handleAssignFinish = async (values: { gridMemberId: string; remark?: string }) => {
    if (!currentTask) return false;
    try {
      setLoading(true);
      await patrolService.assignTask(currentTask.id, values.gridMemberId, values.remark);
      message.success('任务下发成功');
      actionRef.current?.reload();
      setAssignModalVisible(false);
      return true;
    } catch (error: any) {
      message.error(error.message || '下发失败');
      return false;
    } finally {
      setLoading(false);
    }
  };

  const handleOpenEdit = (mode: 'create' | 'edit', record?: PatrolTask) => {
    setEditMode(mode);
    setCurrentTask(record || null);
    if (mode === 'edit' && record) {
      taskForm.setFieldsValue({
        title: record.title,
        type: record.type,
        priority: record.priority,
        description: record.description,
        area: record.area,
        areaId: record.areaId,
        gridMemberId: record.gridMemberId,
        deadline: record.deadline ? dayjs(record.deadline) : undefined,
        requirement: record.requirement,
        remark: record.remark,
        startTime: record.startTime ? dayjs(record.startTime) : undefined,
        endTime: record.endTime ? dayjs(record.endTime) : undefined,
      });
    } else {
      taskForm.resetFields();
    }
    setEditModalVisible(true);
  };

  const handleEditFinish = async (values: CreatePatrolTaskParams) => {
    try {
      setLoading(true);
      const params = {
        ...values,
        deadline: values.deadline ? dayjs(values.deadline as any).format('YYYY-MM-DD HH:mm:ss') : undefined,
        startTime: values.startTime ? dayjs(values.startTime as any).format('YYYY-MM-DD HH:mm:ss') : undefined,
        endTime: values.endTime ? dayjs(values.endTime as any).format('YYYY-MM-DD HH:mm:ss') : undefined,
      };
      if (editMode === 'create') {
        await patrolService.createTask(params);
        message.success('任务创建成功');
      } else if (editMode === 'edit' && currentTask) {
        await patrolService.updateTask(currentTask.id, params);
        message.success('任务更新成功');
      }
      actionRef.current?.reload();
      setEditModalVisible(false);
      return true;
    } catch (error: any) {
      message.error(error.message || '操作失败');
      return false;
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = (record: PatrolTask) => {
    confirm({
      title: '确认删除该任务?',
      icon: <ExclamationCircleOutlined />,
      content: `任务: ${record.title} (${record.taskNo})`,
      okText: '确认删除',
      cancelText: '取消',
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          await patrolService.deleteTask(record.id);
          message.success('删除成功');
          actionRef.current?.reload();
        } catch (error: any) {
          message.error(error.message || '删除失败');
        }
      },
    });
  };

  const handleCancel = (record: PatrolTask) => {
    modal.confirm({
      title: '确认取消该任务?',
      icon: <StopOutlined />,
      content: `任务: ${record.title}`,
      okText: '确认取消',
      cancelText: '返回',
      onOk: async () => {
        try {
          await patrolService.cancelTask(record.id);
          message.success('任务已取消');
          actionRef.current?.reload();
        } catch (error: any) {
          message.error(error.message || '取消失败');
        }
      },
    });
  };

  const columns: ProColumns<PatrolTask>[] = [
    {
      title: '任务编号',
      dataIndex: 'taskNo',
      width: 160,
      copyable: true,
      fixed: 'left',
    },
    {
      title: '任务标题',
      dataIndex: 'title',
      width: 200,
      ellipsis: true,
    },
    {
      title: '任务类型',
      dataIndex: 'type',
      width: 120,
      render: (_, row) => {
        const typeName = row.typeName || TYPE_MAP[row.type] || row.type;
        return <Tag color="blue">{typeName}</Tag>;
      },
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      width: 100,
      render: (_, row) => {
        const priorityName = row.priorityName || PRIORITY_MAP[row.priority] || row.priority;
        return <Tag color={PRIORITY_COLOR[row.priority] || 'default'}>{priorityName}</Tag>;
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (_, row) => {
        const statusName = row.statusName || STATUS_MAP[row.status] || row.status;
        return <Tag color={STATUS_COLOR[row.status] || 'default'}>{statusName}</Tag>;
      },
    },
    {
      title: '所属区域',
      dataIndex: 'area',
      width: 140,
      ellipsis: true,
    },
    {
      title: '网格员',
      dataIndex: 'gridMemberName',
      width: 120,
      render: (_, row) => row.gridMemberName || '-',
    },
    {
      title: '截止时间',
      dataIndex: 'deadline',
      width: 160,
      render: (_, row) => (row.deadline ? dayjs(row.deadline).format('YYYY-MM-DD HH:mm') : '-'),
    },
    {
      title: '创建时间',
      dataIndex: 'createTime',
      width: 160,
      sorter: true,
      render: (_, row) => (row.createTime ? dayjs(row.createTime).format('YYYY-MM-DD HH:mm:ss') : '-'),
    },
    {
      title: '操作',
      key: 'option',
      valueType: 'option',
      width: 260,
      fixed: 'right',
      render: (_, record) => {
        const actions = [];
        actions.push(
          <Button type="link" key="view" icon={<EyeOutlined />} onClick={() => navigate(`/patrol/task/${record.id}`)}>
            查看
          </Button>,
        );
        if (record.status === 'pending') {
          actions.push(
            <Button type="link" key="assign" icon={<SendOutlined />} onClick={() => handleOpenAssign(record)}>
              下发
            </Button>,
          );
        }
        actions.push(
          <Button type="link" key="edit" icon={<EditOutlined />} onClick={() => handleOpenEdit('edit', record)}>
            编辑
          </Button>,
        );
        if (record.status === 'pending') {
          actions.push(
            <Button
              type="link"
              key="delete"
              danger
              icon={<DeleteOutlined />}
              onClick={() => handleDelete(record)}
            >
              删除
            </Button>,
          );
        }
        if (record.status !== 'completed' && record.status !== 'cancelled') {
          actions.push(
            <Button
              type="link"
              key="cancel"
              danger
              icon={<StopOutlined />}
              onClick={() => handleCancel(record)}
            >
              取消
            </Button>,
          );
        }
        return actions;
      },
    },
  ];

  return (
    <>
      <ProTable<PatrolTask>
        columns={columns}
        actionRef={actionRef}
        cardBordered
        rowKey="id"
        search={{
          labelWidth: 'auto',
          defaultCollapsed: false,
        }}
        dateFormatter="string"
        headerTitle="排查任务列表"
        toolBarRender={() => [
          <Button key="reload" icon={<ReloadOutlined />} onClick={() => actionRef.current?.reload()}>
            刷新
          </Button>,
          <Button key="create" type="primary" icon={<PlusOutlined />} onClick={() => handleOpenEdit('create')}>
            新建任务
          </Button>,
        ]}
        request={async (params, sort, filter) => {
          try {
            const startDate = (params as any).createTime?.[0];
            const endDate = (params as any).createTime?.[1];
            const res = await patrolService.getTaskList({
              pageNum: params.current,
              pageSize: params.pageSize,
              keyword: params.keyword as string,
              type: (params.type as string) || undefined,
              status: (params.status as string) || undefined,
              priority: (params.priority as string) || undefined,
              gridMemberId: (params.gridMemberId as string) || undefined,
              startDate,
              endDate,
            });
            const data: any = (res as any)?.data ?? res;
            return {
              data: data.list || [],
              success: true,
              total: data.total || 0,
            };
          } catch (error) {
            return { data: [], success: false, total: 0 };
          }
        }}
        columnsState={{
          persistenceKey: 'patrol-task-list-columns',
          persistenceType: 'localStorage',
        }}
        pagination={{
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total) => `共 ${total} 条记录`,
        }}
        scroll={{ x: 1700 }}
      />

      <ModalForm
        title="下发任务"
        open={assignModalVisible}
        onOpenChange={setAssignModalVisible}
        form={assignForm}
        layout="vertical"
        modalProps={{
          destroyOnClose: true,
          maskClosable: false,
        }}
        onFinish={handleAssignFinish}
        submitter={{
          submitButtonProps: { loading },
        }}
      >
        {currentTask && (
          <div style={{ marginBottom: 16, padding: 12, background: '#f5f5f5', borderRadius: 4 }}>
            <div style={{ fontWeight: 500 }}>{currentTask.title}</div>
            <div style={{ fontSize: 12, color: '#999', marginTop: 4 }}>
              任务编号: {currentTask.taskNo}
            </div>
          </div>
        )}
        <Form.Item
          label="选择网格员"
          name="gridMemberId"
          rules={[{ required: true, message: '请选择网格员' }]}
        >
          <Select
            placeholder="请选择网格员"
            showSearch
            optionFilterProp="label"
            options={memberOptions.map((m) => ({
              label: `${m.name} (${m.area})`,
              value: m.id,
            }))}
          />
        </Form.Item>
        <Form.Item label="备注" name="remark">
          <TextArea rows={3} placeholder="请输入备注信息" maxLength={200} showCount />
        </Form.Item>
      </ModalForm>

      <ModalForm<CreatePatrolTaskParams>
        title={editMode === 'create' ? '新建任务' : '编辑任务'}
        open={editModalVisible}
        onOpenChange={setEditModalVisible}
        form={taskForm}
        layout="vertical"
        autoFocusFirstInput
        modalProps={{
          destroyOnClose: true,
          maskClosable: false,
          width: 600,
        }}
        onFinish={handleEditFinish}
        submitter={{
          submitButtonProps: { loading },
        }}
      >
        <ProFormText
          name="title"
          label="任务标题"
          placeholder="请输入任务标题"
          rules={[
            { required: true, message: '请输入任务标题' },
            { max: 100, message: '标题长度不能超过100个字符' },
          ]}
        />
        <ProFormRadio.Group
          name="type"
          label="任务类型"
          rules={[{ required: true, message: '请选择任务类型' }]}
          options={taskTypes.length > 0
            ? taskTypes.map((t) => ({ label: t.name, value: t.code }))
            : Object.entries(TYPE_MAP).map(([value, label]) => ({ label, value }))
          }
        />
        <ProFormRadio.Group
          name="priority"
          label="优先级"
          rules={[{ required: true, message: '请选择优先级' }]}
          options={Object.entries(PRIORITY_MAP).map(([value, label]) => ({ label, value }))}
        />
        <ProFormSelect
          name="areaId"
          label="所属区域"
          placeholder="请选择所属区域"
          rules={[{ required: true, message: '请选择所属区域' }]}
          options={areaOptions.map((a) => ({ label: a.name, value: a.id }))}
        />
        <ProFormSelect
          name="gridMemberId"
          label="负责网格员"
          placeholder="请选择网格员（可选）"
          options={memberOptions.map((m) => ({
            label: `${m.name} (${m.area})`,
            value: m.id,
          }))}
        />
        <ProFormDatePicker
          name="startTime"
          label="开始时间"
          placeholder="请选择开始时间"
        />
        <ProFormDatePicker
          name="endTime"
          label="结束时间"
          placeholder="请选择结束时间"
        />
        <ProFormDatePicker
          name="deadline"
          label="截止时间"
          placeholder="请选择截止时间"
        />
        <ProFormTextArea
          name="description"
          label="任务描述"
          placeholder="请输入任务描述"
          rows={3}
        />
        <ProFormTextArea
          name="requirement"
          label="排查要求"
          placeholder="请输入排查要求"
          rows={3}
        />
        <ProFormTextArea
          name="remark"
          label="备注"
          placeholder="请输入备注信息"
          rows={2}
        />
      </ModalForm>
    </>
  );
};

export default TaskList;
