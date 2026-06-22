import React, { useRef, useState, useEffect } from 'react';
import { Button, Space, App, Modal, Form, Input, Switch, Tree } from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  ReloadOutlined,
  FolderOutlined,
  FolderOpenOutlined,
} from '@ant-design/icons';
import {
  ModalForm,
  ProFormText,
  ProFormSwitch,
  ProFormDigit,
  ProFormTextArea,
  ProFormSelect,
} from '@ant-design/pro-components';
import type { DataNode } from 'antd/es/tree';
import { patrolService, GiftCategory } from '../../services/patrol';
import dayjs from 'dayjs';

const GiftCategoryPage: React.FC = () => {
  const { message, modal } = App.useApp();
  const [categoryTree, setCategoryTree] = useState<GiftCategory[]>([]);
  const [loading, setLoading] = useState(false);
  const [treeLoading, setTreeLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editMode, setEditMode] = useState<'create' | 'edit'>('create');
  const [currentRecord, setCurrentRecord] = useState<GiftCategory | null>(null);
  const [parentId, setParentId] = useState<string | null>(null);
  const [form] = Form.useForm();

  useEffect(() => {
    loadCategoryTree();
  }, []);

  const loadCategoryTree = async () => {
    setTreeLoading(true);
    try {
      const res = await patrolService.getGiftCategoryTree();
      const data: any = (res as any)?.data ?? res;
      if (Array.isArray(data)) {
        setCategoryTree(data);
      }
    } catch (error: any) {
      message.error(error.message || '加载分类树失败');
    } finally {
      setTreeLoading(false);
    }
  };

  const buildTreeData = (categories: GiftCategory[]): DataNode[] => {
    return categories.map((cat) => ({
      key: cat.id,
      title: (
        <Space>
          <span>{cat.name}</span>
          {cat.status !== 'active' && (
            <span style={{ color: '#999', fontSize: 12 }}>(已禁用)</span>
          )}
        </Space>
      ),
      isLeaf: !cat.children || cat.children.length === 0,
      children: cat.children ? buildTreeData(cat.children) : undefined,
    }));
  };

  const findCategoryById = (
    categories: GiftCategory[],
    id: string,
  ): GiftCategory | null => {
    for (const cat of categories) {
      if (cat.id === id) return cat;
      if (cat.children) {
        const found = findCategoryById(cat.children, id);
        if (found) return found;
      }
    }
    return null;
  };

  const flattenCategories = (categories: GiftCategory[]): GiftCategory[] => {
    let result: GiftCategory[] = [];
    for (const cat of categories) {
      result.push(cat);
      if (cat.children) {
        result = result.concat(flattenCategories(cat.children));
      }
    }
    return result;
  };

  const handleSelect = (selectedKeys: React.Key[]) => {
    if (selectedKeys.length > 0) {
      const id = selectedKeys[0] as string;
      const category = findCategoryById(categoryTree, id);
      if (category) {
        setCurrentRecord(category);
      }
    }
  };

  const handleModalOpen = (
    mode: 'create' | 'edit',
    record?: GiftCategory,
    parent?: string | null,
  ) => {
    setEditMode(mode);
    setCurrentRecord(record || null);
    setParentId(parent || null);
    if (mode === 'edit' && record) {
      form.setFieldsValue({
        name: record.name,
        code: record.code,
        parentId: record.parentId,
        sort: record.sort,
        status: record.status === 'active',
        description: record.description,
      });
    } else {
      form.resetFields();
      form.setFieldsValue({
        parentId: parent || undefined,
        status: true,
        sort: 0,
      });
    }
    setModalOpen(true);
  };

  const handleModalFinish = async (values: any) => {
    try {
      setLoading(true);
      const params = {
        ...values,
        status: values.status ? 'active' : 'inactive',
        parentId: parentId || values.parentId || null,
      };
      if (editMode === 'create') {
        await patrolService.createGiftCategory(params);
        message.success('分类创建成功');
      } else if (editMode === 'edit' && currentRecord) {
        await patrolService.updateGiftCategory(currentRecord.id, params);
        message.success('分类更新成功');
      }
      loadCategoryTree();
      setModalOpen(false);
      return true;
    } catch (error: any) {
      message.error(error.message || '操作失败');
      return false;
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = (record: GiftCategory) => {
    modal.confirm({
      title: '确认删除该分类?',
      icon: <DeleteOutlined />,
      content: `分类: ${record.name} (${record.code})`,
      okText: '确认删除',
      cancelText: '取消',
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          await patrolService.deleteGiftCategory(record.id);
          message.success('删除成功');
          loadCategoryTree();
          setCurrentRecord(null);
        } catch (error: any) {
          message.error(error.message || '删除失败');
          return Promise.reject();
        }
      },
    });
  };

  const handleToggleStatus = async (record: GiftCategory, checked: boolean) => {
    try {
      await patrolService.updateGiftCategory(record.id, {
        status: checked ? 'active' : 'inactive',
      });
      message.success(`分类已${checked ? '启用' : '禁用'}`);
      loadCategoryTree();
    } catch (error: any) {
      message.error(error.message || '操作失败');
      loadCategoryTree();
    }
  };

  const treeData = buildTreeData(categoryTree);
  const flatCategories = flattenCategories(categoryTree);

  return (
    <div style={{ display: 'flex', gap: 16, height: '100%' }}>
      <div
        style={{
          width: 320,
          background: '#fff',
          border: '1px solid #f0f0f0',
          borderRadius: 8,
          padding: 16,
        }}
      >
        <div
          style={{
            marginBottom: 12,
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
          }}
        >
          <span style={{ fontWeight: 500 }}>礼品分类</span>
          <Space>
            <Button
              size="small"
              icon={<ReloadOutlined />}
              onClick={loadCategoryTree}
            />
            <Button
              size="small"
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => handleModalOpen('create', null, null)}
            >
              新增
            </Button>
          </Space>
        </div>
        <Tree
          showLine
          showIcon
          loading={treeLoading}
          blockNode
          onSelect={handleSelect}
          selectedKeys={currentRecord ? [currentRecord.id] : []}
          treeData={treeData}
          icon={({ expanded }) =>
            expanded ? <FolderOpenOutlined /> : <FolderOutlined />
          }
          defaultExpandAll
        />
      </div>

      <div style={{ flex: 1 }}>
        <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <span style={{ fontWeight: 500, fontSize: 16 }}>
            {currentRecord ? `分类详情: ${currentRecord.name}` : '请选择分类查看详情'}
          </span>
          {currentRecord && (
            <Space>
              <Button
                icon={<PlusOutlined />}
                onClick={() => handleModalOpen('create', null, currentRecord.id)}
              >
                添加子分类
              </Button>
              <Button
                type="primary"
                icon={<EditOutlined />}
                onClick={() => handleModalOpen('edit', currentRecord)}
              >
                编辑
              </Button>
              <Button
                danger
                icon={<DeleteOutlined />}
                onClick={() => handleDelete(currentRecord)}
              >
                删除
              </Button>
            </Space>
          )}
        </div>

        {currentRecord ? (
          <div style={{ padding: 24, background: '#fff', border: '1px solid #f0f0f0', borderRadius: 8 }}>
            <div style={{ marginBottom: 16 }}>
              <div style={{ fontSize: 12, color: '#999', marginBottom: 4 }}>分类名称</div>
              <div style={{ fontSize: 16, fontWeight: 500 }}>{currentRecord.name}</div>
            </div>
            <div style={{ marginBottom: 16 }}>
              <div style={{ fontSize: 12, color: '#999', marginBottom: 4 }}>分类编码</div>
              <div>{currentRecord.code}</div>
            </div>
            <div style={{ marginBottom: 16 }}>
              <div style={{ fontSize: 12, color: '#999', marginBottom: 4 }}>状态</div>
              <div>
                <Switch
                  checked={currentRecord.status === 'active'}
                  checkedChildren="启用"
                  unCheckedChildren="禁用"
                  onChange={(checked) => handleToggleStatus(currentRecord, checked)}
                />
              </div>
            </div>
            <div style={{ marginBottom: 16 }}>
              <div style={{ fontSize: 12, color: '#999', marginBottom: 4 }}>排序</div>
              <div>{currentRecord.sort || 0}</div>
            </div>
            <div style={{ marginBottom: 16 }}>
              <div style={{ fontSize: 12, color: '#999', marginBottom: 4 }}>描述</div>
              <div>{currentRecord.description || '-'}</div>
            </div>
            <div style={{ marginBottom: 16 }}>
              <div style={{ fontSize: 12, color: '#999', marginBottom: 4 }}>创建时间</div>
              <div>{currentRecord.createTime ? dayjs(currentRecord.createTime).format('YYYY-MM-DD HH:mm:ss') : '-'}</div>
            </div>
          </div>
        ) : (
          <div style={{ padding: 100, textAlign: 'center', color: '#999', background: '#fff', border: '1px solid #f0f0f0', borderRadius: 8 }}>
            请从左侧选择一个分类
          </div>
        )}
      </div>

      <ModalForm
        title={editMode === 'create' ? '新增礼品分类' : '编辑礼品分类'}
        open={modalOpen}
        onOpenChange={setModalOpen}
        form={form}
        layout="vertical"
        autoFocusFirstInput
        modalProps={{
          destroyOnClose: true,
          maskClosable: false,
          width: 500,
        }}
        onFinish={handleModalFinish}
        submitter={{
          submitButtonProps: {
            loading,
          },
        }}
      >
        <ProFormText
          name="name"
          label="分类名称"
          placeholder="请输入分类名称"
          rules={[
            { required: true, message: '请输入分类名称' },
            { max: 50, message: '分类名称长度不能超过50个字符' },
          ]}
        />
        <ProFormText
          name="code"
          label="分类编码"
          placeholder="请输入分类编码，如：ELECTRONICS"
          rules={[
            { required: true, message: '请输入分类编码' },
            { max: 50, message: '分类编码长度不能超过50个字符' },
          ]}
        />
        {!parentId && (
          <ProFormSelect
            name="parentId"
            label="上级分类"
            placeholder="请选择上级分类（不选则为顶级分类）"
            allowClear
            options={flatCategories
              .filter((c) => !currentRecord || c.id !== currentRecord.id)
              .map((c) => ({ label: c.name, value: c.id }))}
            fieldProps={{
              showSearch: true,
              optionFilterProp: 'label',
            }}
          />
        )}
        <ProFormDigit
          name="sort"
          label="排序"
          placeholder="请输入排序号，数字越小越靠前"
          min={0}
          max={999}
          initialValue={0}
        />
        <ProFormSwitch
          name="status"
          label="状态"
          checkedChildren="启用"
          unCheckedChildren="禁用"
          initialValue={true}
        />
        <ProFormTextArea
          name="description"
          label="描述"
          placeholder="请输入分类描述"
          rows={3}
        />
      </ModalForm>
    </div>
  );
};

export default GiftCategoryPage;
