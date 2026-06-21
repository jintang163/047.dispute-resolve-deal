import React, { useRef, useState } from 'react';
import {
  Button, Tag, Space, App, Modal, Rate, Input, Select, Tabs, Card, Descriptions, Empty,
} from 'antd';
import {
  BookOutlined, SearchOutlined, StarOutlined, CopyOutlined, ArchiveOutlined,
  PlusOutlined, ReloadOutlined, UndoOutlined, ThunderboltOutlined,
} from '@ant-design/icons';
import {
  ProTable, ProFormSelect, ProFormText, ModalForm,
} from '@ant-design/pro-components';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { useNavigate } from 'react-router-dom';
import { caseLibraryService, CaseLibraryItem, CaseSearchResult } from '../../services/caseLibrary';
import dayjs from 'dayjs';

const { TextArea } = Input;

const difficultyMap: Record<number, { text: string; color: string }> = {
  1: { text: '简单', color: 'green' },
  2: { text: '一般', color: 'blue' },
  3: { text: '复杂', color: 'orange' },
  4: { text: '疑难', color: 'red' },
};

const statusMap: Record<number, { text: string; color: string }> = {
  0: { text: '禁用', color: 'default' },
  1: { text: '启用', color: 'success' },
  2: { text: '已归档', color: 'warning' },
};

