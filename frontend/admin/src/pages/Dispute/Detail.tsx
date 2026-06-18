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
} from 'antd';
import {
  ArrowLeftOutlined,
  EditOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  CloseCircleOutlined,
  SafetyCertificateOutlined,
} from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import { ProDescriptions } from '@ant-design/pro-components';
import { disputeService, DisputeDetail } from '../../services/dispute';
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
  const { message } = App.useApp();
  const [loading, setLoading] = useState(false);
  const [detail, setDetail] = useState<DisputeDetail | null>(null);

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
              <Button type="primary">分配调解员</Button>
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
    </Spin>
  );
};

export default DisputeDetail;
