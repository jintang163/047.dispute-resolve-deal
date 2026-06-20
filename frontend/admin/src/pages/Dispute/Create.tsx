import React, { useState, useRef, useCallback, useEffect, useMemo } from 'react';
import { Card, Row, Col, App, Tag, Space, Tooltip, Spin, Cascader, Alert, DatePicker, Radio } from 'antd';
import { SaveOutlined, ThunderboltOutlined, CloseOutlined, RobotOutlined } from '@ant-design/icons';
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
import { disputeService, DisputeTypeNode, KeywordExtractResult } from '../../services/dispute';
import dayjs from 'dayjs';

const keywordColorPalette = ['red', 'volcano', 'orange', 'gold', 'lime', 'green', 'cyan', 'blue', 'geekblue', 'magenta'];

interface CascaderOption {
  value: number;
  label: string;
  children?: CascaderOption[];
}

const buildCascaderOptions = (nodes: DisputeTypeNode[]): CascaderOption[] =>
  nodes.map(n => ({
    value: n.id,
    label: n.typeName,
    children: n.children && n.children.length > 0 ? buildCascaderOptions(n.children) : undefined,
  }));

const findLeafPath = (nodes: DisputeTypeNode[], targetId: number, path: number[] = []): number[] => {
  for (const n of nodes) {
    const np = [...path, n.id];
    if (n.id === targetId) return np;
    if (n.children && n.children.length > 0) {
      const r = findLeafPath(n.children, targetId, np);
      if (r.length > 0) return r;
    }
  }
  return [];
};

