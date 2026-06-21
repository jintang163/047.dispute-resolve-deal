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
  InputNumber,
  Select,
  Switch,
  Drawer,
  Row,
  Col,
  Descriptions,
  Image,
} from 'antd';
import {
  PlusOutlined,
  EyeOutlined,
  EditOutlined,
  DeleteOutlined,
  ExclamationCircleOutlined,
  UserOutlined,
  WarningOutlined,
  StarOutlined,
  CalendarOutlined,
} from '@ant-design/icons';
import {
  ProTable,
  ProFormSelect,
  ProFormText,
  ProFormTextArea,
  ProFormDigit,
  ModalForm,
  DrawerForm,
  ProFormSwitch,
} from '@ant-design/pro-components';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { useNavigate } from 'react-router-dom';
import { counselingService, Counselor } from '../../services/counseling';
import dayjs from 'dayjs';

const { confirm } = Modal;
const { Option } = Select;
const { TextArea } = Input;

const STATUS_MAP: Record<number, string> = {
  0: '停用',
  1: '启用',
};

const STATUS_COLOR: Record<number, string> = {
  0: 'default',
  1: 'success',
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

const pick = <T,>(obj: T, keys: (keyof T | string)[]): any => {
  const o = obj as any;
  for (const k of keys) {
    if (o[k] !== undefined && o[k] !== null && o[k] !== '') return o[k];
  }
  return undefined;
};

const CounselorList: React.FC = () => {
  const navigate = useNavigate();
  const { message } = App.useApp();
  const actionRef = useRef<ActionType>();
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [editDrawerOpen, setEditDrawerOpen] = useState(false);
  const [currentCounselor, setCurrentCounselor] = useState<Counselor | null>(null);
  const [detailDrawerOpen, setDetailDrawerOpen] = useState(false);
  const [detailData, setDetailData] = useState<Counselor | null>(null);
  const [createForm] = Form.useForm();
  const [editForm] = Form.useForm();

  const columns: ProColumns<Counselor>[] = [
    {
      title: '编号',
      dataIndex: 'counselorNo',
      width: 140,
      copyable: true,
      fixed: 'left',
    },
    {
      title: '头像',
      dataIndex: 'avatar',
      width: 80,
      search: false,
      render: (_, row) => {
        const avatar = pick(row, ['avatar']);
        return (
          <Avatar
            size={40}
            src={avatar}
            icon={<UserOutlined />}
          />
        );
      },
    },
    {
      title: '姓名',
      dataIndex: 'realName',
      width: 120,
      render: (_, row) => {
        const name = pick(row, ['real_name', 'realName']);
        const title = pick(row, ['title']);
        return (
          <Space direction="vertical" size={0}>
            <span style={{ fontWeight: 500 }}>{name}</span>
            {title && <span style={{ color: '#999', fontSize: 12 }}>{title}</span>}
          </Space>
        );
      },
    },
    {
      title: '专业领域',
      dataIndex: 'specialty',
      width: 200,
      search: false,
      render: (_, row) => {
        const tags = pick(row, ['specialtyTagList', 'specialty_tag_list']);
        if (tags && Array.isArray(tags) && tags.length > 0) {
          return (
            <Space size={[4, 4]} wrap>
              {tags.slice(0, 3).map((tag: string, idx: number) => (
                <Tag
                  key={tag}
                  color={['magenta', 'red', 'volcano', 'orange', 'gold', 'lime', 'green', 'cyan'][idx % 8]}
                  style={{ fontSize: 11, padding: '0 4px', margin: 0 }}
                >
                  {tag}
                </Tag>
              ))}
              {tags.length > 3 && <Tag style={{ fontSize: 11 }}>+{tags.length - 3}</Tag>}
            </Space>
          );
        }
        const specialty = pick(row, ['specialty']);
        return <span style={{ color: '#999' }}>{specialty || '-'}</span>;
      },
    },
    {
      title: '咨询方式',
      dataIndex: 'consultationTypes',
      width: 160,
      search: false,
      render: (_, row) => {
        const types = pick(row, ['consultationTypeList', 'consultation_type_list']);
        if (types && Array.isArray(types) && types.length > 0) {
          return (
            <Space size={4}>
              {types.map((t: any) => (
                <Tag
                  key={t.type}
                  color={CONSULT_TYPE_COLOR[t.type] || 'default'}
                >
                  {t.name || CONSULT_TYPE_MAP[t.type]}
                </Tag>
              ))}
            </Space>
          );
        }
        return <span style={{ color: '#999' }}>-</span>;
      },
    },
    {
      title: '从业年限',
      dataIndex: 'yearsOfExperience',
      width: 100,
      search: false,
      render: (_, row) => {
        const years = pick(row, ['years_of_experience', 'yearsOfExperience']);
        return years ? `${years} 年` : '-';
      },
    },
    {
      title: '评分',
      dataIndex: 'ratingAvg',
      width: 140,
      search: false,
      render: (_, row) => {
        const avg = pick(row, ['rating_avg', 'ratingAvg']) || 0;
        const count = pick(row, ['rating_count', 'ratingCount']) || 0;
        return (
          <Space>
            <Rate disabled allowHalf value={Number(avg)} style={{ fontSize: 14 }} />
            <span style={{ color: '#faad14', fontWeight: 500 }}>{Number(avg).toFixed(1)}</span>
            <span style={{ color: '#999', fontSize: 12 }}>({count})</span>
          </Space>
        );
      },
    },
    {
      title: '紧急干预',
      dataIndex: 'isEmergencyAvailable',
      width: 100,
      search: false,
      render: (_, row) => {
        const val = pick(row, ['is_emergency_available', 'isEmergencyAvailable']);
        if (val == 1) {
          return (
            <Tag color="red" icon={<WarningOutlined />}>
              可接紧急
            </Tag>
          );
        }
        return <Tag color="default">否</Tag>;
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      valueEnum: {
        0: { text: '停用', status: 'Default' },
        1: { text: '启用', status: 'Success' },
      },
      render: (_, row) => {
        const s = pick(row, ['status']) as number;
        const name = pick(row, ['status_name', 'statusName']) || STATUS_MAP[s];
        return <Tag color={STATUS_COLOR[s] || 'default'}>{name}</Tag>;
      },
    },
    {
      title: '所属机构',
      dataIndex: 'organizationName',
      width: 160,
      ellipsis: true,
      render: (_, row) => pick(row, ['org_name', 'organizationName']) || '-',
    },
    {
      title: '费用',
      dataIndex: 'price',
      width: 100,
      search: false,
      render: (_, row) => {
        const price = pick(row, ['price']);
        return price ? `¥${Number(price).toFixed(2)}` : '免费';
      },
    },
    {
      title: '操作',
      key: 'option',
      valueType: 'option',
      width: 240,
      fixed: 'right',
      render: (_, record) => {
        const id = pick(record, ['id']);
        return [
          <Button
            type="link"
            key="view"
            icon={<EyeOutlined />}
            onClick={async () => {
              try {
                const res: any = await counselingService.getCounselorDetail(id);
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
          <Button
            type="link"
            key="appointment"
            icon={<CalendarOutlined />}
            onClick={() => {
              navigate(`/counseling/appointment?counselorId=${id}`);
            }}
          >
            预约
          </Button>,
          <Button
            type="link"
            key="edit"
            icon={<EditOutlined />}
            onClick={() => {
              setCurrentCounselor(record);
              editForm.setFieldsValue({
                realName: pick(record, ['real_name', 'realName']),
                gender: pick(record, ['gender']) || 0,
                phone: pick(record, ['phone']),
                email: pick(record, ['email']),
                avatar: pick(record, ['avatar']),
                title: pick(record, ['title']),
                licenseNo: pick(record, ['license_no', 'licenseNo']),
                specialty: pick(record, ['specialty']),
                specialtyTags: pick(record, ['specialty_tags', 'specialtyTags']),
                yearsOfExperience: pick(record, ['years_of_experience', 'yearsOfExperience']),
                education: pick(record, ['education']),
                introduction: pick(record, ['introduction']),
                consultationTypes: pick(record, ['consultation_types', 'consultationTypes']),
                workDays: pick(record, ['work_days', 'workDays']),
                workStartTime: pick(record, ['work_start_time', 'workStartTime']),
                workEndTime: pick(record, ['work_end_time', 'workEndTime']),
                sessionDuration: pick(record, ['session_duration', 'sessionDuration']) || 50,
                price: pick(record, ['price']) || 0,
                organizationName: pick(record, ['org_name', 'organizationName']),
                isEmergencyAvailable: pick(record, ['is_emergency_available', 'isEmergencyAvailable']) == 1,
                status: pick(record, ['status']) || 1,
                sortOrder: pick(record, ['sort_order', 'sortOrder']) || 0,
              });
              setEditDrawerOpen(true);
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
                title: '确认删除该心理咨询师?',
                icon: <ExclamationCircleOutlined />,
                content: '删除后将无法恢复，请谨慎操作。',
                okText: '确认删除',
                cancelText: '取消',
                okButtonProps: { danger: true },
                onOk: async () => {
                  try {
                    await counselingService.deleteCounselor(id);
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
        ];
      },
    },
  ];

  return (
    <>
      <ProTable<Counselor>
        columns={columns}
        actionRef={actionRef}
        cardBordered
        rowKey="id"
        search={{
          labelWidth: 'auto',
          defaultCollapsed: false,
        }}
        dateFormatter="string"
        headerTitle="心理咨询师管理"
        toolBarRender={() => [
          <Button
            key="create"
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => {
              createForm.resetFields();
              setCreateModalOpen(true);
            }}
          >
            新增心理咨询师
          </Button>,
          <Button
            key="appointment"
            icon={<CalendarOutlined />}
            onClick={() => navigate('/counseling/appointment')}
          >
            预约管理
          </Button>,
        ]}
        request={async (params, sort, filter) => {
          try {
            const res: any = await counselingService.getCounselorList({
              page: params.current,
              pageSize: params.pageSize,
              keyword: params.keyword as string,
              status: (params.status as unknown as number) || undefined,
              specialty: (params.specialty as string) || undefined,
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
          persistenceKey: 'counselor-list-columns',
          persistenceType: 'localStorage',
        }}
        pagination={{
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: total => `共 ${total} 条记录`,
        }}
        scroll={{ x: 1800 }}
      />

      <ModalForm
        title="新增心理咨询师"
        open={createModalOpen}
        onOpenChange={setCreateModalOpen}
        form={createForm}
        width={720}
        modalProps={{ destroyOnClose: true, maskClosable: false }}
        onFinish={async (values: any) => {
          try {
            await counselingService.createCounselor({
              ...values,
              isEmergencyAvailable: values.isEmergencyAvailable ? 1 : 0,
            });
            message.success('创建成功');
            actionRef.current?.reload();
            return true;
          } catch (error: any) {
            message.error(error.message || '创建失败');
            return false;
          }
        }}
      >
        <Row gutter={16}>
          <Col span={12}>
            <Form.Item label="姓名" name="realName" rules={[{ required: true, message: '请输入姓名' }]}>
              <Input placeholder="请输入姓名" />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="性别" name="gender" initialValue={0}>
              <Select>
                <Option value={0}>未知</Option>
                <Option value={1}>男</Option>
                <Option value={2}>女</Option>
              </Select>
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="联系电话" name="phone">
              <Input placeholder="请输入联系电话" />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="邮箱" name="email">
              <Input placeholder="请输入邮箱" />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="职称" name="title">
              <Input placeholder="如：高级心理咨询师" />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="执业证书编号" name="licenseNo">
              <Input placeholder="请输入执业证书编号" />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="从业年限" name="yearsOfExperience" initialValue={0}>
              <InputNumber min={0} style={{ width: '100%' }} addonAfter="年" />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="学历" name="education">
              <Select allowClear>
                <Option value="高中及以下">高中及以下</Option>
                <Option value="大专">大专</Option>
                <Option value="学士">学士</Option>
                <Option value="硕士">硕士</Option>
                <Option value="博士">博士</Option>
              </Select>
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="专业领域" name="specialty">
              <Input placeholder="逗号分隔，如：婚姻家庭,创伤后应激" />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="擅长标签" name="specialtyTags">
              <Input placeholder="逗号分隔，用于推荐匹配" />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="咨询费用" name="price" initialValue={0}>
              <InputNumber min={0} style={{ width: '100%' }} addonBefore="¥" addonAfter="/次" precision={2} />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="单次时长" name="sessionDuration" initialValue={50}>
              <InputNumber min={15} max={180} style={{ width: '100%' }} addonAfter="分钟" />
            </Form.Item>
          </Col>
          <Col span={24}>
            <Form.Item label="咨询方式" name="consultationTypes" initialValue="1,2,3">
              <Select mode="tags" placeholder="选择咨询方式">
                <Option value="1">线上视频</Option>
                <Option value="2">线上语音</Option>
                <Option value="3">线下面谈</Option>
              </Select>
            </Form.Item>
          </Col>
          <Col span={24}>
            <Form.Item label="可预约工作日" name="workDays" initialValue="1,2,3,4,5">
              <Select mode="tags" placeholder="选择工作日">
                <Option value="1">周一</Option>
                <Option value="2">周二</Option>
                <Option value="3">周三</Option>
                <Option value="4">周四</Option>
                <Option value="5">周五</Option>
                <Option value="6">周六</Option>
                <Option value="7">周日</Option>
              </Select>
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="每日开始时间" name="workStartTime" initialValue="09:00:00">
              <Input placeholder="如：09:00:00" />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="每日结束时间" name="workEndTime" initialValue="18:00:00">
              <Input placeholder="如：18:00:00" />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="所属机构" name="organizationName">
              <Input placeholder="请输入所属机构" />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="排序" name="sortOrder" initialValue={0}>
              <InputNumber style={{ width: '100%' }} />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="状态" name="status" initialValue={1}>
              <Select>
                <Option value={1}>启用</Option>
                <Option value={0}>停用</Option>
              </Select>
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="接受紧急干预" name="isEmergencyAvailable" valuePropName="checked" initialValue={false}>
              <Switch />
            </Form.Item>
          </Col>
          <Col span={24}>
            <Form.Item label="头像URL" name="avatar">
              <Input placeholder="请输入头像图片地址" />
            </Form.Item>
          </Col>
          <Col span={24}>
            <Form.Item label="个人简介" name="introduction">
              <TextArea rows={4} placeholder="请输入个人简介" />
            </Form.Item>
          </Col>
        </Row>
      </ModalForm>

      <DrawerForm
        title="编辑心理咨询师"
        open={editDrawerOpen}
        onOpenChange={setEditDrawerOpen}
        form={editForm}
        width={640}
        drawerProps={{ destroyOnClose: true, maskClosable: false }}
        onFinish={async (values: any) => {
          if (!currentCounselor) return false;
          try {
            await counselingService.updateCounselor(currentCounselor.id as any, {
              ...values,
              isEmergencyAvailable: values.isEmergencyAvailable ? 1 : 0,
            });
            message.success('更新成功');
            actionRef.current?.reload();
            return true;
          } catch (error: any) {
            message.error(error.message || '更新失败');
            return false;
          }
        }}
      >
        <Row gutter={16}>
          <Col span={12}>
            <Form.Item label="姓名" name="realName" rules={[{ required: true, message: '请输入姓名' }]}>
              <Input />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="性别" name="gender">
              <Select>
                <Option value={0}>未知</Option>
                <Option value={1}>男</Option>
                <Option value={2}>女</Option>
              </Select>
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="联系电话" name="phone">
              <Input />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="邮箱" name="email">
              <Input />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="职称" name="title">
              <Input />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="执业证书编号" name="licenseNo">
              <Input />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="从业年限" name="yearsOfExperience">
              <InputNumber min={0} style={{ width: '100%' }} addonAfter="年" />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="学历" name="education">
              <Select allowClear>
                <Option value="高中及以下">高中及以下</Option>
                <Option value="大专">大专</Option>
                <Option value="学士">学士</Option>
                <Option value="硕士">硕士</Option>
                <Option value="博士">博士</Option>
              </Select>
            </Form.Item>
          </Col>
          <Col span={24}>
            <Form.Item label="专业领域" name="specialty">
              <Input placeholder="逗号分隔" />
            </Form.Item>
          </Col>
          <Col span={24}>
            <Form.Item label="擅长标签" name="specialtyTags">
              <Input placeholder="逗号分隔" />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="咨询费用" name="price">
              <InputNumber min={0} style={{ width: '100%' }} addonBefore="¥" addonAfter="/次" precision={2} />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="单次时长" name="sessionDuration">
              <InputNumber min={15} max={180} style={{ width: '100%' }} addonAfter="分钟" />
            </Form.Item>
          </Col>
          <Col span={24}>
            <Form.Item label="咨询方式" name="consultationTypes">
              <Select mode="tags">
                <Option value="1">线上视频</Option>
                <Option value="2">线上语音</Option>
                <Option value="3">线下面谈</Option>
              </Select>
            </Form.Item>
          </Col>
          <Col span={24}>
            <Form.Item label="可预约工作日" name="workDays">
              <Select mode="tags">
                <Option value="1">周一</Option>
                <Option value="2">周二</Option>
                <Option value="3">周三</Option>
                <Option value="4">周四</Option>
                <Option value="5">周五</Option>
                <Option value="6">周六</Option>
                <Option value="7">周日</Option>
              </Select>
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="每日开始时间" name="workStartTime">
              <Input />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="每日结束时间" name="workEndTime">
              <Input />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="所属机构" name="organizationName">
              <Input />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="排序" name="sortOrder">
              <InputNumber style={{ width: '100%' }} />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="状态" name="status">
              <Select>
                <Option value={1}>启用</Option>
                <Option value={0}>停用</Option>
              </Select>
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="接受紧急干预" name="isEmergencyAvailable" valuePropName="checked">
              <Switch />
            </Form.Item>
          </Col>
          <Col span={24}>
            <Form.Item label="头像URL" name="avatar">
              <Input />
            </Form.Item>
          </Col>
          <Col span={24}>
            <Form.Item label="个人简介" name="introduction">
              <TextArea rows={4} />
            </Form.Item>
          </Col>
        </Row>
      </DrawerForm>

      <Drawer
        title="心理咨询师详情"
        width={720}
        open={detailDrawerOpen}
        onClose={() => setDetailDrawerOpen(false)}
        destroyOnClose
      >
        {detailData && (
          <Space direction="vertical" size={16} style={{ width: '100%' }}>
            <Card>
              <Space align="start" size={16}>
                <Avatar size={80} src={detailData.avatar} icon={<UserOutlined />} />
                <Space direction="vertical" size={4}>
                  <Space>
                    <span style={{ fontSize: 20, fontWeight: 600 }}>{detailData.realName || detailData.real_name}</span>
                    {detailData.title && <Tag color="blue">{detailData.title}</Tag>}
                    {(detailData.is_emergency_available == 1 || detailData.isEmergencyAvailable == 1) && (
                      <Tag color="red" icon={<WarningOutlined />}>可接紧急</Tag>
                    )}
                  </Space>
                  <Space>
                    <Rate disabled allowHalf value={Number(detailData.rating_avg || detailData.ratingAvg || 0)} />
                    <span style={{ color: '#faad14', fontWeight: 500 }}>
                      {Number(detailData.rating_avg || detailData.ratingAvg || 0).toFixed(1)}
                    </span>
                    <span style={{ color: '#999' }}>
                      ({detailData.rating_count || detailData.ratingCount || 0} 条评价)
                    </span>
                  </Space>
                  <Space size={16}>
                    <span style={{ color: '#666' }}>从业 {detailData.years_of_experience || detailData.yearsOfExperience || 0} 年</span>
                    <span style={{ color: '#666' }}>{detailData.education || '-'}</span>
                    <span style={{ color: '#666' }}>{detailData.org_name || detailData.organizationName || '-'}</span>
                  </Space>
                </Space>
              </Space>
            </Card>

            <Card title="基本信息" size="small">
              <Descriptions column={2} size="small">
                <Descriptions.Item label="联系电话">{detailData.phone || '-'}</Descriptions.Item>
                <Descriptions.Item label="邮箱">{detailData.email || '-'}</Descriptions.Item>
                <Descriptions.Item label="咨询费用">
                  ¥{Number(detailData.price || 0).toFixed(2)}/次
                </Descriptions.Item>
                <Descriptions.Item label="单次时长">
                  {detailData.session_duration || detailData.sessionDuration || 50} 分钟
                </Descriptions.Item>
                <Descriptions.Item label="执业证书">
                  {detailData.license_no || detailData.licenseNo || '-'}
                </Descriptions.Item>
                <Descriptions.Item label="累计咨询">
                  {detailData.completed_count || detailData.completedCount || 0} 次
                </Descriptions.Item>
              </Descriptions>
            </Card>

            <Card title="专业领域" size="small">
              <Space size={[4, 8]} wrap>
                {(detailData.specialtyTagList || detailData.specialty_tag_list || []).map((tag: string, idx: number) => (
                  <Tag
                    key={tag}
                    color={['magenta', 'red', 'volcano', 'orange', 'gold', 'lime', 'green', 'cyan'][idx % 8]}
                  >
                    {tag}
                  </Tag>
                ))}
              </Space>
            </Card>

            <Card title="咨询方式" size="small">
              <Space size={8}>
                {(detailData.consultationTypeList || detailData.consultation_type_list || []).map((t: any) => (
                  <Tag key={t.type} color={CONSULT_TYPE_COLOR[t.type]}>
                    {t.name}
                  </Tag>
                ))}
              </Space>
            </Card>

            <Card title="工作时间" size="small">
              <Descriptions column={1} size="small">
                <Descriptions.Item label="工作日">
                  {(detailData.workDayList || detailData.work_day_list || ['周一', '周二', '周三', '周四', '周五']).join('、')}
                </Descriptions.Item>
                <Descriptions.Item label="工作时间">
                  {detailData.work_start_time || detailData.workStartTime || '09:00'} - {detailData.work_end_time || detailData.workEndTime || '18:00'}
                </Descriptions.Item>
              </Descriptions>
            </Card>

            {detailData.introduction && (
              <Card title="个人简介" size="small">
                <div style={{ whiteSpace: 'pre-wrap', lineHeight: 1.8 }}>
                  {detailData.introduction}
                </div>
              </Card>
            )}

            {(detailData.recentRatings || []).length > 0 && (
              <Card title="最新评价" size="small">
                <Space direction="vertical" size={12} style={{ width: '100%' }}>
                  {(detailData.recentRatings || []).map((r: any, idx: number) => (
                    <Card key={idx} size="small" variant="borderless" style={{ background: '#fafafa' }}>
                      <Space direction="vertical" size={4} style={{ width: '100%' }}>
                        <Space size={8} align="center">
                          <Avatar size={28} icon={<UserOutlined />} />
                          <span style={{ fontWeight: 500 }}>{r.raterName || r.rater_name || '匿名用户'}</span>
                          <Rate disabled allowHalf value={Number(r.overallScore || r.overall_score || 0)} style={{ fontSize: 12 }} />
                          <span style={{ color: '#999', fontSize: 12 }}>
                            {dayjs(r.createdAt || r.created_at).format('YYYY-MM-DD')}
                          </span>
                        </Space>
                        {r.content && <div style={{ color: '#333', paddingLeft: 36 }}>{r.content}</div>}
                        {(r.tagList || r.tag_list || []).length > 0 && (
                          <Space size={4} wrap style={{ paddingLeft: 36 }}>
                            {(r.tagList || r.tag_list).map((t: string) => (
                              <Tag key={t} color="blue" style={{ fontSize: 11 }}>{t}</Tag>
                            ))}
                          </Space>
                        )}
                      </Space>
                    </Card>
                  ))}
                </Space>
              </Card>
            )}

            <Space style={{ width: '100%', justifyContent: 'flex-end' }}>
              <Button onClick={() => setDetailDrawerOpen(false)}>关闭</Button>
              <Button
                type="primary"
                icon={<CalendarOutlined />}
                onClick={() => {
                  setDetailDrawerOpen(false);
                  navigate(`/counseling/appointment?counselorId=${detailData.id}`);
                }}
              >
                立即预约
              </Button>
            </Space>
          </Space>
        )}
      </Drawer>
    </>
  );
};

export default CounselorList;
