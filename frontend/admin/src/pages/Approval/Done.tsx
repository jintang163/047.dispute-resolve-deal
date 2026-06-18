import React, { useRef } from 'react';
import { Button, Tag, Space } from 'antd';
import {
  EyeOutlined,
  ArrowRightOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons';
import {
  ProTable,
  ProFormSelect,
  ProFormDateRangePicker,
  ProFormText,
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

const statusColorMap: Record<string, string> = {
  approved: 'green',
  rejected: 'red',
};

const statusTextMap: Record<string, string> = {
  approved: '已通过',
  rejected: '已驳回',
};

const DoneApproval: React.FC = () => {
  const navigate = useNavigate();
  const actionRef = useRef<ActionType>();

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
      title: '审批结果',
      dataIndex: 'status',
      width: 120,
      render: (_, record) => (
        <Tag icon={record.status === 'approved' ? <CheckCircleOutlined /> : <CloseCircleOutlined />} color={statusColorMap[record.status || ''] || 'default'}>
          {record.statusName || statusTextMap[record.status || ''] || '-'}
        </Tag>
      ),
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
      title: '审批完成时间',
      dataIndex: 'updateTime',
      width: 180,
      sorter: true,
      render: (_, record) => (record as any).updateTime ? dayjs((record as any).updateTime).format('YYYY-MM-DD HH:mm:ss') : '-',
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
      title: '操作',
      key: 'option',
      valueType: 'option',
      width: 180,
      fixed: 'right',
      render: (_, record) => [
        <Button
          type="link"
          key="view"
          icon={<EyeOutlined />}
          onClick={() => {
            console.log('查看审批详情:', record);
          }}
        >
          详情
        </Button>,
        <Button
          type="link"
          key="case"
          icon={<ArrowRightOutlined />}
          onClick={() => navigate(`/dispute/${record.caseId}`)}
        >
          案件
        </Button>,
      ],
    },
  ];

  return (
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
      headerTitle="已审批列表"
      request={async (params, sort, filter) => {
        try {
          const startDate = (params as any).submitTime?.[0];
          const endDate = (params as any).submitTime?.[1];
          const res = await approvalService.getDoneList({
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
        persistenceKey: 'done-approval-columns',
        persistenceType: 'localStorage',
      }}
      pagination={{
        showSizeChanger: true,
        showQuickJumper: true,
        showTotal: (total) => `共 ${total} 条记录`,
      }}
      scroll={{ x: 1500 }}
    />
  );
};

export default DoneApproval;
