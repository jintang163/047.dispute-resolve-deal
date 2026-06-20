import React, { useState, useRef, useEffect } from 'react';
import { Button, Tag, Space, App, Modal, Card } from 'antd';
import { PlusOutlined, EyeOutlined, EditOutlined, DeleteOutlined, ExclamationCircleOutlined, FireOutlined } from '@ant-design/icons';
import {
  ProTable,
  ProFormSelect,
  ProFormDateRangePicker,
  ProFormText,
} from '@ant-design/pro-components';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { useNavigate } from 'react-router-dom';
import { disputeService, DisputeCase, DisputeTypeNode } from '../../services/dispute';
import dayjs from 'dayjs';

const { confirm } = Modal;

const STATUS_MAP: Record<string, string> = {
  '10': '待分派',
  '20': '调解中',
  '30': '待审批',
  '40': '审批中',
  '50': '已结案',
  '99': '已取消',
};

const STATUS_COLOR: Record<string, string> = {
  '10': 'default',
  '20': 'blue',
  '30': 'orange',
  '40': 'processing',
  '50': 'success',
  '99': 'default',
};

const keywordCategoryColor: Record<string, string> = {
  '纠纷性质': 'red',
  '行为': 'orange',
  '对象': 'blue',
  '程度': 'purple',
};

const pick = <T,>(obj: T, keys: (keyof T | string)[]): any => {
  const o = obj as any;
  for (const k of keys) {
    if (o[k] !== undefined && o[k] !== null && o[k] !== '') return o[k];
  }
  return undefined;
};

