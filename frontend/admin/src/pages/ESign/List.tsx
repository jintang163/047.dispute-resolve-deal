import React, { useRef, useState } from 'react';
import { Button, Tag, Space, App, Modal, Progress, Descriptions, Tooltip } from 'antd';
import {
  FileProtectOutlined,
  EyeOutlined,
  StopOutlined,
  SendOutlined,
  QrcodeOutlined,
  SafetyCertificateOutlined,
} from '@ant-design/icons';
import {
  ProTable,
  ProFormSelect,
  ProFormText,
} from '@ant-design/pro-components';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { useNavigate } from 'react-router-dom';
import { esignService, EsignFlow } from '../../services/esign';
import dayjs from 'dayjs';

const statusColorMap: Record<number, string> = {
  0: 'default',
  10: 'processing',
  20: 'warning',
  30: 'success',
  40: 'error',
  50: 'default',
};

const statusTextMap: Record<number, string> = {
  0: '草稿',
  10: '待签署',
  20: '签署中',
  30: '已完成',
  40: '已过期',
  50: '已撤销',
};

const EsignList: React.FC = () => {
  const navigate = useNavigate();
  const { message, modal } = App.useApp();
  const actionRef = useRef<ActionType>();
  const [revokeModalOpen, setRevokeModalOpen] = useState(false);
  const [revokeFlowId, setRevokeFlowId] = useState<string>('');
  const [revokeReason, setRevokeReason] = useState('');
  const [progressModalOpen, setProgressModalOpen] = useState(false);
  const [progressData, setProgressData] = useState<any>(null);

  const columns: ProColumns<EsignFlow>[] = [
    {
      title: '签署编号',
      dataIndex: 'flowId',
      width: 160,
      copyable: true,
      fixed: 'left',
    },
    {
      title: '案件编号',
      dataIndex: 'caseNo',
      width: 140,
    },
    {
      title: '文档标题',
      dataIndex: 'docTitle',
      width: 200,
      ellipsis: true,
    },
    {
      title: '签署进度',
      width: 180,
      render: (_, record) => {
        const total = record.signerCount || 0;
        const signed = record.signedCount || 0;
        const percent = total > 0 ? Math.round((signed / total) * 100) : 0;
        return (
          <Space direction="vertical" size={0} style={{ width: '100%' }}>
            <Progress percent={percent} size="small" status={record.status === 50 ? 'exception' : undefined} />
            <span style={{ fontSize: 12, color: '#666' }}>{signed}/{total} 已签署</span>
          </Space>
        );
      },
    },
    {
      title: '骑缝章',
      dataIndex: 'crossPageSeal',
      width: 80,
      render: (_, record) => (
        record.crossPageSeal === 1
          ? <Tag color="blue">已加盖</Tag>
          : <Tag>未加盖</Tag>
      ),
    },
    {
      title: '区块链存证',
      width: 100,
      render: (_, record) => {
        if (record.bcStatus === 1) {
          return <Tag color="green" icon={<SafetyCertificateOutlined />}>已存证</Tag>;
        }
        if (record.bcStatus === 2) {
          return <Tag color="red">存证失败</Tag>;
        }
        return <Tag>未存证</Tag>;
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      valueEnum: statusTextMap,
      render: (_, record) => (
        <Tag color={statusColorMap[record.status] || 'default'}>
          {record.statusName || statusTextMap[record.status] || '-'}
        </Tag>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      width: 180,
      valueType: 'dateTime',
      sorter: true,
      render: (_, record) =>
        record.createdAt ? dayjs(record.createdAt).format('YYYY-MM-DD HH:mm:ss') : '-',
    },
    {
      title: '操作',
      key: 'option',
      valueType: 'option',
      width: 200,
      fixed: 'right',
      render: (_, record) => [
        <Button
          type="link"
          key="detail"
          icon={<EyeOutlined />}
          onClick={() => navigate(`/esign/${record.caseId}/${record.flowId}`)}
        >
          详情
        </Button>,
        <Button
          type="link"
          key="progress"
          icon={<FileProtectOutlined />}
          onClick={async () => {
            try {
              const res = await esignService.getProgress(record.caseId, record.flowId);
              setProgressData(res);
              setProgressModalOpen(true);
            } catch {
              message.error('获取签署进度失败');
            }
          }}
        >
          进度
        </Button>,
        record.status < 30 && (
          <Button
            type="link"
            key="revoke"
            danger
            icon={<StopOutlined />}
            onClick={() => {
              setRevokeFlowId(record.flowId);
              setRevokeModalOpen(true);
            }}
          >
            撤销
          </Button>
        ),
      ].filter(Boolean),
    },
  ];

  return (
    <>
      <ProTable<EsignFlow>
        columns={columns}
        actionRef={actionRef}
        cardBordered
        rowKey="flowId"
        search={{
          labelWidth: 'auto',
          defaultCollapsed: false,
        }}
        dateFormatter="string"
        headerTitle="电子签章管理"
        request={async (params) => {
          try {
            const res = await esignService.getList(0, {
              status: params.status as number,
            });
            const data = (res as any)?.data || res || [];
            return {
              data: Array.isArray(data) ? data : [],
              success: true,
              total: Array.isArray(data) ? data.length : 0,
            };
          } catch {
            return { data: [], success: false, total: 0 };
          }
        }}
        columnsState={{
          persistenceKey: 'esign-list-columns',
          persistenceType: 'localStorage',
        }}
        pagination={{
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total) => `共 ${total} 条记录`,
        }}
        scroll={{ x: 1400 }}
      />

      <Modal
        title="撤销签署流程"
        open={revokeModalOpen}
        onOk={async () => {
          if (!revokeReason.trim()) {
            message.warning('请输入撤销原因');
            return;
          }
          try {
            await esignService.revokeFlow(0, revokeFlowId, revokeReason);
            message.success('撤销成功');
            setRevokeModalOpen(false);
            setRevokeReason('');
            actionRef.current?.reload();
          } catch {
            message.error('撤销失败');
          }
        }}
        onCancel={() => {
          setRevokeModalOpen(false);
          setRevokeReason('');
        }}
      >
        <div style={{ marginBottom: 16 }}>
          <label>撤销原因：</label>
          <textarea
            style={{ width: '100%', minHeight: 80, padding: 8, borderRadius: 6, border: '1px solid #d9d9d9' }}
            value={revokeReason}
            onChange={(e) => setRevokeReason(e.target.value)}
            placeholder="请输入撤销原因"
          />
        </div>
      </Modal>

      <Modal
        title="签署进度"
        open={progressModalOpen}
        onCancel={() => setProgressModalOpen(false)}
        footer={null}
        width={600}
      >
        {progressData && (
          <div>
            <Descriptions column={2} bordered size="small">
              <Descriptions.Item label="签署状态">
                <Tag color={statusColorMap[progressData.status]}>
                  {statusTextMap[progressData.status]}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="签署进度">
                {progressData.signedCount}/{progressData.totalCount}
              </Descriptions.Item>
            </Descriptions>
            <div style={{ marginTop: 16 }}>
              <h4>签署人列表</h4>
              {(progressData.signers || []).map((signer: any, index: number) => (
                <div key={index} style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  padding: '8px 12px',
                  borderBottom: '1px solid #f0f0f0',
                }}>
                  <Space>
                    <span>{signer.userName || signer.user_name}</span>
                    <Tag color={signer.signStatus === 1 || signer.sign_status === 1 ? 'green' : 'default'}>
                      {signer.signStatus === 1 || signer.sign_status === 1 ? '已签署' : '待签署'}
                    </Tag>
                  </Space>
                  <Space>
                    {(signer.notifyStatus > 0 || signer.notify_status > 0) && (
                      <Tooltip title={signer.notifyStatusName || signer.notify_status_name}>
                        <SendOutlined style={{ color: '#1890ff' }} />
                      </Tooltip>
                    )}
                    {signer.fadadaSignUrl && (
                      <Tooltip title="法大大签署链接">
                        <QrcodeOutlined style={{ color: '#722ed1' }} />
                      </Tooltip>
                    )}
                  </Space>
                </div>
              ))}
            </div>
          </div>
        )}
      </Modal>
    </>
  );
};

export default EsignList;
