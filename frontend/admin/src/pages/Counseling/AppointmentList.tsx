import React, { useState, useRef, useEffect } from 'react';
import {
  Button,
  Tag,
  Space,
  App,
  Modal,
  Card,
  Avatar,
  Rate,
  Form,
  Input,
  Select,
  Switch,
  Drawer,
  Row,
  Col,
  Descriptions,
  Calendar,
  Badge,
  Radio,
  Alert,
  List,
  Tooltip,
  Divider,
  Empty,
  Statistic,
} from 'antd';
import {
  PlusOutlined,
  EyeOutlined,
  UserOutlined,
  WarningOutlined,
  CalendarOutlined,
  CheckOutlined,
  CloseOutlined,
  TeamOutlined,
  EyeInvisibleOutlined,
  BulbOutlined,
  StopOutlined,
} from '@ant-design/icons';
import {
  ProTable,
  ProFormSelect,
  ProFormText,
  ProFormTextArea,
  ModalForm,
  DrawerForm,
  ProFormSwitch,
  ProFormDatePicker,
  ProFormRadio,
} from '@ant-design/pro-components';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { useNavigate, useSearchParams } from 'react-router-dom';
import {
  counselingService,
  CounselorAppointment,
  Counselor,
  AvailableSlot,
} from '../../services/counseling';
import dayjs, { Dayjs } from 'dayjs';

const { confirm } = Modal;
const { Option } = Select;
const { TextArea } = Input;

const STATUS_MAP: Record<number, string> = {
  10: '待确认',
  20: '已确认',
  30: '咨询中',
  40: '已完成',
  50: '已取消',
  60: '已过期',
};

const STATUS_COLOR: Record<number, string> = {
  10: 'orange',
  20: 'blue',
  30: 'processing',
  40: 'success',
  50: 'default',
  60: 'default',
};

const CONSULT_TYPE_MAP: Record<number, string> = {
  1: '线上视频',
  2: '线上语音',
  3: '线下面谈',
};

const CONSULT_TYPE_COLOR: Record<number, string> = {
  1: 'blue',
  2: 'cyan',
  3: 'purple',
};

const EMERGENCY_LEVEL_MAP: Record<number, string> = {
  0: '普通',
  1: '关注',
  2: '紧急',
  3: '高危',
};

const EMERGENCY_LEVEL_COLOR: Record<number, string> = {
  0: 'default',
  1: 'orange',
  2: 'red',
  3: 'magenta',
};

const CONCERN_TYPES = [
  { label: '家庭暴力', value: '家庭暴力' },
  { label: '心理创伤', value: '心理创伤' },
  { label: '焦虑抑郁', value: '焦虑抑郁' },
  { label: '家庭关系', value: '家庭关系' },
  { label: '青少年心理', value: '青少年心理' },
  { label: '职场压力', value: '职场压力' },
  { label: '情绪管理', value: '情绪管理' },
  { label: '其他', value: '其他' },
];

const pick = <T,>(obj: T, keys: (keyof T | string)[]): any => {
  const o = obj as any;
  for (const k of keys) {
    if (o[k] !== undefined && o[k] !== null && o[k] !== '') return o[k];
  }
  return undefined;
};

