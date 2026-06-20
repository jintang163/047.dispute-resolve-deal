import React, { useState, useEffect } from 'react';
import {
  Table,
  Button,
  Space,
  Modal,
  Form,
  Input,
  Select,
  DatePicker,
  Tag,
  message,
  Popconfirm,
  Descriptions,
  Drawer,
  Badge,
  Card,
  Row,
  Col,
  Progress,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import {
  PhoneOutlined,
  ReloadOutlined,
  PlayCircleOutlined,
  StopOutlined,
  SyncOutlined,
  DeleteOutlined,
  EyeOutlined,
  FileAudioOutlined,
  SearchOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import {
  callbackService,
  CallbackRecord,
  CallbackStatusMap,
  CallStatusMap,
  EmotionMap,
  CallbackStatusEnum,
  SentimentDetail,
} from '../../services/callback';
import { disputeService } from '../../services/dispute';

const { RangePicker } = DatePicker;
const { Option } = Select;
const { TextArea } = Input;

const CallbackList: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<CallbackRecord[]>([]);
  const [total, setTotal] = useState(0);
  const [pagination, setPagination] = useState({ current: 1, pageSize: 10 });
  const [form] = Form.useForm();
  const [detailDrawer, setDetailDrawer] = useState(false);
  const [currentRecord, setCurrentRecord] = useState<CallbackRecord | null>(null);
  const [sentimentDetail, setSentimentDetail] = useState<SentimentDetail | null>(null);
  const [createModal, setCreateModal] = useState(false);
  const [caseList, setCaseList] = useState<any[]>([]);

  const fetchData = async () => {
    setLoading(true);
    try {
      const values = form.getFieldsValue();
      const params = {
        page: pagination.current,
        pageSize: pagination.pageSize,
        ...values,
        startTime: values.timeRange?.[0]?.format('YYYY-MM-DD HH:mm:ss'),
        endTime: values.timeRange?.[1]?.format('YYYY-MM-DD HH:mm:ss'),
      };
      delete params.timeRange;

      const response = await callbackService.getList(params);
      setData(response.list);
      setTotal(response.total);
    } catch (error) {
      message.error('获取回访记录失败');
    } finally {
      setLoading(false);
    }
  };

  const fetchCaseList = async () => {
    try {
      const response = await disputeService.getList({ pageSize: 100 });
      setCaseList(response.list);
    } catch (error) {
      console.error('获取案件列表失败', error);
    }
  };

  useEffect(() => {
    fetchData();
    fetchCaseList();
  }, [pagination.current, pagination.pageSize]);

  const handleSearch = () => {
    setPagination({ ...pagination, current: 1 });
    fetchData();
  };

  const handleReset = () => {
    form.resetFields();
    setPagination({ current: 1, pageSize: 10 });
    fetchData();
  };

  const handleInitiate = async (id: string) => {
    try {
      await callbackService.initiate(id);
      message.success('回访电话已发起');
      fetchData();
    } catch (error) {
      message.error('发起回访失败');
    }
  };

  const handleRetry = async (id: string) => {
    try {
      await callbackService.retry(id);
      message.success('回访重试已发起');
      fetchData();
    } catch (error) {
      message.error('重试失败');
    }
  };

  const handleCancel = async (id: string) => {
    try {
      await callbackService.cancel(id);
      message.success('回访已取消');
      fetchData();
    } catch (error) {
      message.error('取消失败');
    }
  };

  const handleRefresh = async (id: string) => {
    try {
      await callbackService.refresh(id);
      message.success('回访结果已刷新');
      fetchData();
    } catch (error) {
      message.error('刷新失败');
    }
  };

  const handleArchive = async (id: string) => {
    try {
      await callbackService.archiveRecording(id);
      message.success('录音已归档');
      fetchData();
    } catch (error) {
      message.error('归档失败');
    }
  };

  const handleViewDetail = async (record: CallbackRecord) => {
    setCurrentRecord(record);
    if (record.sentimentResult) {
      try {
        const detail = JSON.parse(record.sentimentResult);
        setSentimentDetail(detail);
      } catch (e) {
        setSentimentDetail(null);
      }
    } else {
      setSentimentDetail(null);
    }
    setDetailDrawer(true);
  };

  const handleCreate = async (values: any) => {
    try {
      await callbackService.create(values.caseId);
      message.success('回访任务创建成功');
      setCreateModal(false);
      fetchData();
    } catch (error) {
      message.error('创建失败');
    }
  };

  const renderScore = (score?: number) => {
    if (!score) return '-';
    return (
      <Space>
        {Array.from({ length: 5 }).map((_, i) => (
          <span key={i} style={{ color: i < score ? '#fadb14' : '#d9d9d9' }}>
            ★
          </span>
        ))}
        <span className="ml-2">{score}分</span>
      </Space>
    );
  };

  const columns: ColumnsType<CallbackRecord> = [
    {
      title: '案件信息',
      dataIndex: 'caseNo',
      key: 'caseNo',
      render: (_, record) => (
        <div>
          <div className="font-medium">{record.caseNo}</div>
          <div className="text-sm text-gray-500">{record.caseTitle}</div>
        </div>
      ),
    },
    {
      title: '申请人',
      dataIndex: 'applicantName',
      key: 'applicantName',
      render: (_, record) => (
        <div>
          <div>{record.applicantName}</div>
          <div className="text-sm text-gray-500">{record.applicantPhone}</div>
        </div>
      ),
    },
    {
      title: '回访状态',
      dataIndex: 'status',
      key: 'status',
      render: (status) => {
        const info = CallbackStatusMap[status] || { text: '未知', color: 'default' };
        return <Tag color={info.color}>{info.text}</Tag>;
      },
    },
    {
      title: '通话状态',
      dataIndex: 'callStatus',
      key: 'callStatus',
      render: (status) => {
        const info = CallStatusMap[status] || { text: '未知', color: 'default' };
        return <Tag color={info.color}>{info.text}</Tag>;
      },
    },
    {
      title: '情绪分析',
      dataIndex: 'emotion',
      key: 'emotion',
      render: (emotion) => {
        if (!emotion) return '-';
        const info = EmotionMap[emotion] || { text: '未知', color: 'default' };
        return <Tag color={info.color}>{info.text}</Tag>;
      },
    },
    {
      title: '满意度',
      dataIndex: 'satisfactionScore',
      key: 'satisfactionScore',
      render: (score) => renderScore(score),
    },
    {
      title: '回访时间',
      dataIndex: 'callTime',
      key: 'callTime',
      render: (time, record) => (
        <div>
          <div>{time ? dayjs(time).format('YYYY-MM-DD HH:mm') : '-'}</div>
          {record.callDuration ? (
            <div className="text-sm text-gray-500">通话时长: {record.callDuration}秒</div>
          ) : null}
        </div>
      ),
    },
    {
      title: '重试次数',
      dataIndex: 'retryCount',
      key: 'retryCount',
      render: (count, record) => `${count}/${record.maxRetryCount}`,
    },
    {
      title: '计划时间',
      dataIndex: 'scheduledTime',
      key: 'scheduledTime',
      render: (time) => (time ? dayjs(time).format('YYYY-MM-DD HH:mm') : '-'),
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
          {record.status === CallbackStatusEnum.PENDING && (
            <Button
              type="link"
              size="small"
              icon={<PlayCircleOutlined />}
              onClick={() => handleInitiate(record.id)}
            >
              发起
            </Button>
          )}
          {record.status === CallbackStatusEnum.FAILED &&
            record.retryCount < record.maxRetryCount && (
              <Button
                type="link"
                size="small"
                icon={<SyncOutlined />}
                onClick={() => handleRetry(record.id)}
              >
                重试
              </Button>
            )}
          {record.status === CallbackStatusEnum.CALLING && (
            <Button
              type="link"
              size="small"
              icon={<ReloadOutlined />}
              onClick={() => handleRefresh(record.id)}
            >
              刷新
            </Button>
          )}
          {(record.status === CallbackStatusEnum.PENDING ||
            record.status === CallbackStatusEnum.CALLING) && (
            <Popconfirm
              title="确认取消此回访任务？"
              onConfirm={() => handleCancel(record.id)}
            >
              <Button type="link" size="small" danger icon={<StopOutlined />}>
                取消
              </Button>
            </Popconfirm>
          )}
          {record.recordingUrl && (
            <Button
              type="link"
              size="small"
              icon={<FileAudioOutlined />}
              onClick={() => handleArchive(record.id)}
            >
              归档
            </Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">
          <PhoneOutlined className="mr-2" />
          自动回访管理
        </h1>
        <Button type="primary" icon={<PhoneOutlined />} onClick={() => setCreateModal(true)}>
          创建回访任务
        </Button>
      </div>

      <Card className="mb-4">
        <Form form={form} layout="inline" onFinish={handleSearch}>
          <Form.Item name="keyword" label="关键词">
            <Input placeholder="案件号/申请人/手机号" allowClear style={{ width: 200 }} />
          </Form.Item>
          <Form.Item name="status" label="回访状态">
            <Select placeholder="请选择" allowClear style={{ width: 150 }}>
              {Object.entries(CallbackStatusMap).map(([key, value]) => (
                <Option key={key} value={Number(key)}>
                  {value.text}
                </Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item name="callStatus" label="通话状态">
            <Select placeholder="请选择" allowClear style={{ width: 150 }}>
              {Object.entries(CallStatusMap).map(([key, value]) => (
                <Option key={key} value={Number(key)}>
                  {value.text}
                </Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item name="timeRange" label="创建时间">
            <RangePicker showTime style={{ width: 350 }} />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" icon={<SearchOutlined />}>
                搜索
              </Button>
              <Button onClick={handleReset}>重置</Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>

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
          showTotal: (total) => `共 ${total} 条记录`,
          onChange: (page, pageSize) => setPagination({ current: page, pageSize }),
        }}
        scroll={{ x: 1400 }}
      />

      <Drawer
        title="回访详情"
        placement="right"
        width={800}
        open={detailDrawer}
        onClose={() => setDetailDrawer(false)}
      >
        {currentRecord && (
          <div className="space-y-6">
            <Descriptions title="基本信息" bordered column={2} size="small">
              <Descriptions.Item label="案件编号">{currentRecord.caseNo}</Descriptions.Item>
              <Descriptions.Item label="案件标题">{currentRecord.caseTitle}</Descriptions.Item>
              <Descriptions.Item label="申请人">{currentRecord.applicantName}</Descriptions.Item>
              <Descriptions.Item label="联系电话">{currentRecord.applicantPhone}</Descriptions.Item>
              <Descriptions.Item label="回访状态">
                <Tag color={CallbackStatusMap[currentRecord.status]?.color}>
                  {CallbackStatusMap[currentRecord.status]?.text}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="通话状态">
                <Tag color={CallStatusMap[currentRecord.callStatus]?.color}>
                  {CallStatusMap[currentRecord.callStatus]?.text}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="呼叫时间">
                {currentRecord.callTime
                  ? dayjs(currentRecord.callTime).format('YYYY-MM-DD HH:mm:ss')
                  : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="通话时长">
                {currentRecord.callDuration ? `${currentRecord.callDuration}秒` : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="重试次数">
                {currentRecord.retryCount}/{currentRecord.maxRetryCount}
              </Descriptions.Item>
              <Descriptions.Item label="计划时间">
                {currentRecord.scheduledTime
                  ? dayjs(currentRecord.scheduledTime).format('YYYY-MM-DD HH:mm')
                  : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="创建时间" span={2}>
                {dayjs(currentRecord.createdAt).format('YYYY-MM-DD HH:mm:ss')}
              </Descriptions.Item>
            </Descriptions>

            {currentRecord.status === CallbackStatusEnum.SUCCESS && sentimentDetail && (
              <Card title="情绪分析结果" size="small">
                <Row gutter={16} className="mb-4">
                  <Col span={8}>
                    <div className="text-center">
                      <div className="text-sm text-gray-500 mb-1">整体情绪</div>
                      <Tag
                        color={EmotionMap[sentimentDetail.emotion]?.color}
                        className="text-lg px-4 py-1"
                      >
                        {sentimentDetail.emotionLabel}
                      </Tag>
                    </div>
                  </Col>
                  <Col span={8}>
                    <div className="text-center">
                      <div className="text-sm text-gray-500 mb-1">情绪评分</div>
                      <Progress
                        type="dashboard"
                        percent={Math.round((sentimentDetail.sentimentScore + 1) * 50)}
                        size={80}
                      />
                    </div>
                  </Col>
                  <Col span={8}>
                    <div className="text-center">
                      <div className="text-sm text-gray-500 mb-1">置信度</div>
                      <div className="text-2xl font-bold">
                        {(sentimentDetail.confidence * 100).toFixed(1)}%
                      </div>
                    </div>
                  </Col>
                </Row>
                <Row gutter={16} className="mb-4">
                  <Col span={12}>
                    <div className="mb-2">
                      <span className="text-sm text-gray-500">满意度评分：</span>
                      {renderScore(sentimentDetail.satisfaction)}
                    </div>
                  </Col>
                  <Col span={12}>
                    <div className="mb-2">
                      <span className="text-sm text-gray-500">履约评分：</span>
                      {renderScore(sentimentDetail.performance)}
                    </div>
                  </Col>
                </Row>
                {sentimentDetail.positiveKeywords?.length > 0 && (
                  <div className="mb-2">
                    <span className="text-sm text-gray-500 mr-2">正面关键词：</span>
                    {sentimentDetail.positiveKeywords.map((kw, i) => (
                      <Tag key={i} color="success">
                        {kw}
                      </Tag>
                    ))}
                  </div>
                )}
                {sentimentDetail.negativeKeywords?.length > 0 && (
                  <div className="mb-2">
                    <span className="text-sm text-gray-500 mr-2">负面关键词：</span>
                    {sentimentDetail.negativeKeywords.map((kw, i) => (
                      <Tag key={i} color="error">
                        {kw}
                      </Tag>
                    ))}
                  </div>
                )}
                <div className="mt-4">
                  <div className="text-sm text-gray-500 mb-1">摘要总结</div>
                  <div className="p-3 bg-gray-50 rounded">{sentimentDetail.summary}</div>
                </div>
                {sentimentDetail.keyPoints?.length > 0 && (
                  <div className="mt-4">
                    <div className="text-sm text-gray-500 mb-2">关键点分析</div>
                    {sentimentDetail.keyPoints.map((point, i) => (
                      <div key={i} className="flex items-start mb-2">
                        <Badge
                          status={point.sentiment === 'positive' ? 'success' : point.sentiment === 'negative' ? 'error' : 'default'}
                          className="mt-1 mr-2"
                        />
                        <span className="flex-1">{point.content}</span>
                      </div>
                    ))}
                  </div>
                )}
              </Card>
            )}

            {currentRecord.transcriptText && (
              <Card title="语音转文字内容" size="small">
                <div className="p-3 bg-gray-50 rounded whitespace-pre-wrap">
                  {currentRecord.transcriptText}
                </div>
              </Card>
            )}

            {currentRecord.recordingUrl && (
              <Card title="回访录音" size="small">
                <audio controls src={currentRecord.recordingUrl} className="w-full" />
                <div className="mt-2 text-sm text-gray-500">
                  文件大小: {currentRecord.recordingSize ? `${(currentRecord.recordingSize / 1024 / 1024).toFixed(2)} MB` : '-'}
                </div>
              </Card>
            )}

            {currentRecord.remark && (
              <Card title="备注" size="small">
                <div className="p-3 bg-gray-50 rounded">{currentRecord.remark}</div>
              </Card>
            )}
          </div>
        )}
      </Drawer>

      <Modal
        title="创建回访任务"
        open={createModal}
        onCancel={() => setCreateModal(false)}
        footer={null}
      >
        <Form layout="vertical" onFinish={handleCreate}>
          <Form.Item
            name="caseId"
            label="选择案件"
            rules={[{ required: true, message: '请选择案件' }]}
          >
            <Select placeholder="请选择已结案的案件" showSearch optionFilterProp="children">
              {caseList
                .filter((c) => c.status === '50' || c.statusName === '已结案')
                .map((c) => (
                  <Option key={c.id} value={c.id}>
                    {c.caseNo} - {c.title}
                  </Option>
                ))}
            </Select>
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                创建
              </Button>
              <Button onClick={() => setCreateModal(false)}>取消</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default CallbackList;
