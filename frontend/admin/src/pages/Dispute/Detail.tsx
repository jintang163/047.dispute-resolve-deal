import React, { useEffect, useState } from 'react';
import {
  Descriptions,
  Card,
  Row,
  Col,
  Timeline,
  Tag,
  Button,
  Space,
  Spin,
  Divider,
  App,
  Drawer,
  Modal,
  List,
  Avatar,
  Alert,
  Rate,
  Empty,
  Typography,
} from 'antd';
const { Text } = Typography;
import {
  ArrowLeftOutlined,
  EditOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  CloseCircleOutlined,
  SafetyCertificateOutlined,
  RobotOutlined,
  FileTextOutlined,
  UserOutlined,
  TeamOutlined,
  WarningOutlined,
  BookOutlined,
  CopyOutlined,
  StarOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import { ProDescriptions } from '@ant-design/pro-components';
import { disputeService, DisputeDetail, MediatorOption } from '../../services/dispute';
import { caseLibraryService, CaseSearchResult } from '../../services/caseLibrary';
import ProtocolGenerator from '../Mediation/ProtocolGenerator';
import MediationRecordEditor, { MediationRecordFormData } from '../../components/MediationRecordEditor';
import { MediationRecord } from '../../services/user';
import dayjs from 'dayjs';

const statusColorMap: Record<string, string> = {
  pending: 'default',
  assigning: 'processing',
  mediating: 'blue',
  pending_approval: 'orange',
  approved: 'green',
  rejected: 'red',
  completed: 'success',
  closed: 'default',
};

const statusTextMap: Record<string, string> = {
  pending: '待分配',
  assigning: '分配中',
  mediating: '调解中',
  pending_approval: '待审批',
  approved: '审批通过',
  rejected: '审批驳回',
  completed: '已完成',
  closed: '已结案',
};

const typeTextMap: Record<string, string> = {
  civil: '民事纠纷',
  labor: '劳动争议',
  family: '家庭纠纷',
  neighborhood: '邻里纠纷',
  contract: '合同纠纷',
  property: '物业纠纷',
  other: '其他纠纷',
};

const getTimelineIcon = (status: string) => {
  switch (status) {
    case 'approved':
    case 'completed':
      return <CheckCircleOutlined style={{ color: '#52c41a', fontSize: 16 }} />;
    case 'rejected':
      return <CloseCircleOutlined style={{ color: '#ff4d4f', fontSize: 16 }} />;
    case 'pending':
    case 'pending_approval':
      return <ClockCircleOutlined style={{ color: '#faad14', fontSize: 16 }} />;
    default:
      return <SafetyCertificateOutlined style={{ color: '#1677ff', fontSize: 16 }} />;
  }
};

const getTimelineColor = (status: string) => {
  switch (status) {
    case 'approved':
    case 'completed':
      return 'green';
    case 'rejected':
      return 'red';
    case 'pending':
    case 'pending_approval':
      return 'orange';
    default:
      return 'blue';
  }
};

const DisputeDetail: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const { message, modal } = App.useApp();
  const [loading, setLoading] = useState(false);
  const [detail, setDetail] = useState<DisputeDetail | null>(null);
  const [protocolDrawerOpen, setProtocolDrawerOpen] = useState(false);

  const [assignModalOpen, setAssignModalOpen] = useState(false);
  const [mediators, setMediators] = useState<MediatorOption[]>([]);
  const [mediatorsLoading, setMediatorsLoading] = useState(false);
  const [selectedMediator, setSelectedMediator] = useState<MediatorOption | null>(null);
  const [assigning, setAssigning] = useState(false);

  const [similarCases, setSimilarCases] = useState<CaseSearchResult[]>([]);
  const [similarCasesLoading, setSimilarCasesLoading] = useState(false);
  const [scoreModalOpen, setScoreModalOpen] = useState(false);
  const [scoreCaseId, setScoreCaseId] = useState<number>(0);
  const [scoreValue, setScoreValue] = useState(3);

  const [recordEditorOpen, setRecordEditorOpen] = useState(false);
  const [editingRecord, setEditingRecord] = useState<MediationRecord | undefined>(undefined);
  const [mediationRecords, setMediationRecords] = useState<MediationRecord[]>([]);
  const [recordsLoading, setRecordsLoading] = useState(false);

  useEffect(() => {
    if (id) {
      fetchDetail();
      fetchSimilarCases();
      fetchMediationRecords();
    }
  }, [id]);

  const fetchMediationRecords = async () => {
    if (!id) return;
    setRecordsLoading(true);
    try {
      const res = await disputeService.getMediationRecords(id);
      const list: any[] = (res as any)?.data?.list || (res as any)?.data || [];
      setMediationRecords(list);
    } catch {
    } finally {
      setRecordsLoading(false);
    }
  };

  const fetchSimilarCases = async () => {
    if (!id) return;
    setSimilarCasesLoading(true);
    try {
      const res = await caseLibraryService.searchSimilar(undefined, 5, Number(id));
      setSimilarCases(res.data?.data || []);
    } catch (error: any) {
      message.warning(error.message || '相似案例检索失败');
    } finally {
      setSimilarCasesLoading(false);
    }
  };

  const handleQuote = async (record: CaseSearchResult, quoteType: number) => {
    if (!id) return;
    modal.confirm({
      title: '引用案例',
      content: `确认引用「${record.title}」的${quoteType === 1 ? '话术' : quoteType === 2 ? '策略' : '全文'}到当前案件的调解记录？`,
      onOk: async () => {
        try {
          const res = await caseLibraryService.quote(record.caseId || record.id, Number(id), quoteType);
          const quoteContent = res.data?.data?.quoteContent || '';
          const mediationRecordId = res.data?.data?.mediationRecordId;
          if (quoteContent) {
            await navigator.clipboard.writeText(quoteContent);
            message.success(
              `引用成功，已写入调解记录${mediationRecordId ? '（记录ID：' + mediationRecordId + '）' : ''}，同时已复制到剪贴板`,
            );
          } else {
            message.success('引用成功');
          }
        } catch (error: any) {
          message.error(error.message || '引用失败');
        }
      },
    });
  };

  const handleScore = async () => {
    try {
      await caseLibraryService.score(scoreCaseId, scoreValue, Number(id));
      message.success('评分成功，感谢您的反馈');
      setScoreModalOpen(false);
    } catch {
      message.error('评分失败');
    }
  };

  const openRecordEditor = (record?: MediationRecord) => {
    setEditingRecord(record);
    setRecordEditorOpen(true);
  };

  const handleSubmitRecord = async (data: MediationRecordFormData, _isDraft: boolean) => {
    if (!id) return { id: '' };

    const submitPayload: any = {
      recordType: data.recordType,
      mediationTime: data.mediationTime,
      place: data.mediationPlace,
      mediationDuration: data.mediationDuration,
      participants: Array.isArray(data.participants) ? data.participants : [],
      processContent: data.processContent,
      disputeFocus: data.disputeFocus,
      mediationOpinion: data.mediationOpinion,
      agreementContent: data.agreementContent,
      nextStep: data.nextStep,
      result: data.result,
      isDraft: data.isDraft,
      templateId: data.templateId,
      templateName: data.templateName,
    };

    let res;
    if (data.id) {
      res = await disputeService.updateMediationRecord(id, String(data.id), submitPayload);
    } else {
      res = await disputeService.createMediationRecord(id, submitPayload);
    }

    await fetchMediationRecords();
    return (res as any)?.data || res;
  };

  const fetchDetail = async () => {
    try {
      setLoading(true);
      const res = await disputeService.getDetail(id!);
      const data = res.data || res;
      setDetail(data);
    } catch (error: any) {
      message.error(error.message || '获取案件详情失败');
    } finally {
      setLoading(false);
    }
  };

  const openAssignModal = async () => {
    setAssignModalOpen(true);
    setSelectedMediator(null);
    setMediatorsLoading(true);
    try {
      const res = await disputeService.getMediatorsForAssign();
      const data: any = (res as any)?.data ?? res;
      const list: MediatorOption[] = Array.isArray(data) ? data : data.list || [];
      const normalized = list.map((m: any) => ({
        ...m,
        pendingCaseCount: m.pending_case_count ?? m.pendingCaseCount ?? 0,
        isHighLoad: m.is_high_load ?? m.isHighLoad ?? false,
      }));
      setMediators(normalized.sort((a, b) => (a.pendingCaseCount || 0) - (b.pendingCaseCount || 0)));
    } catch (error: any) {
      message.error(error.message || '获取调解员列表失败');
    } finally {
      setMediatorsLoading(false);
    }
  };

  const handleSelectMediator = (mediator: MediatorOption) => {
    if (mediator.isHighLoad) {
      modal.warning({
        title: '调解员负载较高',
        icon: <WarningOutlined style={{ color: '#faad14' }} />,
        content: (
          <div>
            <p>调解员 <strong>{mediator.realName}</strong> 当前待办案件为 <strong style={{ color: '#ff4d4f' }}>{mediator.pendingCaseCount}</strong> 件，超过推荐上限 10 件。</p>
            <p style={{ marginTop: 8 }}>该调解员负载较高，建议选择其他人，避免任务分配不均导致案件积压。</p>
            <p style={{ marginTop: 8, color: '#999', fontSize: 12 }}>是否仍确认选择该调解员？</p>
          </div>
        ),
        okText: '确认选择',
        cancelText: '重新选择',
        onOk: () => {
          setSelectedMediator(mediator);
        },
      });
    } else {
      setSelectedMediator(mediator);
    }
  };

  const handleConfirmAssign = async () => {
    if (!selectedMediator) {
      message.warning('请选择调解员');
      return;
    }
    setAssigning(true);
    try {
      await disputeService.assignMediator(id!, String(selectedMediator.id));
      message.success('分派成功');
      setAssignModalOpen(false);
      fetchDetail();
    } catch (error: any) {
      message.error(error.message || '分派失败');
    } finally {
      setAssigning(false);
    }
  };

  return (
    <Spin spinning={loading}>
      <Space direction="vertical" style={{ width: '100%' }} size={16}>
        <Card
          bordered={false}
          style={{ borderRadius: 8 }}
          extra={
            <Space>
              <Button icon={<ArrowLeftOutlined />} onClick={() => navigate(-1)}>
                返回
              </Button>
              <Button icon={<EditOutlined />}>编辑</Button>
              <Button
                icon={<RobotOutlined />}
                onClick={() => setProtocolDrawerOpen(true)}
              >
                AI生成协议
              </Button>
              <Button
                icon={<FileTextOutlined />}
                onClick={() => openRecordEditor()}
                type="primary"
              >
                录入调解记录
              </Button>
              <Button type="primary" icon={<TeamOutlined />} onClick={openAssignModal}>
                分配调解员
              </Button>
            </Space>
          }
          title={
            <Space>
              <span>案件详情</span>
              <Tag color={statusColorMap[detail?.caseInfo?.status || ''] || 'default'}>
                {detail?.caseInfo?.statusName || statusTextMap[detail?.caseInfo?.status || ''] || '-'}
              </Tag>
            </Space>
          }
        >
          <ProDescriptions
            column={2}
            dataSource={detail?.caseInfo}
            title="基本信息"
            size="default"
          >
            <ProDescriptions.Item label="案件编号" dataIndex="caseNo" copyable />
            <ProDescriptions.Item
              label="案件类型"
              dataIndex="type"
              render={(val) => <Tag color="blue">{typeTextMap[val] || val}</Tag>}
            />
            <ProDescriptions.Item label="案件标题" dataIndex="title" span={2} />
            <ProDescriptions.Item label="所属组织" dataIndex="orgName" />
            <ProDescriptions.Item label="调解员" dataIndex="mediatorName" />
            <ProDescriptions.Item label="甲方" dataIndex="partyA" />
            <ProDescriptions.Item label="甲方电话" dataIndex="partyAPhone" />
            <ProDescriptions.Item label="乙方" dataIndex="partyB" />
            <ProDescriptions.Item label="乙方电话" dataIndex="partyBPhone" />
            <ProDescriptions.Item label="纠纷地点" dataIndex="address" span={2} />
            <ProDescriptions.Item label="案件描述" dataIndex="description" span={2} />
            <ProDescriptions.Item
              label="创建时间"
              dataIndex="createTime"
              render={(val) => (val ? dayjs(val).format('YYYY-MM-DD HH:mm:ss') : '-')}
            />
            <ProDescriptions.Item
              label="更新时间"
              dataIndex="updateTime"
              render={(val) => (val ? dayjs(val).format('YYYY-MM-DD HH:mm:ss') : '-')}
            />
          </ProDescriptions>
        </Card>

        <Row gutter={16}>
          <Col span={12}>
            <Card title="当事人信息" bordered={false} style={{ borderRadius: 8 }}>
              <Row gutter={[16, 16]}>
                {detail?.parties?.map((party, index) => (
                  <Col span={12} key={party.id || index}>
                    <Card
                      size="small"
                      title={
                        <Tag color={party.type === 'A' ? 'blue' : 'green'}>
                          {party.type === 'A' ? '甲方' : '乙方'}
                        </Tag>
                      }
                    >
                      <Descriptions column={1} size="small">
                        <Descriptions.Item label="姓名">{party.name}</Descriptions.Item>
                        <Descriptions.Item label="联系电话">{party.phone || '-'}</Descriptions.Item>
                        <Descriptions.Item label="身份证">{party.idCard || '-'}</Descriptions.Item>
                        <Descriptions.Item label="住址">{party.address || '-'}</Descriptions.Item>
                      </Descriptions>
                    </Card>
                  </Col>
                ))}
                {(!detail?.parties || detail.parties.length === 0) && (
                  <Col span={24}>
                    <div style={{ textAlign: 'center', padding: 24, color: '#999' }}>暂无当事人信息</div>
                  </Col>
                )}
              </Row>
            </Card>
          </Col>
          <Col span={12}>
            <Card title="审批进度" bordered={false} style={{ borderRadius: 8 }}>
              <Timeline
                items={
                  detail?.approvalRecords && detail.approvalRecords.length > 0
                    ? detail.approvalRecords.map((record) => ({
                        dot: getTimelineIcon(record.status),
                        color: getTimelineColor(record.status),
                        children: (
                          <div style={{ paddingLeft: 8 }}>
                            <div style={{ fontWeight: 500 }}>{record.stepName}</div>
                            <div style={{ fontSize: 12, color: '#666', marginTop: 4 }}>
                              审批人: {record.approverName || '待处理'}
                            </div>
                            {record.opinion && (
                              <div style={{ fontSize: 12, color: '#666', marginTop: 4 }}>
                                意见: {record.opinion}
                              </div>
                            )}
                            <div style={{ fontSize: 12, color: '#999', marginTop: 4 }}>
                              {record.updateTime || record.createTime
                                ? dayjs(record.updateTime || record.createTime).format(
                                    'YYYY-MM-DD HH:mm:ss',
                                  )
                                : '-'}
                            </div>
                          </div>
                        ),
                      }))
                    : [
                        {
                          children: (
                            <div style={{ color: '#999', padding: 8 }}>暂无审批记录</div>
                          ),
                        },
                      ]
                }
              />
            </Card>
          </Col>
        </Row>

        <Card
          bordered={false}
          style={{ borderRadius: 8 }}
          title={
            <Space>
              <FileTextOutlined style={{ color: '#1677ff' }} />
              <span>调解记录</span>
              {mediationRecords.filter((r) => r.isDraft).length > 0 && (
                <Tag color="orange">
                  {mediationRecords.filter((r) => r.isDraft).length} 份草稿
                </Tag>
              )}
            </Space>
          }
          extra={
            <Button
              type="primary"
              size="small"
              icon={<PlusOutlined />}
              onClick={() => openRecordEditor()}
            >
              新增记录
            </Button>
          }
        >
          <Spin spinning={recordsLoading}>
            {mediationRecords.length > 0 ? (
              <Space direction="vertical" style={{ width: '100%' }} size={12}>
                {mediationRecords.map((record) => (
                  <Card
                    key={record.id}
                    size="small"
                    type={record.isDraft ? 'inner' : undefined}
                    style={{
                      borderRadius: 6,
                      borderLeft: record.isDraft ? '3px solid #faad14' : undefined,
                    }}
                    title={
                      <Space>
                        {record.isDraft && (
                          <Tag color="orange" icon={<ClockCircleOutlined />}>
                            草稿
                          </Tag>
                        )}
                        {record.recordTypeName && <Tag color="blue">{record.recordTypeName}</Tag>}
                        <Text strong>
                          {record.mediationTime
                            ? dayjs(record.mediationTime).format('YYYY-MM-DD HH:mm')
                            : '未记录时间'}
                        </Text>
                        {record.templateName && (
                          <Tag color="purple" icon={<BookOutlined />}>
                            套用模板：{record.templateName}
                          </Tag>
                        )}
                        {record.resultName && (
                          <Tag
                            color={
                              record.result === 'success' || record.result === '1'
                                ? 'green'
                                : record.result === 'failed' || record.result === '2'
                                ? 'red'
                                : record.result === 'partial' || record.result === '3'
                                ? 'blue'
                                : 'default'
                            }
                          >
                            {record.resultName}
                          </Tag>
                        )}
                        {record.duration && (
                          <Tag>时长 {record.duration} 分钟</Tag>
                        )}
                      </Space>
                    }
                    extra={
                      <Space>
                        <Button
                          type="link"
                          size="small"
                          icon={<EditOutlined />}
                          onClick={() => openRecordEditor(record)}
                        >
                          {record.isDraft ? '继续编辑' : '编辑'}
                        </Button>
                      </Space>
                    }
                  >
                    {record.processContent && (
                      <div style={{ marginBottom: 6 }}>
                        <Text type="secondary" style={{ fontSize: 12 }}>调解过程：</Text>
                        <div
                          style={{
                            fontSize: 13,
                            lineHeight: 1.6,
                            color: '#333',
                            marginTop: 4,
                            whiteSpace: 'pre-wrap',
                            maxHeight: 80,
                            overflow: 'hidden',
                          }}
                        >
                          {record.processContent}
                        </div>
                      </div>
                    )}
                    {record.agreementContent && (
                      <div>
                        <Text type="secondary" style={{ fontSize: 12 }}>协议内容：</Text>
                        <div
                          style={{
                            fontSize: 13,
                            lineHeight: 1.6,
                            color: '#1677ff',
                            marginTop: 4,
                            whiteSpace: 'pre-wrap',
                            maxHeight: 60,
                            overflow: 'hidden',
                          }}
                        >
                          {record.agreementContent}
                        </div>
                      </div>
                    )}
                  </Card>
                ))}
              </Space>
            ) : (
              !recordsLoading && (
                <Empty
                  description={
                    <Space direction="vertical" style={{ textAlign: 'center' }}>
                      <span>暂无调解记录</span>
                      <Button type="primary" size="small" onClick={() => openRecordEditor()}>
                        录入第一条调解记录
                      </Button>
                    </Space>
                  }
                />
              )
            )}
          </Spin>
        </Card>

        <Card
          bordered={false}
          style={{ borderRadius: 8 }}
          title={
            <Space>
              <BookOutlined style={{ color: '#1677ff' }} />
              <span>相似案例推荐</span>
              <Tag color="blue" style={{ marginLeft: 8 }}>
                基于案件描述 AI 语义检索
              </Tag>
            </Space>
          }
          extra={
            <Button
              type="text"
              icon={<ReloadOutlined />}
              loading={similarCasesLoading}
              onClick={fetchSimilarCases}
            >
              重新检索
            </Button>
          }
        >
          <Spin spinning={similarCasesLoading}>
            {similarCases.length > 0 ? (
              <Row gutter={[12, 12]}>
                {similarCases.map((item, idx) => (
                  <Col span={12} key={item.caseId || item.id || idx}>
                    <Card
                      size="small"
                      style={{ height: '100%' }}
                      title={
                        <Space>
                          <Tag color="blue">#{idx + 1}</Tag>
                          <span
                            style={{
                              maxWidth: 180,
                              overflow: 'hidden',
                              textOverflow: 'ellipsis',
                              whiteSpace: 'nowrap',
                              display: 'inline-block',
                            }}
                          >
                            {item.title}
                          </span>
                          <Tag color="green">相似度 {(item.score * 100).toFixed(1)}%</Tag>
                          {item.difficultyLevel && (
                            <Tag color={['', 'green', 'blue', 'orange', 'red'][item.difficultyLevel] || 'default'}>
                              {['', '简单', '一般', '复杂', '疑难'][item.difficultyLevel] || ''}
                            </Tag>
                          )}
                        </Space>
                      }
                      extra={
                        <Button
                          type="link"
                          size="small"
                          icon={<StarOutlined />}
                          onClick={() => {
                            setScoreCaseId(item.caseId || item.id);
                            setScoreValue(3);
                            setScoreModalOpen(true);
                          }}
                        >
                          评分
                        </Button>
                      }
                    >
                      {item.disputeType && (
                        <div style={{ fontSize: 12, color: '#666', marginBottom: 6 }}>
                          纠纷类型：{item.disputeType}
                        </div>
                      )}
                      {item.mediationTactics && (
                        <div style={{ marginBottom: 6 }}>
                          <div style={{ fontSize: 12, color: '#888', marginBottom: 2 }}>调解话术：</div>
                          <div
                            style={{
                              fontSize: 13,
                              maxHeight: 52,
                              overflow: 'hidden',
                              lineHeight: 1.4,
                              color: '#333',
                            }}
                          >
                            {item.mediationTactics}
                          </div>
                        </div>
                      )}
                      {item.keyPoints && (
                        <div style={{ marginBottom: 8 }}>
                          <div style={{ fontSize: 12, color: '#888', marginBottom: 2 }}>调解要点：</div>
                          <div
                            style={{
                              fontSize: 13,
                              maxHeight: 40,
                              overflow: 'hidden',
                              lineHeight: 1.4,
                              color: '#333',
                            }}
                          >
                            {item.keyPoints}
                          </div>
                        </div>
                      )}
                      <Space wrap style={{ marginTop: 4 }}>
                        <Button
                          size="small"
                          icon={<CopyOutlined />}
                          onClick={() => handleQuote(item, 1)}
                        >
                          引用话术
                        </Button>
                        <Button
                          size="small"
                          icon={<CopyOutlined />}
                          onClick={() => handleQuote(item, 2)}
                        >
                          引用策略
                        </Button>
                        <Button
                          size="small"
                          type="primary"
                          icon={<FileTextOutlined />}
                          onClick={() => handleQuote(item, 3)}
                        >
                          全文引用
                        </Button>
                        <Button
                          type="link"
                          size="small"
                          onClick={() => navigate(`/case-library/${item.caseId || item.id}`)}
                        >
                          查看详情
                        </Button>
                      </Space>
                    </Card>
                  </Col>
                ))}
              </Row>
            ) : (
              !similarCasesLoading && (
                <Empty description="暂无相似案例，可尝试录入更多典型案例或点击重新检索" />
              )
            )}
          </Spin>
        </Card>
      </Space>

      <Drawer
        title={
          <Space>
            <RobotOutlined style={{ color: '#533483' }} />
            <span>AI智能生成调解协议</span>
          </Space>
        }
        width={1400}
        open={protocolDrawerOpen}
        onClose={() => setProtocolDrawerOpen(false)}
        destroyOnClose
      >
        {id && <ProtocolGenerator caseId={id} onClose={() => setProtocolDrawerOpen(false)} />}
      </Drawer>

      <Modal
        title={
          <Space>
            <TeamOutlined />
            <span>分配调解员</span>
          </Space>
        }
        width={760}
        open={assignModalOpen}
        onCancel={() => setAssignModalOpen(false)}
        onOk={handleConfirmAssign}
        okText="确认分派"
        cancelText="取消"
        confirmLoading={assigning}
        okButtonProps={{ disabled: !selectedMediator }}
        destroyOnClose
      >
        <Alert
          type="info"
          showIcon
          style={{ marginBottom: 16 }}
          message="负载均衡提示"
          description={
            <span>
              系统实时显示各调解员当前待办案件数量。待办超过 <strong>10 件</strong> 的调解员将高亮警示，
              建议优先选择负载较低的调解员，避免任务分配不均。
            </span>
          }
        />
        {selectedMediator && (
          <Alert
            type={selectedMediator.isHighLoad ? 'warning' : 'success'}
            showIcon
            style={{ marginBottom: 16 }}
            message={
              <Space>
                <span>已选择：</span>
                <Avatar size="small" icon={<UserOutlined />} src={selectedMediator.avatar} />
                <strong>{selectedMediator.realName}</strong>
                <Tag color={selectedMediator.isHighLoad ? 'red' : 'green'}>
                  待办 {selectedMediator.pendingCaseCount || 0} 件
                </Tag>
                {selectedMediator.orgName && <Tag color="blue">{selectedMediator.orgName}</Tag>}
              </Space>
            }
          />
        )}
        <Spin spinning={mediatorsLoading}>
          <List
            dataSource={mediators}
            grid={{ gutter: 12, xs: 1, sm: 2, md: 2, lg: 2, xl: 3, xxl: 3 }}
            renderItem={(mediator) => {
              const isSelected = selectedMediator?.id === mediator.id;
              const highLoad = !!mediator.isHighLoad;
              return (
                <List.Item
                  onClick={() => handleSelectMediator(mediator)}
                  style={{ cursor: 'pointer' }}
                >
                  <Card
                    size="small"
                    style={{
                      borderWidth: 2,
                      borderColor: isSelected ? '#1677ff' : highLoad ? '#ff4d4f' : '#f0f0f0',
                      boxShadow: isSelected ? '0 2px 8px rgba(22,119,255,0.25)' : 'none',
                      transition: 'all 0.2s',
                    }}
                    hoverable
                    bodyStyle={{ padding: '12px 14px' }}
                  >
                    <Space align="start" style={{ width: '100%' }}>
                      <Avatar size={40} icon={<UserOutlined />} src={mediator.avatar} />
                      <div style={{ flex: 1, minWidth: 0 }}>
                        <Space style={{ marginBottom: 4 }} wrap>
                          <strong>{mediator.realName}</strong>
                          {highLoad ? (
                            <Tag color="red" icon={<WarningOutlined />}>
                              高负载
                            </Tag>
                          ) : (
                            <Tag color="green">正常</Tag>
                          )}
                        </Space>
                        <Space size={8} style={{ marginBottom: 4 }} wrap>
                          <span style={{ color: '#666', fontSize: 12 }}>
                            <UserOutlined style={{ marginRight: 4 }} />
                            待办 <strong style={{ color: highLoad ? '#ff4d4f' : '#1677ff' }}>
                              {mediator.pendingCaseCount || 0}
                            </strong> 件
                          </span>
                          {mediator.phone && (
                            <span style={{ color: '#999', fontSize: 12 }}>{mediator.phone}</span>
                          )}
                        </Space>
                        {mediator.specialty && (
                          <div style={{ color: '#888', fontSize: 12 }}>
                            专长：{mediator.specialty}
                          </div>
                        )}
                        {mediator.orgName && (
                          <div style={{ color: '#999', fontSize: 12, marginTop: 2 }}>
                            {mediator.orgName}
                          </div>
                        )}
                      </div>
                    </Space>
                  </Card>
                </List.Item>
              );
            }}
          />
        </Spin>
      </Modal>

      <MediationRecordEditor
        open={recordEditorOpen}
        onClose={() => setRecordEditorOpen(false)}
        caseId={id || ''}
        caseInfo={{
          typeName: detail?.caseInfo?.typeName,
          mediatorId: detail?.caseInfo?.mediatorId,
          mediatorName: detail?.caseInfo?.mediatorName,
        }}
        initialData={editingRecord}
        onSubmit={handleSubmitRecord}
      />

      <Modal
        title="案例评分"
        open={scoreModalOpen}
        onCancel={() => setScoreModalOpen(false)}
        onOk={handleScore}
        okText="提交评分"
        destroyOnClose
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          <div>请对该推荐案例的有用性进行评分：</div>
          <Rate value={scoreValue} onChange={setScoreValue} />
          <div style={{ color: '#999' }}>{scoreValue} 分</div>
        </Space>
      </Modal>
    </Spin>
  );
};

export default DisputeDetail;
