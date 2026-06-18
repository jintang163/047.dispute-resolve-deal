import React, { useRef, useState } from 'react';
import { Button, Tag, Space, App, Drawer, Form, Input, Radio, message } from 'antd';
import {
  EyeOutlined,
  CheckOutlined,
  CloseOutlined,
  SafetyCertificateOutlined,
  ArrowRightOutlined,
} from '@ant-design/icons';
import {
  ProTable,
  ProFormSelect,
  ProFormDateRangePicker,
  ProFormText,
  ProDescriptions,
} from '@ant-design/pro-components';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { useNavigate } from 'react-router-dom';
import { approvalService, ApprovalItem } from '../../services/approval';
import dayjs from 'dayjs';

const priorityColorMap: Record<string, string> = {
  low: 'default',
  normal: 'blue',
  high: 'orange',
  urgent: 'red',
};

const priorityTextMap: Record<string, string> = {
  low: '低',
  normal: '普通',
  high: '高',
  urgent: '紧急',
};

const statusTextMap: Record<string, string> = {
  pending: '待审批',
  approved: '已通过',
  rejected: '已驳回',
};

const TodoApproval: React.FC = () => {
  const navigate = useNavigate();
  const { message: appMessage, modal } = App.useApp();
  const actionRef = useRef<ActionType>();
  const [detailOpen, setDetailOpen] = useState(false);
  const [currentItem, setCurrentItem] = useState<ApprovalItem | null>(null);
  const [approveForm] = Form.useForm();
  const [approveLoading, setApproveLoading] = useState(false);

  const showApproveModal = (record: ApprovalItem, type: 'approved' | 'rejected') => {
    modal.confirm({
      title: type === 'approved' ? '审批通过' : '审批驳回',
      icon: <SafetyCertificateOutlined />,
      content: (
        <Form form={approveForm} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item label={`案件: ${record.caseNo} - ${record.caseTitle}`}>
            <div style={{ color: '#666', fontSize: 14, padding: '8px 0' }}>
              审批节点: <strong>{record.workflowNodeName}</strong>
            </div>
          </Form.Item>
          <Form.Item name="opinion" label="审批意见" rules={[{ max: 500, message: '意见长度不超过500字' }]}>
            <Input.TextArea rows={4} placeholder="请输入审批意见（选填）" maxLength={500} showCount />
          </Form.Item>
        </Form>
      ),
      okText: type === 'approved' ? '确认通过' : '确认驳回',
      okButtonProps: {
        danger: type === 'rejected' ? true : false,
        type: type === 'approved' ? 'primary' : 'default',
      },
      cancelText: '取消',
      onOk: async () => {
        try {
          setApproveLoading(true);
          const values = await approveForm.validateFields();
          await approvalService[type === 'approved' ? 'approve' : 'reject']({
            id: record.id,
            status: type,
            opinion: values.opinion,
          });
          appMessage.success(type === 'approved' ? '审批通过成功' : '审批驳回成功');
          actionRef.current?.reload();
          approveForm.resetFields();
        } catch (error: any) {
          appMessage.error(error.message || '操作失败');
          return Promise.reject();
        } finally {
          setApproveLoading(false);
        }
      },
      confirmLoading: approveLoading,
    });
  };

  const columns: ProColumns<ApprovalItem>[] = [
    {
      title: '审批编号',
      dataIndex: 'id',
      width: 160,
      copyable: true,
      fixed: 'left',
      render: (_, record) => `AP${record.id?.slice(-8)}`,
    },
    {
      title: '关联案件',
      dataIndex: 'caseNo',
      width: 200,
      render: (_, record) => (
        <Space direction="vertical" size={0}>
          <span
            style={{ color: '#1677ff', cursor: 'pointer' }}
            onClick={() => navigate(`/dispute/${record.caseId}`)}
          >
            {record.caseNo}
          </span>
          <span style={{ fontSize: 12, color: '#666' }}>{record.caseTitle}</span>
        </Space>
      ),
    },
    {
      title: '审批节点',
      dataIndex: 'workflowNodeName',
      width: 140,
    },
    {
      title: '提交人',
      dataIndex: 'submitterName',
      width: 100,
    },
    {
      title: '提交时间',
      dataIndex: 'submitTime',
      width: 180,
      sorter: true,
      render: (_, record) =>
        record.submitTime ? dayjs(record.submitTime).format('YYYY-MM-DD HH:mm:ss') : '-',
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      width: 100,
      valueEnum: priorityTextMap,
      render: (_, record) => (
        <Tag color={priorityColorMap[record.priority || ''] || 'default'}>
          {priorityTextMap[record.priority || 'normal']}
        </Tag>
      ),
    },
    {
      title: '截止时间',
      dataIndex: 'deadline',
      width: 180,
      render: (_, record) =>
        record.deadline ? dayjs(record.deadline).format('YYYY-MM-DD HH:mm') : '-',
    },
    {
      title: '操作',
      key: 'option',
      valueType: 'option',
      width: 240,
      fixed: 'right',
      render: (_, record) => [
        <Button
          type="link"
          key="view"
          icon={<EyeOutlined />}
          onClick={() => {
            setCurrentItem(record);
            setDetailOpen(true);
          }}
        >
          查看
        </Button>,
        <Button
          type="link"
          key="case"
          icon={<ArrowRightOutlined />}
          onClick={() => navigate(`/dispute/${record.caseId}`)}
        >
          案件
        </Button>,
        <Button
          type="link"
          key="approve"
          style={{ color: '#52c41a' }}
          icon={<CheckOutlined />}
          onClick={() => showApproveModal(record, 'approved')}
        >
          通过
        </Button>,
        <Button
          type="link"
          key="reject"
          danger
          icon={<CloseOutlined />}
          onClick={() => showApproveModal(record, 'rejected')}
        >
          驳回
        </Button>,
      ],
    },
  ];

  return (
    <>
      <ProTable<ApprovalItem>
        columns={columns}
        actionRef={actionRef}
        cardBordered
        rowKey="id"
        search={{
          labelWidth: 'auto',
          defaultCollapsed: false,
        }}
        dateFormatter="string"
        headerTitle="待审批列表"
        request={async (params, sort, filter) => {
          try {
            const startDate = (params as any).submitTime?.[0];
            const endDate = (params as any).submitTime?.[1];
            const res = await approvalService.getTodoList({
              pageNum: params.current,
              pageSize: params.pageSize,
              keyword: params.keyword,
              workflowNode: params.workflowNode,
              priority: params.priority,
              startDate,
              endDate,
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
          persistenceKey: 'todo-approval-columns',
          persistenceType: 'localStorage',
        }}
        pagination={{
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total) => `共 ${total} 条记录`,
        }}
        scroll={{ x: 1400 }}
      />

      <Drawer
        title="审批详情"
        width={640}
        onClose={() => setDetailOpen(false)}
        open={detailOpen}
        extra={
          <Space>
            <Button onClick={() => setDetailOpen(false)}>关闭</Button>
            {currentItem && (
              <>
                <Button danger icon={<CloseOutlined />} onClick={() => showApproveModal(currentItem, 'rejected')}>
                  驳回
                </Button>
                <Button type="primary" icon={<CheckOutlined />} onClick={() => showApproveModal(currentItem, 'approved')}>
                  通过
                </Button>
              </>
            )}
          </Space>
        }
      >
        {currentItem && (
          <ProDescriptions column={1} bordered size="small">
            <ProDescriptions.Item label="审批编号">{`AP${currentItem.id?.slice(-8)}`}</ProDescriptions.Item>
            <ProDescriptions.Item label="案件编号">{currentItem.caseNo}</ProDescriptions.Item>
            <ProDescriptions.Item label="案件标题">{currentItem.caseTitle}</ProDescriptions.Item>
            <ProDescriptions.Item label="审批节点">{currentItem.workflowNodeName}</ProDescriptions.Item>
            <ProDescriptions.Item label="提交人">{currentItem.submitterName}</ProDescriptions.Item>
            <ProDescriptions.Item label="提交时间">
              {currentItem.submitTime ? dayjs(currentItem.submitTime).format('YYYY-MM-DD HH:mm:ss') : '-'}
            </ProDescriptions.Item>
            <ProDescriptions.Item label="优先级">
              <Tag color={priorityColorMap[currentItem.priority || ''] || 'default'}>
                {priorityTextMap[currentItem.priority || 'normal']}
              </Tag>
            </ProDescriptions.Item>
            <ProDescriptions.Item label="截止时间">
              {currentItem.deadline ? dayjs(currentItem.deadline).format('YYYY-MM-DD HH:mm') : '-'}
            </ProDescriptions.Item>
          </ProDescriptions>
        )}
      </Drawer>
    </>
  );
};

export default TodoApproval;
