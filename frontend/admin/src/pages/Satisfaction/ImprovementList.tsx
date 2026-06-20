import React, { useState, useEffect } from 'react';
import {
  Table,
  Button,
  Space,
  Tag,
  Modal,
  Form,
  Input,
  Select,
  Descriptions,
  Drawer,
  message,
  Popconfirm,
  Card,
  Badge,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import {
  ToolOutlined,
  CheckCircleOutlined,
  EyeOutlined,
  SendOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import {
  improvementService,
  ImprovementOrder,
  ImprovementStatusMap,
  IssueTypeMap,
  PriorityMap,
} from '../../services/satisfaction';

const { TextArea } = Input;
const { Option } = Select;

const ImprovementList: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<ImprovementOrder[]>([]);
  const [total, setTotal] = useState(0);
  const [pagination, setPagination] = useState({ current: 1, pageSize: 10 });
  const [statusFilter, setStatusFilter] = useState<number | undefined>();
  const [detailDrawer, setDetailDrawer] = useState(false);
  const [currentOrder, setCurrentOrder] = useState<ImprovementOrder | null>(null);
  const [rectifyModal, setRectifyModal] = useState(false);
  const [reviewModal, setReviewModal] = useState(false);
  const [rectifyForm] = Form.useForm();
  const [reviewForm] = Form.useForm();

  const fetchData = async () => {
    setLoading(true);
    try {
      const response = await improvementService.getList({
        status: statusFilter,
        page: pagination.current,
        pageSize: pagination.pageSize,
      });
      setData(response.list);
      setTotal(response.total);
    } catch (error) {
      message.error('获取改进工单失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [pagination.current, pagination.pageSize, statusFilter]);

  const handleViewDetail = async (record: ImprovementOrder) => {
    try {
      const detail = await improvementService.getDetail(record.id);
      setCurrentOrder(detail);
      setDetailDrawer(true);
    } catch (error) {
      message.error('获取工单详情失败');
    }
  };

  const handleSubmitRectify = async (values: any) => {
    if (!currentOrder) return;
    try {
      await improvementService.submitRectification(
        currentOrder.id,
        values.content,
        values.result
      );
      message.success('整改报告提交成功');
      setRectifyModal(false);
      rectifyForm.resetFields();
      fetchData();
    } catch (error) {
      message.error('提交失败');
    }
  };

  const handleReview = async (values: any) => {
    if (!currentOrder) return;
    try {
      await improvementService.review(currentOrder.id, values.opinion, values.approved);
      message.success(values.approved ? '审核通过' : '审核退回');
      setReviewModal(false);
      reviewForm.resetFields();
      fetchData();
    } catch (error) {
      message.error('审核失败');
    }
  };

  const handleClose = async (id: string) => {
    try {
      await improvementService.close(id);
      message.success('工单已关闭');
      fetchData();
    } catch (error) {
      message.error('关闭失败');
    }
  };

  const columns: ColumnsType<ImprovementOrder> = [
    {
      title: '工单编号',
      dataIndex: 'orderNo',
      key: 'orderNo',
      width: 180,
      render: (no, record) => (
        <Button type="link" size="small" onClick={() => handleViewDetail(record)}>
          {no}
        </Button>
      ),
    },
    {
      title: '关联案件',
      dataIndex: 'caseNo',
      key: 'caseNo',
      width: 150,
      render: (_, record) => (
        <div>
          <div>{record.caseNo}</div>
          <div className="text-sm text-gray-500 truncate" style={{ maxWidth: 150 }}>
            {record.caseTitle}
          </div>
        </div>
      ),
    },
    {
      title: '调解员',
      dataIndex: 'mediatorName',
      key: 'mediatorName',
      width: 80,
    },
    {
      title: '问题类型',
      dataIndex: 'issueType',
      key: 'issueType',
      width: 100,
      render: (type) => <Tag>{IssueTypeMap[type] || type}</Tag>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status) => {
        const info = ImprovementStatusMap[status] || { text: '未知', color: 'default' };
        return <Tag color={info.color}>{info.text}</Tag>;
      },
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      key: 'priority',
      width: 70,
      render: (priority) => {
        const info = PriorityMap[priority] || { text: '未知', color: 'default' };
        return <Tag color={info.color}>{info.text}</Tag>;
      },
    },
    {
      title: '扣分',
      dataIndex: 'deductionScore',
      key: 'deductionScore',
      width: 70,
      render: (score) =>
        score > 0 ? (
          <span className="text-red-500 font-bold">-{score.toFixed(1)}</span>
        ) : (
          '-'
        ),
    },
    {
      title: '截止时间',
      dataIndex: 'deadline',
      key: 'deadline',
      width: 110,
      render: (time) => {
        if (!time) return '-';
        const isOverdue = dayjs(time).isBefore(dayjs());
        return (
          <span className={isOverdue ? 'text-red-500' : ''}>
            {dayjs(time).format('YYYY-MM-DD')}
            {isOverdue && ' (逾期)'}
          </span>
        );
      },
    },
    {
      title: '操作',
      key: 'action',
      fixed: 'right',
      width: 200,
      render: (_, record) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => handleViewDetail(record)}
          >
            详情
          </Button>
          {(record.status === 10 || record.status === 20) && (
            <Button
              type="link"
              size="small"
              icon={<SendOutlined />}
              onClick={() => {
                setCurrentOrder(record);
                setRectifyModal(true);
              }}
            >
              整改
            </Button>
          )}
          {record.status === 30 && (
            <Button
              type="link"
              size="small"
              icon={<CheckCircleOutlined />}
              onClick={() => {
                setCurrentOrder(record);
                setReviewModal(true);
              }}
            >
              审核
            </Button>
          )}
          {record.status === 40 && (
            <Popconfirm
              title="确认关闭此工单？"
              onConfirm={() => handleClose(record.id)}
            >
              <Button type="link" size="small" danger>
                关闭
              </Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">
          <ToolOutlined className="mr-2" />
          改进工单管理
        </h1>
        <Select
          placeholder="筛选状态"
          allowClear
          style={{ width: 150 }}
          value={statusFilter}
          onChange={(val) => {
            setStatusFilter(val);
            setPagination({ ...pagination, current: 1 });
          }}
        >
          {Object.entries(ImprovementStatusMap).map(([key, value]) => (
            <Option key={key} value={Number(key)}>
              {value.text}
            </Option>
          ))}
        </Select>
      </div>

      <Table
        columns={columns}
        dataSource={data}
        rowKey="id"
        loading={loading}
        pagination={{
          ...pagination,
          total,
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (t) => `共 ${t} 条记录`,
          onChange: (page, pageSize) => setPagination({ current: page, pageSize }),
        }}
        scroll={{ x: 1100 }}
      />

      <Drawer
        title="工单详情"
        placement="right"
        width={700}
        open={detailDrawer}
        onClose={() => setDetailDrawer(false)}
      >
        {currentOrder && (
          <div className="space-y-4">
            <Descriptions title="基本信息" bordered column={2} size="small">
              <Descriptions.Item label="工单编号">{currentOrder.orderNo}</Descriptions.Item>
              <Descriptions.Item label="状态">
                <Tag color={ImprovementStatusMap[currentOrder.status]?.color}>
                  {ImprovementStatusMap[currentOrder.status]?.text}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="案件编号">{currentOrder.caseNo}</Descriptions.Item>
              <Descriptions.Item label="案件标题">{currentOrder.caseTitle}</Descriptions.Item>
              <Descriptions.Item label="调解员">{currentOrder.mediatorName}</Descriptions.Item>
              <Descriptions.Item label="组织">{currentOrder.orgName}</Descriptions.Item>
              <Descriptions.Item label="问题类型">
                <Tag>{IssueTypeMap[currentOrder.issueType] || currentOrder.issueType}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="优先级">
                <Tag color={PriorityMap[currentOrder.priority]?.color}>
                  {PriorityMap[currentOrder.priority]?.text}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="截止时间" span={2}>
                {currentOrder.deadline ? dayjs(currentOrder.deadline).format('YYYY-MM-DD HH:mm') : '-'}
              </Descriptions.Item>
            </Descriptions>

            <Card title="评价内容" size="small">
              <div className="mb-2">
                <span className="text-gray-500 mr-2">满意度评分：</span>
                {Array.from({ length: 5 }).map((_, i) => (
                  <span key={i} style={{ color: i < currentOrder.satisfactionScore ? '#fadb14' : '#d9d9d9' }}>
                    ★
                  </span>
                ))}
              </div>
              <div className="mb-2">
                <span className="text-gray-500 mr-2">群众评语：</span>
                {currentOrder.satisfactionComment}
              </div>
              <div className="mb-2">
                <span className="text-gray-500 mr-2">情感分析：</span>
                <Tag color={currentOrder.sentimentEmotion === 'negative' ? 'error' : currentOrder.sentimentEmotion === 'positive' ? 'success' : 'warning'}>
                  {currentOrder.sentimentEmotion === 'negative' ? '负面' : currentOrder.sentimentEmotion === 'positive' ? '正面' : '中性'}
                </Tag>
                <span className="ml-2">评分: {currentOrder.sentimentScore?.toFixed(3)}</span>
              </div>
              {currentOrder.sentimentSummary && (
                <div>
                  <span className="text-gray-500 mr-2">分析摘要：</span>
                  {currentOrder.sentimentSummary}
                </div>
              )}
            </Card>

            <Card title="问题与改进建议" size="small">
              <div className="mb-2">
                <span className="text-gray-500 mr-2">问题描述：</span>
                {currentOrder.issueDescription || '-'}
              </div>
              <div>
                <span className="text-gray-500 mr-2">改进建议：</span>
                {currentOrder.improvementSuggestion || '-'}
              </div>
            </Card>

            <Card title="绩效考核" size="small">
              <div className="mb-2">
                <span className="text-gray-500 mr-2">扣分：</span>
                <span className="text-red-500 font-bold">
                  {currentOrder.deductionScore > 0 ? `-${currentOrder.deductionScore.toFixed(1)}` : '无'}
                </span>
              </div>
              <div className="mb-2">
                <span className="text-gray-500 mr-2">扣分原因：</span>
                {currentOrder.deductionReason || '-'}
              </div>
              <div>
                <span className="text-gray-500 mr-2">是否已扣分：</span>
                <Tag color={currentOrder.isDeductionApplied === 1 ? 'success' : 'default'}>
                  {currentOrder.isDeductionApplied === 1 ? '已生效' : '未生效'}
                </Tag>
              </div>
            </Card>

            {currentOrder.rectifyContent && (
              <Card title="整改信息" size="small">
                <div className="mb-2">
                  <span className="text-gray-500 mr-2">整改内容：</span>
                  {currentOrder.rectifyContent}
                </div>
                <div className="mb-2">
                  <span className="text-gray-500 mr-2">整改结果：</span>
                  {currentOrder.rectifyResult}
                </div>
                <div>
                  <span className="text-gray-500 mr-2">整改时间：</span>
                  {currentOrder.rectifiedAt ? dayjs(currentOrder.rectifiedAt).format('YYYY-MM-DD HH:mm') : '-'}
                </div>
              </Card>
            )}

            {currentOrder.reviewOpinion && (
              <Card title="审核信息" size="small">
                <div className="mb-2">
                  <span className="text-gray-500 mr-2">审核人：</span>
                  {currentOrder.reviewedByName}
                </div>
                <div className="mb-2">
                  <span className="text-gray-500 mr-2">审核意见：</span>
                  {currentOrder.reviewOpinion}
                </div>
                <div>
                  <span className="text-gray-500 mr-2">审核时间：</span>
                  {currentOrder.reviewedAt ? dayjs(currentOrder.reviewedAt).format('YYYY-MM-DD HH:mm') : '-'}
                </div>
              </Card>
            )}
          </div>
        )}
      </Drawer>

      <Modal
        title="提交整改报告"
        open={rectifyModal}
        onCancel={() => {
          setRectifyModal(false);
          rectifyForm.resetFields();
        }}
        onOk={() => rectifyForm.submit()}
      >
        <Form form={rectifyForm} layout="vertical" onFinish={handleSubmitRectify}>
          <Form.Item name="content" label="整改内容" rules={[{ required: true, message: '请填写整改内容' }]}>
            <TextArea rows={4} placeholder="请描述您采取的整改措施" />
          </Form.Item>
          <Form.Item name="result" label="整改结果" rules={[{ required: true, message: '请填写整改结果' }]}>
            <TextArea rows={4} placeholder="请描述整改后的效果" />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="审核整改"
        open={reviewModal}
        onCancel={() => {
          setReviewModal(false);
          reviewForm.resetFields();
        }}
        onOk={() => reviewForm.submit()}
      >
        <Form form={reviewForm} layout="vertical" onFinish={handleReview}>
          <Form.Item name="approved" label="审核结果" rules={[{ required: true, message: '请选择审核结果' }]}>
            <Select placeholder="请选择">
              <Option value={true}>通过</Option>
              <Option value={false}>退回重新整改</Option>
            </Select>
          </Form.Item>
          <Form.Item name="opinion" label="审核意见" rules={[{ required: true, message: '请填写审核意见' }]}>
            <TextArea rows={4} placeholder="请填写审核意见" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default ImprovementList;
