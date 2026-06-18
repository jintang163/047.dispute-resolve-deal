import React, { useRef, useState, useEffect } from 'react';
import { Button, Tag, Space, App, Modal, Form, Input, Tree, Switch, Card } from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  ReloadOutlined,
  TeamOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons';
import {
  ProTable,
  ProFormSelect,
  ProFormText,
  ModalForm,
} from '@ant-design/pro-components';
import type { ActionType, ProColumns, DataNode } from '@ant-design/pro-components';
import { orgService, Organization } from '../../services/user';

const typeTextMap: Record<string, string> = {
  center: '综治中心',
  street: '街道办事处',
  committee: '调解委员会',
  other: '其他',
};

const typeColorMap: Record<string, string> = {
  center: 'purple',
  street: 'blue',
  committee: 'green',
  other: 'default',
};

const OrgList: React.FC = () => {
  const { message, modal } = App.useApp();
  const actionRef = useRef<ActionType>();
  const [modalOpen, setModalOpen] = useState(false);
  const [editMode, setEditMode] = useState<'create' | 'edit'>('create');
  const [currentRecord, setCurrentRecord] = useState<Organization | null>(null);
  const [treeData, setTreeData] = useState<any[]>([]);
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [expandedKeys, setExpandedKeys] = useState<React.Key[]>([]);

  useEffect(() => {
    fetchTreeData();
  }, []);

  const fetchTreeData = async () => {
    try {
      const res = await orgService.getTree();
      const data = res.data || res;
      const buildTree = (items: Organization[]): any[] => {
        return items.map((item) => ({
          title: (
            <Space>
              <span>{item.name}</span>
              {item.status === 'disabled' && <Tag color="default" style={{ fontSize: 11 }}>已禁用</Tag>}
            </Space>
          ),
          key: item.id,
          children: item.children ? buildTree(item.children) : undefined,
        }));
      };
      setTreeData(buildTree(data || []));
    } catch (error) {
      console.error('Fetch org tree error:', error);
    }
  };

  const handleModalOpen = (mode: 'create' | 'edit', record?: Organization) => {
    setEditMode(mode);
    setCurrentRecord(record || null);
    if (mode === 'edit' && record) {
      form.setFieldsValue(record);
    } else {
      form.resetFields();
    }
    setModalOpen(true);
  };

  const handleModalFinish = async (values: any) => {
    try {
      setLoading(true);
      if (editMode === 'create') {
        await orgService.create(values);
        message.success('组织创建成功');
      } else if (editMode === 'edit' && currentRecord) {
        await orgService.update(currentRecord.id, values);
        message.success('组织更新成功');
      }
      actionRef.current?.reload();
      fetchTreeData();
      setModalOpen(false);
    } catch (error: any) {
      message.error(error.message || '操作失败');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = (record: Organization) => {
    modal.confirm({
      title: '确认删除该组织?',
      icon: <ExclamationCircleOutlined />,
      content: (
        <div>
          <p>组织名称: <strong>{record.name}</strong></p>
          <p style={{ color: '#faad14' }}>注意: 删除后该组织下的子组织也将被删除，请谨慎操作。</p>
        </div>
      ),
      okText: '确认删除',
      cancelText: '取消',
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          await orgService.delete(record.id);
          message.success('删除成功');
          actionRef.current?.reload();
          fetchTreeData();
        } catch (error: any) {
          message.error(error.message || '删除失败');
          return Promise.reject();
        }
      },
    });
  };

  const handleToggleStatus = async (record: Organization, checked: boolean) => {
    try {
      await orgService.update(record.id, { status: checked ? 'active' : 'disabled' });
      message.success(`组织已${checked ? '启用' : '禁用'}`);
      actionRef.current?.reload();
      fetchTreeData();
    } catch (error: any) {
      message.error(error.message || '操作失败');
      actionRef.current?.reload();
    }
  };

  const columns: ProColumns<Organization>[] = [
    {
      title: '组织名称',
      dataIndex: 'name',
      width: 220,
      fixed: 'left',
      render: (_, record) => (
        <Space>
          <div
            style={{
              width: 32,
              height: 32,
              borderRadius: 8,
              background: '#52c41a15',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: '#52c41a',
              fontSize: 16,
            }}
          >
            <TeamOutlined />
          </div>
          <div>
            <div style={{ fontWeight: 500 }}>{record.name}</div>
            <div style={{ fontSize: 12, color: '#999' }}>编码: {record.code}</div>
          </div>
        </Space>
      ),
    },
    {
      title: '组织类型',
      dataIndex: 'type',
      width: 120,
      valueEnum: typeTextMap,
      render: (_, record) => (
        <Tag color={typeColorMap[record.type || ''] || 'default'}>
          {typeTextMap[record.type || 'other'] || record.type}
        </Tag>
      ),
    },
    {
      title: '负责人',
      dataIndex: 'leader',
      width: 120,
    },
    {
      title: '联系电话',
      dataIndex: 'leaderPhone',
      width: 140,
    },
    {
      title: '地址',
      dataIndex: 'address',
      width: 200,
      ellipsis: true,
    },
    {
      title: '排序',
      dataIndex: 'sort',
      width: 80,
      sorter: true,
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (_, record) => (
        <Switch
          checked={record.status === 'active'}
          checkedChildren="启用"
          unCheckedChildren="禁用"
          onChange={(checked) => handleToggleStatus(record, checked)}
        />
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
          key="add"
          icon={<PlusOutlined />}
          onClick={() => {
            handleModalOpen('create');
            form.setFieldsValue({ parentId: record.id });
          }}
        >
          新增子级
        </Button>,
        <Button
          type="link"
          key="edit"
          icon={<EditOutlined />}
          onClick={() => handleModalOpen('edit', record)}
        >
          编辑
        </Button>,
        <Button
          type="link"
          key="delete"
          danger
          icon={<DeleteOutlined />}
          onClick={() => handleDelete(record)}
        >
          删除
        </Button>,
      ],
    },
  ];

  const buildFlatTree = (items: Organization[], level: number = 0): Organization[] => {
    const result: Organization[] = [];
    items.forEach((item) => {
      result.push({ ...item, name: `${'　'.repeat(level)}${level > 0 ? '└ ' : ''}${item.name}` });
      if (item.children && item.children.length > 0) {
        result.push(...buildFlatTree(item.children, level + 1));
      }
    });
    return result;
  };

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Space size={16} style={{ width: '100%', alignItems: 'stretch' }}>
        <Card
          title="组织架构"
          bordered={false}
          style={{ width: 320, borderRadius: 12, flexShrink: 0 }}
          extra={
            <Button
              size="small"
              icon={<ReloadOutlined />}
              onClick={fetchTreeData}
            />
          }
        >
          <Tree
            showLine
            expandedKeys={expandedKeys}
            onExpand={(keys) => setExpandedKeys(keys)}
            treeData={treeData}
            defaultExpandAll
          />
        </Card>

        <div style={{ flex: 1 }}>
          <ProTable<Organization>
            columns={columns}
            actionRef={actionRef}
            cardBordered
            rowKey="id"
            search={false}
            dateFormatter="string"
            headerTitle="组织管理"
            toolBarRender={() => [
              <Button
                key="reload"
                icon={<ReloadOutlined />}
                onClick={() => {
                  actionRef.current?.reload();
                  fetchTreeData();
                }}
              >
                刷新
              </Button>,
              <Button
                key="create"
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => handleModalOpen('create')}
              >
                新增组织
              </Button>,
            ]}
            request={async (params, sort, filter) => {
              try {
                const res = await orgService.getTree();
                const data = res.data || res;
                const flatList = buildFlatTree(data || []);
                return {
                  data: flatList,
                  success: true,
                  total: flatList.length,
                };
              } catch (error) {
                return {
                  data: [],
                  success: false,
                  total: 0,
                };
              }
            }}
            columnsState={{
              persistenceKey: 'org-list-columns',
              persistenceType: 'localStorage',
            }}
            pagination={false}
            scroll={{ x: 1200 }}
          />
        </div>
      </Space>

      <ModalForm
        title={editMode === 'create' ? '新增组织' : '编辑组织'}
        open={modalOpen}
        onOpenChange={setModalOpen}
        form={form}
        layout="vertical"
        autoFocusFirstInput
        modalProps={{
          destroyOnClose: true,
          maskClosable: false,
        }}
        submitTimeout={2000}
        onFinish={handleModalFinish}
        submitter={{
          submitButtonProps: {
            loading,
          },
          resetButtonProps: {
            onClick: () => {
              form.resetFields();
            },
          },
        }}
      >
        <Form.Item
          label="组织名称"
          name="name"
          rules={[{ required: true, message: '请输入组织名称' }]}
        >
          <Input placeholder="请输入组织名称" />
        </Form.Item>
        <Form.Item
          label="组织编码"
          name="code"
          rules={[{ required: true, message: '请输入组织编码' }]}
        >
          <Input placeholder="请输入组织编码（唯一标识）" disabled={editMode === 'edit'} />
        </Form.Item>
        <Form.Item
          label="上级组织"
          name="parentId"
        >
          <ProFormSelect
            placeholder="请选择上级组织（不选为顶级）"
            allowClear
            debounceTime={300}
            request={async () => {
              try {
                const res = await orgService.getTree();
                const data = res.data || res;
                const options: { label: string; value: string }[] = [];
                const buildOptions = (items: Organization[], level: number = 0) => {
                  items.forEach((item) => {
                    if (editMode === 'edit' && currentRecord && item.id === currentRecord.id) {
                      return;
                    }
                    options.push({
                      label: `${'　'.repeat(level)}${level > 0 ? '└ ' : ''}${item.name}`,
                      value: item.id,
                    });
                    if (item.children && item.children.length > 0) {
                      buildOptions(item.children, level + 1);
                    }
                  });
                };
                buildOptions(data || []);
                return options;
              } catch {
                return [];
              }
            }}
          />
        </Form.Item>
        <Form.Item
          label="组织类型"
          name="type"
          rules={[{ required: true, message: '请选择组织类型' }]}
        >
          <Select
            placeholder="请选择组织类型"
            options={Object.entries(typeTextMap).map(([value, label]) => ({ value, label }))}
          />
        </Form.Item>
        <Form.Item
          label="负责人"
          name="leader"
        >
          <Input placeholder="请输入负责人姓名" />
        </Form.Item>
        <Form.Item
          label="联系电话"
          name="leaderPhone"
          rules={[
            {
              pattern: /^1[3-9]\d{9}$/,
              message: '请输入正确的手机号码',
            },
          ]}
        >
          <Input placeholder="请输入联系电话" />
        </Form.Item>
        <Form.Item
          label="地址"
          name="address"
        >
          <Input placeholder="请输入组织地址" />
        </Form.Item>
        <Form.Item
          label="排序"
          name="sort"
          initialValue={0}
        >
          <Input type="number" placeholder="数字越小越靠前" />
        </Form.Item>
        {editMode === 'edit' && (
          <Form.Item
            label="状态"
            name="status"
            initialValue="active"
          >
            <Select
              placeholder="请选择状态"
              options={[
                { value: 'active', label: '启用' },
                { value: 'disabled', label: '禁用' },
              ]}
            />
          </Form.Item>
        )}
      </ModalForm>
    </Space>
  );
};

export default OrgList;
