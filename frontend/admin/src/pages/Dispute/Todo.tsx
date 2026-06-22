import React, { useRef, useState } from 'react';
import { Button, Tag, Space, App, Card, Modal, Form, Input, Radio, Badge, Tooltip, Dropdown } from 'antd';
import {
  EyeOutlined,
  WarningOutlined,
  AlertOutlined,
  ExclamationCircleOutlined,
  ThunderboltOutlined,
  FlagOutlined,
  MoreOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { ProTable, ProFormSelect } from '@ant-design/pro-components';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { useNavigate } from 'react-router-dom';
import { disputeService, DisputeCase } from '../../services/dispute';
import dayjs from 'dayjs';

const STATUS_MAP: Record<string, string> = {
  '10': '待分派',
  '20': '调解中',
};

const STATUS_COLOR: Record<string, string> = {
  '10': 'default',
  '20': 'blue',
};

const RISK_FLAG_MAP: Record<number, { label: string; color: string; icon: React.ReactNode }> = {
  1: { label: '扬言上访', color: '#ff4d4f', icon: <WarningOutlined /> },
  2: { label: '极端行为', color: '#cf1322', icon: <AlertOutlined /> },
  3: { label: '群体事件', color: '#a8071a', icon: <ExclamationCircleOutlined /> },
};

const PRIORITY_LABEL_MAP: Record<string, { label: string; color: string; bgColor: string }> = {
  '风险预警': { label: '风险预警', color: '#fff', bgColor: '#ff4d4f' },
  '超时未处理': { label: '超时未处理', color: '#874d00', bgColor: '#ffd591' },
  '新分配': { label: '新分配', color: '#0958d9', bgColor: '#adc6ff' },
};

const pick = <T,>(obj: T, keys: (keyof T | string)[]): any => {
  const o = obj as any;
  for (const k of keys) {
    if (o[k] !== undefined && o[k] !== null && o[k] !== '') return o[k];
  }
  return undefined;
};

const TodoList: React.FC = () => {
  const navigate = useNavigate();
  const { message, modal } = App.useApp();
  const actionRef = useRef<ActionType>();
  const [riskModalVisible, setRiskModalVisible] = useState(false);
  const [currentCase, setCurrentCase] = useState<DisputeCase | null>(null);
  const [riskForm] = Form.useForm();

  const handleSetRiskFlag = (record: DisputeCase) => {
    setCurrentCase(record);
    riskForm.setFieldsValue({
      riskFlag: pick(record, ['risk_flag', 'riskFlag']) || 0,
      reason: pick(record, ['risk_reason', 'riskReason']) || '',
    });
    setRiskModalVisible(true);
  };

  const handleRiskSubmit = async () => {
    try {
      const values = await riskForm.validateFields();
      const id = pick(currentCase!, ['id']);
      await disputeService.setRiskFlag(id, values.riskFlag, values.reason);
      message.success('风险标记设置成功');
      setRiskModalVisible(false);
      actionRef.current?.reload();
    } catch (error: any) {
      if (error.errorFields) return;
      message.error(error.message || '设置失败');
    }
  };

  const renderPriorityBadge = (record: any) => {
    const priorityLabel = record.priority_label || '';
    const riskFlag = pick(record, ['risk_flag', 'riskFlag']) || 0;
    const sortPriority = record.sort_priority || 99;

    if (riskFlag > 0 && RISK_FLAG_MAP[riskFlag]) {
      const risk = RISK_FLAG_MAP[riskFlag];
      return (
        <Tooltip title={`风险标记: ${risk.label}`}>
          <Tag
            icon={risk.icon}
            color="error"
            style={{
              animation: 'riskBlink 1.5s ease-in-out infinite',
              fontWeight: 600,
              fontSize: 12,
            }}
          >
            {risk.label}
          </Tag>
        </Tooltip>
      );
    }

    const config = PRIORITY_LABEL_MAP[priorityLabel];
    if (config) {
      return (
        <Tag
          style={{
            color: config.color,
            backgroundColor: config.bgColor,
            borderColor: config.bgColor,
            fontSize: 12,
          }}
        >
          {config.label}
        </Tag>
      );
    }

    if (sortPriority <= 2) {
      return <Tag color="warning">超时未处理</Tag>;
    }

    return null;
  };

  const columns: ProColumns<DisputeCase>[] = [
    {
      title: '优先级',
      dataIndex: 'sort_priority',
      width: 130,
      search: false,
      fixed: 'left',
      render: (_, record: any) => renderPriorityBadge(record),
    },
    {
      title: '案件编号',
      dataIndex: 'caseNo',
      width: 180,
      copyable: true,
      fixed: 'left',
      render: (_, record) => pick(record, ['case_no', 'caseNo']),
    },
    {
      title: '案件标题',
      dataIndex: 'title',
      width: 200,
      ellipsis: true,
      render: (_, record) => {
        const title = pick(record, ['title']);
        const riskFlag = pick(record, ['risk_flag', 'riskFlag']) || 0;
        return (
          <span style={{ fontWeight: riskFlag > 0 ? 700 : 400, color: riskFlag > 0 ? '#ff4d4f' : undefined }}>
            {title}
          </span>
        );
      },
    },
    {
      title: '纠纷类型',
      dataIndex: 'type',
      width: 120,
      render: (_, record) => {
        const name = pick(record, ['type_name', 'typeName']);
        return <Tag color="blue">{name || '-'}</Tag>;
      },
    },
    {
      title: '紧急程度',
      dataIndex: 'caseLevel',
      width: 90,
      search: false,
      render: (_, record) => {
        const level = pick(record, ['case_level', 'caseLevel', 'level']);
        const levelMap: Record<number, { label: string; color: string }> = {
          1: { label: '特急', color: '#ff4d4f' },
          2: { label: '紧急', color: '#fa8c16' },
          3: { label: '一般', color: '#1890ff' },
          4: { label: '普通', color: '#d9d9d9' },
        };
        const config = levelMap[level] || { label: level || '-', color: '#d9d9d9' };
        return (
          <Tag color={config.color} style={level === 1 ? { fontWeight: 700 } : undefined}>
            {config.label}
          </Tag>
        );
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 90,
      render: (_, record) => {
        const raw = pick(record, ['status']);
        const s = typeof raw === 'number' ? String(raw) : String(raw || '');
        const name = pick(record, ['status_name', 'statusName']) || STATUS_MAP[s] || s;
        return <Tag color={STATUS_COLOR[s] || 'default'}>{name}</Tag>;
      },
    },
    {
      title: '当事人',
      dataIndex: 'applicantName',
      width: 120,
      ellipsis: true,
      render: (_, record) => {
        const applicant = pick(record, ['applicant_name', 'applicantName']) || '-';
        const respondent = pick(record, ['respondent_name', 'respondentName']) || '-';
        return (
          <Tooltip title={`${applicant} vs ${respondent}`}>
            <span>{applicant}</span>
          </Tooltip>
        );
      },
    },
    {
      title: '所属组织',
      dataIndex: 'orgName',
      width: 140,
      ellipsis: true,
      search: false,
      render: (_, record) => pick(record, ['org_name', 'orgName']) || '-',
    },
    {
      title: '分配时间',
      dataIndex: 'mediatorTime',
      width: 170,
      search: false,
      render: (_, record) => {
        const v = pick(record, ['mediator_time', 'mediatorTime']);
        return v ? dayjs(v).format('YYYY-MM-DD HH:mm') : '-';
      },
    },
    {
      title: '超时',
      dataIndex: 'isOverdue',
      width: 80,
      search: false,
      render: (_, record) => {
        const mediatorTime = pick(record, ['mediator_time', 'mediatorTime']);
        const status = pick(record, ['status']);
        if (!mediatorTime || (status !== 10 && status !== 20)) return '-';
        const hours = dayjs().diff(dayjs(mediatorTime), 'hour');
        if (hours > 72) {
          return (
            <Tooltip title={`已超时 ${hours} 小时`}>
              <Tag color="error" style={{ fontWeight: 600 }}>
                {Math.floor(hours / 24)}天
              </Tag>
            </Tooltip>
          );
        }
        const remaining = 72 - hours;
        return (
          <Tooltip title={`剩余 ${remaining} 小时`}>
            <Tag color={remaining < 24 ? 'warning' : 'default'}>
              {Math.floor(remaining / 24)}天{remaining % 24}时
            </Tag>
          </Tooltip>
        );
      },
    },
    {
      title: '催办',
      dataIndex: 'urgencyCount',
      width: 70,
      search: false,
      render: (_, record) => {
        const count = pick(record, ['urgency_count', 'urgencyCount']) || 0;
        if (count === 0) return '-';
        return (
          <Tag color={count >= 3 ? 'error' : 'warning'}>
            {count}次
          </Tag>
        );
      },
    },
    {
      title: '操作',
      key: 'option',
      valueType: 'option',
      width: 140,
      fixed: 'right',
      render: (_, record) => {
        const id = pick(record, ['id']);
        const riskFlag = pick(record, ['risk_flag', 'riskFlag']) || 0;

        return [
          <Button type="link" key="view" icon={<EyeOutlined />} onClick={() => navigate(`/dispute/${id}`)}>
            处理
          </Button>,
          <Dropdown
            key="more"
            menu={{
              items: [
                {
                  key: 'risk',
                  icon: <FlagOutlined />,
                  label: riskFlag > 0 ? '修改风险标记' : '设置风险标记',
                  danger: riskFlag === 0,
                  onClick: () => handleSetRiskFlag(record),
                },
                {
                  key: 'clearRisk',
                  icon: <FlagOutlined />,
                  label: '清除风险标记',
                  disabled: riskFlag === 0,
                  onClick: async () => {
                    modal.confirm({
                      title: '确认清除风险标记?',
                      content: '清除后该案件将不再显示风险预警标识。',
                      onOk: async () => {
                        await disputeService.setRiskFlag(id, 0, '');
                        message.success('风险标记已清除');
                        actionRef.current?.reload();
                      },
                    });
                  },
                },
              ],
            }}
          >
            <Button type="text" icon={<MoreOutlined />} />
          </Dropdown>,
        ];
      },
    },
  ];

  return (
    <>
      <Card
        size="small"
        style={{ marginBottom: 12 }}
        bodyStyle={{ padding: '12px 16px' }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 16, flexWrap: 'wrap' }}>
          <Space>
            <ThunderboltOutlined style={{ color: '#fa8c16', fontSize: 16 }} />
            <span style={{ fontWeight: 600, fontSize: 14 }}>智能排序规则：</span>
          </Space>
          <Space size={8}>
            <Tag
              icon={<WarningOutlined />}
              color="error"
              style={{ fontSize: 12, fontWeight: 600 }}
            >
              1. 风险预警（扬言上访/极端行为/群体事件）
            </Tag>
            <Tag color="warning" style={{ fontSize: 12 }}>
              2. 超时未处理（超过72小时）
            </Tag>
            <Tag color="processing" style={{ fontSize: 12 }}>
              3. 新分配案件
            </Tag>
          </Space>
        </div>
      </Card>

      <ProTable<DisputeCase>
        columns={columns}
        actionRef={actionRef}
        cardBordered
        rowKey="id"
        search={{
          labelWidth: 'auto',
          defaultCollapsed: false,
        }}
        dateFormatter="string"
        headerTitle="待办案件"
        toolBarRender={() => [
          <Button
            key="refresh"
            icon={<ReloadOutlined />}
            onClick={() => actionRef.current?.reload()}
          >
            刷新
          </Button>,
        ]}
        request={async (params) => {
          try {
            const res = await disputeService.getTodoList({
              pageNum: params.current,
              pageSize: params.pageSize,
              status: params.status as number,
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
        rowClassName={(record) => {
          const riskFlag = pick(record, ['risk_flag', 'riskFlag']) || 0;
          if (riskFlag > 0) return 'risk-alert-row';
          const sortPriority = (record as any).sort_priority || 99;
          if (sortPriority === 2) return 'overdue-row';
          return '';
        }}
        pagination={{
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: total => `共 ${total} 条待办`,
          defaultPageSize: 20,
        }}
        scroll={{ x: 1600 }}
      />

      <Modal
        title={
          <span>
            <FlagOutlined style={{ color: '#ff4d4f', marginRight: 8 }} />
            设置风险标记
          </span>
        }
        open={riskModalVisible}
        onOk={handleRiskSubmit}
        onCancel={() => setRiskModalVisible(false)}
        okText="确认设置"
        cancelText="取消"
        width={520}
        destroyOnClose
      >
        {currentCase && (
          <div style={{ marginBottom: 16, padding: '8px 12px', background: '#f5f5f5', borderRadius: 6 }}>
            <div><strong>案件编号：</strong>{pick(currentCase, ['case_no', 'caseNo'])}</div>
            <div><strong>案件标题：</strong>{pick(currentCase, ['title'])}</div>
          </div>
        )}
        <Form form={riskForm} layout="vertical">
          <Form.Item
            name="riskFlag"
            label="风险类型"
            rules={[{ required: true, message: '请选择风险类型' }]}
          >
            <Radio.Group>
              <Radio value={0}>正常（清除标记）</Radio>
              <Radio value={1}>
                <span style={{ color: '#ff4d4f' }}>⚠️ 扬言上访</span>
              </Radio>
              <Radio value={2}>
                <span style={{ color: '#cf1322' }}>🚨 极端行为</span>
              </Radio>
              <Radio value={3}>
                <span style={{ color: '#a8071a' }}>👥 群体事件</span>
              </Radio>
            </Radio.Group>
          </Form.Item>
          <Form.Item
            name="reason"
            label="标记原因"
            rules={[
              { required: true, message: '请输入标记原因' },
              { max: 500, message: '原因不超过500字' },
            ]}
          >
            <Input.TextArea
              rows={4}
              placeholder="请输入设置风险标记的原因，如：当事人多次扬言要上访、存在过激行为倾向等"
              maxLength={500}
              showCount
            />
          </Form.Item>
        </Form>
      </Modal>

      <style>{`
        @keyframes riskBlink {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.6; }
        }
        .risk-alert-row {
          background-color: #fff1f0 !important;
          border-left: 3px solid #ff4d4f;
        }
        .risk-alert-row:hover > td {
          background-color: #ffccc7 !important;
        }
        .overdue-row {
          background-color: #fffbe6 !important;
          border-left: 3px solid #faad14;
        }
        .overdue-row:hover > td {
          background-color: #fff1b8 !important;
        }
      `}</style>
    </>
  );
};

export default TodoList;