const DisputeCreate: React.FC = () => {
  const navigate = useNavigate();
  const { message } = App.useApp();
  const [loading, setLoading] = useState(false);
  const [extractedKeywords, setExtractedKeywords] = useState<string[]>([]);
  const [extracting, setExtracting] = useState(false);
  const [suggestedType, setSuggestedType] = useState<{
    typeId: number;
    typeName: string;
    reason: string;
  } | null>(null);
  const [typeTree, setTypeTree] = useState<DisputeTypeNode[]>([]);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const formRef = useRef<any>();

  useEffect(() => {
    disputeService
      .getTypes()
      .then(res => {
        const data: any = (res as any)?.data ?? res;
        if (Array.isArray(data)) setTypeTree(data);
      })
      .catch(() => {});
  }, []);

  const cascaderOptions = useMemo(() => buildCascaderOptions(typeTree), [typeTree]);

  const handleExtractKeywords = useCallback(async () => {
    const values = formRef.current?.getFieldsValue?.();
    const description = values?.description || '';
    const title = values?.title || '';
    const typePath = values?.typeId as number[];
    const lastTypeId = typePath?.length ? typePath[typePath.length - 1] : undefined;

    const combined = (title + ' ' + description).trim();
    if (combined.length < 4) {
      message.warning('请先输入纠纷标题或描述（至少4个字）');
      return;
    }

    setExtracting(true);
    try {
      const res = await disputeService.extractKeywords(description, title, lastTypeId);
      const payload: any = (res as any)?.data ?? (res as KeywordExtractResult);
      const data = payload as KeywordExtractResult;

      if (data?.keywords?.length > 0) {
        setExtractedKeywords(data.keywords);
      }

      if (data?.suggestedTypeId && data.suggestedTypeId > 0) {
        setSuggestedType({
          typeId: data.suggestedTypeId,
          typeName: data.suggestedTypeName || '',
          reason: data.reason || '',
        });
        const path = findLeafPath(typeTree, data.suggestedTypeId);
        if (path.length > 0) {
          formRef.current?.setFieldsValue?.({ typeId: path });
        }
        message.success(
          `AI已提取 ${data.keywords?.length || 0} 个关键词并建议分类「${data.suggestedTypeName || data.suggestedTypeId}」`,
        );
      } else if (data?.keywords?.length > 0) {
        message.success(`AI已提取 ${data.keywords.length} 个关键词标签`);
      } else {
        message.info('未能提取到有效关键词');
      }
    } catch (error: any) {
      message.error(error.message || '关键词提取失败');
    } finally {
      setExtracting(false);
    }
  }, [message, typeTree]);

  const handleDescriptionChange = useCallback(() => {
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => {
      const values = formRef.current?.getFieldsValue?.();
      const description = values?.description || '';
      const title = values?.title || '';
      const combined = (title + ' ' + description).trim();
      if (combined.length >= 10) handleExtractKeywords();
    }, 1500);
  }, [handleExtractKeywords]);

  const removeKeyword = useCallback((kw: string) => {
    setExtractedKeywords(prev => prev.filter(k => k !== kw));
  }, []);

  const onFinish = async (values: any) => {
    try {
      setLoading(true);
      const typePath = values.typeId as number[];
      const typeId = typePath?.length ? typePath[typePath.length - 1] : 0;
      const caseLevelMap: Record<string, number> = { critical: 1, urgent: 2, normal: 3 };
      const caseLevel = caseLevelMap[values.urgency || 'normal'] ?? 3;

      const params = {
        title: values.title,
        typeId,
        description: values.description,
        occurAddress: values.occurAddress,
        occurTime: values.occurTime ? dayjs(values.occurTime).format('YYYY-MM-DD HH:mm:ss') : undefined,
        expectation: values.expectation,
        caseLevel,
        caseSource: 4,
        reporterName: values.reporterName,
        reporterPhone: values.reporterPhone,
        reporterAddress: values.reporterAddress,
        reporterIdCard: values.reporterIdCard,
        respondentName: values.respondentName,
        respondentPhone: values.respondentPhone,
        respondentAddress: values.respondentAddress,
        organizationId: values.organizationId ? Number(values.organizationId) : undefined,
        longitude: values.longitude,
        latitude: values.latitude,
        keywords: extractedKeywords.length > 0 ? extractedKeywords : undefined,
      };

      await disputeService.create(params as any);
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
          resetButtonProps: { size: 'large' },
        }}
        initialValues={{
          urgency: 'normal',
          caseSource: 4,
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
              <ProForm.Item
                label={
                  <Space>
                    纠纷类型（三级分类）
                    <Tag color="blue" icon={<RobotOutlined />}>
                      AI自动推断
                    </Tag>
                  </Space>
                }
                name="typeId"
                rules={[{ required: true, message: '请选择纠纷类型，或输入描述后由AI自动推断' }]}
              >
                <Cascader
                  size="large"
                  options={cascaderOptions}
                  placeholder="请选择纠纷类型（建议输入描述后由AI自动匹配）"
                  changeOnSelect
                  showSearch
                  expandTrigger="hover"
                />
              </ProForm.Item>
              {suggestedType && (
                <Alert
                  style={{ marginTop: 6 }}
                  type="info"
                  showIcon
                  icon={<RobotOutlined />}
                  message={
                    <Space size="small">
                      <strong>AI推荐分类：</strong>
                      <Tag color="geekblue">{suggestedType.typeName || `ID ${suggestedType.typeId}`}</Tag>
                      {suggestedType.reason && (
                        <span style={{ color: '#666', fontSize: 12 }}>{suggestedType.reason}</span>
                      )}
                    </Space>
                  }
                />
              )}
            </Col>

            <Col xs={24} md={12}>
              <ProForm.Item label="紧急程度" name="urgency" initialValue="normal">
                <Radio.Group size="large" buttonStyle="solid">
                  <Radio.Button value="normal">普通</Radio.Button>
                  <Radio.Button value="urgent">紧急</Radio.Button>
                  <Radio.Button value="critical">特急</Radio.Button>
                </Radio.Group>
              </ProForm.Item>
            </Col>

            <Col xs={24} md={12}>
              <ProFormSelect
                label="所属组织"
                name="organizationId"
                placeholder="请选择所属组织"
                valueEnum={{
                  '1': '市综治中心',
                  '2': '朝阳区街道办',
                  '3': '海淀区街道办',
                  '4': '朝阳门社区',
                  '5': '建国门社区',
                }}
                fieldProps={{ size: 'large', showSearch: true }}
              />
            </Col>

            <Col xs={24} md={12}>
              <ProFormText
                label="纠纷发生地点"
                name="occurAddress"
                placeholder="请输入纠纷发生地点"
                fieldProps={{ size: 'large' }}
              />
            </Col>

            <Col xs={24} md={12}>
              <ProForm.Item label="发生时间" name="occurTime">
                <DatePicker
                  showTime
                  style={{ width: '100%' }}
                  size="large"
                  placeholder="请选择纠纷发生时间"
                />
              </ProForm.Item>
            </Col>

            <Col xs={24} md={24}>
              <ProFormTextArea
                label={
                  <Space>
                    案件描述
                    <span style={{ color: '#999', fontSize: 12, fontWeight: 400 }}>
                      输入后 AI 将自动提取关键词并推断三级分类
                    </span>
                  </Space>
                }
                name="description"
                placeholder="请详细描述案件情况，如：楼上装修噪音持续一周，多次沟通无果，影响正常休息..."
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
                    {extracting && <Spin size="small" tip="AI分析中..." />}
                  </Space>
                }
                extra={
                  <Tooltip title="手动触发AI从纠纷描述中提取关键词并推断三级分类">
                    <a
                      onClick={handleExtractKeywords}
                      style={{ opacity: extracting ? 0.5 : 1, pointerEvents: extracting ? 'none' : 'auto' }}
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
                        color={keywordColorPalette[idx % keywordColorPalette.length]}
                        closable
                        onClose={() => removeKeyword(kw)}
                        closeIcon={<CloseOutlined style={{ fontSize: 10 }} />}
                        style={{ margin: 2, fontSize: 13, padding: '2px 8px' }}
                      >
                        {kw}
                      </Tag>
                    ))}
                    <Tag
                      style={{ borderStyle: 'dashed', cursor: 'pointer', background: '#f0f5ff' }}
                      onClick={handleExtractKeywords}
                    >
                      + 添加/重新提取
                    </Tag>
                  </Space>
                ) : (
                  <div style={{ color: '#999', fontSize: 13 }}>
                    输入纠纷描述后，AI将自动提取核心关键词标签（如"噪音扰民""欠薪3个月""漏水赔偿"等）
                  </div>
                )}
              </Card>
            </Col>

            <Col xs={24} md={24}>
              <ProFormTextArea
                label="当事人期望"
                name="expectation"
                placeholder="请输入当事人期望解决方式"
                fieldProps={{ rows: 2, showCount: true, maxLength: 500 }}
              />
            </Col>
          </Row>
        </ProCard>

        <ProCard title="报案人（甲方）信息" headerBordered collapsible style={{ marginTop: 16 }}>
          <Row gutter={24}>
            <Col xs={24} md={12}>
              <ProFormText
                label="姓名/单位名称"
                name="reporterName"
                placeholder="请输入报案人姓名或单位名称"
                rules={[{ required: true, message: '请输入报案人姓名/单位' }]}
                fieldProps={{ size: 'large' }}
              />
            </Col>
            <Col xs={24} md={12}>
              <ProFormText
                label="联系电话"
                name="reporterPhone"
                placeholder="请输入联系电话"
                fieldProps={{ size: 'large' }}
                rules={[{ pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号码' }]}
              />
            </Col>
            <Col xs={24} md={12}>
              <ProFormText label="身份证号" name="reporterIdCard" fieldProps={{ size: 'large' }} />
            </Col>
            <Col xs={24} md={12}>
              <ProFormText label="联系地址" name="reporterAddress" fieldProps={{ size: 'large' }} />
            </Col>
          </Row>
        </ProCard>

        <ProCard title="对方（乙方）信息" headerBordered collapsible style={{ marginTop: 16 }}>
          <Row gutter={24}>
            <Col xs={24} md={12}>
              <ProFormText
                label="姓名/单位名称"
                name="respondentName"
                placeholder="请输入对方姓名或单位名称"
                rules={[{ required: true, message: '请输入对方姓名/单位' }]}
                fieldProps={{ size: 'large' }}
              />
            </Col>
            <Col xs={24} md={12}>
              <ProFormText
                label="联系电话"
                name="respondentPhone"
                placeholder="请输入对方联系电话"
                fieldProps={{ size: 'large' }}
                rules={[{ pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号码' }]}
              />
            </Col>
            <Col xs={24} md={24}>
              <ProFormText label="联系地址" name="respondentAddress" fieldProps={{ size: 'large' }} />
            </Col>
          </Row>
        </ProCard>

        <div style={{ height: 64 }} />
      </ProForm>
    </div>
  );
};

export default DisputeCreate;