const DisputeList: React.FC = () => {
  const navigate = useNavigate();
  const { message } = App.useApp();
  const actionRef = useRef<ActionType>();
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [activeTagKeyword, setActiveTagKeyword] = useState<string>('');
  const [hotKeywords, setHotKeywords] = useState<Array<{ keyword: string; category?: string; frequency?: number }>>([]);
  const [typesFlat, setTypesFlat] = useState<Map<number, string>>(new Map());

  useEffect(() => {
    Promise.all([
      disputeService.getHotKeywords({ days: 30, limit: 15 }),
      disputeService.getTypes(),
    ]).then(([hotRes, typesRes]) => {
      const hotData: any = (hotRes as any)?.data ?? hotRes;
      if (Array.isArray(hotData)) setHotKeywords(hotData);

      const flat = new Map<number, string>();
      const walk = (nodes: DisputeTypeNode[]) => {
        for (const n of nodes) {
          flat.set(n.id, n.typeName);
          if (n.children && n.children.length > 0) walk(n.children);
        }
      };
      const typesData: any = (typesRes as any)?.data ?? typesRes;
      if (Array.isArray(typesData)) walk(typesData);
      setTypesFlat(flat);
    }).catch(() => {});
  }, []);

  const handleTagKeywordClick = (kw: string) => {
    if (activeTagKeyword === kw) {
      setActiveTagKeyword('');
    } else {
      setActiveTagKeyword(kw);
    }
    actionRef.current?.reload();
  };

  const columns: ProColumns<DisputeCase>[] = [
    {
      title: '案件编号',
      dataIndex: 'caseNo',
      width: 180,
      copyable: true,
      fixed: 'left',
      render: (_, row) => pick(row, ['case_no', 'caseNo']),
    },
    {
      title: '案件标题',
      dataIndex: 'title',
      width: 200,
      ellipsis: true,
    },
    {
      title: '纠纷类型',
      dataIndex: 'type',
      width: 140,
      render: (_, row) => {
        const typeId = pick(row, ['type_id', 'typeId', 'type']);
        let name = pick(row, ['type_name', 'typeName']);
        if (!name && typesFlat.size > 0 && typeof typeId === 'number') {
          name = typesFlat.get(typeId);
        }
        if (!name) name = typeof typeId === 'string' ? typeId : '-';
        return <Tag color="blue">{name}</Tag>;
      },
    },
    {
      title: '关键词标签',
      dataIndex: 'keywords',
      width: 240,
      search: false,
      render: (_, row) => {
        const kws = row.keywords;
        if (!kws || !Array.isArray(kws) || kws.length === 0) {
          return <span style={{ color: '#ccc' }}>-</span>;
        }
        return (
          <Space size={[2, 4]} wrap>
            {kws.slice(0, 4).map((kw: string, idx: number) => (
              <Tag
                key={kw}
                color={['red', 'volcano', 'orange', 'gold', 'lime', 'green', 'cyan', 'blue'][idx % 8]}
                style={{
                  fontSize: 11,
                  padding: '0 4px',
                  margin: 0,
                  cursor: 'pointer',
                }}
                onClick={() => handleTagKeywordClick(kw)}
              >
                {kw}
              </Tag>
            ))}
            {kws.length > 4 && <Tag style={{ fontSize: 11 }}>+{kws.length - 4}</Tag>}
          </Space>
        );
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (_, row) => {
        const raw = pick(row, ['status']);
        const s = typeof raw === 'number' ? String(raw) : String(raw || '');
        const name = pick(row, ['status_name', 'statusName']) || STATUS_MAP[s] || s;
        return <Tag color={STATUS_COLOR[s] || 'default'}>{name}</Tag>;
      },
    },
    {
      title: '报案人',
      dataIndex: 'partyA',
      width: 120,
      ellipsis: true,
      render: (_, row) => pick(row, ['reporter_name', 'reporterName', 'partyA']) || '-',
    },
    {
      title: '对方',
      dataIndex: 'partyB',
      width: 120,
      ellipsis: true,
      render: (_, row) => pick(row, ['respondent_name', 'respondentName', 'partyB']) || '-',
    },
    {
      title: '发生地点',
      dataIndex: 'address',
      width: 180,
      ellipsis: true,
      render: (_, row) => pick(row, ['occur_address', 'occurAddress', 'address']) || '-',
    },
    {
      title: '所属组织',
      dataIndex: 'orgName',
      width: 160,
      ellipsis: true,
      render: (_, row) => pick(row, ['org_name', 'orgName']) || '-',
    },
    {
      title: '调解员',
      dataIndex: 'mediatorName',
      width: 120,
      ellipsis: true,
      render: (_, row) => pick(row, ['mediator_name', 'mediatorName']) || '-',
    },
    {
      title: '创建时间',
      dataIndex: 'createTime',
      width: 180,
      sorter: true,
      render: (_, row) => {
        const v = pick(row, ['created_at', 'createdAt', 'create_time', 'createTime']);
        return v ? dayjs(v).format('YYYY-MM-DD HH:mm:ss') : '-';
      },
    },
    {
      title: '操作',
      key: 'option',
      valueType: 'option',
      width: 180,
      fixed: 'right',
      render: (_, record) => {
        const id = pick(record, ['id']);
        return [
          <Button type="link" key="view" icon={<EyeOutlined />} onClick={() => navigate(`/dispute/${id}`)}>
            查看
          </Button>,
          <Button
            type="link"
            key="edit"
            icon={<EditOutlined />}
            onClick={() => message.info('编辑功能开发中')}
          >
            编辑
          </Button>,
          <Button
            type="link"
            key="delete"
            danger
            icon={<DeleteOutlined />}
            onClick={() => {
              confirm({
                title: '确认删除该案件?',
                icon: <ExclamationCircleOutlined />,
                content: '删除后将无法恢复，请谨慎操作。',
                okText: '确认删除',
                cancelText: '取消',
                okButtonProps: { danger: true },
                onOk: async () => {
                  try {
                    await disputeService.delete(id);
                    message.success('删除成功');
                    actionRef.current?.reload();
                  } catch (error: any) {
                    message.error(error.message || '删除失败');
                  }
                },
              });
            }}
          >
            删除
          </Button>,
        ];
      },
    },
  ];

  return (
    <>
      {hotKeywords.length > 0 && (
        <Card size="small" style={{ marginBottom: 12 }} bodyStyle={{ padding: '8px 16px' }}>
          <Space size={4} wrap>
            <FireOutlined style={{ color: '#ff4d4f', marginRight: 4 }} />
            <span style={{ color: '#666', marginRight: 8, fontSize: 13 }}>热门标签:</span>
            {hotKeywords.map(item => (
              <Tag
                key={item.keyword}
                color={
                  activeTagKeyword === item.keyword
                    ? '#1890ff'
                    : (item.category ? keywordCategoryColor[item.category] : undefined) || 'default'
                }
                style={{
                  cursor: 'pointer',
                  fontSize: 12,
                  padding: '0 6px',
                  opacity: activeTagKeyword && activeTagKeyword !== item.keyword ? 0.4 : 1,
                }}
                onClick={() => handleTagKeywordClick(item.keyword)}
              >
                {item.keyword}
                {(item.frequency ?? 0) > 1 && (
                  <span style={{ marginLeft: 2, opacity: 0.6 }}>({item.frequency})</span>
                )}
              </Tag>
            ))}
            {activeTagKeyword && (
              <Tag
                color="default"
                style={{ cursor: 'pointer', fontSize: 12 }}
                onClick={() => {
                  setActiveTagKeyword('');
                  actionRef.current?.reload();
                }}
              >
                ✕ 清除筛选
              </Tag>
            )}
          </Space>
        </Card>
      )}

      <ProTable<DisputeCase>
        columns={columns}
        actionRef={actionRef}
        cardBordered
        rowKey="id"
        search={{
          labelWidth: 'auto',
          defaultCollapsed: false,
        }}
        rowSelection={{
          onChange: keys => {
            setSelectedRowKeys(keys);
          },
        }}
        dateFormatter="string"
        headerTitle="纠纷案件列表"
        toolBarRender={() => [
          <Button key="create" type="primary" icon={<PlusOutlined />} onClick={() => navigate('/dispute/create')}>
            新增案件
          </Button>,
        ]}
        request={async (params, sort, filter) => {
          try {
            const startDate = (params as any).createTime?.[0];
            const endDate = (params as any).createTime?.[1];
            const res = await disputeService.getList({
              pageNum: params.current,
              pageSize: params.pageSize,
              keyword: params.keyword as string,
              tagKeyword: activeTagKeyword || undefined,
              type: (params.type as string) || undefined,
              status: (params.status as string) || undefined,
              startDate,
              endDate,
              ...filter,
            });
            const data: any = (res as any)?.data ?? res;
            return {
              data: data.list || [],
              success: true,
              total: data.total || 0,
            };
          } catch (error) {
            return { data: [], success: false, total: 0 };
          }
        }}
        columnsState={{
          persistenceKey: 'dispute-list-columns',
          persistenceType: 'localStorage',
        }}
        pagination={{
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: total => `共 ${total} 条记录`,
        }}
        scroll={{ x: 1800 }}
      />
    </>
  );
};

export default DisputeList;
