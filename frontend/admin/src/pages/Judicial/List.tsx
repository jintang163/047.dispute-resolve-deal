import React, { useState, useRef } from 'react';
import { Button, Tag, Space, App, Modal, Dropdown, MenuProps } from 'antd';
import {
  PlusOutlined,
  EyeOutlined,
  SendOutlined,
  FileSearchOutlined,
  FileTextOutlined,
  SafetyOutlined,
  BellOutlined,
  WarningOutlined,
  MoreOutlined,
} from '@ant-design/icons';
import {
  ProTable,
  ProFormSelect,
  ProFormDateRangePicker,
  ProFormText,
  ModalForm,
  ProForm,
} from '@ant-design/pro-components';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { useNavigate } from 'react-router-dom';
import {
  judicialService,
  JudicialConfirmation,
  JudicialStatusMap,
  JudicialStatusColorMap,
  CreateJudicialParams,
} from '../../services/judicial';
import dayjs from 'dayjs';

const { confirm } = Modal;

const JudicialList: React.FC = () => {
  const navigate = useNavigate();
  const { message } = App.useApp();
  const actionRef = useRef<ActionType>();
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [courtOptions, setCourtOptions] = useState<{ label: string; value: number }[]>([]);

  React.useEffect(() => {
    loadCourtOptions();
  }, []);

  const loadCourtOptions = async () => {
    try {
      const res = await judicialService.getCourtOptions();
      if (res) {
        setCourtOptions(res.map((item: any) => ({
          label: item.courtName,
          value: item.id,
        })));
      }
    } catch (error) {
      console.error('Load court options failed:', error);
    }
  };

  const handleSubmitToCourt = async (record: JudicialConfirmation) => {
    confirm({
      title: '确认提交',
      icon: <SendOutlined />,
      content: `确定要将司法确认申请"${record.confirmNo}"提交到法院吗？`,
      okText: '确认提交',
      cancelText: '取消',
      onOk: async () => {
        try {
          await judicialService.submitToCourt(record.id);
          message.success('提交成功');
          actionRef.current?.reload();
        } catch (error: any) {
          message.error(error.message || '提交失败');
        }
      },
    });
  };

  const handleGenerateDocument = async (record: JudicialConfirmation) => {
    try {
      await judicialService.generateDocument(record.id);
      message.success('确认书生成成功');
      actionRef.current?.reload();
    } catch (error: any) {
      message.error(error.message || '生成失败');
    }
  };

  const handleSealDocument = async (record: JudicialConfirmation) => {
    confirm({
      title: '确认签章',
      icon: <SafetyOutlined />,
      content: `确定要对确认书"${record.confirmNo}"进行电子签章吗？`,
      okText: '确认签章',
      cancelText: '取消',
      onOk: async () => {
        try {
          await judicialService.sealDocument(record.id);
          message.success('签章成功');
          actionRef.current?.reload();
        } catch (error: any) {
          message.error(error.message || '签章失败');
        }
      },
    });
  };

  const handleSendPerformanceReminder = async (record: JudicialConfirmation) => {
    confirm({
      title: '发送履行提醒',
      icon: <BellOutlined />,
      content: `确定要向当事人发送履行提醒吗？`,
      okText: '确认发送',
      cancelText: '取消',
      onOk: async () => {
        try {
          await judicialService.sendPerformanceReminder(record.id);
          message.success('提醒发送成功');
          actionRef.current?.reload();
        } catch (error: any) {
          message.error(error.message || '发送失败');
        }
      },
    });
  };

  const handleSendExpirationReminder = async (record: JudicialConfirmation) => {
    confirm({
      title: '发送失效提醒',
      icon: <WarningOutlined />,
      content: `确认书已超过履行期限，确定要发送失效提醒并更新状态吗？`,
      okText: '确认发送',
      cancelText: '取消',
      onOk: async () => {
        try {
          await judicialService.sendExpirationReminder(record.id);
          message.success('失效提醒发送成功');
          actionRef.current?.reload();
        } catch (error: any) {
          message.error(error.message || '发送失败');
        }
      },
    });
  };

  const getActionMenu = (record: JudicialConfirmation): MenuProps => {
    const items: MenuProps['items'] = [];

    if (record.status === 10) {
      items.push({
        key: 'submit',
        label: '提交法院',
        icon: <SendOutlined />,
        onClick: () => handleSubmitToCourt(record),
      });
    }

    if (record.status === 20) {
      items.push({
        key: 'query-status',
        label: '查询法院状态',
        icon: <FileSearchOutlined />,
        onClick: async () => {
          try {
            await judicialService.queryCourtStatus(record.id);
            message.success('状态同步成功');
            actionRef.current?.reload();
          } catch (error: any) {
            message.error(error.message || '查询失败');
          }
        },
      });
    }

    if (record.status === 30 && !record.documentUrl) {
      items.push({
        key: 'generate-doc',
        label: '生成确认书',
        icon: <FileTextOutlined />,
        onClick: () => handleGenerateDocument(record),
      });
    }

    if (record.status === 30 && record.documentUrl && !record.sealTime) {
      items.push({
        key: 'seal',
        label: '电子签章',
        icon: <SafetyOutlined />,
        onClick: () => handleSealDocument(record),
      });
    }

    if (record.status === 30 && record.daysLeft && record.daysLeft <= 7 && record.daysLeft > 0) {
      items.push({
        key: 'performance-remind',
        label: '发送履行提醒',
        icon: <BellOutlined />,
        onClick: () => handleSendPerformanceReminder(record),
      });
    }

    if (record.status === 30 && record.daysLeft !== undefined && record.daysLeft <= 0) {
      items.push({
        key: 'expiration-remind',
        label: '发送失效提醒',
        icon: <WarningOutlined />,
        onClick: () => handleSendExpirationReminder(record),
      });
    }

    return { items };
  };

  const columns: ProColumns<JudicialConfirmation>[] = [
    {
      title: '确认编号',
      dataIndex: 'confirmNo',
      width: 180,
      copyable: true,
      fixed: 'left',
    },
    {
      title: '关联案件',
      dataIndex: 'caseTitle',
      width: 200,
      ellipsis: true,
      render: (_, record) => (
        <div>
          <div>{record.caseTitle}</div>
          <div style={{ color: '#999', fontSize: 12 }}>{record.caseNo}</div>
        </div>
      ),
    },
    {
      title: '申请人',
      dataIndex: 'applicantName',
      width: 100,
      ellipsis: true,
    },
    {
      title: '被申请人',
      dataIndex: 'respondentName',
      width: 100,
      ellipsis: true,
    },
    {
      title: '法院',
      dataIndex: 'courtName',
      width: 150,
      ellipsis: true,
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      valueEnum: JudicialStatusMap,
      render: (_, entity) => (
        <Space direction="vertical" size={2}>
          <Tag color={JudicialStatusColorMap[entity.status] || 'default'}>
            {entity.statusName || JudicialStatusMap[entity.status] || entity.status}
          </Tag>
          {entity.daysLeft !== undefined && entity.status === 30 && (
            <span style={{ fontSize: 12, color: entity.daysLeft <= 0 ? '#ff4d4f' : entity.daysLeft <= 7 ? '#faad14' : '#52c41a' }}>
              {entity.daysLeft > 0 ? `剩余${entity.daysLeft}天` : '已逾期'}
            </span>
          )}
        </Space>
      ),
    },
    {
      title: '确认金额',
      dataIndex: 'confirmAmount',
      width: 120,
      render: (val) => val ? `¥${val.toFixed(2)}` : '-',
    },
    {
      title: '履行期限',
      dataIndex: 'performanceDeadline',
      width: 180,
      valueType: 'date',
      render: (_, entity) => entity.performanceDeadline ? dayjs(entity.performanceDeadline).format('YYYY-MM-DD') : '-',
    },
    {
      title: '创建时间',
      dataIndex: 'createTime',
      width: 180,
      valueType: 'dateTime',
      sorter: true,
      render: (_, entity) => entity.createTime ? dayjs(entity.createTime).format('YYYY-MM-DD HH:mm:ss') : '-',
    },
    {
      title: '操作',
      key: 'option',
      valueType: 'option',
      width: 180,
      fixed: 'right',
      render: (_, record) => [
        <Button
          type="link"
          key="view"
          icon={<EyeOutlined />}
          onClick={() => navigate(`/judicial/${record.id}`)}
        >
          详情
        </Button>,
        getActionMenu(record).items && getActionMenu(record).items!.length > 0 && (
          <Dropdown key="more" menu={getActionMenu(record)} trigger={['click']}>
            <Button type="link" icon={<MoreOutlined />}>
              更多
            </Button>
          </Dropdown>
        ),
      ],
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      <ProTable<JudicialConfirmation>
        headerTitle="司法确认管理"
        actionRef={actionRef}
        rowKey="id"
        search={{
          labelWidth: 120,
        }}
        toolBarRender={() => [
          <Button
            key="create"
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => setCreateModalVisible(true)}
          >
            新建申请
          </Button>,
        ]}
        request={async (params, sort) => {
          const res = await judicialService.getList({
            page: params.current,
            pageSize: params.pageSize,
            status: params.status,
            keyword: params.keyword,
            startTime: params.createTime?.[0] ? dayjs(params.createTime[0]).format('YYYY-MM-DD') : undefined,
            endTime: params.createTime?.[1] ? dayjs(params.createTime[1]).format('YYYY-MM-DD') : undefined,
          });
          return {
            data: res.list,
            success: true,
            total: res.total,
          };
        }}
        columns={columns}
        pagination={{
          showSizeChanger: true,
          showQuickJumper: true,
        }}
      />

      <ModalForm<CreateJudicialParams>
        title="新建司法确认申请"
        open={createModalVisible}
        onOpenChange={setCreateModalVisible}
        width={800}
        layout="vertical"
        onFinish={async (values) => {
          try {
            await judicialService.create(values as any);
            message.success('创建成功');
            actionRef.current?.reload();
            return true;
          } catch (error: any) {
            message.error(error.message || '创建失败');
            return false;
          }
        }}
      >
        <ProForm.Group title="案件信息">
          <ProFormText
            name="caseId"
            label="关联案件ID"
            rules={[{ required: true, message: '请输入关联案件ID' }]}
            placeholder="请输入关联案件ID"
          />
          <ProFormText
            name="caseTitle"
            label="案件标题"
            placeholder="请输入案件标题"
          />
        </ProForm.Group>

        <ProForm.Group title="申请人信息">
          <ProFormText
            name="applicantName"
            label="姓名"
            rules={[{ required: true, message: '请输入申请人姓名' }]}
            placeholder="请输入申请人姓名"
          />
          <ProFormText
            name="applicantPhone"
            label="手机号"
            rules={[{ required: true, message: '请输入申请人手机号' }]}
            placeholder="请输入申请人手机号"
          />
          <ProFormText
            name="applicantIdCard"
            label="身份证号"
            placeholder="请输入身份证号"
          />
          <ProFormText
            name="applicantAddress"
            label="地址"
            placeholder="请输入地址"
          />
        </ProForm.Group>

        <ProForm.Group title="被申请人信息">
          <ProFormText
            name="respondentName"
            label="姓名"
            rules={[{ required: true, message: '请输入被申请人姓名' }]}
            placeholder="请输入被申请人姓名"
          />
          <ProFormText
            name="respondentPhone"
            label="手机号"
            rules={[{ required: true, message: '请输入被申请人手机号' }]}
            placeholder="请输入被申请人手机号"
          />
          <ProFormText
            name="respondentIdCard"
            label="身份证号"
            placeholder="请输入身份证号"
          />
          <ProFormText
            name="respondentAddress"
            label="地址"
            placeholder="请输入地址"
          />
        </ProForm.Group>

        <ProForm.Group title="确认信息">
          <ProFormSelect
            name="courtId"
            label="选择法院"
            rules={[{ required: true, message: '请选择法院' }]}
            options={courtOptions}
            placeholder="请选择法院"
          />
          <ProFormText
            name="performanceDeadline"
            label="履行期限"
            placeholder="请选择履行期限"
          />
          <ProFormText
            name="confirmAmount"
            label="确认金额(元)"
            placeholder="请输入确认金额"
          />
          <ProFormText
            name="agreementContent"
            label="协议内容"
            rules={[{ required: true, message: '请输入协议内容' }]}
            fieldProps={{
              autoSize: { minRows: 4, maxRows: 8 },
            }}
            placeholder="请输入协议内容"
          />
        </ProForm.Group>

        <ProFormText
          name="remark"
          label="备注"
          fieldProps={{
            autoSize: { minRows: 2, maxRows: 4 },
          }}
          placeholder="请输入备注"
        />
      </ModalForm>
    </div>
  );
};

export default JudicialList;
