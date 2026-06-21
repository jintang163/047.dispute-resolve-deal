import React, { useState, useRef } from 'react';
import { Button, Tag, Space, App, Modal, Descriptions } from 'antd';
import { DownloadOutlined, EyeOutlined, ReloadOutlined, WarningOutlined } from '@ant-design/icons';
import {
  ProTable,
  ProFormSelect,
  ProFormDateRangePicker,
  ProFormText,
} from '@ant-design/pro-components';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import dayjs from 'dayjs';
import { exportService, DataExportLog } from '../../services/export';

const EXPORT_TYPE_MAP: Record<number, string> = {
  1: '纠纷案件',
  2: '绩效统计',
  3: '证据材料',
  4: '其他',
};

const EXPORT_STATUS_MAP: Record<number, { label: string; color: string }> = {
  10: { label: '处理中', color: 'blue' },
  20: { label: '成功', color: 'green' },
  30: { label: '失败', color: 'red' },
};

const SMS_STATUS_MAP: Record<number, { label: string; color: string }> = {
  0: { label: '待发送', color: 'default' },
  1: { label: '已发送', color: 'green' },
  2: { label: '发送失败', color: 'red' },
};

const formatFileSize = (bytes: number) => {
  if (!bytes) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

const ExportLogList: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const { message, modal } = App.useApp();
  const [detailVisible, setDetailVisible] = useState(false);
  const [detailRecord, setDetailRecord] = useState<DataExportLog | null>(null);
  const [downloadingId, setDownloadingId] = useState<string | number | null>(null);

  const showDetail = async (record: DataExportLog) => {
    try {
      const res: any = await exportService.getExportDetail(record.id);
      setDetailRecord(res?.data ?? res);
      setDetailVisible(true);
    } catch (error: any) {
      message.error(error.message || '获取详情失败');
    }
  };

  const handleDownload = async (record: DataExportLog) => {
    if (!record.exportNo) {
      message.warning('该导出记录无文件信息');
      return;
    }
    const isExpired = record.expiredAt && dayjs().isAfter(dayjs(record.expiredAt));
    if (isExpired) {
      message.error('该导出文件已过期，无法下载');
      return;
    }
    modal.confirm({
      title: '确认下载',
      icon: <WarningOutlined style={{ color: '#faad14' }} />,
      content: (
        <div>
          <div>您即将下载加密压缩包，请确认已收到解压密码短信。</div>
          <div style={{ marginTop: 8, color: '#8c8c8c' }}>
            文件格式：AES-256 加密 ZIP 包，有效期至 {dayjs(record.expiredAt).format('YYYY-MM-DD HH:mm:ss')}
          </div>
        </div>
      ),
      okText: '确认下载',
      cancelText: '取消',
      onOk: async () => {
        try {
          setDownloadingId(record.id);
          await exportService.downloadExport(record.id, record.exportNo);
          message.success('文件下载成功，请使用短信中的密码解压');
        } catch (error: any) {
          message.error(error.message || '下载失败');
        } finally {
          setDownloadingId(null);
        }
      },
    });
  };

  const columns: ProColumns<DataExportLog>[] = [
    {
      title: '导出单号',
      dataIndex: 'exportNo',
      width: 200,
      fixed: 'left',
      render: (text, record) => (
        <a onClick={() => showDetail(record)}>{text}</a>
      ),
    },
    {
      title: '导出类型',
      dataIndex: 'exportType',
      width: 120,
      valueType: 'select',
      valueEnum: EXPORT_TYPE_MAP,
      render: (_, record) => {
        const type = record.exportType;
        const label = EXPORT_TYPE_MAP[type] || '未知';
        return <Tag color="blue">{label}</Tag>;
      },
    },
    {
      title: '导出名称',
      dataIndex: 'exportName',
      width: 200,
      ellipsis: true,
    },
    {
      title: '记录数',
      dataIndex: 'recordCount',
      width: 100,
      sorter: true,
      align: 'right',
      render: (val) => `${val ?? 0} 条`,
    },
    {
      title: '文件大小',
      dataIndex: 'fileSize',
      width: 120,
      align: 'right',
      render: (val) => formatFileSize(val || 0),
    },
    {
      title: '加密算法',
      dataIndex: 'encryptionAlgorithm',
      width: 140,
      render: (val) => val || 'AES-256-GCM',
    },
    {
      title: '导出状态',
      dataIndex: 'exportStatus',
      width: 100,
      valueType: 'select',
      valueEnum: Object.fromEntries(
        Object.entries(EXPORT_STATUS_MAP).map(([k, v]) => [k, v.label]),
      ),
      render: (_, record) => {
        const status = record.exportStatus;
        const info = EXPORT_STATUS_MAP[status] || { label: '未知', color: 'default' };
        return <Tag color={info.color}>{info.label}</Tag>;
      },
    },
    {
      title: '密码短信',
      dataIndex: 'passwordSmsSent',
      width: 100,
      render: (_, record) => {
        const status = record.passwordSmsSent;
        const info = SMS_STATUS_MAP[status] || { label: '未知', color: 'default' };
        return <Tag color={info.color}>{info.label}</Tag>;
      },
    },
    {
      title: '操作人',
      dataIndex: 'operatorName',
      width: 100,
    },
    {
      title: '所属机构',
      dataIndex: 'orgName',
      width: 140,
      ellipsis: true,
    },
    {
      title: 'IP地址',
      dataIndex: 'ipAddress',
      width: 140,
    },
    {
      title: '完成时间',
      dataIndex: 'completedAt',
      width: 180,
      valueType: 'dateTime',
      render: (_, record) => record.completedAt ? dayjs(record.completedAt).format('YYYY-MM-DD HH:mm:ss') : '-',
    },
    {
      title: '过期时间',
      dataIndex: 'expiredAt',
      width: 180,
      valueType: 'dateTime',
      render: (_, record) => {
        if (!record.expiredAt) return '-';
        const expired = dayjs().isAfter(dayjs(record.expiredAt));
        return (
          <span style={{ color: expired ? '#ff4d4f' : undefined }}>
            {dayjs(record.expiredAt).format('YYYY-MM-DD HH:mm:ss')}
            {expired && <Tag color="red" style={{ marginLeft: 4 }}>已过期</Tag>}
          </span>
        );
      },
    },
    {
      title: '操作',
      valueType: 'option',
      key: 'option',
      width: 180,
      fixed: 'right',
      render: (_, record) => {
        const isSuccess = record.exportStatus === 20;
        const isExpired = record.expiredAt && dayjs().isAfter(dayjs(record.expiredAt));
        const canDownload = isSuccess && !isExpired;
        return [
          <Button
            key="detail"
            type="link"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => showDetail(record)}
          >
            详情
          </Button>,
          <Button
            key="download"
            type="link"
            size="small"
            icon={<DownloadOutlined />}
            disabled={!canDownload}
            loading={downloadingId === record.id}
            onClick={() => handleDownload(record)}
          >
            下载
          </Button>,
        ];
      },
    },
  ];

  return (
    <>
      <ProTable<DataExportLog>
        columns={columns}
        actionRef={actionRef}
        cardBordered
        rowKey="id"
        search={{
          labelWidth: 'auto',
          defaultCollapsed: false,
        }}
        dateFormatter="string"
        headerTitle="数据导出记录"
        toolBarRender={() => [
          <Button
            key="refresh"
            icon={<ReloadOutlined />}
            onClick={() => actionRef.current?.reload()}
          >
            刷新
          </Button>,
        ]}
        request={async (params) => {
          try {
            const startDate = (params as any).createdAt?.[0];
            const endDate = (params as any).createdAt?.[1];
            const res: any = await exportService.getExportList({
              page: params.current,
              pageSize: params.pageSize,
              keyword: params.keyword as string,
              exportType: params.exportType as number,
              exportStatus: params.exportStatus as number,
              startTime: startDate ? dayjs(startDate).format('YYYY-MM-DD') : undefined,
              endTime: endDate ? dayjs(endDate).format('YYYY-MM-DD') : undefined,
            });
            const data = res?.data ?? res;
            return {
              data: data.list || [],
              success: true,
              total: data.total || 0,
            };
          } catch {
            return { data: [], success: false, total: 0 };
          }
        }}
        columnsState={{
          persistenceKey: 'export-log-columns',
          persistenceType: 'localStorage',
        }}
        pagination={{
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total) => `共 ${total} 条记录`,
        }}
        scroll={{ x: 2000 }}
      />

      <Modal
        title="导出记录详情"
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={[
          <Button key="close" onClick={() => setDetailVisible(false)}>
            关闭
          </Button>,
        ]}
        width={720}
        destroyOnClose
      >
        {detailRecord && (
          <Descriptions bordered column={2} size="small">
            <Descriptions.Item label="导出单号" span={2}>
              {detailRecord.exportNo}
            </Descriptions.Item>
            <Descriptions.Item label="导出类型">
              {EXPORT_TYPE_MAP[detailRecord.exportType] || '未知'}
            </Descriptions.Item>
            <Descriptions.Item label="导出名称">
              {detailRecord.exportName}
            </Descriptions.Item>
            <Descriptions.Item label="导出状态">
              {(() => {
                const info = EXPORT_STATUS_MAP[detailRecord.exportStatus];
                return <Tag color={info?.color || 'default'}>{info?.label || '未知'}</Tag>;
              })()}
            </Descriptions.Item>
            <Descriptions.Item label="密码短信">
              {(() => {
                const info = SMS_STATUS_MAP[detailRecord.passwordSmsSent];
                return <Tag color={info?.color || 'default'}>{info?.label || '未知'}</Tag>;
              })()}
            </Descriptions.Item>
            <Descriptions.Item label="记录数">{detailRecord.recordCount} 条</Descriptions.Item>
            <Descriptions.Item label="文件大小">
              {formatFileSize(detailRecord.fileSize || 0)}
            </Descriptions.Item>
            <Descriptions.Item label="加密算法">
              {detailRecord.encryptionAlgorithm || 'AES-256-GCM'}
            </Descriptions.Item>
            <Descriptions.Item label="文件名称">
              {detailRecord.fileName || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="操作人">{detailRecord.operatorName}</Descriptions.Item>
            <Descriptions.Item label="联系电话">
              {detailRecord.operatorPhone || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="所属机构">
              {detailRecord.orgName || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="IP地址">
              {detailRecord.ipAddress || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="完成时间">
              {detailRecord.completedAt ? dayjs(detailRecord.completedAt).format('YYYY-MM-DD HH:mm:ss') : '-'}
            </Descriptions.Item>
            <Descriptions.Item label="过期时间">
              {detailRecord.expiredAt ? dayjs(detailRecord.expiredAt).format('YYYY-MM-DD HH:mm:ss') : '-'}
            </Descriptions.Item>
            <Descriptions.Item label="创建时间" span={2}>
              {detailRecord.createdAt ? dayjs(detailRecord.createdAt).format('YYYY-MM-DD HH:mm:ss') : '-'}
            </Descriptions.Item>
            <Descriptions.Item label="筛选条件" span={2}>
              {detailRecord.filterConditions ? (
                <pre style={{ margin: 0, whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
                  {detailRecord.filterConditions}
                </pre>
              ) : (
                '-'
              )}
            </Descriptions.Item>
            {detailRecord.errorMessage && (
              <Descriptions.Item label="错误信息" span={2}>
                <span style={{ color: '#ff4d4f' }}>{detailRecord.errorMessage}</span>
              </Descriptions.Item>
            )}
          </Descriptions>
        )}
      </Modal>
    </>
  );
};

export default ExportLogList;