const AppointmentList: React.FC = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { message } = App.useApp();
  const actionRef = useRef<ActionType>();
  const [viewMode, setViewMode] = useState<'list' | 'calendar'>('list');
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [detailDrawerOpen, setDetailDrawerOpen] = useState(false);
  const [detailData, setDetailData] = useState<CounselorAppointment | null>(null);
  const [cancelModalOpen, setCancelModalOpen] = useState(false);
  const [currentAppointment, setCurrentAppointment] = useState<CounselorAppointment | null>(null);
  const [ratingModalOpen, setRatingModalOpen] = useState(false);
  const [createForm] = Form.useForm();
  const [cancelForm] = Form.useForm();
  const [ratingForm] = Form.useForm();

  const [counselors, setCounselors] = useState<Counselor[]>([]);
  const [selectedCounselor, setSelectedCounselor] = useState<Counselor | null>(null);
  const [selectedDate, setSelectedDate] = useState<Dayjs>(dayjs());
  const [availableSlots, setAvailableSlots] = useState<AvailableSlot[]>([]);
  const [selectedSlot, setSelectedSlot] = useState<AvailableSlot | null>(null);
  const [loadingSlots, setLoadingSlots] = useState(false);
  const [emergencyDetected, setEmergencyDetected] = useState<{ triggered: boolean; words: string[] }>({
    triggered: false,
    words: [],
  });
  const [calendarAppointments, setCalendarAppointments] = useState<CounselorAppointment[]>([]);

  const initialCounselorId = searchParams.get('counselorId');

  useEffect(() => {
    counselingService
      .getCounselorList({ status: 1, pageSize: 100 })
      .then((res: any) => {
        const data: any = (res as any)?.data ?? res;
        const list: Counselor[] = data.list || [];
        setCounselors(list);
        if (initialCounselorId) {
          const found = list.find((c) => String(c.id) === String(initialCounselorId));
          if (found) {
            setSelectedCounselor(found);
            createForm.setFieldsValue({ counselorId: String(found.id) });
            loadAvailableSlots(found.id as any, dayjs().format('YYYY-MM-DD'));
          }
        }
      })
      .catch(() => {});
  }, [initialCounselorId]);

  const loadAvailableSlots = (counselorId: string | number, date: string) => {
    setLoadingSlots(true);
    counselingService
      .getCounselorAvailableSlots(counselorId, date)
      .then((res: any) => {
        const data: any = (res as any)?.data ?? res;
        setAvailableSlots(data.slots || []);
      })
      .finally(() => setLoadingSlots(false));
  };

  const loadCalendarAppointments = (month: Dayjs) => {
    const start = month.startOf('month').format('YYYY-MM-DD');
    const end = month.endOf('month').format('YYYY-MM-DD');
    counselingService
      .getAppointmentList({ startDate: start, endDate: end, pageSize: 1000 })
      .then((res: any) => {
        const data: any = (res as any)?.data ?? res;
        setCalendarAppointments(data.list || []);
      })
      .catch(() => {});
  };

  const checkEmergencyKeywords = (text: string) => {
    const keywords = [
      '自杀', '想死', '不想活', '活不下去', '结束生命', '自残', '割腕',
      '跳楼', '跳河', '自伤', '自我伤害', '不想活了', '活够了',
    ];
    const found: string[] = [];
    keywords.forEach((kw) => {
      if (text.includes(kw)) found.push(kw);
    });
    setEmergencyDetected({ triggered: found.length > 0, words: found });
    return found.length > 0;
  };

  const columns: ProColumns<CounselorAppointment>[] = [
    {
      title: '预约编号',
      dataIndex: 'appointmentNo',
      width: 160,
      copyable: true,
      fixed: 'left',
    },
    {
      title: '紧急程度',
      dataIndex: 'isEmergency',
      width: 100,
      search: false,
      render: (_, row) => {
        const isEm = pick(row, ['is_emergency', 'isEmergency']) == 1;
        const level = pick(row, ['emergency_level', 'emergencyLevel']) || 0;
        const levelName = pick(row, ['emergency_level_name', 'emergencyLevelName']) || EMERGENCY_LEVEL_MAP[level as number];
        if (isEm) {
          return (
            <Tag color={EMERGENCY_LEVEL_COLOR[level as number] || 'red'} icon={<WarningOutlined />}>
              {levelName}
            </Tag>
          );
        }
        return <Tag color="default">普通</Tag>;
      },
    },
    {
      title: '心理咨询师',
      dataIndex: 'counselorName',
      width: 140,
      render: (_, row) => {
        const name = pick(row, ['counselor_name', 'counselorRealName', 'counselorName']);
        const title = pick(row, ['counselorTitle']);
        return (
          <Space direction="vertical" size={0}>
            <span style={{ fontWeight: 500 }}>{name}</span>
            {title && <span style={{ color: '#999', fontSize: 12 }}>{title}</span>}
          </Space>
        );
      },
    },
    {
      title: '当事人',
      dataIndex: 'partyName',
      width: 140,
      search: false,
      render: (_, row) => {
        const isAnon = pick(row, ['is_anonymous', 'isAnonymous']) == 1;
        const display = pick(row, ['party_name_display', 'partyNameDisplay']);
        const name = pick(row, ['party_name', 'partyName']);
        const code = pick(row, ['anonymous_code', 'anonymousCode']);
        if (isAnon) {
          return (
            <Space>
              <Tag color="purple" icon={<EyeInvisibleOutlined />}>
                匿名模式
              </Tag>
              <span style={{ color: '#666' }}>{display || code || '匿名用户'}</span>
            </Space>
          );
        }
        return (
          <Space direction="vertical" size={0}>
            <span>{name || '-'}</span>
            <span style={{ color: '#999', fontSize: 12 }}>
              {pick(row, ['party_phone_display', 'partyPhoneDisplay']) || pick(row, ['party_phone', 'partyPhone']) || ''}
            </span>
          </Space>
        );
      },
    },
    {
      title: '预约日期',
      dataIndex: 'appointmentDate',
      width: 120,
      render: (_, row) => pick(row, ['appointment_date', 'appointmentDate']),
    },
    {
      title: '预约时段',
      dataIndex: 'startTime',
      width: 140,
      search: false,
      render: (_, row) => {
        const start = pick(row, ['start_time', 'startTime']);
        const end = pick(row, ['end_time', 'endTime']);
        return `${start?.slice(0, 5) || ''} - ${end?.slice(0, 5) || ''}`;
      },
    },
    {
      title: '咨询方式',
      dataIndex: 'consultationType',
      width: 100,
      search: false,
      render: (_, row) => {
        const type = pick(row, ['consultation_type', 'consultationType']) as number;
        const name = pick(row, ['consultation_type_name', 'consultationTypeName']) || CONSULT_TYPE_MAP[type];
        return <Tag color={CONSULT_TYPE_COLOR[type] || 'default'}>{name}</Tag>;
      },
    },
    {
      title: '问题类型',
      dataIndex: 'concernType',
      width: 120,
      search: false,
      render: (_, row) => {
        const type = pick(row, ['concern_type', 'concernType']);
        return type || '-';
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      valueEnum: {
        10: { text: '待确认', status: 'Warning' },
        20: { text: '已确认', status: 'Processing' },
        30: { text: '咨询中', status: 'Running' },
        40: { text: '已完成', status: 'Success' },
        50: { text: '已取消', status: 'Default' },
        60: { text: '已过期', status: 'Default' },
      },
      render: (_, row) => {
        const s = pick(row, ['status']) as number;
        const name = pick(row, ['status_name', 'statusName']) || STATUS_MAP[s];
        return <Tag color={STATUS_COLOR[s] || 'default'}>{name}</Tag>;
      },
    },
    {
      title: '操作',
      key: 'option',
      valueType: 'option',
      width: 260,
      fixed: 'right',
      render: (_, record) => {
        const id = pick(record, ['id']);
        const status = pick(record, ['status']) as number;
        const ratingSubmitted = pick(record, ['rating_submitted', 'ratingSubmitted']) == 1;
        return [
          <Button
            type="link"
            key="view"
            icon={<EyeOutlined />}
            onClick={async () => {
              try {
                const res: any = await counselingService.getAppointmentDetail(id);
                const data: any = (res as any)?.data ?? res;
                setDetailData(data);
                setDetailDrawerOpen(true);
              } catch (e: any) {
                message.error(e.message || '获取详情失败');
              }
            }}
          >
            查看
          </Button>,
          status === 10 && (
            <Button
              type="link"
              key="confirm"
              icon={<CheckOutlined />}
              onClick={async () => {
                try {
                  await counselingService.updateAppointment(id, { status: 20 });
                  message.success('已确认预约');
                  actionRef.current?.reload();
                } catch (e: any) {
                  message.error(e.message || '操作失败');
                }
              }}
            >
              确认
            </Button>
          ),
          status === 20 && (
            <Button
              type="link"
              key="start"
              icon={<CalendarOutlined />}
              onClick={async () => {
                try {
                  await counselingService.updateAppointment(id, { status: 30 });
                  message.success('已开始咨询');
                  actionRef.current?.reload();
                } catch (e: any) {
                  message.error(e.message || '操作失败');
                }
              }}
            >
              开始咨询
            </Button>
          ),
          status === 30 && (
            <Button
              type="link"
              key="complete"
              icon={<CheckOutlined />}
              onClick={async () => {
                try {
                  await counselingService.updateAppointment(id, { status: 40 });
                  message.success('已完成咨询');
                  actionRef.current?.reload();
                } catch (e: any) {
                  message.error(e.message || '操作失败');
                }
              }}
            >
              完成咨询
            </Button>
          ),
          (status === 10 || status === 20) && (
            <Button
              type="link"
              key="cancel"
              danger
              icon={<CloseOutlined />}
              onClick={() => {
                setCurrentAppointment(record);
                cancelForm.resetFields();
                setCancelModalOpen(true);
              }}
            >
              取消
            </Button>
          ),
          status === 40 && !ratingSubmitted && (
            <Button
              type="link"
              key="rate"
              icon={<BulbOutlined />}
              onClick={() => {
                setCurrentAppointment(record);
                ratingForm.resetFields();
                setRatingModalOpen(true);
              }}
            >
              评价
            </Button>
          ),
        ];
      },
    },
  ];

  const dateCellRender = (value: Dayjs) => {
    const dateStr = value.format('YYYY-MM-DD');
    const listData = calendarAppointments.filter(
      (a) => (a.appointmentDate || (a as any).appointment_date) === dateStr,
    );
    return (
      <ul className="events">
        {listData.slice(0, 3).map((item) => {
          const status = (item.status || (item as any).status) as number;
          const isEm = (item.isEmergency || (item as any).is_emergency) == 1;
          const counselor = item.counselorName || (item as any).counselor_name || '';
          const start = (item.startTime || (item as any).start_time || '').slice(0, 5);
          return (
            <li key={String(item.id)}>
              <Badge
                status={isEm ? 'error' : status === 40 ? 'success' : status === 10 ? 'warning' : 'processing'}
                text={
                  <Tooltip title={`${start} ${counselor}`}>
                    <span style={{ fontSize: 11 }}>{start} {counselor.slice(0, 2)}</span>
                  </Tooltip>
                }
              />
            </li>
          );
        })}
        {listData.length > 3 && (
          <li>
            <span style={{ color: '#999', fontSize: 11 }}>+{listData.length - 3} 更多</span>
          </li>
        )}
      </ul>
    );
  };

  return (
    <>
      <Card
        style={{ marginBottom: 16 }}
        bodyStyle={{ padding: '12px 16px' }}
        title={
          <Space>
            <span style={{ fontSize: 16, fontWeight: 600 }}>心理咨询预约管理</span>
            <Radio.Group
              value={viewMode}
              onChange={(e) => {
                setViewMode(e.target.value);
                if (e.target.value === 'calendar') {
                  loadCalendarAppointments(dayjs());
                }
              }}
              size="small"
            >
              <Radio.Button value="list">列表模式</Radio.Button>
              <Radio.Button value="calendar">日历模式</Radio.Button>
            </Radio.Group>
          </Space>
        }
        extra={
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => {
              createForm.resetFields();
              setSelectedSlot(null);
              setEmergencyDetected({ triggered: false, words: [] });
              if (initialCounselorId) {
                setSelectedDate(dayjs());
                loadAvailableSlots(initialCounselorId, dayjs().format('YYYY-MM-DD'));
              }
              setCreateModalOpen(true);
            }}
          >
            新增预约
          </Button>
        }
      />

      {viewMode === 'list' ? (
        <ProTable<CounselorAppointment>
          columns={columns}
          actionRef={actionRef}
          cardBordered
          rowKey="id"
          search={{
            labelWidth: 'auto',
            defaultCollapsed: false,
          }}
          dateFormatter="string"
          headerTitle="预约列表"
          request={async (params, sort, filter) => {
            try {
              const startDate = (params as any).appointmentDate?.[0];
              const endDate = (params as any).appointmentDate?.[1];
              const res: any = await counselingService.getAppointmentList({
                page: params.current,
                pageSize: params.pageSize,
                keyword: params.keyword as string,
                status: (params.status as unknown as number) || undefined,
                counselorId: (params.counselorId as string) || undefined,
                isEmergency: (params.isEmergency as unknown as number) || undefined,
                startDate,
                endDate,
                sortBy: 'is_emergency',
                sortOrder: 'desc',
              });
              const data: any = (res as any)?.data ?? res;
              return {
                data: data.list || [],
                success: true,
                total: data.total || 0,
              };
            } catch (error) {
              return { data: [], success: false, total: 0 };
            }
          }}
          columnsState={{
            persistenceKey: 'counseling-appointment-columns',
            persistenceType: 'localStorage',
          }}
          pagination={{
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: total => `共 ${total} 条记录`,
          }}
          scroll={{ x: 1800 }}
        />
      ) : (
        <Card>
          <Calendar
            cellRender={dateCellRender}
            onPanelChange={(value) => loadCalendarAppointments(value)}
          />
        </Card>
      )}

      <ModalForm
        title="新增心理咨询预约"
        open={createModalOpen}
        onOpenChange={setCreateModalOpen}
        form={createForm}
        width={800}
        modalProps={{ destroyOnClose: true, maskClosable: false }}
        onFinish={async (values: any) => {
          if (!selectedSlot) {
            message.warning('请选择预约时段');
            return false;
          }
          try {
            const res: any = await counselingService.createAppointment({
              counselorId: values.counselorId,
              caseId: values.caseId ? Number(values.caseId) : undefined,
              partyName: values.partyName,
              partyPhone: values.partyPhone,
              partyIdCard: values.partyIdCard,
              isAnonymous: values.isAnonymous ? 1 : 0,
              appointmentDate: selectedDate.format('YYYY-MM-DD'),
              startTime: selectedSlot.startTime,
              endTime: selectedSlot.endTime,
              consultationType: values.consultationType,
              concernType: values.concernType,
              concernDescription: values.concernDescription,
              isEmergency: emergencyDetected.triggered || values.isEmergency ? 1 : 0,
            });
            const data: any = (res as any)?.data ?? res;
            if (data?.warning) {
              Modal.warning({
                title: '紧急心理风险预警',
                icon: <WarningOutlined style={{ color: '#ff4d4f' }} />,
                content: (
                  <Space direction="vertical">
                    <Alert
                      type="error"
                      showIcon
                      message={data.warning}
                      description={`检测到风险关键词：${data.emergencyWords || emergencyDetected.words.join('、')}`}
                    />
                    <p style={{ marginBottom: 0 }}>
                      该预约已自动标记为
                      <Tag color="red" style={{ margin: '0 4px' }}>
                        紧急
                      </Tag>
                      级别，请优先联系处理！
                    </p>
                  </Space>
                ),
              });
            } else {
              message.success('预约创建成功');
            }
            actionRef.current?.reload();
            return true;
          } catch (error: any) {
            message.error(error.message || '创建失败');
            return false;
          }
        }}
      >
        <Row gutter={16}>
          <Col span={24}>
            <Form.Item
              label="选择心理咨询师"
              name="counselorId"
              rules={[{ required: true, message: '请选择心理咨询师' }]}
            >
              <Select
                placeholder="请选择心理咨询师"
                showSearch
                optionFilterProp="children"
                onChange={(val) => {
                  const counselor = counselors.find((c) => String(c.id) === String(val));
                  setSelectedCounselor(counselor || null);
                  setSelectedSlot(null);
                  loadAvailableSlots(val, selectedDate.format('YYYY-MM-DD'));
                }}
              >
                {counselors.map((c) => (
                  <Option key={String(c.id)} value={String(c.id)}>
                    <Space>
                      <span>{c.realName}</span>
                      {c.title && <Tag color="blue" style={{ fontSize: 11 }}>{c.title}</Tag>}
                      {(c as any).is_emergency_available == 1 && (
                        <Tag color="red" icon={<WarningOutlined />} style={{ fontSize: 11 }}>
                          可接紧急
                        </Tag>
                      )}
                      <Rate disabled allowHalf value={Number(c.ratingAvg || (c as any).rating_avg || 0)} style={{ fontSize: 12 }} />
                      <span style={{ color: '#999', fontSize: 12 }}>
                        ¥{Number(c.price || 0).toFixed(0)}
                      </span>
                    </Space>
                  </Option>
                ))}
              </Select>
            </Form.Item>
          </Col>

          <Col span={24}>
            <Form.Item label="选择日期" required>
              <div style={{ display: 'flex', gap: 16, alignItems: 'flex-start', flexWrap: 'wrap' }}>
                <Calendar
                  fullscreen={false}
                  style={{ width: 320, border: '1px solid #f0f0f0', borderRadius: 8 }}
                  value={selectedDate}
                  disabledDate={(current) => current.isBefore(dayjs().startOf('day'))}
                  onSelect={(date) => {
                    setSelectedDate(date);
                    setSelectedSlot(null);
                    if (selectedCounselor) {
                      loadAvailableSlots(selectedCounselor.id as any, date.format('YYYY-MM-DD'));
                    }
                  }}
                />
                <div style={{ flex: 1, minWidth: 280 }}>
                  <div style={{ marginBottom: 8, fontWeight: 500 }}>
                    {selectedDate.format('YYYY年MM月DD日')} 可预约时段
                  </div>
                  {!selectedCounselor ? (
                    <Empty description="请先选择心理咨询师" image={Empty.PRESENTED_IMAGE_SIMPLE} />
                  ) : loadingSlots ? (
                    <div style={{ padding: 24, textAlign: 'center', color: '#999' }}>加载中...</div>
                  ) : availableSlots.length === 0 ? (
                    <Empty description="当日无可预约时段" image={Empty.PRESENTED_IMAGE_SIMPLE} />
                  ) : (
                    <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
                      {availableSlots.map((slot, idx) => (
                        <Button
                          key={idx}
                          type={selectedSlot?.startTime === slot.startTime ? 'primary' : 'default'}
                          disabled={!slot.available}
                          danger={slot.available && emergencyDetected.triggered}
                          style={{
                            minWidth: 100,
                            borderColor: selectedSlot?.startTime === slot.startTime ? undefined : '#d9d9d9',
                          }}
                          onClick={() => slot.available && setSelectedSlot(slot)}
                        >
                          {slot.startTime.slice(0, 5)} - {slot.endTime.slice(0, 5)}
                          {!slot.available && <StopOutlined style={{ marginLeft: 4, fontSize: 11 }} />}
                        </Button>
                      ))}
                    </div>
                  )}
                  {selectedSlot && (
                    <Alert
                      type="success"
                      showIcon
                      style={{ marginTop: 12 }}
                      message={`已选择：${selectedSlot.startTime.slice(0, 5)} - ${selectedSlot.endTime.slice(0, 5)}`}
                    />
                  )}
                </div>
              </div>
            </Form.Item>
          </Col>

          <Col span={12}>
            <Form.Item label="当事人姓名" name="partyName">
              <Input placeholder="请输入当事人姓名" />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="联系电话" name="partyPhone">
              <Input placeholder="请输入联系电话" />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="身份证号" name="partyIdCard">
              <Input placeholder="选填" />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="关联案件ID" name="caseId">
              <Input placeholder="选填，关联纠纷案件" />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="咨询方式" name="consultationType" initialValue={1}>
              <Select>
                <Option value={1}>线上视频</Option>
                <Option value={2}>线上语音</Option>
                <Option value={3}>线下面谈</Option>
              </Select>
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="匿名模式" name="isAnonymous" valuePropName="checked" initialValue={false}>
              <Switch
                checkedChildren={<EyeInvisibleOutlined />}
                unCheckedChildren={<UserOutlined />}
                onChange={(checked) => {
                  if (checked) {
                    message.info('匿名模式：隐藏当事人姓名、电话、身份证，仅显示匿名编号');
                  }
                }}
              />
            </Form.Item>
          </Col>

          <Col span={24}>
            <Form.Item label="问题类型" name="concernType">
              <Select placeholder="请选择问题类型" allowClear>
                {CONCERN_TYPES.map((t) => (
                  <Option key={t.value} value={t.value}>{t.label}</Option>
                ))}
              </Select>
            </Form.Item>
          </Col>

          <Col span={24}>
            <Form.Item label="问题描述" name="concernDescription">
              <TextArea
                rows={4}
                placeholder="请详细描述需要咨询的问题。系统会自动检测紧急心理风险关键词（如：自杀、自残等）"
                onChange={(e) => checkEmergencyKeywords(e.target.value)}
              />
            </Form.Item>
          </Col>

          {emergencyDetected.triggered && (
            <Col span={24}>
              <Alert
                type="error"
                showIcon
                icon={<WarningOutlined />}
                message="检测到紧急心理风险！"
                description={
                  <Space direction="vertical">
                    <p style={{ marginBottom: 4 }}>
                      检测到风险关键词：
                      <Space size={4}>
                        {emergencyDetected.words.map((w) => (
                          <Tag key={w} color="red">{w}</Tag>
                        ))}
                      </Space>
                    </p>
                    <p style={{ marginBottom: 0 }}>
                      提交后该预约将被自动标记为
                      <Tag color="red" style={{ margin: '0 4px' }}>紧急</Tag>
                      ，建议优先处理！
                    </p>
                    <Form.Item
                      name="isEmergency"
                      valuePropName="checked"
                      initialValue={true}
                      style={{ marginTop: 8, marginBottom: 0 }}
                    >
                      <Switch
                        checkedChildren="紧急预约"
                        unCheckedChildren="普通预约"
                      />
                    </Form.Item>
                  </Space>
                }
              />
            </Col>
          )}

          <Col span={24}>
            {selectedCounselor && (
              <Card size="small" style={{ background: '#fafafa' }}>
                <Descriptions size="small" column={2} title={
                  <Space>
                    <Avatar size={32} src={selectedCounselor.avatar} icon={<UserOutlined />} />
                    <span style={{ fontWeight: 500 }}>{selectedCounselor.realName}</span>
                    {selectedCounselor.title && <Tag color="blue">{selectedCounselor.title}</Tag>}
                  </Space>
                }>
                  <Descriptions.Item label="专业领域">
                    <Space size={4} wrap>
                      {((selectedCounselor as any).specialtyTagList || (selectedCounselor as any).specialty_tag_list || []).map((t: string) => (
                        <Tag key={t} color="blue" style={{ fontSize: 11 }}>{t}</Tag>
                      ))}
                    </Space>
                  </Descriptions.Item>
                  <Descriptions.Item label="从业年限">
                    {selectedCounselor.yearsOfExperience || (selectedCounselor as any).years_of_experience || 0} 年
                  </Descriptions.Item>
                  <Descriptions.Item label="咨询方式">
                    <Space>
                      {((selectedCounselor as any).consultationTypeList || (selectedCounselor as any).consultation_type_list || []).map((t: any) => (
                        <Tag key={t.type} color={CONSULT_TYPE_COLOR[t.type]}>{t.name}</Tag>
                      ))}
                    </Space>
                  </Descriptions.Item>
                  <Descriptions.Item label="评分">
                    <Rate disabled allowHalf value={Number(selectedCounselor.ratingAvg || (selectedCounselor as any).rating_avg || 0)} style={{ fontSize: 14 }} />
                    <span style={{ color: '#faad14', marginLeft: 4 }}>
                      {Number(selectedCounselor.ratingAvg || (selectedCounselor as any).rating_avg || 0).toFixed(1)}
                    </span>
                  </Descriptions.Item>
                </Descriptions>
              </Card>
            )}
          </Col>
        </Row>
      </ModalForm>

      <ModalForm
        title="取消预约"
        open={cancelModalOpen}
        onOpenChange={setCancelModalOpen}
        form={cancelForm}
        width={480}
        modalProps={{ destroyOnClose: true }}
        onFinish={async (values: any) => {
          if (!currentAppointment) return false;
          try {
            await counselingService.cancelAppointment(currentAppointment.id as any, values.reason);
            message.success('已取消预约');
            actionRef.current?.reload();
            return true;
          } catch (error: any) {
            message.error(error.message || '操作失败');
            return false;
          }
        }}
      >
        <Form.Item label="取消原因" name="reason" rules={[{ required: true, message: '请输入取消原因' }]}>
          <TextArea rows={4} placeholder="请输入取消预约的原因" />
        </Form.Item>
      </ModalForm>

      <ModalForm
        title="咨询评价"
        open={ratingModalOpen}
        onOpenChange={setRatingModalOpen}
        form={ratingForm}
        width={560}
        modalProps={{ destroyOnClose: true }}
        onFinish={async (values: any) => {
          if (!currentAppointment) return false;
          try {
            await counselingService.createRating({
              appointmentId: currentAppointment.id as any,
              counselorId: currentAppointment.counselorId || (currentAppointment as any).counselor_id as any,
              overallScore: values.overallScore,
              professionalScore: values.professionalScore,
              attitudeScore: values.attitudeScore,
              empathyScore: values.empathyScore,
              helpfulScore: values.helpfulScore,
              content: values.content,
              tags: values.tags?.join(','),
              isAnonymousRating: values.isAnonymousRating ? 1 : 0,
            });
            message.success('评价提交成功');
            actionRef.current?.reload();
            return true;
          } catch (error: any) {
            message.error(error.message || '提交失败');
            return false;
          }
        }}
      >
        <Form.Item
          label="总体评分"
          name="overallScore"
          rules={[{ required: true, message: '请评分' }]}
        >
          <Rate />
        </Form.Item>
        <Divider style={{ margin: '8px 0' }} />
        <Row gutter={16}>
          <Col span={12}>
            <Form.Item label="专业度" name="professionalScore" initialValue={0}>
              <Rate />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="态度" name="attitudeScore" initialValue={0}>
              <Rate />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="共情能力" name="empathyScore" initialValue={0}>
              <Rate />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="帮助程度" name="helpfulScore" initialValue={0}>
              <Rate />
            </Form.Item>
          </Col>
        </Row>
        <Divider style={{ margin: '8px 0' }} />
        <Form.Item label="评价标签" name="tags">
          <Select
            mode="tags"
            placeholder="选择或输入评价标签"
            style={{ width: '100%' }}
            options={[
              { label: '专业', value: '专业' },
              { label: '耐心', value: '耐心' },
              { label: '温暖', value: '温暖' },
              { label: '有帮助', value: '有帮助' },
              { label: '共情', value: '共情' },
              { label: '建议实用', value: '建议实用' },
            ]}
          />
        </Form.Item>
        <Form.Item label="评价内容" name="content">
          <TextArea rows={4} placeholder="分享您的咨询体验..." />
        </Form.Item>
        <Form.Item label="匿名评价" name="isAnonymousRating" valuePropName="checked" initialValue={false}>
          <Switch />
        </Form.Item>
      </ModalForm>

      <Drawer
        title="预约详情"
        width={640}
        open={detailDrawerOpen}
        onClose={() => setDetailDrawerOpen(false)}
        destroyOnClose
      >
        {detailData && (
          <Space direction="vertical" size={16} style={{ width: '100%' }}>
            {pick(detailData, ['is_emergency', 'isEmergency']) == 1 && (
              <Alert
                type="error"
                showIcon
                icon={<WarningOutlined />}
                message="紧急预约"
                description={`紧急级别：${pick(detailData, ['emergency_level_name', 'emergencyLevelName']) || EMERGENCY_LEVEL_MAP[pick(detailData, ['emergency_level', 'emergencyLevel']) as number] || '紧急'}`}
              />
            )}

            <Card>
              <Space align="start" size={16}>
                <Avatar
                  size={64}
                  src={pick(detailData, ['counselor_avatar', 'counselorAvatar'])}
                  icon={<UserOutlined />}
                />
                <Space direction="vertical" size={4}>
                  <Space>
                    <span style={{ fontSize: 18, fontWeight: 600 }}>
                      {pick(detailData, ['counselor_real_name', 'counselorRealName', 'counselor_name', 'counselorName'])}
                    </span>
                    {pick(detailData, ['counselor_title', 'counselorTitle']) && (
                      <Tag color="blue">{pick(detailData, ['counselor_title', 'counselorTitle'])}</Tag>
                    )}
                  </Space>
                  <Tag color={STATUS_COLOR[pick(detailData, ['status']) as number] || 'default'}>
                    {pick(detailData, ['status_name', 'statusName']) || STATUS_MAP[pick(detailData, ['status']) as number]}
                  </Tag>
                </Space>
              </Space>
            </Card>

            <Card title="基本信息" size="small">
              <Descriptions column={2} size="small">
                <Descriptions.Item label="预约编号">
                  {pick(detailData, ['appointment_no', 'appointmentNo'])}
                </Descriptions.Item>
                <Descriptions.Item label="咨询方式">
                  <Tag color={CONSULT_TYPE_COLOR[pick(detailData, ['consultation_type', 'consultationType']) as number]}>
                    {pick(detailData, ['consultation_type_name', 'consultationTypeName']) ||
                      CONSULT_TYPE_MAP[pick(detailData, ['consultation_type', 'consultationType']) as number]}
                  </Tag>
                </Descriptions.Item>
                <Descriptions.Item label="预约日期">
                  {pick(detailData, ['appointment_date', 'appointmentDate'])}
                </Descriptions.Item>
                <Descriptions.Item label="时段">
                  {(pick(detailData, ['start_time', 'startTime']) || '').slice(0, 5)} -
                  {(pick(detailData, ['end_time', 'endTime']) || '').slice(0, 5)}
                </Descriptions.Item>
                <Descriptions.Item label="问题类型">
                  {pick(detailData, ['concern_type', 'concernType']) || '-'}
                </Descriptions.Item>
                <Descriptions.Item label="匿名模式">
                  {pick(detailData, ['is_anonymous', 'isAnonymous']) == 1 ? (
                    <Tag color="purple" icon={<EyeInvisibleOutlined />}>是</Tag>
                  ) : (
                    <Tag>否</Tag>
                  )}
                </Descriptions.Item>
              </Descriptions>
            </Card>

            <Card title="当事人信息" size="small">
              <Descriptions column={2} size="small">
                <Descriptions.Item label="姓名">
                  {pick(detailData, ['party_name_display', 'partyNameDisplay']) || pick(detailData, ['party_name', 'partyName']) || '-'}
                </Descriptions.Item>
                <Descriptions.Item label="电话">
                  {pick(detailData, ['party_phone_display', 'partyPhoneDisplay']) || pick(detailData, ['party_phone', 'partyPhone']) || '-'}
                </Descriptions.Item>
                <Descriptions.Item label="身份证号">
                  {pick(detailData, ['party_id_card_display', 'partyIdCardDisplay']) || pick(detailData, ['party_id_card', 'partyIdCard']) || '-'}
                </Descriptions.Item>
                <Descriptions.Item label="关联案件">
                  {pick(detailData, ['case_id', 'caseId']) || '-'}
                </Descriptions.Item>
              </Descriptions>
            </Card>

            {pick(detailData, ['concern_description', 'concernDescription']) && (
              <Card title="问题描述" size="small">
                <div style={{ whiteSpace: 'pre-wrap', lineHeight: 1.8 }}>
                  {pick(detailData, ['concern_description', 'concernDescription'])}
                </div>
              </Card>
            )}

            {pick(detailData, ['consultation_summary', 'consultationSummary']) && (
              <Card title="咨询摘要" size="small">
                <div style={{ whiteSpace: 'pre-wrap', lineHeight: 1.8 }}>
                  {pick(detailData, ['consultation_summary', 'consultationSummary'])}
                </div>
              </Card>
            )}

            {pick(detailData, ['follow_up_suggestion', 'followUpSuggestion']) && (
              <Card title="后续建议" size="small">
                <div style={{ whiteSpace: 'pre-wrap', lineHeight: 1.8 }}>
                  {pick(detailData, ['follow_up_suggestion', 'followUpSuggestion'])}
                </div>
              </Card>
            )}

            <Card title="时间记录" size="small">
              <Descriptions column={2} size="small">
                <Descriptions.Item label="创建时间">
                  {pick(detailData, ['created_at', 'createdAt'])}
                </Descriptions.Item>
                <Descriptions.Item label="确认时间">
                  {pick(detailData, ['confirmed_at', 'confirmedAt']) || '-'}
                </Descriptions.Item>
                <Descriptions.Item label="开始时间">
                  {pick(detailData, ['started_at', 'startedAt']) || '-'}
                </Descriptions.Item>
                <Descriptions.Item label="完成时间">
                  {pick(detailData, ['completed_at', 'completedAt']) || '-'}
                </Descriptions.Item>
              </Descriptions>
            </Card>
          </Space>
        )}
      </Drawer>
    </>
  );
};

export default AppointmentList;
