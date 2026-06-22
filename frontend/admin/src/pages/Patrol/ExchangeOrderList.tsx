import React, { useRef, useState } from 'react';
import {
  Button,
  Tag,
  Space,
  App,
  Modal,
  Form,
  Input,
  Select,
  Image,
  Card,
  Row,
  Col,
  Statistic,
  Descriptions,
  Typography,
} from 'antd';
const { Text } = Typography;
import {
  ReloadOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  TruckOutlined,
  EyeOutlined,
  UserOutlined,
  PhoneOutlined,
  EnvironmentOutlined,
  ShoppingCartOutlined,
  ClockCircleOutlined,
  StopOutlined,
  CheckOutlined,
} from '@ant-design/icons';
import {
  ProTable,
  ProFormSelect,
  ProFormDateRangePicker,
  ProFormText,
  ModalForm,
  ProFormTextArea,
} from '@ant-design/pro-components';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { patrolService, ExchangeOrder } from '../../services/patrol';
import dayjs from 'dayjs';

const STATUS_MAP: Record<string, string> = {
  pending: '待审核',
  approved: '已通过',
  rejected: '已拒绝',
  delivering: '已发货',
  completed: '已完成',
  cancelled: '已取消',
};

const STATUS_COLOR: Record<string, string> = {
  pending: 'orange',
  approved: 'blue',
  rejected: 'red',
  delivering: 'processing',
  completed: 'success',
  cancelled: 'default',
};

const expressCompanyOptions = [
  { label: '顺丰速运', value: 'SF' },
  { label: '京东物流', value: 'JD' },
  { label: '中通快递', value: 'ZTO' },
  { label: '圆通速递', value: 'YTO' },
  { label: '申通快递', value: 'STO' },
  { label: '韵达快递', value: 'YD' },
  { label: '邮政EMS', value: 'EMS' },
  { label: '其他', value: 'OTHER' },
];

