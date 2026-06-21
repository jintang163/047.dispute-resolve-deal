import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Descriptions, Card, Tag, Rate, Button, Space, App, Spin } from 'antd';
import { ArrowLeftOutlined, CopyOutlined, StarOutlined, ArchiveOutlined, ThunderboltOutlined } from '@ant-design/icons';
import { caseLibraryService, CaseLibraryItem } from '../../services/caseLibrary';
import dayjs from 'dayjs';

const difficultyMap: Record<number, { text: string; color: string }> = {
  1: { text: '简单', color: 'green' },
  2: { text: '一般', color: 'blue' },
  3: { text: '复杂', color: 'orange' },
  4: { text: '疑难', color: 'red' },
};

const CaseLibraryDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { message, modal } = App.useApp();
  const [loading, setLoading] = useState(true);
  const [data, setData] = useState<CaseLibraryItem | null>(null);

  useEffect(() => {
    if (id) {
      loadData();
    }
  }, [id]);

  const loadData = async () => {
    setLoading(true);
    try {
      const res = await caseLibraryService.getDetail(Number(id));
      setData(res.data?.data || null);
    } catch {
      message.error('加载案例详情失败');
    } finally {
      setLoading(false);
    }
  };

  const handleCopyTactics = () => {
    if (data?.mediationTactics) {
      navigator.clipboard.writeText(data.mediationTactics);
      message.success('话术已复制到剪贴板');
    }
  };

  const handleCopyKeyPoints = () => {
    if (data?.keyPoints) {
      navigator.clipboard.writeText(data.keyPoints);
      message.success('要点已复制到剪贴板');
    }
  };

  const handleArchive = () => {
    modal.confirm({
      title: '确认归档',
      content: '归档后案例将移至历史库，确认归档？',
      onOk: async () => {
        try {
          await caseLibraryService.archive(Number(id));
          message.success('归档成功');
          loadData();
        } catch { message.error('归档失败'); }
      },
    });
  };

  if (loading) return <Spin size="large" style={{ display: 'block', marginTop: 100 }} />;
  if (!data) return <div>案例不存在</div>;

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/case-library')}>返回列表</Button>
        <Button icon={<CopyOutlined />} onClick={handleCopyTactics}>复制话术</Button>
        <Button icon={<CopyOutlined />} onClick={handleCopyKeyPoints}>复制要点</Button>
        <Button icon={<ThunderboltOutlined />} onClick={async () => {
          try {
            await caseLibraryService.vectorize(Number(id));
            message.success('已开始向量化处理');
          } catch { message.error('操作失败'); }
        }}>重新向量化</Button>
        {data.status === 1 && (
          <Button danger icon={<ArchiveOutlined />} onClick={handleArchive}>归档</Button>
        )}
        {data.status === 2 && (
          <Button type="primary" onClick={async () => {
            try {
              await caseLibraryService.restore(Number(id));
              message.success('恢复成功');
              loadData();
            } catch { message.error('恢复失败'); }
          }}>从归档恢复</Button>
        )}
      </Space>

      <Card title="基本信息">
        <Descriptions column={2} bordered>
          <Descriptions.Item label="案例编号">{data.caseNo}</Descriptions.Item>
          <Descriptions.Item label="标题">{data.title}</Descriptions.Item>
          <Descriptions.Item label="纠纷类型">{data.disputeType}</Descriptions.Item>
          <Descriptions.Item label="难度等级">
            <Tag color={difficultyMap[data.difficultyLevel]?.color}>
              {difficultyMap[data.difficultyLevel]?.text}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="调解结果">
            <Tag color={data.isSuccess === 1 ? 'success' : 'error'}>
              {data.isSuccess === 1 ? '调解成功' : '调解未成'}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="调解员">{data.mediatorName || '-'}</Descriptions.Item>
          <Descriptions.Item label="调解组织">{data.orgName || '-'}</Descriptions.Item>
          <Descriptions.Item label="关键词">{data.keywords || '-'}</Descriptions.Item>
          <Descriptions.Item label="标签">{data.tags || '-'}</Descriptions.Item>
          <Descriptions.Item label="来源案件ID">{data.sourceCaseId || '-'}</Descriptions.Item>
          <Descriptions.Item label="案例描述" span={2}>{data.description || '-'}</Descriptions.Item>
        </Descriptions>
      </Card>

      <Card title="调解经验" style={{ marginTop: 16 }}>
        <Descriptions column={1} bordered>
          <Descriptions.Item label="调解话术">
            <div style={{ whiteSpace: 'pre-wrap' }}>{data.mediationTactics || '-'}</div>
          </Descriptions.Item>
          <Descriptions.Item label="调解要点">
            <div style={{ whiteSpace: 'pre-wrap' }}>{data.keyPoints || '-'}</div>
          </Descriptions.Item>
          <Descriptions.Item label="结果摘要">
            <div style={{ whiteSpace: 'pre-wrap' }}>{data.resultSummary || '-'}</div>
          </Descriptions.Item>
        </Descriptions>
      </Card>

      <Card title="评价统计" style={{ marginTop: 16 }}>
        <Descriptions column={3} bordered>
          <Descriptions.Item label="平均评分">
            <Space>
              <Rate disabled value={Math.round(data.avgScore)} />
              <span>{data.avgScore?.toFixed(1) || '0.0'}</span>
            </Space>
          </Descriptions.Item>
          <Descriptions.Item label="评分次数">{data.scoreCount}</Descriptions.Item>
          <Descriptions.Item label="引用次数">{data.referenceCount}</Descriptions.Item>
          <Descriptions.Item label="最后使用时间">
            {data.lastUsedAt ? dayjs(data.lastUsedAt).format('YYYY-MM-DD HH:mm') : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="向量化状态">
            {data.vectorStatus === 0 && <Tag>未处理</Tag>}
            {data.vectorStatus === 1 && <Tag color="processing">处理中</Tag>}
            {data.vectorStatus === 2 && <Tag color="success">已完成</Tag>}
            {data.vectorStatus === 3 && <Tag color="error">失败</Tag>}
          </Descriptions.Item>
          <Descriptions.Item label="状态">
            {data.status === 0 && <Tag>禁用</Tag>}
            {data.status === 1 && <Tag color="success">启用</Tag>}
            {data.status === 2 && <Tag color="warning">已归档</Tag>}
          </Descriptions.Item>
        </Descriptions>
      </Card>
    </div>
  );
};

export default CaseLibraryDetail;
