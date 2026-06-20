import React, { useState, useRef, useCallback } from 'react';
import { Card, Row, Col, App, Tag, Space, Tooltip, Spin } from 'antd';
import { ArrowLeftOutlined, SaveOutlined, ThunderboltOutlined, CloseOutlined } from '@ant-design/icons';
import {
  ProForm,
  ProFormText,
  ProFormTextArea,
  ProFormSelect,
  ProFormDigit,
  FooterToolbar,
  ProCard,
} from '@ant-design/pro-components';
import { useNavigate } from 'react-router-dom';
import { disputeService } from '../../services/dispute';

const typeOptions = [
  { label: '民事纠纷', value: 'civil' },
  { label: '劳动争议', value: 'labor' },
  { label: '家庭纠纷', value: 'family' },
  { label: '邻里纠纷', value: 'neighborhood' },
  { label: '合同纠纷', value: 'contract' },
  { label: '物业纠纷', value: 'property' },
  { label: '其他纠纷', value: 'other' },
];

const urgencyOptions = [
  { label: '普通', value: 'normal' },
  { label: '紧急', value: 'urgent' },
  { label: '特急', value: 'critical' },
];

const orgOptions = [
  { label: '综治中心', value: 'org_001' },
  { label: '东街街道办事处', value: 'org_002' },
  { label: '西街街道办事处', value: 'org_003' },
  { label: '南区调解委员会', value: 'org_004' },
  { label: '北区调解委员会', value: 'org_005' },
];

const keywordColorMap: Record<string, string> = {
  '纠纷性质': 'red',
  '行为': 'orange',
  '对象': 'blue',
  '程度': 'purple',
};

