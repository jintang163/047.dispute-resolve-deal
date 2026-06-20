import React, { useRef, useState } from 'react';
import { Button, Tag, Space, App, Drawer } from 'antd';
import {
  EyeOutlined,
  FileTextOutlined,
  PlusOutlined,
  RobotOutlined,
} from '@ant-design/icons';
import {
  ProTable,
  ProFormSelect,
  ProFormDateRangePicker,
  ProFormText,
} from '@ant-design/pro-components';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { useNavigate } from 'react-router-dom';
import { mediationService, MediationRecord } from '../../services/user';
import ProtocolGenerator from '../Mediation/ProtocolGenerator';
import dayjs from 'dayjs';

const resultColorMap: Record<string, string> = {
  success: 'green',
  partial: 'blue',
  failed: 'red',
  pending: 'default',
};

const resultTextMap: Record<string, string> = {
  success: '调解成功',
  partial: '部分达成',
  failed: '调解失败',
  pending: '待调解',
};

const MediationList: React.FC = () => {
  const navigate = useNavigate();
  const { message } = App.useApp();
  const actionRef = useRef<ActionType>();
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [protocolDrawerOpen, setProtocolDrawerOpen] = useState(false);
  const [protocolCaseId, setProtocolCaseId] = useState<string>('');

  const columns: ProColumns<MediationRecord>[] = [
    {
      title: '调解编号',
      dataIndex: 'id',
      width: 160,
      copyable: true,
      fixed: 'left',
      render: (_, record) => `MT${record.id?.slice(-8)}`,
    },
    {
      title: '关联案件',
      dataIndex: 'caseNo',
      width: 160,
      render: (_, record) => (
        <Space direction="vertical" size={0}>
          <span style={{ color: '#1677ff', cursor: 'pointer' }} onClick={() => navigate(`/dispute/${record.caseId}`)}>
            {record.caseNo}
          </span>
          <span style={{ fontSize: 12, color: '#666' }}>{record.caseTitle}</span>
        </Space>
      ),
    },
    {
      title: '调解员',
      dataIndex: 'mediatorName',
      width: 120,
    },
    {
      title: '调解时间',
      dataIndex: 'mediationTime',
      width: 180,
      sorter: true,
      render: (_, record) =>
        record.mediationTime ? dayjs(record.mediationTime).format('YYYY-MM-DD HH:mm') : '-',
    },
    {
      title: '调解地点',
      dataIndex: 'place',
      width: 160,
      ellipsis: true,
    },
    {
      title: '调解时长',
      dataIndex: 'duration',
      width: 100,
      render: (_, record) => (record.duration ? `${record.duration} 分钟` : '-'),
    },
    {
      title: '调解结果',
      dataIndex: 'result',
      width: 120,
      valueEnum: resultTextMap,
      render: (_, record) => (
        <Tag color={resultColorMap[record.result || ''] || 'default'}>
          {record.resultName || resultTextMap[record.result || ''] || '-'}
        </Tag>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'createTime',
      width: 180,
      valueType: 'dateTime',
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
          key="view"
          icon={<EyeOutlined />}
          onClick={() => {
            message.info('查看调解详情');
          }}
        >
          详情
        </Button>,
        <Button
          type="link"
          key="protocol"
          icon={<FileTextOutlined />}
          onClick={() => {
            setProtocolCaseId(String(record.caseId));
            setProtocolDrawerOpen(true);
          }}
        >
          协议
        </Button>,
      ],
    },
  ];

  return (
    <>
    <ProTable<MediationRecord>
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
      headerTitle="调解记录列表"
      toolBarRender={() => [
        <Button
          key="create"
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => {
            message.info('新增调解记录');
          }}
        >
          新增调解记录
        </Button>,
      ]}
      request={async (params, sort, filter) => {
        try {
          const startDate = (params as any).mediationTime?.[0];
          const endDate = (params as any).mediationTime?.[1];
          const res = await mediationService.getList({
            pageNum: params.current,
            pageSize: params.pageSize,
            keyword: params.keyword,
            result: params.result,
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
        persistenceKey: 'mediation-list-columns',
        persistenceType: 'localStorage',
      }}
      pagination={{
        showSizeChanger: true,
        showQuickJumper: true,
        showTotal: (total) => `共 ${total} 条记录`,
      }}
      scroll={{ x: 1300 }}
    />

      <Drawer
        title={
          <Space>
            <RobotOutlined style={{ color: '#533483' }} />
            <span>AI智能生成调解协议</span>
          </Space>
        }
        width={1400}
        open={protocolDrawerOpen}
        onClose={() => setProtocolDrawerOpen(false)}
        destroyOnClose
      >
        {protocolCaseId && (
          <ProtocolGenerator
            caseId={protocolCaseId}
            onClose={() => setProtocolDrawerOpen(false)}
          />
        )}
      </Drawer>
    </>
  );
};

export default MediationList;
