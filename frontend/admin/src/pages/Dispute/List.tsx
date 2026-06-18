import React, { useState, useRef } from 'react';
import { Button, Tag, Space, App, Modal } from 'antd';
import { PlusOutlined, EyeOutlined, EditOutlined, DeleteOutlined, ExclamationCircleOutlined } from '@ant-design/icons';
import {
  ProTable,
  ProFormSelect,
  ProFormDateRangePicker,
  ProFormText,
} from '@ant-design/pro-components';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { useNavigate } from 'react-router-dom';
import { disputeService, DisputeCase } from '../../services/dispute';
import dayjs from 'dayjs';

const { confirm } = Modal;

const statusColorMap: Record<string, string> = {
  pending: 'default',
  assigning: 'processing',
  mediating: 'blue',
  pending_approval: 'orange',
  approved: 'green',
  rejected: 'red',
  completed: 'success',
  closed: 'default',
};

const statusTextMap: Record<string, string> = {
  pending: '待分配',
  assigning: '分配中',
  mediating: '调解中',
  pending_approval: '待审批',
  approved: '审批通过',
  rejected: '审批驳回',
  completed: '已完成',
  closed: '已结案',
};

const typeTextMap: Record<string, string> = {
  civil: '民事纠纷',
  labor: '劳动争议',
  family: '家庭纠纷',
  neighborhood: '邻里纠纷',
  contract: '合同纠纷',
  property: '物业纠纷',
  other: '其他纠纷',
};

const DisputeList: React.FC = () => {
  const navigate = useNavigate();
  const { message } = App.useApp();
  const actionRef = useRef<ActionType>();
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);

  const columns: ProColumns<DisputeCase>[] = [
    {
      title: '案件编号',
      dataIndex: 'caseNo',
      width: 160,
      copyable: true,
      fixed: 'left',
    },
    {
      title: '案件标题',
      dataIndex: 'title',
      width: 200,
      ellipsis: true,
    },
    {
      title: '案件类型',
      dataIndex: 'type',
      width: 120,
      valueEnum: typeTextMap,
      render: (_, entity) => {
        return <Tag color="blue">{typeTextMap[entity.type] || entity.type}</Tag>;
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      valueEnum: statusTextMap,
      render: (_, entity) => {
        return (
          <Tag color={statusColorMap[entity.status] || 'default'}>
            {entity.statusName || statusTextMap[entity.status] || entity.status}
          </Tag>
        );
      },
    },
    {
      title: '甲方',
      dataIndex: 'partyA',
      width: 120,
      ellipsis: true,
    },
    {
      title: '乙方',
      dataIndex: 'partyB',
      width: 120,
      ellipsis: true,
    },
    {
      title: '所属组织',
      dataIndex: 'orgName',
      width: 160,
      ellipsis: true,
    },
    {
      title: '调解员',
      dataIndex: 'mediatorName',
      width: 120,
      ellipsis: true,
    },
    {
      title: '创建时间',
      dataIndex: 'createTime',
      width: 180,
      valueType: 'dateTime',
      sorter: true,
      render: (_, entity) => dayjs(entity.createTime).format('YYYY-MM-DD HH:mm:ss'),
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
          onClick={() => navigate(`/dispute/${record.id}`)}
        >
          查看
        </Button>,
        <Button
          type="link"
          key="edit"
          icon={<EditOutlined />}
          onClick={() => {
            message.info('编辑功能开发中');
          }}
        >
          编辑
        </Button>,
        <Button
          type="link"
          key="delete"
          danger
          icon={<DeleteOutlined />}
          onClick={() => {
            confirm({
              title: '确认删除该案件?',
              icon: <ExclamationCircleOutlined />,
              content: '删除后将无法恢复，请谨慎操作。',
              okText: '确认删除',
              cancelText: '取消',
              okButtonProps: { danger: true },
              onOk: async () => {
                try {
                  await disputeService.delete(record.id);
                  message.success('删除成功');
                  actionRef.current?.reload();
                } catch (error: any) {
                  message.error(error.message || '删除失败');
                }
              },
            });
          }}
        >
          删除
        </Button>,
      ],
    },
  ];

  return (
    <ProTable<DisputeCase>
      columns={columns}
      actionRef={actionRef}
      cardBordered
      rowKey="id"
      search={{
        labelWidth: 'auto',
        defaultCollapsed: false,
      }}
      rowSelection={{
        onChange: (keys) => {
          setSelectedRowKeys(keys);
        },
      }}
      dateFormatter="string"
      headerTitle="纠纷案件列表"
      toolBarRender={() => [
        <Button
          key="create"
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => navigate('/dispute/create')}
        >
          新增案件
        </Button>,
      ]}
      request={async (params, sort, filter) => {
        try {
          const startDate = (params as any).createTime?.[0];
          const endDate = (params as any).createTime?.[1];
          const res = await disputeService.getList({
            pageNum: params.current,
            pageSize: params.pageSize,
            keyword: params.keyword,
            type: params.type,
            status: params.status,
            startDate,
            endDate,
            ...filter,
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
        persistenceKey: 'dispute-list-columns',
        persistenceType: 'localStorage',
      }}
      pagination={{
        showSizeChanger: true,
        showQuickJumper: true,
        showTotal: (total) => `共 ${total} 条记录`,
      }}
      scroll={{ x: 1400 }}
    />
  );
};

export default DisputeList;