const CaseLibraryList: React.FC = () => {
  const navigate = useNavigate();
  const { message, modal } = App.useApp();
  const actionRef = useRef<ActionType>();
  const [searchModalOpen, setSearchModalOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<CaseSearchResult[]>([]);
  const [searchLoading, setSearchLoading] = useState(false);
  const [scoreModalOpen, setScoreModalOpen] = useState(false);
  const [scoreCaseId, setScoreCaseId] = useState<number>(0);
  const [scoreValue, setScoreValue] = useState(3);

  const handleSearch = async () => {
    if (!searchQuery.trim()) {
      message.warning('请输入查询内容');
      return;
    }
    setSearchLoading(true);
    try {
      const res = await caseLibraryService.searchSimilar(searchQuery, 10);
      setSearchResults(res.data?.data || []);
    } catch {
      message.error('搜索失败');
    } finally {
      setSearchLoading(false);
    }
  };

  const handleQuote = async (record: CaseSearchResult | CaseLibraryItem, quoteType: number) => {
    modal.confirm({
      title: '引用案例',
      content: `确认引用「${record.title}」的${quoteType === 1 ? '话术' : quoteType === 2 ? '策略' : '全文'}到当前案件？`,
      onOk: async () => {
        try {
          const res = await caseLibraryService.quote(record.id || (record as CaseSearchResult).caseId, 0, quoteType);
          const quoteContent = res.data?.data?.quoteContent || '';
          if (quoteContent) {
            await navigator.clipboard.writeText(quoteContent);
            message.success('已复制到剪贴板，可粘贴到当前记录中');
          } else {
            message.success('引用成功');
          }
        } catch {
          message.error('引用失败');
        }
      },
    });
  };

  const handleArchive = (id: number) => {
    modal.confirm({
      title: '确认归档',
      content: '归档后案例将移至历史库，可通过恢复功能重新启用。确认归档？',
      onOk: async () => {
        try {
          await caseLibraryService.archive(id);
          message.success('归档成功');
          actionRef.current?.reload();
        } catch {
          message.error('归档失败');
        }
      },
    });
  };

  const handleScore = async () => {
    try {
      await caseLibraryService.score(scoreCaseId, scoreValue);
      message.success('评分成功');
      setScoreModalOpen(false);
      actionRef.current?.reload();
    } catch {
      message.error('评分失败');
    }
  };

  const columns: ProColumns<CaseLibraryItem>[] = [
    {
      title: '案例编号',
      dataIndex: 'caseNo',
      width: 140,
      copyable: true,
    },
    {
      title: '标题',
      dataIndex: 'title',
      width: 200,
      ellipsis: true,
    },
    {
      title: '纠纷类型',
      dataIndex: 'disputeType',
      width: 120,
    },
    {
      title: '难度',
      dataIndex: 'difficultyLevel',
      width: 80,
      render: (_, record) => {
        const d = difficultyMap[record.difficultyLevel];
        return d ? <Tag color={d.color}>{d.text}</Tag> : '-';
      },
    },
    {
      title: '结果',
      dataIndex: 'isSuccess',
      width: 80,
      render: (_, record) => (
        <Tag color={record.isSuccess === 1 ? 'success' : 'error'}>
          {record.isSuccess === 1 ? '成功' : '未成'}
        </Tag>
      ),
    },
    {
      title: '平均评分',
      dataIndex: 'avgScore',
      width: 120,
      render: (_, record) => (
        <Space>
          <Rate disabled value={Math.round(record.avgScore)} style={{ fontSize: 14 }} />
          <span>{record.avgScore?.toFixed(1) || '0.0'}</span>
        </Space>
      ),
    },
    {
      title: '引用次数',
      dataIndex: 'referenceCount',
      width: 80,
      sorter: true,
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 80,
      render: (_, record) => {
        const s = statusMap[record.status];
        return s ? <Tag color={s.color}>{s.text}</Tag> : '-';
      },
    },
    {
      title: '最后使用',
      dataIndex: 'lastUsedAt',
      width: 120,
      render: (_, record) => record.lastUsedAt ? dayjs(record.lastUsedAt).format('YYYY-MM-DD') : '-',
    },
    {
      title: '操作',
      valueType: 'option',
      width: 200,
      fixed: 'right',
      render: (_, record) => (
        <Space size={0} wrap>
          <Button type="link" size="small" onClick={() => navigate(`/case-library/${record.id}`)}>
            查看
          </Button>
          <Button type="link" size="small" onClick={() => {
            setScoreCaseId(record.id);
            setScoreValue(3);
            setScoreModalOpen(true);
          }}>
            评分
          </Button>
          <Button type="link" size="small" onClick={() => handleQuote(record, 1)}>
            引用话术
          </Button>
          {record.status === 1 && (
            <Button type="link" size="small" danger onClick={() => handleArchive(record.id)}>
              归档
            </Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div>
      <ProTable<CaseLibraryItem>
        columns={columns}
        actionRef={actionRef}
        request={async (params) => {
          const res = await caseLibraryService.getList({
            page: params.current,
            pageSize: params.pageSize,
            keyword: params.keyword,
            disputeType: params.disputeType,
            difficultyLevel: params.difficultyLevel,
            status: params.status,
          });
          const data = res.data?.data || {};
          return {
            data: data.list || [],
            total: data.total || 0,
            success: true,
          };
        }}
        rowKey="id"
        search={{
          labelWidth: 'auto',
        }}
        toolBarRender={() => [
          <Button key="search" type="primary" icon={<SearchOutlined />} onClick={() => setSearchModalOpen(true)}>
            相似案例检索
          </Button>,
          <Button key="create" type="primary" icon={<PlusOutlined />} onClick={() => navigate('/case-library/create')}>
            新增案例
          </Button>,
          <Button key="vectorize" icon={<ThunderboltOutlined />} onClick={async () => {
            try {
              await caseLibraryService.vectorizeAll();
              message.success('已开始批量向量化处理');
            } catch { message.error('操作失败'); }
          }}>
            批量向量化
          </Button>,
        ]}
        pagination={{ defaultPageSize: 10 }}
        scroll={{ x: 1300 }}
      />

      <Modal
        title="相似案例检索"
        open={searchModalOpen}
        onCancel={() => setSearchModalOpen(false)}
        width={800}
        footer={null}
      >
        <Space direction="vertical" style={{ width: '100%' }} size="middle">
          <Space.Compact style={{ width: '100%' }}>
            <TextArea
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="输入纠纷描述，检索相似案例..."
              rows={3}
              style={{ flex: 1 }}
            />
            <Button type="primary" icon={<SearchOutlined />} loading={searchLoading} onClick={handleSearch}>
              检索
            </Button>
          </Space.Compact>
          {searchResults.length > 0 ? (
            searchResults.map((item, idx) => (
              <Card key={item.caseId} size="small" title={
                <Space>
                  <Tag color="blue">#{idx + 1}</Tag>
                  <span>{item.title}</span>
                  <Tag>相似度: {(item.score * 100).toFixed(1)}%</Tag>
                </Space>
              } extra={
                <Space>
                  <Button size="small" onClick={() => handleQuote(item, 1)}>引用话术</Button>
                  <Button size="small" onClick={() => handleQuote(item, 2)}>引用策略</Button>
                  <Button size="small" type="primary" onClick={() => handleQuote(item, 3)}>全文引用</Button>
                </Space>
              }>
                <Descriptions column={2} size="small">
                  <Descriptions.Item label="纠纷类型">{item.disputeType}</Descriptions.Item>
                  <Descriptions.Item label="难度">{difficultyMap[item.difficultyLevel]?.text}</Descriptions.Item>
                  {item.mediationTactics && (
                    <Descriptions.Item label="调解话术" span={2}>
                      <div style={{ maxHeight: 80, overflow: 'auto' }}>{item.mediationTactics}</div>
                    </Descriptions.Item>
                  )}
                  {item.keyPoints && (
                    <Descriptions.Item label="调解要点" span={2}>
                      <div style={{ maxHeight: 80, overflow: 'auto' }}>{item.keyPoints}</div>
                    </Descriptions.Item>
                  )}
                </Descriptions>
              </Card>
            ))
          ) : (
            searchQuery && !searchLoading && <Empty description="暂无搜索结果" />
          )}
        </Space>
      </Modal>

      <Modal
        title="案例评分"
        open={scoreModalOpen}
        onCancel={() => setScoreModalOpen(false)}
        onOk={handleScore}
        okText="提交评分"
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          <div>请对该推荐案例的有用性进行评分：</div>
          <Rate value={scoreValue} onChange={setScoreValue} />
          <div style={{ color: '#999' }}>{scoreValue} 分</div>
        </Space>
      </Modal>
    </div>
  );
};

export default CaseLibraryList;
