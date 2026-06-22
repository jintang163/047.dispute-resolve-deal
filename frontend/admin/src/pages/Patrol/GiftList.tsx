import React, { useRef, useState, useEffect } from 'react';
import { Button, Tag, Space, Switch, App, Modal, Form, Input, InputNumber, Image, Upload, Card, Row, Col, Statistic } from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  ReloadOutlined,
  UploadOutlined,
  FireOutlined,
  StarOutlined,
  ShoppingOutlined,
} from '@ant-design/icons';
import {
  ProTable,
  ProFormSelect,
  ProFormText,
  ModalForm,
  ProFormSwitch,
  ProFormTextArea,
  ProFormDigit,
  ProFormRadio,
} from '@ant-design/pro-components';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { patrolService, Gift, GiftCategory, CreateGiftParams } from '../../services/patrol';
import dayjs from 'dayjs';

const statusTextMap: Record<string, string> = {
  active: '上架',
  inactive: '下架',
};

const GiftList: React.FC = () => {
  const { message, modal } = App.useApp();
  const actionRef = useRef<ActionType>();
  const [modalOpen, setModalOpen] = useState(false);
  const [editMode, setEditMode] = useState<'create' | 'edit'>('create');
  const [currentRecord, setCurrentRecord] = useState<Gift | null>(null);
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [categoryOptions, setCategoryOptions] = useState<{ id: string; name: string }[]>([]);
  const [stats, setStats] = useState({
    totalGifts: 0,
    totalStock: 0,
    totalSold: 0,
    totalPoints: 0,
  });

  useEffect(() => {
    loadCategoryOptions();
  }, []);

  const loadCategoryOptions = async () => {
    try {
      const res = await patrolService.getGiftCategoryTree();
      const data: any = (res as any)?.data ?? res;
      if (Array.isArray(data)) {
        const flatten = (categories: GiftCategory[]): GiftCategory[] => {
          let result: GiftCategory[] = [];
          for (const cat of categories) {
            if (cat.status === 'active') {
              result.push(cat);
            }
            if (cat.children) {
              result = result.concat(flatten(cat.children));
            }
          }
          return result;
        };
        setCategoryOptions(flatten(data).map((c) => ({ id: c.id, name: c.name })));
      }
    } catch (error) {
      console.error('加载分类选项失败:', error);
    }
  };

  const handleModalOpen = (mode: 'create' | 'edit', record?: Gift) => {
    setEditMode(mode);
    setCurrentRecord(record || null);
    if (mode === 'edit' && record) {
      form.setFieldsValue({
        name: record.name,
        categoryId: record.categoryId,
        points: record.points,
        originalPrice: record.originalPrice,
        stock: record.stock,
        description: record.description,
        images: record.images,
        status: record.status === 'active',
        sort: record.sort,
        isHot: record.isHot,
        isNew: record.isNew,
      });
    } else {
      form.resetFields();
      form.setFieldsValue({
        status: true,
        sort: 0,
        isHot: false,
        isNew: false,
      });
    }
    setModalOpen(true);
  };

  const handleModalFinish = async (values: CreateGiftParams) => {
    try {
      setLoading(true);
      const params = {
        ...values,
        status: values.status ? 'active' : 'inactive',
      };
      if (editMode === 'create') {
        await patrolService.createGift(params);
        message.success('礼品创建成功');
      } else if (editMode === 'edit' && currentRecord) {
        await patrolService.updateGift(currentRecord.id, params);
        message.success('礼品更新成功');
      }
      actionRef.current?.reload();
      setModalOpen(false);
      return true;
    } catch (error: any) {
      message.error(error.message || '操作失败');
      return false;
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = (record: Gift) => {
    modal.confirm({
      title: '确认删除该礼品?',
      icon: <DeleteOutlined />,
      content: `礼品: ${record.name} (${record.giftNo})`,
      okText: '确认删除',
      cancelText: '取消',
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          await patrolService.deleteGift(record.id);
          message.success('删除成功');
          actionRef.current?.reload();
        } catch (error: any) {
          message.error(error.message || '删除失败');
          return Promise.reject();
        }
      },
    });
  };

  const handleToggleStatus = async (record: Gift, checked: boolean) => {
    try {
      await patrolService.updateGift(record.id, { status: checked ? 'active' : 'inactive' });
      message.success(`礼品已${checked ? '上架' : '下架'}`);
      actionRef.current?.reload();
    } catch (error: any) {
      message.error(error.message || '操作失败');
      actionRef.current?.reload();
    }
  };

  const columns: ProColumns<Gift>[] = [
    {
      title: '礼品图片',
      dataIndex: 'images',
      width: 80,
      render: (_, record) => {
        const img = record.images && record.images.length > 0 ? record.images[0] : undefined;
        return img ? (
          <Image
            width={50}
            height={50}
            src={img}
            style={{ objectFit: 'cover', borderRadius: 4 }}
          />
        ) : (
          <div
            style={{
              width: 50,
              height: 50,
              background: '#f5f5f5',
              borderRadius: 4,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: '#999',
            }}
          >
            <ShoppingOutlined />
          </div>
        );
      },
    },
    {
      title: '礼品信息',
      dataIndex: 'name',
      width: 220,
      ellipsis: true,
      render: (_, record) => (
        <Space direction="vertical" size={0}>
          <Space>
            <span style={{ fontWeight: 500 }}>{record.name}</span>
            {record.isHot && <Tag color="red" icon={<FireOutlined />}>热销</Tag>}
            {record.isNew && <Tag color="green" icon={<StarOutlined />}>新品</Tag>}
          </Space>
          <span style={{ fontSize: 12, color: '#999' }}>{record.giftNo}</span>
        </Space>
      ),
    },
    {
      title: '分类',
      dataIndex: 'categoryName',
      width: 120,
      render: (_, record) => (
        <Tag color="blue">{record.categoryName || '-'}</Tag>
      ),
    },
    {
      title: '积分',
      dataIndex: 'points',
      width: 100,
      render: (_, record) => (
        <span style={{ color: '#faad14', fontWeight: 500, fontSize: 16 }}>
          {record.points}
        </span>
      ),
    },
    {
      title: '原价',
      dataIndex: 'originalPrice',
      width: 100,
      render: (_, record) => (
        record.originalPrice ? <span style={{ textDecoration: 'line-through', color: '#999' }}>¥{record.originalPrice}</span> : '-'
      ),
    },
    {
      title: '库存',
      dataIndex: 'stock',
      width: 100,
      render: (_, record) => (
        <span style={{ color: record.stock < 10 ? '#ff4d4f' : '#52c41a', fontWeight: 500 }}>
          {record.stock}
        </span>
      ),
    },
    {
      title: '已售',
      dataIndex: 'soldCount',
      width: 100,
      render: (_, record) => <span>{record.soldCount || 0}</span>,
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
          checkedChildren="上架"
          unCheckedChildren="下架"
          onChange={(checked) => handleToggleStatus(record, checked)}
        />
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'createTime',
      width: 160,
      sorter: true,
      render: (_, record) =>
        record.createTime ? dayjs(record.createTime).format('YYYY-MM-DD HH:mm:ss') : '-',
    },
    {
      title: '操作',
      key: 'option',
      valueType: 'option',
      width: 160,
      fixed: 'right',
      render: (_, record) => [
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

  const normFile = (e: any) => {
    if (Array.isArray(e)) {
      return e;
    }
    return e?.fileList;
  };

  return (
    <>
      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={12} md={6}>
          <Card>
            <Statistic
              title="礼品总数"
              value={stats.totalGifts}
              prefix={<ShoppingOutlined style={{ color: '#1677ff' }} />}
              valueStyle={{ color: '#1677ff' }}
            />
          </Card>
        </Col>
        <Col xs={12} md={6}>
          <Card>
            <Statistic
              title="总库存"
              value={stats.totalStock}
              prefix={<Tag color="green">件</Tag>}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={12} md={6}>
          <Card>
            <Statistic
              title="总销量"
              value={stats.totalSold}
              prefix={<Tag color="orange">件</Tag>}
              valueStyle={{ color: '#fa8c16' }}
            />
          </Card>
        </Col>
        <Col xs={12} md={6}>
          <Card>
            <Statistic
              title="总积分价值"
              value={stats.totalPoints}
              prefix={<Tag color="gold">P</Tag>}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
      </Row>

      <ProTable<Gift>
        columns={columns}
        actionRef={actionRef}
        cardBordered
        rowKey="id"
        search={{
          labelWidth: 'auto',
          defaultCollapsed: false,
        }}
        dateFormatter="string"
        headerTitle="礼品列表"
        toolBarRender={() => [
          <Button
            key="reload"
            icon={<ReloadOutlined />}
            onClick={() => actionRef.current?.reload()}
          >
            刷新
          </Button>,
          <Button
            key="create"
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => handleModalOpen('create')}
          >
            新增礼品
          </Button>,
        ]}
        request={async (params, sort, filter) => {
          try {
            const res = await patrolService.getGiftList({
              pageNum: params.current,
              pageSize: params.pageSize,
              keyword: params.keyword as string,
              categoryId: (params.categoryId as string) || undefined,
              status: (params.status as string) || undefined,
            });
            const data: any = (res as any)?.data ?? res;
            if (data.stats) {
              setStats(data.stats);
            }
            return {
              data: data.list || [],
              success: true,
              total: data.total || 0,
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
          persistenceKey: 'gift-list-columns',
          persistenceType: 'localStorage',
        }}
        pagination={{
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total) => `共 ${total} 条记录`,
        }}
        scroll={{ x: 1600 }}
      />

      <ModalForm
        title={editMode === 'create' ? '新增礼品' : '编辑礼品'}
        open={modalOpen}
        onOpenChange={setModalOpen}
        form={form}
        layout="vertical"
        autoFocusFirstInput
        modalProps={{
          destroyOnClose: true,
          maskClosable: false,
          width: 600,
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
          label="礼品名称"
          placeholder="请输入礼品名称"
          rules={[
            { required: true, message: '请输入礼品名称' },
            { max: 100, message: '礼品名称长度不能超过100个字符' },
          ]}
        />
        <ProFormSelect
          name="categoryId"
          label="礼品分类"
          placeholder="请选择礼品分类"
          rules={[{ required: true, message: '请选择礼品分类' }]}
          options={categoryOptions.map((c) => ({ label: c.name, value: c.id }))}
          fieldProps={{
            showSearch: true,
            optionFilterProp: 'label',
          }}
        />
        <ProFormDigit
          name="points"
          label="所需积分"
          placeholder="请输入所需积分"
          min={1}
          max={100000}
          rules={[{ required: true, message: '请输入所需积分' }]}
          fieldProps={{
            precision: 0,
            addonBefore: 'P',
          }}
        />
        <ProFormDigit
          name="originalPrice"
          label="原价"
          placeholder="请输入原价"
          min={0}
          max={100000}
          fieldProps={{
            precision: 2,
            addonBefore: '¥',
          }}
        />
        <ProFormDigit
          name="stock"
          label="库存数量"
          placeholder="请输入库存数量"
          min={0}
          max={100000}
          rules={[{ required: true, message: '请输入库存数量' }]}
          fieldProps={{
            precision: 0,
          }}
        />
        <ProFormDigit
          name="sort"
          label="排序"
          placeholder="请输入排序号，数字越小越靠前"
          min={0}
          max={999}
          initialValue={0}
        />
        <Row gutter={16}>
          <Col span={12}>
            <ProFormSwitch
              name="isHot"
              label="是否热销"
              checkedChildren="是"
              unCheckedChildren="否"
              initialValue={false}
            />
          </Col>
          <Col span={12}>
            <ProFormSwitch
              name="isNew"
              label="是否新品"
              checkedChildren="是"
              unCheckedChildren="否"
              initialValue={false}
            />
          </Col>
        </Row>
        <ProFormSwitch
          name="status"
          label="状态"
          checkedChildren="上架"
          unCheckedChildren="下架"
          initialValue={true}
        />
        <ProFormTextArea
          name="description"
          label="礼品描述"
          placeholder="请输入礼品描述"
          rows={3}
        />
        <Form.Item
          name="images"
          label="礼品图片"
          valuePropName="fileList"
          getValueFromEvent={normFile}
          extra="支持上传多张图片，建议尺寸 400x400px"
        >
          <Upload
            listType="picture-card"
            multiple
            maxCount={5}
            beforeUpload={() => false}
          >
            <div>
              <UploadOutlined />
              <div style={{ marginTop: 8 }}>上传</div>
            </div>
          </Upload>
        </Form.Item>
      </ModalForm>
    </>
  );
};

export default GiftList;
