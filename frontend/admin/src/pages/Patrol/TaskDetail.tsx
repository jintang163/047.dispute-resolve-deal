import React, { useEffect, useState } from 'react';
import {
  Descriptions,
  Card,
  Row,
  Col,
  Timeline,
  Tag,
  Button,
  Space,
  Spin,
  Divider,
  App,
  List,
  Avatar,
  Empty,
  Typography,
  Image,
} from 'antd';
const { Text } = Typography;
import {
  ArrowLeftOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  CloseCircleOutlined,
  UserOutlined,
  EnvironmentOutlined,
  FileTextOutlined,
  SendOutlined,
  StopOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import { ProDescriptions } from '@ant-design/pro-components';
import {
  patrolService,
  PatrolTaskDetail,
  PatrolCheckRecord,
  PatrolTask,
  GridMember,
} from '../../services/patrol';
import dayjs from 'dayjs';

const statusColorMap: Record<string, string> = {
  pending: 'default',
  assigned: 'blue',
  in_progress: 'processing',
  completed: 'success',
  cancelled: 'default',
};

const statusTextMap: Record<string, string> = {
  pending: '待下发',
  assigned: '已下发',
  in_progress: '进行中',
  completed: '已完成',
  cancelled: '已取消',
};

const priorityColorMap: Record<string, string> = {
  low: 'default',
  medium: 'blue',
  high: 'orange',
  urgent: 'red',
};

const priorityTextMap: Record<string, string> = {
  low: '低',
  medium: '中',
  high: '高',
  urgent: '紧急',
};

const typeTextMap: Record<string, string> = {
  routine: '日常排查',
  key_area: '重点区域排查',
  special: '专项排查',
  emergency: '应急排查',
  complaint: '投诉核查',
};

const getTimelineIcon = (status: string) => {
  switch (status) {
    case 'completed':
      return <CheckCircleOutlined style={{ color: '#52c41a', fontSize: 16 }} />;
    case 'cancelled':
      return <CloseCircleOutlined style={{ color: '#ff4d4f', fontSize: 16 }} />;
    case 'pending':
      return <ClockCircleOutlined style={{ color: '#faad14', fontSize: 16 }} />;
    case 'assigned':
      return <SendOutlined style={{ color: '#1677ff', fontSize: 16 }} />;
    default:
      return <ClockCircleOutlined style={{ color: '#1677ff', fontSize: 16 }} />;
  }
};

const getTimelineColor = (status: string) => {
  switch (status) {
    case 'completed':
      return 'green';
    case 'cancelled':
      return 'red';
    case 'pending':
      return 'orange';
    case 'assigned':
      return 'blue';
    default:
      return 'blue';
  }
};

const TaskDetail: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const { message, modal } = App.useApp();
  const [loading, setLoading] = useState(false);
  const [detail, setDetail] = useState<PatrolTaskDetail | null>(null);
  const [checkRecords, setCheckRecords] = useState<PatrolCheckRecord[]>([]);
  const [recordsLoading, setRecordsLoading] = useState(false);

  useEffect(() => {
    if (id) {
      fetchDetail();
      fetchCheckRecords();
    }
  }, [id]);

  const fetchDetail = async () => {
    if (!id) return;
    setLoading(true);
    try {
      const res = await patrolService.getTaskDetail(id);
      const data: any = (res as any)?.data ?? res;
      setDetail(data);
    } catch (error: any) {
      message.error(error.message || '获取任务详情失败');
    } finally {
      setLoading(false);
    }
  };

  const fetchCheckRecords = async () => {
    if (!id) return;
    setRecordsLoading(true);
    try {
      const res = await patrolService.getTaskCheckRecords(id);
      const data: any = (res as any)?.data ?? res;
      if (Array.isArray(data)) {
        setCheckRecords(data);
      }
    } catch (error) {
      console.error('获取排查记录失败:', error);
    } finally {
      setRecordsLoading(false);
    }
  };

  const handleCancelTask = () => {
    if (!detail?.taskInfo) return;
    modal.confirm({
      title: '确认取消该任务?',
      icon: <StopOutlined />,
      content: `任务: ${detail.taskInfo.title}`,
      okText: '确认取消',
      cancelText: '返回',
      onOk: async () => {
        try {
          await patrolService.cancelTask(detail.taskInfo.id);
          message.success('任务已取消');
          fetchDetail();
        } catch (error: any) {
          message.error(error.message || '取消失败');
        }
      },
    });
  };

  const handleRetry = () => {
    fetchDetail();
    fetchCheckRecords();
  };

  const taskInfo = detail?.taskInfo;
  const gridMember = detail?.gridMember;

  const buildTimeline = (task: PatrolTask) => {
    const items: { time: string; status: string; title: string; description?: string }[] = [];
    if (task.createTime) {
      items.push({
        time: task.createTime,
        status: 'pending',
        title: '任务创建',
        description: `由 ${task.creatorName || '系统'} 创建`,
      });
    }
    if (task.status !== 'pending' && task.startTime) {
      items.push({
        time: task.startTime,
        status: 'assigned',
        title: '任务下发',
        description: `已下发给 ${task.gridMemberName || '网格员'}`,
      });
    }
    if (task.actualStartTime) {
      items.push({
        time: task.actualStartTime,
        status: 'in_progress',
        title: '开始排查',
        description: '网格员已开始执行排查任务',
      });
    }
    if (task.status === 'completed' && task.actualEndTime) {
      items.push({
        time: task.actualEndTime,
        status: 'completed',
        title: '任务完成',
        description: task.result || '排查任务已完成',
      });
    }
    if (task.status === 'cancelled') {
      items.push({
        time: task.updateTime || new Date().toISOString(),
        status: 'cancelled',
        title: '任务取消',
        description: task.remark || '任务已取消',
      });
    }
    return items;
  };

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '100px 0' }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!detail || !taskInfo) {
    return (
      <div style={{ padding: 24 }}>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate(-1)} style={{ marginBottom: 16 }}>
          返回列表
        </Button>
        <Empty description="任务不存在或已删除" />
      </div>
    );
  }

  const timelineItems = buildTimeline(taskInfo);

  return (
    <div style={{ padding: 24 }}>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/patrol/task')}>
            返回列表
          </Button>
          <Button icon={<ReloadOutlined />} onClick={handleRetry}>
            刷新
          </Button>
        </Space>
        <Space>
          {taskInfo.status !== 'completed' && taskInfo.status !== 'cancelled' && (
            <Button danger icon={<StopOutlined />} onClick={handleCancelTask}>
              取消任务
            </Button>
          )}
        </Space>
      </div>

      <Row gutter={[16, 16]}>
        <Col xs={24} lg={16}>
          <Card title="任务信息" bordered>
            <ProDescriptions<PatrolTask>
              column={2}
              dataSource={taskInfo}
              bordered
              size="middle"
            >
              <ProDescriptions.Item label="任务编号" dataIndex="taskNo" copyable />
              <ProDescriptions.Item label="任务状态">
                <Tag color={statusColorMap[taskInfo.status] || 'default'}>
                  {taskInfo.statusName || statusTextMap[taskInfo.status] || taskInfo.status}
                </Tag>
              </ProDescriptions.Item>
              <ProDescriptions.Item label="任务标题" dataIndex="title" span={2} />
              <ProDescriptions.Item label="任务类型">
                <Tag color="blue">
                  {taskInfo.typeName || typeTextMap[taskInfo.type] || taskInfo.type}
                </Tag>
              </ProDescriptions.Item>
              <ProDescriptions.Item label="优先级">
                <Tag color={priorityColorMap[taskInfo.priority] || 'default'}>
                  {taskInfo.priorityName || priorityTextMap[taskInfo.priority] || taskInfo.priority}
                </Tag>
              </ProDescriptions.Item>
              <ProDescriptions.Item label="所属区域">
                <Space>
                  <EnvironmentOutlined />
                  {taskInfo.area || '-'}
                </Space>
              </ProDescriptions.Item>
              <ProDescriptions.Item label="负责网格员">
                <Space>
                  <UserOutlined />
                  {taskInfo.gridMemberName || '-'}
                  {taskInfo.gridMemberPhone && (
                    <Text type="secondary">({taskInfo.gridMemberPhone})</Text>
                  )}
                </Space>
              </ProDescriptions.Item>
              <ProDescriptions.Item label="计划开始时间">
                {taskInfo.startTime ? dayjs(taskInfo.startTime).format('YYYY-MM-DD HH:mm') : '-'}
              </ProDescriptions.Item>
              <ProDescriptions.Item label="计划结束时间">
                {taskInfo.endTime ? dayjs(taskInfo.endTime).format('YYYY-MM-DD HH:mm') : '-'}
              </ProDescriptions.Item>
              <ProDescriptions.Item label="截止时间">
                {taskInfo.deadline ? dayjs(taskInfo.deadline).format('YYYY-MM-DD HH:mm') : '-'}
              </ProDescriptions.Item>
              <ProDescriptions.Item label="实际开始时间">
                {taskInfo.actualStartTime ? dayjs(taskInfo.actualStartTime).format('YYYY-MM-DD HH:mm') : '-'}
              </ProDescriptions.Item>
              <ProDescriptions.Item label="实际结束时间">
                {taskInfo.actualEndTime ? dayjs(taskInfo.actualEndTime).format('YYYY-MM-DD HH:mm') : '-'}
              </ProDescriptions.Item>
              <ProDescriptions.Item label="任务描述" dataIndex="description" span={2} />
              <ProDescriptions.Item label="排查要求" dataIndex="requirement" span={2} />
              <ProDescriptions.Item label="排查结果" dataIndex="result" span={2} />
              <ProDescriptions.Item label="备注" dataIndex="remark" span={2} />
              <ProDescriptions.Item label="创建时间">
                {taskInfo.createTime ? dayjs(taskInfo.createTime).format('YYYY-MM-DD HH:mm:ss') : '-'}
              </ProDescriptions.Item>
              <ProDescriptions.Item label="创建人" dataIndex="creatorName" />
            </ProDescriptions>
          </Card>

          <Divider />

          <Card
            title={
              <Space>
                <FileTextOutlined />
                排查记录
              </Space>
            }
            bordered
          >
            {recordsLoading ? (
              <div style={{ textAlign: 'center', padding: '40px 0' }}>
                <Spin size="small" />
              </div>
            ) : checkRecords.length > 0 ? (
              <List
                dataSource={checkRecords}
                renderItem={(record) => (
                  <List.Item key={record.id}>
                    <List.Item.Meta
                      avatar={<Avatar icon={<EnvironmentOutlined />} />}
                      title={
                        <Space>
                          <span>{record.location || '未知位置'}</span>
                          {record.issueLevel && (
                            <Tag color={record.issueLevel === 'high' ? 'red' : record.issueLevel === 'medium' ? 'orange' : 'default'}>
                              {record.issueLevelName || record.issueLevel}
                            </Tag>
                          )}
                        </Space>
                      }
                      description={
                        <div>
                          <div style={{ marginBottom: 8 }}>
                            <Text type="secondary">
                              {record.checkTime ? dayjs(record.checkTime).format('YYYY-MM-DD HH:mm:ss') : ''}
                            </Text>
                          </div>
                          {record.description && (
                            <div style={{ marginBottom: 8 }}>{record.description}</div>
                          )}
                          {record.issue && (
                            <div style={{ marginBottom: 8, color: '#ff4d4f' }}>
                              发现问题: {record.issue}
                            </div>
                          )}
                          {record.handleResult && (
                            <div style={{ color: '#52c41a' }}>
                              处理结果: {record.handleResult}
                            </div>
                          )}
                          {record.images && record.images.length > 0 && (
                            <div style={{ marginTop: 8 }}>
                              <Image.PreviewGroup>
                                <Space wrap>
                                  {record.images.map((img, idx) => (
                                    <Image
                                      key={idx}
                                      width={80}
                                      height={80}
                                      src={img}
                                      style={{ objectFit: 'cover', borderRadius: 4 }}
                                    />
                                  ))}
                                </Space>
                              </Image.PreviewGroup>
                            </div>
                          )}
                        </div>
                      }
                    />
                  </List.Item>
                )}
              />
            ) : (
              <Empty description="暂无排查记录" />
            )}
          </Card>
        </Col>

        <Col xs={24} lg={8}>
          {gridMember && (
            <Card
              title={
                <Space>
                  <UserOutlined />
                  网格员信息
                </Space>
              }
              bordered
              style={{ marginBottom: 16 }}
            >
              <div style={{ textAlign: 'center', marginBottom: 16 }}>
                <Avatar
                  size={64}
                  src={gridMember.avatar}
                  style={{ fontSize: 28, background: '#1677ff' }}
                  icon={<UserOutlined />}
                />
                <div style={{ marginTop: 8, fontWeight: 500, fontSize: 16 }}>
                  {gridMember.name}
                </div>
                <div style={{ color: '#999', fontSize: 12 }}>
                  {gridMember.memberNo}
                </div>
              </div>
              <Descriptions column={1} size="small" bordered>
                <Descriptions.Item label="联系电话">{gridMember.phone || '-'}</Descriptions.Item>
                <Descriptions.Item label="所属区域">{gridMember.area || '-'}</Descriptions.Item>
                <Descriptions.Item label="当前积分">{gridMember.points || 0}</Descriptions.Item>
                <Descriptions.Item label="累计积分">{gridMember.totalPoints || 0}</Descriptions.Item>
                <Descriptions.Item label="完成任务">{gridMember.completedTaskCount || 0} 个</Descriptions.Item>
                <Descriptions.Item label="加入时间">
                  {gridMember.joinDate ? dayjs(gridMember.joinDate).format('YYYY-MM-DD') : '-'}
                </Descriptions.Item>
              </Descriptions>
            </Card>
          )}

          <Card
            title={
              <Space>
                <ClockCircleOutlined />
                任务进度
              </Space>
            }
            bordered
          >
            <Timeline
              items={timelineItems.map((item) => ({
                color: getTimelineColor(item.status),
                dot: getTimelineIcon(item.status),
                children: (
                  <div>
                    <div style={{ fontWeight: 500 }}>{item.title}</div>
                    {item.description && (
                      <div style={{ fontSize: 12, color: '#999', marginTop: 4 }}>
                        {item.description}
                      </div>
                    )}
                    <div style={{ fontSize: 12, color: '#666', marginTop: 4 }}>
                      {dayjs(item.time).format('YYYY-MM-DD HH:mm:ss')}
                    </div>
                  </div>
                ),
              }))}
            />
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default TaskDetail;
