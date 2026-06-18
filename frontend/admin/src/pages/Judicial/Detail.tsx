import React, { useState, useEffect } from 'react';
import {
  Card,
  Descriptions,
  Tag,
  Space,
  Button,
  Timeline,
  Row,
  Col,
  App,
  Modal,
  Divider,
} from 'antd';
import {
  ArrowLeftOutlined,
  SendOutlined,
  FileSearchOutlined,
  FileTextOutlined,
  SafetyOutlined,
  BellOutlined,
  WarningOutlined,
  DownloadOutlined,
  EyeOutlined,
} from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import {
  judicialService,
  JudicialConfirmation,
  JudicialConfirmLog,
  JudicialStatusMap,
  JudicialStatusColorMap,
  ActionTypeMap,
  OperatorTypeMap,
} from '../../services/judicial';
import dayjs from 'dayjs';

const { confirm } = Modal;

const JudicialDetail: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const { message } = App.useApp();
  const [detail, setDetail] = useState<JudicialConfirmation | null>(null);
  const [logs, setLogs] = useState<JudicialConfirmLog[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (id) {
      loadDetail(parseInt(id));
      loadLogs(parseInt(id));
    }
  }, [id]);

  const loadDetail = async (confirmId: number) => {
    setLoading(true);
    try {
      const res = await judicialService.getDetail(confirmId);
      setDetail(res);
    } catch (error: any) {
      message.error(error.message || '加载详情失败');
    } finally {
      setLoading(false);
    }
  };

  const loadLogs = async (confirmId: number) => {
    try {
      const res = await judicialService.getLogs(confirmId);
      setLogs(res || []);
    } catch (error: any) {
      console.error('Load logs failed:', error);
    }
  };

  const handleSubmitToCourt = async () => {
    if (!detail) return;
    confirm({
      title: '确认提交',
      icon: <SendOutlined />,
      content: `确定要将司法确认申请"${detail.confirmNo}"提交到法院吗？`,
      okText: '确认提交',
      cancelText: '取消',
      onOk: async () => {
        try {
          await judicialService.submitToCourt(detail.id);
          message.success('提交成功');
          loadDetail(detail.id);
          loadLogs(detail.id);
        } catch (error: any) {
          message.error(error.message || '提交失败');
        }
      },
    });
  };

  const handleQueryCourtStatus = async () => {
    if (!detail) return;
    try {
      await judicialService.queryCourtStatus(detail.id);
      message.success('状态同步成功');
      loadDetail(detail.id);
      loadLogs(detail.id);
    } catch (error: any) {
      message.error(error.message || '查询失败');
    }
  };

  const handleGenerateDocument = async () => {
    if (!detail) return;
    try {
      const res = await judicialService.generateDocument(detail.id);
      message.success('确认书生成成功');
      loadDetail(detail.id);
      loadLogs(detail.id);
    } catch (error: any) {
      message.error(error.message || '生成失败');
    }
  };

  const handleSealDocument = async () => {
    if (!detail) return;
    confirm({
      title: '确认签章',
      icon: <SafetyOutlined />,
      content: `确定要对确认书"${detail.confirmNo}"进行电子签章吗？`,
      okText: '确认签章',
      cancelText: '取消',
      onOk: async () => {
        try {
          await judicialService.sealDocument(detail.id);
          message.success('签章成功');
          loadDetail(detail.id);
          loadLogs(detail.id);
        } catch (error: any) {
          message.error(error.message || '签章失败');
        }
      },
    });
  };

  const handleSendPerformanceReminder = async () => {
    if (!detail) return;
    confirm({
      title: '发送履行提醒',
      icon: <BellOutlined />,
      content: `确定要向当事人发送履行提醒吗？`,
      okText: '确认发送',
      cancelText: '取消',
      onOk: async () => {
        try {
          await judicialService.sendPerformanceReminder(detail.id);
          message.success('提醒发送成功');
          loadDetail(detail.id);
          loadLogs(detail.id);
        } catch (error: any) {
          message.error(error.message || '发送失败');
        }
      },
    });
  };

  const handleSendExpirationReminder = async () => {
    if (!detail) return;
    confirm({
      title: '发送失效提醒',
      icon: <WarningOutlined />,
      content: `确认书已超过履行期限，确定要发送失效提醒并更新状态吗？`,
      okText: '确认发送',
      cancelText: '取消',
      onOk: async () => {
        try {
          await judicialService.sendExpirationReminder(detail.id);
          message.success('失效提醒发送成功');
          loadDetail(detail.id);
          loadLogs(detail.id);
        } catch (error: any) {
          message.error(error.message || '发送失败');
        }
      },
    });
  };

  const renderActionButtons = () => {
    if (!detail) return null;
    const buttons = [];

    if (detail.status === 10) {
      buttons.push(
        <Button key="submit" type="primary" icon={<SendOutlined />} onClick={handleSubmitToCourt}>
          提交法院
        </Button>
      );
    }

    if (detail.status === 20) {
      buttons.push(
        <Button key="query-status" icon={<FileSearchOutlined />} onClick={handleQueryCourtStatus}>
          查询法院状态
        </Button>
      );
    }

    if (detail.status === 30 && !detail.documentUrl) {
      buttons.push(
        <Button key="generate-doc" icon={<FileTextOutlined />} onClick={handleGenerateDocument}>
          生成确认书
        </Button>
      );
    }

    if (detail.status === 30 && detail.documentUrl && !detail.sealTime) {
      buttons.push(
        <Button key="seal" icon={<SafetyOutlined />} onClick={handleSealDocument}>
          电子签章
        </Button>
      );
    }

    if (detail.status === 30 && detail.daysLeft && detail.daysLeft <= 7 && detail.daysLeft > 0) {
      buttons.push(
        <Button key="performance-remind" icon={<BellOutlined />} onClick={handleSendPerformanceReminder}>
          发送履行提醒
        </Button>
      );
    }

    if (detail.status === 30 && detail.daysLeft !== undefined && detail.daysLeft <= 0) {
      buttons.push(
        <Button key="expiration-remind" danger icon={<WarningOutlined />} onClick={handleSendExpirationReminder}>
          发送失效提醒
        </Button>
      );
    }

    if (detail.documentUrl) {
      buttons.push(
        <Button key="view-doc" icon={<EyeOutlined />} onClick={() => window.open(detail.documentUrl!, '_blank')}>
          查看确认书
        </Button>
      );
      buttons.push(
        <Button key="download-doc" icon={<DownloadOutlined />} onClick={() => window.open(detail.documentUrl!, '_blank')}>
          下载确认书
        </Button>
      );
    }

    return buttons;
  };

  if (!detail && !loading) {
    return (
      <div style={{ padding: 24, textAlign: 'center' }}>
        <p>加载失败，请返回重试</p>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/judicial')}>
          返回列表
        </Button>
      </div>
    );
  }

  return (
    <div style={{ padding: 24 }}>
      <Card
        loading={loading}
        title={
          <Space>
            <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/judicial')}>
              返回
            </Button>
            <span>司法确认详情</span>
            {detail && (
              <Tag color={JudicialStatusColorMap[detail.status]}>
                {detail.statusName || JudicialStatusMap[detail.status]}
              </Tag>
            )}
          </Space>
        }
        extra={<Space>{renderActionButtons()}</Space>}
      >
        {detail && (
          <>
            <Descriptions bordered column={2} size="small">
              <Descriptions.Item label="确认编号">{detail.confirmNo}</Descriptions.Item>
              <Descriptions.Item label="法院案件编号">{detail.courtCaseNo || '-'}</Descriptions.Item>
              <Descriptions.Item label="关联案件">
                {detail.caseTitle}
                <div style={{ color: '#999', fontSize: 12 }}>{detail.caseNo}</div>
              </Descriptions.Item>
              <Descriptions.Item label="所属法院">{detail.courtName}</Descriptions.Item>
              <Descriptions.Item label="确认金额">
                {detail.confirmAmount ? `¥${detail.confirmAmount.toFixed(2)}` : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="履行期限">
                {detail.performanceDeadline
                  ? dayjs(detail.performanceDeadline).format('YYYY-MM-DD')
                  : '-'}
                {detail.daysLeft !== undefined && detail.status === 30 && (
                  <Tag
                    color={detail.daysLeft <= 0 ? 'error' : detail.daysLeft <= 7 ? 'warning' : 'success'}
                    style={{ marginLeft: 8 }}
                  >
                    {detail.daysLeft > 0 ? `剩余${detail.daysLeft}天` : '已逾期'}
                  </Tag>
                )}
              </Descriptions.Item>
              <Descriptions.Item label="文书编号">{detail.documentNo || '-'}</Descriptions.Item>
              <Descriptions.Item label="签章时间">
                {detail.sealTime ? dayjs(detail.sealTime).format('YYYY-MM-DD HH:mm:ss') : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="创建时间">
                {detail.createTime ? dayjs(detail.createTime).format('YYYY-MM-DD HH:mm:ss') : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="更新时间">
                {detail.updateTime ? dayjs(detail.updateTime).format('YYYY-MM-DD HH:mm:ss') : '-'}
              </Descriptions.Item>
            </Descriptions>

            <Divider orientation="left">当事人信息</Divider>
            <Row gutter={24}>
              <Col span={12}>
                <Card title="申请人" size="small">
                  <Descriptions column={1} size="small">
                    <Descriptions.Item label="姓名">{detail.applicantName}</Descriptions.Item>
                    <Descriptions.Item label="手机号">{detail.applicantPhone}</Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
              <Col span={12}>
                <Card title="被申请人" size="small">
                  <Descriptions column={1} size="small">
                    <Descriptions.Item label="姓名">{detail.respondentName}</Descriptions.Item>
                    <Descriptions.Item label="手机号">{detail.respondentPhone}</Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
            </Row>

            <Divider orientation="left">协议内容</Divider>
            <Card size="small">
              <p style={{ whiteSpace: 'pre-wrap', margin: 0 }}>{detail.agreementContent}</p>
            </Card>

            {detail.remark && (
              <>
                <Divider orientation="left">备注</Divider>
                <Card size="small">
                  <p style={{ whiteSpace: 'pre-wrap', margin: 0 }}>{detail.remark}</p>
                </Card>
              </>
            )}

            <Divider orientation="left">进度轨迹</Divider>
            <Card size="small">
              <Timeline
                mode="left"
                items={logs
                  .sort((a, b) => new Date(b.createTime).getTime() - new Date(a.createTime).getTime())
                  .map((log) => ({
                    color:
                      log.actionType === 30 || log.actionType === 50 || log.actionType === 90
                        ? 'green'
                        : log.actionType === 40 || log.actionType === 99
                        ? 'red'
                        : log.actionType === 70 || log.actionType === 80
                        ? 'orange'
                        : 'blue',
                    children: (
                      <Space direction="vertical" size={0}>
                        <Space>
                          <Tag color="blue">{ActionTypeMap[log.actionType] || log.actionType}</Tag>
                          <Tag color="default">{OperatorTypeMap[log.operatorType] || log.operatorType}</Tag>
                          {log.operatorName && <span>操作人：{log.operatorName}</span>}
                        </Space>
                        {log.remark && <div style={{ color: '#666' }}>{log.remark}</div>}
                        <div style={{ color: '#999', fontSize: 12 }}>
                          {dayjs(log.createTime).format('YYYY-MM-DD HH:mm:ss')}
                        </div>
                      </Space>
                    ),
                  }))}
              />
            </Card>
          </>
        )}
      </Card>
    </div>
  );
};

export default JudicialDetail;