const ExchangeOrderList: React.FC = () => {
  const { message, modal } = App.useApp();
  const actionRef = useRef<ActionType>();
  const [auditModalVisible, setAuditModalVisible] = useState(false);
  const [deliveryModalVisible, setDeliveryModalVisible] = useState(false);
  const [detailModalVisible, setDetailModalVisible] = useState(false);
  const [currentOrder, setCurrentOrder] = useState<ExchangeOrder | null>(null);
  const [auditType, setAuditType] = useState<'approve' | 'reject'>('approve');
  const [loading, setLoading] = useState(false);
  const [auditForm] = Form.useForm<{ remark?: string }>();
  const [deliveryForm] = Form.useForm<{
    expressCompany: string;
    expressNo: string;
    remark?: string;
  }>();
  const [stats, setStats] = useState({
    totalOrders: 0,
    pendingOrders: 0,
    deliveringOrders: 0,
    completedOrders: 0,
  });

  const handleOpenAudit = (record: ExchangeOrder, type: 'approve' | 'reject') => {
    setCurrentOrder(record);
    setAuditType(type);
    auditForm.resetFields();
    setAuditModalVisible(true);
  };

  const handleAuditFinish = async (values: { remark?: string }) => {
    if (!currentOrder) return false;
    try {
      setLoading(true);
      await patrolService.auditExchangeOrder({
        orderId: currentOrder.id,
        status: auditType === 'approve' ? 'approved' : 'rejected',
        remark: values.remark,
      });
      message.success(auditType === 'approve' ? '审核通过' : '审核已拒绝');
      actionRef.current?.reload();
      setAuditModalVisible(false);
      return true;
    } catch (error: any) {
      message.error(error.message || '操作失败');
      return false;
    } finally {
      setLoading(false);
    }
  };

  const handleOpenDelivery = (record: ExchangeOrder) => {
    setCurrentOrder(record);
    deliveryForm.resetFields();
    setDeliveryModalVisible(true);
  };

  const handleDeliveryFinish = async (values: {
    expressCompany: string;
    expressNo: string;
    remark?: string;
  }) => {
    if (!currentOrder) return false;
    try {
      setLoading(true);
      await patrolService.deliveryExchangeOrder({
        orderId: currentOrder.id,
        expressCompany: values.expressCompany,
        expressNo: values.expressNo,
        remark: values.remark,
      });
      message.success('发货成功');
      actionRef.current?.reload();
      setDeliveryModalVisible(false);
      return true;
    } catch (error: any) {
      message.error(error.message || '操作失败');
      return false;
    } finally {
      setLoading(false);
    }
  };

  const handleViewDetail = async (record: ExchangeOrder) => {
    try {
      const res = await patrolService.getExchangeOrderDetail(record.id);
      const data: any = (res as any)?.data ?? res;
      setCurrentOrder(data);
      setDetailModalVisible(true);
    } catch (error: any) {
      message.error(error.message || '获取详情失败');
    }
  };

  const handleComplete = (record: ExchangeOrder) => {
    modal.confirm({
      title: '确认标记为已完成?',
      icon: <CheckCircleOutlined />,
      content: `订单: ${record.orderNo}`,
      okText: '确认完成',
      cancelText: '取消',
      onOk: async () => {
        try {
          await patrolService.completeExchangeOrder(record.id);
          message.success('订单已完成');
          actionRef.current?.reload();
        } catch (error: any) {
          message.error(error.message || '操作失败');
          return Promise.reject();
        }
      },
    });
  };

  const handleCancel = (record: ExchangeOrder) => {
    modal.confirm({
      title: '确认取消该订单?',
      icon: <StopOutlined />,
      content: `订单: ${record.orderNo}`,
      okText: '确认取消',
      cancelText: '取消',
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          await patrolService.cancelExchangeOrder(record.id);
          message.success('订单已取消');
          actionRef.current?.reload();
        } catch (error: any) {
          message.error(error.message || '操作失败');
          return Promise.reject();
        }
      },
    });
  };

  const columns: ProColumns<ExchangeOrder>[] = [
    {
      title: '订单编号',
      dataIndex: 'orderNo',
      width: 180,
      copyable: true,
      fixed: 'left',
    },
    {
      title: '礼品信息',
      dataIndex: 'giftName',
      width: 220,
      render: (_, record) => (
        <Space>
          {record.giftImage ? (
            <Image
              width={40}
              height={40}
              src={record.giftImage}
              style={{ objectFit: 'cover', borderRadius: 4 }}
            />
          ) : (
            <div
              style={{
                width: 40,
                height: 40,
                background: '#f5f5f5',
                borderRadius: 4,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: '#999',
              }}
            >
              <ShoppingCartOutlined />
            </div>
          )}
          <Space direction="vertical" size={0}>
            <span style={{ fontWeight: 500 }}>{record.giftName}</span>
            <span style={{ fontSize: 12, color: '#999' }}>
              x{record.quantity} · {record.giftPoints || 0}积分/件
            </span>
          </Space>
        </Space>
      ),
    },
    {
      title: '消耗积分',
      dataIndex: 'totalPoints',
      width: 100,
      render: (_, record) => (
        <span style={{ color: '#faad14', fontWeight: 500, fontSize: 16 }}>
          {record.totalPoints}
        </span>
      ),
    },
    {
      title: '兑换人',
      dataIndex: 'memberName',
      width: 120,
      render: (_, record) => (
        <Space>
          <UserOutlined />
          <span>{record.memberName}</span>
        </Space>
      ),
    },
    {
      title: '联系电话',
      dataIndex: 'memberPhone',
      width: 130,
      render: (_, record) => (
        <Space>
          <PhoneOutlined />
          <span>{record.memberPhone}</span>
        </Space>
      ),
    },
    {
      title: '收货地址',
      dataIndex: 'receiverAddress',
      width: 200,
      ellipsis: true,
      render: (_, record) => (
        <Space>
          <EnvironmentOutlined />
          <span>{record.receiverAddress || '-'}</span>
        </Space>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (_, record) => {
        const statusName = record.statusName || STATUS_MAP[record.status] || record.status;
        return (
          <Tag color={STATUS_COLOR[record.status] || 'default'}>
            {statusName}
          </Tag>
        );
      },
    },
    {
      title: '物流信息',
      dataIndex: 'expressNo',
      width: 160,
      render: (_, record) => {
        if (record.expressCompany && record.expressNo) {
          return (
            <Space direction="vertical" size={0}>
              <span>{record.expressCompany}</span>
              <span style={{ fontSize: 12, color: '#999' }}>{record.expressNo}</span>
            </Space>
          );
        }
        return '-';
      },
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
      width: 240,
      fixed: 'right',
      render: (_, record) => {
        const actions = [];
        actions.push(
          <Button type="link" key="view" icon={<EyeOutlined />} onClick={() => handleViewDetail(record)}>
            详情
          </Button>,
        );
        if (record.status === 'pending') {
          actions.push(
            <Button
              type="link"
              key="approve"
              icon={<CheckCircleOutlined />}
              onClick={() => handleOpenAudit(record, 'approve')}
            >
              通过
            </Button>,
          );
          actions.push(
            <Button
              type="link"
              key="reject"
              danger
              icon={<CloseCircleOutlined />}
              onClick={() => handleOpenAudit(record, 'reject')}
            >
              拒绝
            </Button>,
          );
        }
        if (record.status === 'approved') {
          actions.push(
            <Button
              type="link"
              key="delivery"
              icon={<TruckOutlined />}
              onClick={() => handleOpenDelivery(record)}
            >
              发货
            </Button>,
          );
        }
        if (record.status === 'delivering') {
          actions.push(
            <Button
              type="link"
              key="complete"
              icon={<CheckOutlined />}
              onClick={() => handleComplete(record)}
            >
              完成
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
      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={12} md={6}>
          <Card>
            <Statistic
              title="订单总数"
              value={stats.totalOrders}
              prefix={<ShoppingCartOutlined style={{ color: '#1677ff' }} />}
              valueStyle={{ color: '#1677ff' }}
            />
          </Card>
        </Col>
        <Col xs={12} md={6}>
          <Card>
            <Statistic
              title="待审核"
              value={stats.pendingOrders}
              prefix={<ClockCircleOutlined style={{ color: '#faad14' }} />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col xs={12} md={6}>
          <Card>
            <Statistic
              title="已发货"
              value={stats.deliveringOrders}
              prefix={<TruckOutlined style={{ color: '#1677ff' }} />}
              valueStyle={{ color: '#1677ff' }}
            />
          </Card>
        </Col>
        <Col xs={12} md={6}>
          <Card>
            <Statistic
              title="已完成"
              value={stats.completedOrders}
              prefix={<CheckCircleOutlined style={{ color: '#52c41a' }} />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
      </Row>

      <ProTable<ExchangeOrder>
        columns={columns}
        actionRef={actionRef}
        cardBordered
        rowKey="id"
        search={{
          labelWidth: 'auto',
          defaultCollapsed: false,
        }}
        dateFormatter="string"
        headerTitle="兑换订单列表"
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
            const res = await patrolService.getExchangeOrderList({
              pageNum: params.current,
              pageSize: params.pageSize,
              keyword: params.keyword as string,
              orderNo: (params.orderNo as string) || undefined,
              memberName: (params.memberName as string) || undefined,
              status: (params.status as string) || undefined,
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
          persistenceKey: 'exchange-order-list-columns',
          persistenceType: 'localStorage',
        }}
        pagination={{
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total) => `共 ${total} 条记录`,
        }}
        scroll={{ x: 1800 }}
      />

      <ModalForm
        title={auditType === 'approve' ? '审核通过' : '审核拒绝'}
        open={auditModalVisible}
        onOpenChange={setAuditModalVisible}
        form={auditForm}
        layout="vertical"
        modalProps={{
          destroyOnClose: true,
          maskClosable: false,
        }}
        onFinish={handleAuditFinish}
        submitter={{
          submitButtonProps: {
            loading,
            danger: auditType === 'reject',
          },
        }}
      >
        {currentOrder && (
          <div style={{ marginBottom: 16, padding: 12, background: '#f5f5f5', borderRadius: 4 }}>
            <div style={{ fontWeight: 500 }}>{currentOrder.giftName}</div>
            <div style={{ fontSize: 12, color: '#999', marginTop: 4 }}>
              订单编号: {currentOrder.orderNo} · 兑换人: {currentOrder.memberName}
            </div>
          </div>
        )}
        <ProFormTextArea
          name="remark"
          label="审核备注"
          placeholder={auditType === 'approve' ? '请输入审核备注（可选）' : '请输入拒绝原因'}
          rules={auditType === 'reject' ? [{ required: true, message: '请输入拒绝原因' }] : []}
          rows={3}
        />
      </ModalForm>

      <ModalForm
        title="订单发货"
        open={deliveryModalVisible}
        onOpenChange={setDeliveryModalVisible}
        form={deliveryForm}
        layout="vertical"
        modalProps={{
          destroyOnClose: true,
          maskClosable: false,
        }}
        onFinish={handleDeliveryFinish}
        submitter={{
          submitButtonProps: {
            loading,
          },
        }}
      >
        {currentOrder && (
          <div style={{ marginBottom: 16, padding: 12, background: '#f5f5f5', borderRadius: 4 }}>
            <div style={{ fontWeight: 500 }}>{currentOrder.giftName}</div>
            <div style={{ fontSize: 12, color: '#999', marginTop: 4 }}>
              订单编号: {currentOrder.orderNo}
            </div>
            <div style={{ fontSize: 12, marginTop: 4 }}>
              <Text type="secondary">收货人：</Text>
              {currentOrder.receiverName} {currentOrder.receiverPhone}
            </div>
            <div style={{ fontSize: 12 }}>
              <Text type="secondary">收货地址：</Text>
              {currentOrder.receiverAddress}
            </div>
          </div>
        )}
        <ProFormSelect
          name="expressCompany"
          label="物流公司"
          placeholder="请选择物流公司"
          rules={[{ required: true, message: '请选择物流公司' }]}
          options={expressCompanyOptions}
        />
        <ProFormText
          name="expressNo"
          label="物流单号"
          placeholder="请输入物流单号"
          rules={[{ required: true, message: '请输入物流单号' }]}
        />
        <ProFormTextArea
          name="remark"
          label="发货备注"
          placeholder="请输入发货备注（可选）"
          rows={2}
        />
      </ModalForm>

      <Modal
        title="订单详情"
        open={detailModalVisible}
        onCancel={() => setDetailModalVisible(false)}
        footer={null}
        width={600}
        destroyOnClose
      >
        {currentOrder && (
          <Descriptions column={1} bordered size="small">
            <Descriptions.Item label="订单编号">{currentOrder.orderNo}</Descriptions.Item>
            <Descriptions.Item label="订单状态">
              <Tag color={STATUS_COLOR[currentOrder.status] || 'default'}>
                {currentOrder.statusName || STATUS_MAP[currentOrder.status] || currentOrder.status}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="礼品信息">
              <Space>
                {currentOrder.giftImage && (
                  <Image width={40} height={40} src={currentOrder.giftImage} />
                )}
                <div>
                  <div>{currentOrder.giftName}</div>
                  <div style={{ fontSize: 12, color: '#999' }}>
                    x{currentOrder.quantity} · {currentOrder.giftPoints}积分/件
                  </div>
                </div>
              </Space>
            </Descriptions.Item>
            <Descriptions.Item label="消耗积分">
              <span style={{ color: '#faad14', fontWeight: 500 }}>{currentOrder.totalPoints} 积分</span>
            </Descriptions.Item>
            <Descriptions.Item label="兑换人">{currentOrder.memberName}</Descriptions.Item>
            <Descriptions.Item label="联系电话">{currentOrder.memberPhone}</Descriptions.Item>
            <Descriptions.Item label="收货人">{currentOrder.receiverName}</Descriptions.Item>
            <Descriptions.Item label="收货电话">{currentOrder.receiverPhone}</Descriptions.Item>
            <Descriptions.Item label="收货地址">{currentOrder.receiverAddress}</Descriptions.Item>
            {currentOrder.expressCompany && (
              <Descriptions.Item label="物流公司">{currentOrder.expressCompany}</Descriptions.Item>
            )}
            {currentOrder.expressNo && (
              <Descriptions.Item label="物流单号">{currentOrder.expressNo}</Descriptions.Item>
            )}
            {currentOrder.auditRemark && (
              <Descriptions.Item label="审核备注">{currentOrder.auditRemark}</Descriptions.Item>
            )}
            {currentOrder.remark && (
              <Descriptions.Item label="订单备注">{currentOrder.remark}</Descriptions.Item>
            )}
            <Descriptions.Item label="创建时间">
              {currentOrder.createTime ? dayjs(currentOrder.createTime).format('YYYY-MM-DD HH:mm:ss') : '-'}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Modal>
    </>
  );
};

export default ExchangeOrderList;