const DisputeCreate: React.FC = () => {
  const navigate = useNavigate();
  const { message } = App.useApp();
  const [loading, setLoading] = useState(false);
  const [extractedKeywords, setExtractedKeywords] = useState<string[]>([]);
  const [extracting, setExtracting] = useState(false);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const formRef = useRef<any>();

  const handleExtractKeywords = useCallback(async () => {
    const values = formRef.current?.getFieldsValue?.();
    const description = values?.description || '';
    const title = values?.title || '';
    const combined = (title + ' ' + description).trim();

    if (combined.length < 4) {
      message.warning('请先输入纠纷标题或描述（至少4个字）');
      return;
    }

    setExtracting(true);
    try {
      const res = await disputeService.extractKeywords(
        description,
        title,
        undefined,
      );
      const data = (res as any)?.data;
      if (data?.keywords?.length > 0) {
        setExtractedKeywords(data.keywords);
        message.success(`AI已提取 ${data.keywords.length} 个关键词标签`);
      } else {
        message.info('未能提取到有效关键词');
      }
    } catch (error: any) {
      message.error(error.message || '关键词提取失败');
    } finally {
      setExtracting(false);
    }
  }, [message]);

  const handleDescriptionChange = useCallback(() => {
    if (debounceRef.current) {
      clearTimeout(debounceRef.current);
    }
    debounceRef.current = setTimeout(() => {
      const values = formRef.current?.getFieldsValue?.();
      const description = values?.description || '';
      const title = values?.title || '';
      const combined = (title + ' ' + description).trim();
      if (combined.length >= 10) {
        handleExtractKeywords();
      }
    }, 1500);
  }, [handleExtractKeywords]);

  const removeKeyword = useCallback((kw: string) => {
    setExtractedKeywords(prev => prev.filter(k => k !== kw));
  }, []);

  const onFinish = async (values: any) => {
    try {
      setLoading(true);
      const params = {
        ...values,
        keywords: extractedKeywords.length > 0 ? extractedKeywords : undefined,
      };
      await disputeService.create(params);
      message.success('案件创建成功');
      navigate('/dispute');
    } catch (error: any) {
      message.error(error.message || '案件创建失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <ProForm
        formRef={formRef}
        layout="vertical"
        onFinish={onFinish}
        submitter={{
          render: (_, dom) => <FooterToolbar>{dom}</FooterToolbar>,
          searchConfig: { submitText: '提交案件' },
          submitButtonProps: {
            loading,
            icon: <SaveOutlined />,
            size: 'large',
          },
          resetButtonProps: {
            size: 'large',
          },
        }}
      >
        <ProCard title="基本信息" headerBordered collapsible>
          <Row gutter={24}>
            <Col xs={24} md={12}>
              <ProFormText
                label="案件标题"
                name="title"
                placeholder="请输入案件标题"
                rules={[
                  { required: true, message: '请输入案件标题' },
                  { max: 200, message: '标题长度不超过200个字符' },
                ]}
                fieldProps={{
                  size: 'large',
                  onChange: () => handleDescriptionChange(),
                }}
              />
            </Col>
            <Col xs={24} md={12}>
              <ProFormSelect
                label="案件类型"
                name="type"
                placeholder="请选择案件类型"
                options={typeOptions}
                rules={[{ required: true, message: '请选择案件类型' }]}
                fieldProps={{ size: 'large' }}
              />
            </Col>
            <Col xs={24} md={12}>
              <ProFormSelect
                label="紧急程度"
                name="urgency"
                placeholder="请选择紧急程度"
                options={urgencyOptions}
                initialValue="normal"
                fieldProps={{ size: 'large' }}
              />
            </Col>
            <Col xs={24} md={12}>
              <ProFormSelect
                label="所属组织"
                name="orgId"
                placeholder="请选择所属组织"
                options={orgOptions}
                fieldProps={{ size: 'large', showSearch: true }}
              />
            </Col>
            <Col xs={24} md={24}>
              <ProFormText
                label="纠纷地点"
                name="address"
                placeholder="请输入纠纷发生地点"
                fieldProps={{ size: 'large' }}
              />
            </Col>
            <Col xs={24} md={24}>
              <ProFormTextArea
                label="案件描述"
                name="description"
                placeholder="请详细描述案件情况，AI将自动提取核心关键词标签"
                fieldProps={{
                  size: 'large',
                  rows: 5,
                  showCount: true,
                  maxLength: 2000,
                  onChange: () => handleDescriptionChange(),
                }}
                rules={[{ max: 2000, message: '描述长度不超过2000个字符' }]}
              />
            </Col>
            <Col xs={24} md={24}>
              <Card
                size="small"
                title={
                  <Space>
                    <ThunderboltOutlined style={{ color: '#1890ff' }} />
                    <span>AI关键词标签</span>
                    {extracting && <Spin size="small" />}
                  </Space>
                }
                extra={
                  <Tooltip title="手动触发AI从纠纷描述中提取关键词">
                    <a
                      onClick={handleExtractKeywords}
                      style={{ opacity: extracting ? 0.5 : 1 }}
                    >
                      <ThunderboltOutlined /> 重新提取
                    </a>
                  </Tooltip>
                }
                style={{ marginBottom: 16 }}
              >
                {extractedKeywords.length > 0 ? (
                  <Space wrap>
                    {extractedKeywords.map((kw, idx) => (
                      <Tag
                        key={kw}
                        color={
                          ['red', 'volcano', 'orange', 'gold', 'lime', 'green', 'cyan', 'blue'][idx % 8]
                        }
                        closable
                        onClose={() => removeKeyword(kw)}
                        closeIcon={<CloseOutlined style={{ fontSize: 10 }} />}
                        style={{ margin: 2, fontSize: 13, padding: '2px 8px' }}
                      >
                        {kw}
                      </Tag>
                    ))}
                    <Tag
                      style={{
                        borderStyle: 'dashed',
                        cursor: 'pointer',
                        background: '#f0f5ff',
                      }}
                      onClick={handleExtractKeywords}
                    >
                      + 添加
                    </Tag>
                  </Space>
                ) : (
                  <div style={{ color: '#999', fontSize: 13 }}>
                    输入纠纷描述后，AI将自动提取核心关键词标签（如"噪音扰民""欠薪3个月""漏水赔偿"等）
                  </div>
                )}
              </Card>
            </Col>
          </Row>
        </ProCard>

        <ProCard
          title="甲方信息"
          headerBordered
          collapsible
          style={{ marginTop: 16 }}
        >
          <Row gutter={24}>
            <Col xs={24} md={12}>
              <ProFormText
                label="姓名/单位名称"
                name="partyA"
                placeholder="请输入甲方姓名或单位名称"
                rules={[{ required: true, message: '请输入甲方信息' }]}
                fieldProps={{ size: 'large' }}
              />
            </Col>
            <Col xs={24} md={12}>
              <ProFormText
                label="联系电话"
                name="partyAPhone"
                placeholder="请输入甲方联系电话"
                fieldProps={{ size: 'large' }}
                rules={[
                  {
                    pattern: /^1[3-9]\d{9}$/,
                    message: '请输入正确的手机号码',
                  },
                ]}
              />
            </Col>
          </Row>
        </ProCard>

        <ProCard
          title="乙方信息"
          headerBordered
          collapsible
          style={{ marginTop: 16 }}
        >
          <Row gutter={24}>
            <Col xs={24} md={12}>
              <ProFormText
                label="姓名/单位名称"
                name="partyB"
                placeholder="请输入乙方姓名或单位名称"
                rules={[{ required: true, message: '请输入乙方信息' }]}
                fieldProps={{ size: 'large' }}
              />
            </Col>
            <Col xs={24} md={12}>
              <ProFormText
                label="联系电话"
                name="partyBPhone"
                placeholder="请输入乙方联系电话"
                fieldProps={{ size: 'large' }}
                rules={[
                  {
                    pattern: /^1[3-9]\d{9}$/,
                    message: '请输入正确的手机号码',
                  },
                ]}
              />
            </Col>
          </Row>
        </ProCard>

        <div style={{ height: 64 }} />
      </ProForm>
    </div>
  );
};

export default DisputeCreate;
