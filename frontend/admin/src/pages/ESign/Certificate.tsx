import React, { useRef, useState } from 'react';
import { Button, Tag, Space, App, Card, Descriptions, Modal, Alert } from 'antd';
import {
  SafetyCertificateOutlined,
  EyeOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  DownloadOutlined,
  QrcodeOutlined,
  SearchOutlined,
} from '@ant-design/icons';
import {
  ProTable,
  ProFormSelect,
} from '@ant-design/pro-components';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { blockchainService, BlockchainCertificate } from '../../services/esign';
import dayjs from 'dayjs';

const bcStatusMap: Record<number, { color: string; text: string }> = {
  0: { color: 'default', text: '待存证' },
  1: { color: 'green', text: '已存证' },
  2: { color: 'red', text: '存证失败' },
  3: { color: 'blue', text: '已验证' },
};

const evidenceTypeMap: Record<string, string> = {
  mediation_protocol: '调解协议',
  esign_document: '签章文档',
  evidence: '证据材料',
};

const CertificateList: React.FC = () => {
  const { message } = App.useApp();
  const actionRef = useRef<ActionType>();
  const [verifyModalOpen, setVerifyModalOpen] = useState(false);
  const [verifyResult, setVerifyResult] = useState<any>(null);
  const [verifyLoading, setVerifyLoading] = useState(false);
  const [searchCertNo, setSearchCertNo] = useState('');

  const handleVerify = async (certNo: string) => {
    setVerifyLoading(true);
    try {
      const res = await blockchainService.verifyEvidence(0, certNo);
      setVerifyResult((res as any)?.data || res);
      setVerifyModalOpen(true);
    } catch {
      message.error('验证失败');
    } finally {
      setVerifyLoading(false);
    }
  };

  const handlePublicVerify = async () => {
    if (!searchCertNo.trim()) {
      message.warning('请输入证书编号');
      return;
    }
    setVerifyLoading(true);
    try {
      const res = await blockchainService.verifyEvidence(0, searchCertNo);
      setVerifyResult((res as any)?.data || res);
      setVerifyModalOpen(true);
    } catch {
      message.error('验证失败');
    } finally {
      setVerifyLoading(false);
    }
  };

  const handleDownload = async (certNo: string) => {
    try {
      const res = await blockchainService.downloadCert(0, certNo);
      const data = (res as any)?.data || res;
      if (data.certUrl) {
        window.open(data.certUrl, '_blank');
      } else {
        message.info('证书下载链接暂不可用');
      }
    } catch {
      message.error('下载失败');
    }
  };

  const columns: ProColumns<BlockchainCertificate>[] = [
    {
      title: '证书编号',
      dataIndex: 'certNo',
      width: 180,
      copyable: true,
      fixed: 'left',
    },
    {
      title: '存证名称',
      dataIndex: 'evidenceName',
      width: 200,
      ellipsis: true,
    },
    {
      title: '存证类型',
      dataIndex: 'evidenceType',
      width: 100,
      render: (_, record) => (
        <Tag>{evidenceTypeMap[record.evidenceType] || record.evidenceType}</Tag>
      ),
    },
    {
      title: '案件编号',
      dataIndex: 'caseNo',
      width: 140,
    },
    {
      title: '交易ID',
      dataIndex: 'txId',
      width: 200,
      ellipsis: true,
      render: (_, record) => (
        <span style={{ fontFamily: 'monospace', fontSize: 11 }}>{record.txId}</span>
      ),
    },
    {
      title: '区块高度',
      dataIndex: 'blockHeight',
      width: 100,
    },
    {
      title: '上链时间',
      dataIndex: 'onChainTime',
      width: 180,
      render: (_, record) =>
        record.onChainTime ? dayjs(record.onChainTime).format('YYYY-MM-DD HH:mm:ss') : '-',
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (_, record) => (
        <Tag color={bcStatusMap[record.status]?.color}>
          {bcStatusMap[record.status]?.text || record.statusName}
        </Tag>
      ),
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
          key="verify"
          icon={<SearchOutlined />}
          onClick={() => handleVerify(record.certNo)}
        >
          验证
        </Button>,
        <Button
          type="link"
          key="qrcode"
          icon={<QrcodeOutlined />}
          onClick={() => {
            if (record.verifyUrl) window.open(record.verifyUrl, '_blank');
            else message.info('核验页面暂不可用');
          }}
        >
          核验
        </Button>,
        <Button
          type="link"
          key="download"
          icon={<DownloadOutlined />}
          onClick={() => handleDownload(record.certNo)}
        >
          下载
        </Button>,
      ],
    },
  ];

  return (
    <div>
      <Card
        title={
          <Space>
            <SafetyCertificateOutlined style={{ color: '#722ed1' }} />
            <span>扫码核验真伪</span>
          </Space>
        }
        style={{ marginBottom: 16 }}
      >
        <Space>
          <input
            placeholder="输入存证证书编号"
            value={searchCertNo}
            onChange={(e) => setSearchCertNo(e.target.value)}
            style={{
              padding: '4px 11px',
              borderRadius: 6,
              border: '1px solid #d9d9d9',
              width: 300,
            }}
          />
          <Button
            type="primary"
            icon={<SearchOutlined />}
            onClick={handlePublicVerify}
            loading={verifyLoading}
          >
            核验
          </Button>
        </Space>
      </Card>

      <ProTable<BlockchainCertificate>
        columns={columns}
        actionRef={actionRef}
        cardBordered
        rowKey="certNo"
        search={{
          labelWidth: 'auto',
          defaultCollapsed: false,
        }}
        dateFormatter="string"
        headerTitle="区块链存证证书"
        request={async (params) => {
          try {
            const res = await blockchainService.getCertList(0, {
              evidenceType: params.evidenceType as string,
              page: params.current,
              pageSize: params.pageSize,
            });
            const data = (res as any)?.data || res;
            return {
              data: data?.list || (Array.isArray(data) ? data : []),
              success: true,
              total: data?.total || 0,
            };
          } catch {
            return { data: [], success: false, total: 0 };
          }
        }}
        pagination={{
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total) => `共 ${total} 条记录`,
        }}
        scroll={{ x: 1400 }}
      />

      <Modal
        title="存证核验结果"
        open={verifyModalOpen}
        onCancel={() => setVerifyModalOpen(false)}
        footer={
          <Button onClick={() => setVerifyModalOpen(false)}>关闭</Button>
        }
        width={600}
      >
        {verifyResult && (
          <div>
            {verifyResult.valid ? (
              <Alert
                type="success"
                message="验证通过"
                description="该存证证书在区块链上的记录与提交的数据一致，未被篡改"
                showIcon
                icon={<CheckCircleOutlined />}
                style={{ marginBottom: 16 }}
              />
            ) : (
              <Alert
                type="error"
                message="验证失败"
                description="该存证证书的区块链记录与提交的数据不一致，可能已被篡改"
                showIcon
                icon={<CloseCircleOutlined />}
                style={{ marginBottom: 16 }}
              />
            )}
            <Descriptions column={1} bordered size="small">
              <Descriptions.Item label="证书编号">{verifyResult.certNo}</Descriptions.Item>
              <Descriptions.Item label="证据名称">{verifyResult.evidenceName}</Descriptions.Item>
              <Descriptions.Item label="证据哈希">
                <span style={{ fontFamily: 'monospace', fontSize: 11 }}>{verifyResult.evidenceHash}</span>
              </Descriptions.Item>
              <Descriptions.Item label="区块链交易ID">
                <span style={{ fontFamily: 'monospace', fontSize: 11 }}>{verifyResult.txId}</span>
              </Descriptions.Item>
              <Descriptions.Item label="区块高度">{verifyResult.blockHeight}</Descriptions.Item>
              <Descriptions.Item label="上链时间">
                {verifyResult.onChainTime
                  ? dayjs(verifyResult.onChainTime).format('YYYY-MM-DD HH:mm:ss')
                  : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="数据来源">
                <Tag color={verifyResult.source === 'blockchain' ? 'green' : 'blue'}>
                  {verifyResult.source === 'blockchain' ? '区块链实时验证' : '数据库记录'}
                </Tag>
              </Descriptions.Item>
            </Descriptions>
          </div>
        )}
      </Modal>
    </div>
  );
};

export default CertificateList;
