import React, { useEffect, useState } from 'react';
import {
  Table,
  Card,
  Form,
  Select,
  Button,
  Space,
  Tag,
  Modal,
  Input,
  message,
  Descriptions,
  Drawer,
} from 'antd';
import {
  escalationApi,
  ESCALATION_TYPE_MAP,
  ESCALATION_LEVEL_MAP,
  ESCALATION_STATUS_MAP,
  type EscalationRecord,
  type EscalationListParams,
} from '@/services/escalation';

const { Option } = Select;
const { TextArea } = Input;

const EscalationList: React.FC = () => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [list, setList] = useState<EscalationRecord[]>([]);
  const [total, setTotal] = useState(0);
  const [pagination, setPagination] = useState({ current: 1, pageSize: 10 });

  const [detailVisible, setDetailVisible] = useState(false);
  const [currentDetail, setCurrentDetail] = useState<EscalationRecord | null>(null);

  const [handleModalVisible, setHandleModalVisible] = useState(false);
  const [closeModalVisible, setCloseModalVisible] = useState(false);
  const [selectedId, setSelectedId] = useState<number | null>(null);
  const [actionLoading, setActionLoading] = useState(false);

  const fetchData = async (params?: EscalationListParams) => {
    setLoading(true);
    try {
      const values = form.getFieldsValue();
      const queryParams: EscalationListParams = {
        page: pagination.current,
        pageSize: pagination.pageSize,
        ...values,
        ...params,
      };
      const res = await escalationApi.getList(queryParams);
      setList(res.list);
      setTotal(res.total);
    } catch (error) {
      message.error('获取升级记录列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [pagination.current, pagination.pageSize]);

  const handleSearch = () => {
    setPagination({ ...pagination, current: 1 });
    fetchData({ page: 1 });
  };

  const handleReset = () => {
    form.resetFields();
    setPagination({ current: 1, pageSize: 10 });
    fetchData({ page: 1, pageSize: 10 });
  };

  const handleViewDetail = async (record: EscalationRecord) => {
    try {
      const detail = await escalationApi.getDetail(record.id);
      setCurrentDetail(detail);
      setDetailVisible(true);
    } catch (error) {
      message.error('获取详情失败');
    }
  };

  const openHandleModal = (id: number) => {
    setSelectedId(id);
    setHandleModalVisible(true);
  };

  const openCloseModal = (id: number) => {
    setSelectedId(id);
    setCloseModalVisible(true);
  };

  const handleSubmitHandle = async (values: { remark?: string }) => {
    if (!selectedId) return;
    setActionLoading(true);
    try {
      await escalationApi.handle(selectedId, values.remark);
      message.success('处理成功');
      setHandleModalVisible(false);
      setSelectedId(null);
      fetchData();
    } catch (error) {
      message.error('处理失败');
    } finally {
      setActionLoading(false);
    }
  };

  const handleSubmitClose = async (values: { remark: string }) => {
    if (!selectedId) return;
    setActionLoading(true);
    try {
      await escalationApi.close(selectedId, values.remark);
      message.success('关闭成功');
      setCloseModalVisible(false);
      setSelectedId(null);
      fetchData();
    } catch (error) {
      message.error('关闭失败');
    } finally {
      setActionLoading(false);
    }
  };

  const columns = [
    {
      title: '案件编号',
      dataIndex: 'caseNo',
      key: 'caseNo',
      width: 140,
    },
    {
      title: '升级类型',
      dataIndex: 'escalateType',
      key: 'escalateType',
      width: 120,
      render: (val: number) => (
        <Tag color={val === 1 ? 'orange' : 'red'}>
          {ESCALATION_TYPE_MAP[val] || '未知'}
        </Tag>
      ),
    },
    {
      title: '升级路径',
      key: 'levelPath',
      width: 150,
      render: (_: any, record: EscalationRecord) => (
        <Space>
          <Tag color="blue">{ESCALATION_LEVEL_MAP[record.fromLevel] || '-'}</Tag>
          <span>→</span>
          <Tag color="red">{ESCALATION_LEVEL_MAP[record.toLevel] || '-'}</Tag>
        </Space>
      ),
    },
    {
      title: '原处理人',
      dataIndex: 'fromUserName',
      key: 'fromUserName',
      width: 100,
    },
    {
      title: '升级后处理人',
      dataIndex: 'toUserName',
      key: 'toUserName',
      width: 100,
    },
    {
      title: '催办次数',
      dataIndex: 'urgeCount',
      key: 'urgeCount',
      width: 80,
    },
    {
      title: '超时(小时)',
      dataIndex: 'timeoutHours',
      key: 'timeoutHours',
      width: 80,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (val: number) => {
        const colorMap: Record<number, string> = {
          10: 'orange',
          20: 'blue',
          30: 'green',
          40: 'default',
        };
        return <Tag color={colorMap[val]}>{ESCALATION_STATUS_MAP[val] || '-'}</Tag>;
      },
    },
    {
      title: '升级原因',
      dataIndex: 'reason',
      key: 'reason',
      ellipsis: true,
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      width: 160,
    },
    {
      title: '操作',
      key: 'action',
      width: 180,
      fixed: 'right' as const,
      render: (_: any, record: EscalationRecord) => (
        <Space size="small">
          <Button type="link" size="small" onClick={() => handleViewDetail(record)}>
            详情
          </Button>
          {record.status === 10 && (
            <Button type="link" size="small" onClick={() => openHandleModal(record.id)}>
              处理
            </Button>
          )}
          {record.status !== 40 && (
            <Button type="link" size="small" danger onClick={() => openCloseModal(record.id)}>
              关闭
            </Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Card>
        <Form form={form} layout="inline" onFinish={handleSearch}>
          <Form.Item name="toLevel" label="升级到级别">
            <Select placeholder="请选择" allowClear style={{ width: 150 }}>
              <Option value={1}>组长</Option>
              <Option value={2}>主任</Option>
              <Option value={3}>领导</Option>
            </Select>
          </Form.Item>
          <Form.Item name="status" label="状态">
            <Select placeholder="请选择" allowClear style={{ width: 150 }}>
              <Option value={10}>待处理</Option>
              <Option value={20}>处理中</Option>
              <Option value={30}>已处理</Option>
              <Option value={40}>已关闭</Option>
            </Select>
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                查询
              </Button>
              <Button onClick={handleReset}>重置</Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>

      <Card style={{ marginTop: 16 }}>
        <Table
          rowKey="id"
          loading={loading}
          dataSource={list}
          columns={columns}
          pagination={{
            current: pagination.current,
            pageSize: pagination.pageSize,
            total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (t) => `共 ${t} 条`,
            onChange: (page, pageSize) => setPagination({ current: page, pageSize }),
          }}
          scroll={{ x: 1300 }}
        />
      </Card>

      <Drawer
        title="升级记录详情"
        width={600}
        open={detailVisible}
        onClose={() => setDetailVisible(false)}
      >
        {currentDetail && (
          <Descriptions column={1} bordered size="small">
            <Descriptions.Item label="案件编号">{currentDetail.caseNo}</Descriptions.Item>
            <Descriptions.Item label="升级类型">
              {ESCALATION_TYPE_MAP[currentDetail.escalateType] || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="原处理级别">
              {ESCALATION_LEVEL_MAP[currentDetail.fromLevel] || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="升级到级别">
              {ESCALATION_LEVEL_MAP[currentDetail.toLevel] || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="原处理人">{currentDetail.fromUserName || '-'}</Descriptions.Item>
            <Descriptions.Item label="升级后处理人">{currentDetail.toUserName || '-'}</Descriptions.Item>
            <Descriptions.Item label="归属组织">{currentDetail.toOrgName || '-'}</Descriptions.Item>
            <Descriptions.Item label="催办次数">{currentDetail.urgeCount}</Descriptions.Item>
            <Descriptions.Item label="超时小时数">{currentDetail.timeoutHours}</Descriptions.Item>
            <Descriptions.Item label="首次催办时间">
              {currentDetail.firstUrgeTime || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="升级原因">{currentDetail.reason}</Descriptions.Item>
            <Descriptions.Item label="操作人">{currentDetail.operatorName}</Descriptions.Item>
            <Descriptions.Item label="状态">
              {ESCALATION_STATUS_MAP[currentDetail.status] || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="备注">{currentDetail.remark || '-'}</Descriptions.Item>
            <Descriptions.Item label="创建时间">{currentDetail.createdAt}</Descriptions.Item>
          </Descriptions>
        )}
      </Drawer>

      <Modal
        title="处理升级记录"
        open={handleModalVisible}
        onCancel={() => setHandleModalVisible(false)}
        onOk={() => form.submit()}
        confirmLoading={actionLoading}
        destroyOnClose
      >
        <Form layout="vertical" onFinish={handleSubmitHandle}>
          <Form.Item name="remark" label="处理备注">
            <TextArea rows={4} placeholder="请输入处理备注" />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="关闭升级记录"
        open={closeModalVisible}
        onCancel={() => setCloseModalVisible(false)}
        onOk={() => form.submit()}
        confirmLoading={actionLoading}
        destroyOnClose
      >
        <Form layout="vertical" onFinish={handleSubmitClose}>
          <Form.Item
            name="remark"
            label="关闭原因"
            rules={[{ required: true, message: '请输入关闭原因' }]}
          >
            <TextArea rows={4} placeholder="请输入关闭原因" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default EscalationList;
