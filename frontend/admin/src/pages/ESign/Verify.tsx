import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import {
  Result,
  Descriptions,
  Spin,
  Card,
  Typography,
  Space,
  Tag,
  Divider,
  Button,
} from 'antd';
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  SafetyCertificateOutlined,
  CopyOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';

const { Title, Text } = Typography;

interface VerifyResult {
  valid: boolean;
  certNo: string;
  evidenceName: string;
  evidenceHash: string;
  txId: string;
  blockHeight: number;
  onChainTime: string;
  caseNo: string;
  caseTitle: string;
  evidenceType: string;
}

const evidenceTypeMap: Record<string, string> = {
  mediation_protocol: '调解协议',
  esign_document: '签章文档',
  evidence: '证据材料',
};

const PublicVerify: React.FC = () => {
  const { certNo } = useParams<{ certNo: string }>();
  const [loading, setLoading] = useState(true);
  const [result, setResult] = useState<VerifyResult | null>(null);
  const [error, setError] = useState<string>('');

  useEffect(() => {
    if (!certNo) {
      setError('缺少证书编号');
      setLoading(false);
      return;
    }

    const doVerify = async () => {
      try {
        const res = await fetch(`/api/v1/public/blockchain/verify/${certNo}`);
        const json = await res.json();
        if (json.code === 0 || json.code === 200) {
          setResult(json.data || json);
        } else {
          setError(json.message || json.msg || '核验失败');
        }
      } catch {
        setError('网络请求失败，请稍后重试');
      } finally {
        setLoading(false);
      }
    };

    doVerify();
  }, [certNo]);

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text).catch(() => {});
  };

  if (loading) {
    return (
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        minHeight: '100vh',
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
      }}>
        <Card style={{ textAlign: 'center', padding: 40, borderRadius: 16 }}>
          <Spin size="large" />
          <div style={{ marginTop: 16 }}>
            <Text type="secondary">正在核验存证证书...</Text>
          </div>
        </Card>
      </div>
    );
  }

  if (error) {
    return (
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        minHeight: '100vh',
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
      }}>
        <Card style={{ maxWidth: 480, width: '90%', borderRadius: 16 }}>
          <Result
            status="error"
            title="核验失败"
            subTitle={error}
          />
        </Card>
      </div>
    );
  }

  if (!result) return null;

  return (
    <div style={{
      minHeight: '100vh',
      background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
      padding: '24px 16px',
    }}>
      <div style={{ maxWidth: 560, margin: '0 auto' }}>
        <div style={{ textAlign: 'center', marginBottom: 24 }}>
          <SafetyCertificateOutlined style={{ fontSize: 48, color: '#fff' }} />
          <Title level={3} style={{ color: '#fff', marginTop: 12, marginBottom: 4 }}>
            区块链存证核验
          </Title>
          <Text style={{ color: 'rgba(255,255,255,0.8)' }}>
            司法存证链 · 扫码核验真伪
          </Text>
        </div>

        <Card style={{ borderRadius: 16, boxShadow: '0 8px 32px rgba(0,0,0,0.12)' }}>
          <Result
            status={result.valid ? 'success' : 'error'}
            icon={result.valid ? <CheckCircleOutlined /> : <CloseCircleOutlined />}
            title={result.valid ? '验证通过' : '验证失败'}
            subTitle={result.valid
              ? '该存证证书在区块链上的记录与提交的数据一致，未被篡改'
              : '该存证证书的区块链记录与提交的数据不一致，可能已被篡改'}
          />

          <Divider />

          <Descriptions column={1} bordered size="small" labelStyle={{ width: 120 }}>
            <Descriptions.Item label="证书编号">
              <Space>
                <Text code style={{ fontSize: 13 }}>{result.certNo}</Text>
                <Button
                  type="text"
                  size="small"
                  icon={<CopyOutlined />}
                  onClick={() => copyToClipboard(result.certNo)}
                />
              </Space>
            </Descriptions.Item>
            <Descriptions.Item label="证据名称">
              {result.evidenceName}
            </Descriptions.Item>
            {result.evidenceType && (
              <Descriptions.Item label="证据类型">
                <Tag color="blue">
                  {evidenceTypeMap[result.evidenceType] || result.evidenceType}
                </Tag>
              </Descriptions.Item>
            )}
            {result.caseNo && (
              <Descriptions.Item label="案件编号">
                {result.caseNo}
              </Descriptions.Item>
            )}
            {result.caseTitle && (
              <Descriptions.Item label="案件标题">
                {result.caseTitle}
              </Descriptions.Item>
            )}
            <Descriptions.Item label="证据哈希">
              <Space>
                <Text code style={{ fontSize: 11, wordBreak: 'break-all' }}>
                  {result.evidenceHash}
                </Text>
                <Button
                  type="text"
                  size="small"
                  icon={<CopyOutlined />}
                  onClick={() => copyToClipboard(result.evidenceHash)}
                />
              </Space>
            </Descriptions.Item>
            <Descriptions.Item label="区块链交易ID">
              <Space>
                <Text code style={{ fontSize: 11, wordBreak: 'break-all' }}>
                  {result.txId}
                </Text>
                <Button
                  type="text"
                  size="small"
                  icon={<CopyOutlined />}
                  onClick={() => copyToClipboard(result.txId)}
                />
              </Space>
            </Descriptions.Item>
            <Descriptions.Item label="区块高度">
              {result.blockHeight}
            </Descriptions.Item>
            <Descriptions.Item label="上链时间">
              {result.onChainTime
                ? dayjs(result.onChainTime).format('YYYY-MM-DD HH:mm:ss')
                : '-'}
            </Descriptions.Item>
          </Descriptions>

          <Divider />

          <div style={{ textAlign: 'center' }}>
            <Text type="secondary" style={{ fontSize: 12 }}>
              本核验结果由司法存证链提供，仅供参考。核验时间：{dayjs().format('YYYY-MM-DD HH:mm:ss')}
            </Text>
          </div>
        </Card>
      </div>
    </div>
  );
};

export default PublicVerify;
