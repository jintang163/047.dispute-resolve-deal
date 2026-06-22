import React, { useRef, useState, useEffect } from 'react';
import { Button, Tag, Space, App, Card, Statistic, Row, Col, Form, Select } from 'antd';
import {
  ReloadOutlined,
  ThunderboltOutlined,
  GiftOutlined,
  RiseOutlined,
  FallOutlined,
  DollarOutlined,
} from '@ant-design/icons';
import {
  ProTable,
  ProFormSelect,
  ProFormDateRangePicker,
} from '@ant-design/pro-components';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { patrolService, PointFlow, GridMember } from '../../services/patrol';
import dayjs from 'dayjs';

const typeColorMap: Record<string, string> = {
  earn: 'green',
  spend: 'red',
};

const typeTextMap: Record<string, string> = {
  earn: '获取积分',
  spend: '消耗积分',
};

const PointFlowPage: React.FC = () => {
  const { message } = App.useApp();
  const actionRef = useRef<ActionType>();
  const [memberOptions, setMemberOptions] = useState<GridMember[]>([]);
  const [stats, setStats] = useState({
    totalEarn: 0,
    totalSpend: 0,
    totalBalance: 0,
    totalCount: 0,
  });

  useEffect(() => {
    loadMemberOptions();
  }, []);

  const loadMemberOptions = async () => {
    try {
      const res = await patrolService.getMemberOptions();
      const data: any = (res as any)?.data ?? res;
      if (Array.isArray(data)) {
        setMemberOptions(data);
      }
    } catch (error) {
      console.error('加载网格员选项失败:', error);
    }
  };

  const columns: ProColumns<PointFlow>[] = [
    {
      title: '流水编号',
      dataIndex: 'flowNo',
      width: 180,
      copyable: true,
      fixed: 'left',
    },
    {
      title: '网格员',
      dataIndex: 'memberName',
      width: 120,
      render: (_, record) => record.memberName || '-',
    },
    {
      title: '类型',
      dataIndex: 'type',
      width: 120,
      render: (_, record) => {
        const typeName = record.typeName || typeTextMap[record.type] || record.type;
        return (
          <Tag color={typeColorMap[record.type] || 'default'}>
            <Space size={4}>
              {record.type === 'earn' ? (
                <RiseOutlined />
              ) : (
                <FallOutlined />
              )}
              {typeName}
            </Space>
          </Tag>
        );
      },
    },
    {
      title: '规则名称',
      dataIndex: 'ruleName',
      width: 180,
      ellipsis: true,
      render: (_, record) => record.ruleName || '-',
    },
    {
      title: '积分变动',
      dataIndex: 'points',
      width: 120,
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
      title: '当前余额',
      dataIndex: 'balance',
      width: 120,
      render: (_, record) => (
        <span style={{ fontWeight: 500, color: '#faad14' }}>
          {record.balance}
        </span>
      ),
    },
    {
      title: '关联任务',
      dataIndex: 'taskNo',
      width: 160,
      ellipsis: true,
      render: (_, record) => record.taskNo || '-',
    },
    {
      title: '关联订单',
      dataIndex: 'orderNo',
      width: 160,
      ellipsis: true,
      render: (_, record) => record.orderNo || '-',
    },
    {
      title: '备注',
      dataIndex: 'remark',
      width: 200,
      ellipsis: true,
      render: (_, record) => record.remark || '-',
    },
    {
      title: '操作人',
      dataIndex: 'operatorName',
      width: 100,
      render: (_, record) => record.operatorName || '系统',
    },
    {
      title: '创建时间',
      dataIndex: 'createTime',
      width: 160,
      sorter: true,
      render: (_, record) =>
        record.createTime ? dayjs(record.createTime).format('YYYY-MM-DD HH:mm:ss') : '-',
    },
  ];

  return (
    <>
      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={12} md={6}>
          <Card>
            <Statistic
              title="累计获取积分"
              value={stats.totalEarn}
              prefix={<ThunderboltOutlined style={{ color: '#52c41a' }} />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={12} md={6}>
          <Card>
            <Statistic
              title="累计消耗积分"
              value={stats.totalSpend}
              prefix={<GiftOutlined style={{ color: '#ff4d4f' }} />}
              valueStyle={{ color: '#ff4d4f' }}
            />
          </Card>
        </Col>
        <Col xs={12} md={6}>
          <Card>
            <Statistic
              title="当前总余额"
              value={stats.totalBalance}
              prefix={<DollarOutlined style={{ color: '#faad14' }} />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col xs={12} md={6}>
          <Card>
            <Statistic
              title="流水记录数"
              value={stats.totalCount}
              prefix={<Tag color="blue">条</Tag>}
              valueStyle={{ color: '#1677ff' }}
            />
          </Card>
        </Col>
      </Row>

      <ProTable<PointFlow>
        columns={columns}
        actionRef={actionRef}
        cardBordered
        rowKey="id"
        search={{
          labelWidth: 'auto',
          defaultCollapsed: false,
        }}
        dateFormatter="string"
        headerTitle="积分流水记录"
        toolBarRender={() => [
          <Button
            key="reload"
            icon={<ReloadOutlined />}
            onClick={() => actionRef.current?.reload()}
          >
            刷新
          </Button>,
        ]}
        request={async (params, sort, filter) => {
          try {
            const startDate = (params as any).createTime?.[0];
            const endDate = (params as any).createTime?.[1];
            const res = await patrolService.getPointFlowList({
              pageNum: params.current,
              pageSize: params.pageSize,
              memberId: (params.memberId as string) || undefined,
              memberName: (params.memberName as string) || undefined,
              type: (params.type as string) || undefined,
              startDate,
              endDate,
            });
            const data: any = (res as any)?.data ?? res;
            if (data.stats) {
              setStats(data.stats);
            }
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
          persistenceKey: 'point-flow-list-columns',
          persistenceType: 'localStorage',
        }}
        pagination={{
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total) => `共 ${total} 条记录`,
        }}
        scroll={{ x: 1700 }}
      />
    </>
  );
};

export default PointFlowPage;
