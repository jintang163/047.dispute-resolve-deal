import React, { useState, useRef } from 'react';
import { Button, Tag, Space, App, Modal, Switch, Form } from 'antd';
import {
  PlusOutlined,
  EyeOutlined,
  EditOutlined,
  DeleteOutlined,
  SettingOutlined,
} from '@ant-design/icons';
import {
  ProTable,
  ProFormText,
  ProFormSelect,
  ProFormDigit,
  ModalForm,
  ProForm,
} from '@ant-design/pro-components';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import {
  judicialService,
  CourtConfig,
  CourtConfigParams,
} from '../../services/judicial';
import dayjs from 'dayjs';

const { confirm } = Modal;

const courtLevelOptions = [
  { label: '基层人民法院', value: 1 },
  { label: '中级人民法院', value: 2 },
  { label: '高级人民法院', value: 3 },
  { label: '最高人民法院', value: 4 },
];

const CourtConfigPage: React.FC = () => {
  const { message } = App.useApp();
  const actionRef = useRef<ActionType>();
  const [modalVisible, setModalVisible] = useState(false);
  const [modalType, setModalType] = useState<'create' | 'edit' | 'view'>('create');
  const [currentRecord, setCurrentRecord] = useState<CourtConfig | null>(null);
  const [form] = Form.useForm();

  const handleEdit = (record: CourtConfig) => {
    setCurrentRecord(record);
    setModalType('edit');
    form.setFieldsValue({
      ...record,
    });
    setModalVisible(true);
  };

  const handleView = (record: CourtConfig) => {
    setCurrentRecord(record);
    setModalType('view');
    form.setFieldsValue({
      ...record,
    });
    setModalVisible(true);
  };

  const handleDelete = (record: CourtConfig) => {
    confirm({
      title: '确认删除',
      icon: <DeleteOutlined />,
      content: `确定要删除法院配置"${record.courtName}"吗？此操作不可恢复。`,
      okText: '确认删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: async () => {
        try {
          await judicialService.deleteCourtConfig(record.id);
          message.success('删除成功');
          actionRef.current?.reload();
        } catch (error: any) {
          message.error(error.message || '删除失败');
        }
      },
    });
  };

  const handleStatusChange = async (record: CourtConfig, checked: boolean) => {
    try {
      await judicialService.updateCourtConfig(record.id, {
        status: checked ? 1 : 0,
      });
      message.success('状态更新成功');
      actionRef.current?.reload();
    } catch (error: any) {
      message.error(error.message || '更新失败');
    }
  };

  const handleSubmit = async (values: CourtConfigParams) => {
    try {
      if (modalType === 'create') {
        await judicialService.createCourtConfig(values);
        message.success('创建成功');
      } else if (modalType === 'edit' && currentRecord) {
        await judicialService.updateCourtConfig(currentRecord.id, values);
        message.success('更新成功');
      }
      actionRef.current?.reload();
      return true;
    } catch (error: any) {
      message.error(error.message || '操作失败');
      return false;
    }
  };

  const columns: ProColumns<CourtConfig>[] = [
    {
      title: '法院代码',
      dataIndex: 'courtCode',
      width: 140,
      copyable: true,
    },
    {
      title: '法院名称',
      dataIndex: 'courtName',
      width: 200,
      ellipsis: true,
    },
    {
      title: '法院级别',
      dataIndex: 'courtLevel',
      width: 140,
      valueEnum: {
        1: '基层人民法院',
        2: '中级人民法院',
        3: '高级人民法院',
        4: '最高人民法院',
      },
      render: (_, entity) => {
        const levelMap: Record<number, string> = {
          1: '基层人民法院',
          2: '中级人民法院',
          3: '高级人民法院',
          4: '最高人民法院',
        };
        return levelMap[entity.courtLevel || 1] || '-';
      },
    },
    {
      title: '管辖区域',
      dataIndex: 'jurisdictionArea',
      width: 150,
      ellipsis: true,
    },
    {
      title: '联系方式',
      dataIndex: 'phone',
      width: 140,
    },
    {
      title: '排序',
      dataIndex: 'sortOrder',
      width: 80,
      sorter: true,
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (_, entity) => (
        <Switch
          checked={entity.status === 1}
          onChange={(checked) => handleStatusChange(entity, checked)}
          checkedChildren="启用"
          unCheckedChildren="禁用"
        />
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'createTime',
      width: 180,
      valueType: 'dateTime',
      sorter: true,
      render: (_, entity) => entity.createTime ? dayjs(entity.createTime).format('YYYY-MM-DD HH:mm:ss') : '-',
    },
    {
      title: '操作',
      key: 'option',
      valueType: 'option',
      width: 180,
      fixed: 'right',
      render: (_, record) => [
        <Button
          type="link"
          key="view"
          icon={<EyeOutlined />}
          onClick={() => handleView(record)}
        >
          查看
        </Button>,
        <Button
          type="link"
          key="edit"
          icon={<EditOutlined />}
          onClick={() => handleEdit(record)}
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

  return (
    <div style={{ padding: 24 }}>
      <ProTable<CourtConfig>
        headerTitle="法院配置管理"
        actionRef={actionRef}
        rowKey="id"
        search={{
          labelWidth: 120,
        }}
        toolBarRender={() => [
          <Button
            key="create"
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => {
              setModalType('create');
              setCurrentRecord(null);
              form.resetFields();
              setModalVisible(true);
            }}
          >
            新增法院
          </Button>,
        ]}
        request={async (params) => {
          const res = await judicialService.getCourtConfigList({
            page: params.current,
            pageSize: params.pageSize,
            keyword: params.keyword,
          });
          return {
            data: res.list,
            success: true,
            total: res.total,
          };
        }}
        columns={columns}
        pagination={{
          showSizeChanger: true,
          showQuickJumper: true,
        }}
      />

      <ModalForm<CourtConfigParams>
        title={
          modalType === 'create'
            ? '新增法院配置'
            : modalType === 'edit'
            ? '编辑法院配置'
            : '法院配置详情'
        }
        open={modalVisible}
        onOpenChange={setModalVisible}
        form={form}
        width={700}
        layout="vertical"
        modalProps={{
          destroyOnClose: true,
        }}
        readonly={modalType === 'view'}
        onFinish={async (values) => {
          if (modalType === 'view') return true;
          return handleSubmit(values);
        }}
      >
        <ProForm.Group title="基本信息">
          <ProFormText
            name="courtCode"
            label="法院代码"
            rules={[{ required: true, message: '请输入法院代码' }]}
            placeholder="请输入法院代码"
            disabled={modalType === 'edit'}
          />
          <ProFormText
            name="courtName"
            label="法院名称"
            rules={[{ required: true, message: '请输入法院名称' }]}
            placeholder="请输入法院名称"
          />
          <ProFormSelect
            name="courtLevel"
            label="法院级别"
            rules={[{ required: true, message: '请选择法院级别' }]}
            options={courtLevelOptions}
            placeholder="请选择法院级别"
          />
          <ProFormText
            name="jurisdictionArea"
            label="管辖区域"
            placeholder="请输入管辖区域"
          />
        </ProForm.Group>

        <ProForm.Group title="联系信息">
          <ProFormText
            name="address"
            label="地址"
            placeholder="请输入地址"
          />
          <ProFormText
            name="contact"
            label="联系人"
            placeholder="请输入联系人"
          />
          <ProFormText
            name="phone"
            label="联系电话"
            placeholder="请输入联系电话"
          />
        </ProForm.Group>

        <ProForm.Group title="API配置">
          <ProFormText
            name="apiEndpoint"
            label="API地址"
            placeholder="请输入API地址"
          />
          <ProFormText
            name="apiAppId"
            label="应用ID"
            placeholder="请输入应用ID"
          />
          {modalType !== 'view' && (
            <ProFormText.Password
              name="apiSecret"
              label="应用密钥"
              placeholder="请输入应用密钥"
            />
          )}
          <ProFormText
            name="apiPublicKey"
            label="公钥"
            fieldProps={{
              autoSize: { minRows: 2, maxRows: 4 },
            }}
            placeholder="请输入公钥"
          />
        </ProForm.Group>

        <ProForm.Group title="签章配置">
          <ProFormText
            name="sealCertNo"
            label="印章证书编号"
            placeholder="请输入印章证书编号"
          />
          <ProFormText
            name="sealImageUrl"
            label="印章图片地址"
            placeholder="请输入印章图片地址"
          />
        </ProForm.Group>

        <ProForm.Group title="其他配置">
          <ProFormDigit
            name="sortOrder"
            label="排序"
            min={0}
            placeholder="请输入排序值，数字越小越靠前"
          />
          <ProFormSelect
            name="status"
            label="状态"
            rules={[{ required: true, message: '请选择状态' }]}
            options={[
              { label: '启用', value: 1 },
              { label: '禁用', value: 0 },
            ]}
            placeholder="请选择状态"
          />
        </ProForm.Group>
      </ModalForm>
    </div>
  );
};

export default CourtConfigPage;
