import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Card,
  Descriptions,
  Tag,
  Button,
  Space,
  Steps,
  Timeline,
  App,
  Spin,
  Divider,
  Modal,
  Input,
  Alert,
  Row,
  Col,
} from 'antd';
import {
  ArrowLeftOutlined,
  SafetyCertificateOutlined,
  SendOutlined,
  QrcodeOutlined,
  DownloadOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  FileProtectOutlined,
  StopOutlined,
} from '@ant-design/icons';
import { esignService, blockchainService, EsignFlow, EsignSigner } from '../../services/esign';
import dayjs from 'dayjs';

const statusColorMap: Record<number, string> = {
  0: 'default', 10: 'processing', 20: 'warning', 30: 'success', 40: 'error', 50: 'default',
};

const statusTextMap: Record<number, string> = {
  0: '草稿', 10: '待签署', 20: '签署中', 30: '已完成', 40: '已过期', 50: '已撤销',
};

const bcStatusMap: Record<number, { color: string; text: string }> = {
  0: { color: 'default', text: '待存证' },
  1: { color: 'green', text: '已存证' },
  2: { color: 'red', text: '存证失败' },
  3: { color: 'blue', text: '已验证' },
};

const EsignDetail: React.FC = () => {
  const { caseId, flowId } = useParams<{ caseId: string; flowId: string }>();
  const navigate = useNavigate();
  const { message } = App.useApp();
  const [loading, setLoading] = useState(true);
  const [detail, setDetail] = useState<EsignFlow | null>(null);
  const [verifyModalOpen, setVerifyModalOpen] = useState(false);
  const [verifyCode, setVerifyCode] = useState('');
  const [revokeModalOpen, setRevokeModalOpen] = useState(false);
  const [revokeReason, setRevokeReason] = useState('');
  const [polling, setPolling] = useState(false);

  const fetchDetail = async () => {
    if (!caseId || !flowId) return;
    try {
      setLoading(true);
      const res = await esignService.getDetail(caseId, flowId);
      setDetail((res as any)?.data || res);
    } catch {
      message.error('获取详情失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDetail();
  }, [caseId, flowId]);

  useEffect(() => {
    if (!detail) return;
    const isActive = detail.status === 10 || detail.status === 20;
    if (isActive && !polling) {
      setPolling(true);
      const timer = setInterval(() => {
        fetchDetail();
      }, 15000);
      return () => {
        clearInterval(timer);
        setPolling(false);
      };
    }
  }, [detail?.status]);

  const handleSendCode = async () => {
    if (!caseId || !flowId) return;
    try {
      await esignService.sendVerifyCode(caseId, flowId);
      message.success('验证码已发送');
    } catch {
      message.error('发送验证码失败');
    }
  };

  const handleSign = async () => {
    if (!caseId || !flowId || !detail) return;
    try {
      await esignService.signDocument(caseId, flowId, {
        recordId: parseInt(detail.id || '0'),
        verifyCode,
      });
      message.success('签署成功');
      setVerifyModalOpen(false);
      setVerifyCode('');
      fetchDetail();
    } catch {
      message.error('签署失败');
    }
  };

  const handleRevoke = async () => {
    if (!caseId || !flowId) return;
    try {
      await esignService.revokeFlow(caseId, flowId, revokeReason);
      message.success('撤销成功');
      setRevokeModalOpen(false);
      setRevokeReason('');
      fetchDetail();
    } catch {
      message.error('撤销失败');
    }
  };

  const handleVerifyBlockchain = async () => {
    if (!caseId || !detail?.blockchainCert?.certNo) return;
    try {
      const res = await blockchainService.verifyEvidence(caseId, detail.blockchainCert.certNo);
      const data = (res as any)?.data || res;
      if (data.valid) {
        message.success('区块链存证验证通过');
      } else {
        message.error('区块链存证验证失败，数据可能被篡改');
      }
      fetchDetail();
    } catch {
      message.error('验证失败');
    }
  };

  const handleDownloadCert = async () => {
    if (!caseId || !detail?.blockchainCert?.certNo) return;
    try {
      const res = await blockchainService.downloadCert(caseId, detail.blockchainCert.certNo);
      const data = (res as any)?.data || res;
      if (data.certUrl) {
        window.open(data.certUrl, '_blank');
      } else {
        message.info('证书下载链接暂不可用');
      }
    } catch {
      message.error('下载证书失败');
    }
  };

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '100px 0' }}>
        <Spin size="large" tip="加载中..." />
      </div>
    );
  }

  if (!detail) {
    return <Alert type="error" message="签署流程不存在" />;
  }

  const bcCert = detail.blockchainCert;

  return (
    <div style={{ padding: 24 }}>
      <div style={{ marginBottom: 16 }}>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/esign')}>
          返回列表
        </Button>
      </div>

      <Card title="签署流程详情" style={{ marginBottom: 16 }}>
        <Descriptions column={3} bordered size="small">
          <Descriptions.Item label="签署编号">{detail.flowId}</Descriptions.Item>
          <Descriptions.Item label="案件编号">{detail.caseNo}</Descriptions.Item>
          <Descriptions.Item label="文档标题">{detail.docTitle}</Descriptions.Item>
          <Descriptions.Item label="签署状态">
            <Tag color={statusColorMap[detail.status]}>
              {detail.statusName || statusTextMap[detail.status]}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="签署进度">
            {detail.signedCount}/{detail.signerCount}
          </Descriptions.Item>
          <Descriptions.Item label="骑缝章">
            {detail.crossPageSeal === 1
              ? <Tag color="blue">已自动加盖</Tag>
              : <Tag>未加盖</Tag>}
          </Descriptions.Item>
          <Descriptions.Item label="过期时间">
            {detail.expireTime ? dayjs(detail.expireTime).format('YYYY-MM-DD HH:mm:ss') : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="法大大流程ID">
            {detail.fadadaFlowId || '-'}
          </Descriptions.Item>
          <Descriptions.Item label="创建时间">
            {detail.createdAt ? dayjs(detail.createdAt).format('YYYY-MM-DD HH:mm:ss') : '-'}
          </Descriptions.Item>
        </Descriptions>

        <div style={{ marginTop: 16 }}>
          <Space>
            {(detail.status === 10 || detail.status === 20) && (
              <Button type="primary" icon={<CheckCircleOutlined />} onClick={() => setVerifyModalOpen(true)}>
                签署文件
              </Button>
            )}
            {detail.status < 30 && (
              <Button danger icon={<StopOutlined />} onClick={() => setRevokeModalOpen(true)}>
                撤销流程
              </Button>
            )}
            {(detail.status === 10 || detail.status === 20) && (
              <Button icon={<SendOutlined />} onClick={handleSendCode}>
                发送验证码
              </Button>
            )}
          </Space>
        </div>
      </Card>

      <Card title="签署人列表" style={{ marginBottom: 16 }}>
        <Steps
          current={detail.signedCount}
          items={(detail.signers || []).map((signer: EsignSigner) => ({
            title: signer.userName,
            description: (
              <div>
                <Tag color={signer.signStatus === 1 ? 'green' : 'default'}>
                  {signer.signStatus === 1 ? '已签署' : '待签署'}
                </Tag>
                {signer.signedAt && (
                  <div style={{ fontSize: 12, color: '#999', marginTop: 4 }}>
                    {dayjs(signer.signedAt).format('HH:mm:ss')}
                  </div>
                )}
                <div style={{ marginTop: 4 }}>
                  <Space size={4}>
                    {(signer.notifyStatus || 0) > 0 && (
                      <Tooltip title={`通知方式: ${signer.notifyStatusName}`}>
                        <SendOutlined style={{ color: '#1890ff', fontSize: 12 }} />
                      </Tooltip>
                    )}
                    {signer.fadadaSignUrl && (
                      <Tooltip title="法大大签署链接已生成">
                        <QrcodeOutlined style={{ color: '#722ed1', fontSize: 12 }} />
                      </Tooltip>
                    )}
                  </Space>
                </div>
              </div>
            ),
            icon: signer.signStatus === 1
              ? <CheckCircleOutlined style={{ color: '#52c41a' }} />
              : <ClockCircleOutlined />,
          }))}
        />
      </Card>

      <Card
        title={
          <Space>
            <SafetyCertificateOutlined style={{ color: '#722ed1' }} />
            <span>区块链存证</span>
          </Space>
        }
        style={{ marginBottom: 16 }}
      >
        {bcCert ? (
          <>
            <Alert
              type="success"
              message="文档已上链存证"
              description="该签署文档已完成区块链存证，具备司法效力"
              showIcon
              style={{ marginBottom: 16 }}
            />
            <Descriptions column={2} bordered size="small">
              <Descriptions.Item label="存证证书编号">
                <Space>
                  {bcCert.certNo}
                  <Button
                    type="link"
                    size="small"
                    icon={<QrcodeOutlined />}
                    onClick={() => {
                      if (bcCert.verifyUrl) window.open(bcCert.verifyUrl, '_blank');
                    }}
                  >
                    扫码核验
                  </Button>
                </Space>
              </Descriptions.Item>
              <Descriptions.Item label="区块链交易ID">
                <span style={{ fontFamily: 'monospace', fontSize: 12 }}>{bcCert.txId}</span>
              </Descriptions.Item>
              <Descriptions.Item label="区块高度">{bcCert.blockHeight}</Descriptions.Item>
              <Descriptions.Item label="上链时间">
                {bcCert.onChainTime ? dayjs(bcCert.onChainTime).format('YYYY-MM-DD HH:mm:ss') : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="证据哈希">
                <span style={{ fontFamily: 'monospace', fontSize: 11 }}>{bcCert.evidenceHash}</span>
              </Descriptions.Item>
              <Descriptions.Item label="存证状态">
                <Tag color={bcStatusMap[bcCert.status]?.color}>
                  {bcStatusMap[bcCert.status]?.text}
                </Tag>
              </Descriptions.Item>
            </Descriptions>
            <div style={{ marginTop: 16 }}>
              <Space>
                <Button
                  type="primary"
                  icon={<SafetyCertificateOutlined />}
                  onClick={handleVerifyBlockchain}
                >
                  验证存证
                </Button>
                <Button icon={<DownloadOutlined />} onClick={handleDownloadCert}>
                  下载存证证书
                </Button>
              </Space>
            </div>
          </>
        ) : detail.status === 30 ? (
          <Alert
            type="info"
            message="该文档尚未进行区块链存证"
            description="签署完成后，可手动将文档哈希上链存证"
            showIcon
          />
        ) : (
          <Alert
            type="warning"
            message="签署尚未完成"
            description="签署完成并自动上链后，存证信息将在此显示"
            showIcon
          />
        )}
      </Card>

      <Modal
        title="签署文件"
        open={verifyModalOpen}
        onOk={handleSign}
        onCancel={() => { setVerifyModalOpen(false); setVerifyCode(''); }}
      >
        <div style={{ marginBottom: 16 }}>
          <Alert type="info" message="请先发送验证码到手机，然后输入验证码完成签署" />
        </div>
        <Space>
          <Input
            placeholder="请输入验证码"
            value={verifyCode}
            onChange={(e) => setVerifyCode(e.target.value)}
            style={{ width: 200 }}
          />
          <Button onClick={handleSendCode}>发送验证码</Button>
        </Space>
      </Modal>

      <Modal
        title="撤销签署流程"
        open={revokeModalOpen}
        onOk={handleRevoke}
        onCancel={() => { setRevokeModalOpen(false); setRevokeReason(''); }}
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
    </div>
  );
};

export default EsignDetail;
