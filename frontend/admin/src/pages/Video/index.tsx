import React, { useState, useEffect } from 'react';
import {
  Card,
  Table,
  Button,
  Space,
  Tag,
  Modal,
  Form,
  Input,
  InputNumber,
  DatePicker,
  Select,
  Switch,
  message,
  Tooltip,
  Row,
  Col,
  Statistic,
} from 'antd';
import {
  PlusOutlined,
  VideoCameraOutlined,
  CloudOutlined,
  TeamOutlined,
  FileTextOutlined,
  EyeOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import { videoApi, queueApi } from '../../services/video';

const { RangePicker } = DatePicker;

const VideoMediation: React.FC = () => {
  const [rooms, setRooms] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [roomDetailVisible, setRoomDetailVisible] = useState(false);
  const [currentRoom, setCurrentRoom] = useState<any>(null);
  const [queueStatus, setQueueStatus] = useState<any>(null);
  const [form] = Form.useForm();
  const [selectedCaseId, setSelectedCaseId] = useState<number>(0);

  useEffect(() => {
    fetchRooms();
    fetchQueueStatus();
  }, []);

  const fetchRooms = async () => {
    setLoading(true);
    try {
      const res: any = await videoApi.getRoomList(selectedCaseId || 0);
      setRooms(res.data || []);
    } catch {
      // Ignore
    } finally {
      setLoading(false);
    }
  };

  const fetchQueueStatus = async () => {
    try {
      const res: any = await queueApi.getStatus();
      setQueueStatus(res.data);
    } catch {
      // Ignore
    }
  };

  const handleCreateRoom = async (values: any) => {
    try {
      const res: any = await videoApi.createRoom(values.caseId, {
        title: values.title,
        scheduledTime: values.scheduledTime.format('YYYY-MM-DD HH:mm:ss'),
        participantIds: values.participantIds,
        password: values.password,
        duration: values.duration,
        virtualBg: values.virtualBg,
        beauty: values.beauty,
      });
      message.success('视频房间创建成功');
      setCreateModalVisible(false);
      form.resetFields();
      fetchRooms();
    } catch {
      message.error('创建视频房间失败');
    }
  };

  const handleViewDetail = async (record: any) => {
    try {
      const res: any = await videoApi.getRoomDetail(record.case_id || 0, record.room_id);
      setCurrentRoom(res.data);
      setRoomDetailVisible(true);
    } catch {
      message.error('获取房间详情失败');
    }
  };

  const statusMap: Record<number, { color: string; text: string }> = {
    10: { color: 'default', text: '未开始' },
    20: { color: 'processing', text: '进行中' },
    30: { color: 'success', text: '已结束' },
    40: { color: 'error', text: '已取消' },
  };

  const recordStatusMap: Record<number, { color: string; text: string }> = {
    0: { color: 'default', text: '未录制' },
    1: { color: 'processing', text: '录制中' },
    2: { color: 'success', text: '已结束' },
    3: { color: 'error', text: '失败' },
  };

  const columns = [
    {
      title: '房间号',
      dataIndex: 'room_id',
      key: 'roomId',
      width: 140,
    },
    {
      title: '案件编号',
      dataIndex: 'case_no',
      key: 'caseNo',
      width: 140,
    },
    {
      title: '标题',
      dataIndex: 'title',
      key: 'title',
      ellipsis: true,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: number) => {
        const s = statusMap[status] || { color: 'default', text: '未知' };
        return <Tag color={s.color}>{s.text}</Tag>;
      },
    },
    {
      title: '录制',
      dataIndex: 'record_status',
      key: 'recordStatus',
      width: 100,
      render: (status: number) => {
        const s = recordStatusMap[status] || { color: 'default', text: '未知' };
        return <Tag color={s.color}>{s.text}</Tag>;
      },
    },
    {
      title: '会议纪要',
      dataIndex: 'has_meeting_minutes',
      key: 'minutes',
      width: 100,
      render: (v: number) => v ? <Tag color="green"><FileTextOutlined /> 已生成</Tag> : <Tag>未生成</Tag>,
    },
    {
      title: '预约时间',
      dataIndex: 'scheduled_time',
      key: 'scheduledTime',
      width: 170,
      render: (v: string) => v ? dayjs(v).format('YYYY-MM-DD HH:mm') : '-',
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      render: (_: any, record: any) => (
        <Space>
          <Tooltip title="查看详情">
            <Button size="small" icon={<EyeOutlined />} onClick={() => handleViewDetail(record)} />
          </Tooltip>
        </Space>
      ),
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={6}>
          <Card>
            <Statistic title="视频房间总数" value={rooms.length} prefix={<VideoCameraOutlined />} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="进行中"
              value={rooms.filter((r) => r.status === 20).length}
              valueStyle={{ color: '#1890ff' }}
              prefix={<VideoCameraOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="录制中"
              value={rooms.filter((r) => r.record_status === 1).length}
              valueStyle={{ color: '#cf1322' }}
              prefix={<CloudOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="排队人数"
              value={queueStatus?.queueLength || 0}
              valueStyle={{ color: '#faad14' }}
              prefix={<TeamOutlined />}
              suffix={
                <span style={{ fontSize: 12, color: '#999' }}>
                  (预计等待{queueStatus?.estimatedWait || 0}分钟)
                </span>
              }
            />
          </Card>
        </Col>
      </Row>

      <Card
        title="视频调解管理"
        extra={
          <Space>
            <InputNumber
              placeholder="案件ID筛选"
              style={{ width: 140 }}
              value={selectedCaseId || undefined}
              onChange={(v) => setSelectedCaseId(v || 0)}
            />
            <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateModalVisible(true)}>
              创建视频调解
            </Button>
          </Space>
        }
      >
        <Table
          columns={columns}
          dataSource={rooms}
          rowKey="room_id"
          loading={loading}
          pagination={{ pageSize: 10 }}
        />
      </Card>

      <Modal
        title="创建视频调解"
        open={createModalVisible}
        onCancel={() => setCreateModalVisible(false)}
        onOk={() => form.submit()}
        width={600}
      >
        <Form form={form} layout="vertical" onFinish={handleCreateRoom}>
          <Form.Item name="caseId" label="案件ID" rules={[{ required: true }]}>
            <InputNumber style={{ width: '100%' }} placeholder="请输入案件ID" />
          </Form.Item>
          <Form.Item name="title" label="调解标题" rules={[{ required: true }]}>
            <Input placeholder="请输入调解标题" />
          </Form.Item>
          <Form.Item name="scheduledTime" label="预约时间" rules={[{ required: true }]}>
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="duration" label="时长(分钟)" initialValue={60}>
            <InputNumber min={10} max={240} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="participantIds" label="参与人ID" rules={[{ required: true }]}>
            <Select mode="multiple" placeholder="选择参与人" />
          </Form.Item>
          <Form.Item name="password" label="房间密码">
            <Input placeholder="留空自动生成" />
          </Form.Item>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="virtualBg" label="虚拟背景" valuePropName="checked" initialValue={false}>
                <Switch />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="beauty" label="美颜" valuePropName="checked" initialValue={false}>
                <Switch />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Modal>

      <Modal
        title="视频房间详情"
        open={roomDetailVisible}
        onCancel={() => setRoomDetailVisible(false)}
        footer={null}
        width={700}
      >
        {currentRoom && (
          <div>
            <Row gutter={16}>
              <Col span={12}>
                <Card size="small" title="基本信息">
                  <p><strong>房间号:</strong> {currentRoom.room_id}</p>
                  <p><strong>TRTC房间号:</strong> {currentRoom.trtc_room_id}</p>
                  <p><strong>状态:</strong> <Tag color={statusMap[currentRoom.status]?.color}>{statusMap[currentRoom.status]?.text}</Tag></p>
                  <p><strong>录制状态:</strong> <Tag color={recordStatusMap[currentRoom.record_status]?.color}>{recordStatusMap[currentRoom.record_status]?.text}</Tag></p>
                </Card>
              </Col>
              <Col span={12}>
                <Card size="small" title="功能设置">
                  <p>虚拟背景: {currentRoom.virtual_bg_enabled ? <Tag color="green">已开启</Tag> : <Tag>未开启</Tag>}</p>
                  <p>美颜: {currentRoom.beauty_enabled ? <Tag color="green">已开启</Tag> : <Tag>未开启</Tag>}</p>
                  <p>屏幕共享: {currentRoom.screen_share_user_id ? <Tag color="blue">进行中</Tag> : <Tag>无</Tag>}</p>
                  <p>会议纪要: {currentRoom.has_meeting_minutes ? <Tag color="green">已生成</Tag> : <Tag>未生成</Tag>}</p>
                </Card>
              </Col>
            </Row>
            {currentRoom.participants && (
              <Card size="small" title="参与人" style={{ marginTop: 16 }}>
                {currentRoom.participants.map((p: any, i: number) => (
                  <Tag key={i} color={p.join_status === 20 ? 'green' : 'default'}>
                    {p.user_name} {p.is_creator ? '(主持人)' : ''}
                  </Tag>
                ))}
              </Card>
            )}
            {currentRoom.recordSegments && currentRoom.recordSegments.length > 0 && (
              <Card size="small" title="录制分段" style={{ marginTop: 16 }}>
                {currentRoom.recordSegments.map((seg: any, i: number) => (
                  <Tag key={i} color={seg.status === 1 ? 'processing' : 'success'}>
                    第{seg.segment_index + 1}段 {seg.duration_sec ? `(${Math.floor(seg.duration_sec / 60)}分${seg.duration_sec % 60}秒)` : ''}
                  </Tag>
                ))}
              </Card>
            )}
          </div>
        )}
      </Modal>
    </div>
  );
};

export default VideoMediation;
