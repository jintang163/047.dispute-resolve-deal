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
} from 'antd';
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
} from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import { ProDescriptions } from '@ant-design/pro-components';
import { disputeService, DisputeDetail, MediatorOption } from '../../services/dispute';
import ProtocolGenerator from '../Mediation/ProtocolGenerator';
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

  useEffect(() => {
    if (id) {
      fetchDetail();
    }
  }, [id]);

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
    </Spin>
  );
};

export default DisputeDetail;
