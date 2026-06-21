import React, { useEffect, useState } from 'react';
import {
  Button,
  Space,
  App,
  Drawer,
  Form,
  Input,
  Select,
  DatePicker,
  InputNumber,
  Radio,
  Row,
  Col,
  Tag,
  Spin,
  Alert,
  Card,
  Typography,
  Tooltip,
  Empty,
  Divider,
  Cascader,
} from 'antd';
import {
  FileTextOutlined,
  PlusOutlined,
  EditOutlined,
  RobotOutlined,
  SaveOutlined,
  SaveFilled,
  BulbOutlined,
  CheckCircleTwoTone,
} from '@ant-design/icons';
import dayjs, { Dayjs } from 'dayjs';
import type { Dayjs as DayjsType } from 'dayjs';
import {
  mediationTemplateService,
  MediationTemplate,
  MediationTemplateCategory,
  ApplyMediationTemplateResponse,
} from '../../services/mediationTemplate';
import { MediationRecord } from '../../services/user';
import { disputeService } from '../../services/dispute';

const { TextArea } = Input;
const { Title, Paragraph, Text } = Typography;

export interface MediationRecordFormData {
  id?: string;
  recordType?: number;
  mediationTime?: Dayjs;
  mediationPlace?: string;
  mediationDuration?: number;
  participants?: string[];
  processContent?: string;
  disputeFocus?: string;
  mediationOpinion?: string;
  agreementContent?: string;
  result?: number;
  nextStep?: string;
  assistMediators?: number[];
  isDraft?: number;
  templateId?: string;
  templateName?: string;
}

interface MediationRecordEditorProps {
  open: boolean;
  onClose: () => void;
  caseId: string;
  caseInfo?: {
    typeName?: string;
    mediatorId?: string;
    mediatorName?: string;
  };
  initialData?: MediationRecord;
  onSubmit: (data: MediationRecordFormData, isDraft: boolean) => Promise<{ id: string } | any>;
}

const recordTypeOptions = [
  { label: '初次调解', value: 1 },
  { label: '再次调解', value: 2 },
  { label: '补充调解', value: 3 },
];

const resultOptions = [
  { label: '进行中', value: 0 },
  { label: '达成协议', value: 1 },
  { label: '未达成', value: 2 },
  { label: '转审/转法援', value: 3 },
];

const MediationRecordEditor: React.FC<MediationRecordEditorProps> = ({
  open,
  onClose,
  caseId,
  caseInfo,
  initialData,
  onSubmit,
}) => {
  const { message, modal } = App.useApp();
  const [form] = Form.useForm<MediationRecordFormData>();
  const [submitting, setSubmitting] = useState(false);

  const [templateDrawerOpen, setTemplateDrawerOpen] = useState(false);
  const [templateCategories, setTemplateCategories] = useState<MediationTemplateCategory[]>([]);
  const [templateList, setTemplateList] = useState<MediationTemplate[]>([]);
  const [templatesLoading, setTemplatesLoading] = useState(false);
  const [selectedTemplate, setSelectedTemplate] = useState<MediationTemplate | null>(null);
  const [templateCategory, setTemplateCategory] = useState<string>('');
  const [applyingTemplate, setApplyingTemplate] = useState(false);

  useEffect(() => {
    if (open) {
      form.resetFields();
      if (initialData) {
        form.setFieldsValue({
          id: initialData.id,
          recordType: initialData.recordType || 1,
          mediationTime: initialData.mediationTime ? dayjs(initialData.mediationTime) : dayjs(),
          mediationPlace: initialData.place,
          mediationDuration: initialData.duration || 30,
          participants: initialData.participantNames
            ? initialData.participantNames.split(',').filter(Boolean)
            : [],
          processContent: initialData.processContent,
          disputeFocus: initialData.disputeFocus,
          mediationOpinion: initialData.mediationOpinion,
          agreementContent: initialData.agreementContent,
          result: initialData.result ? Number(initialData.result) : 0,
          nextStep: initialData.nextStep,
          isDraft: initialData.isDraft ? 1 : 0,
          templateId: initialData.templateId,
          templateName: initialData.templateName,
        });
      } else {
        form.setFieldsValue({
          recordType: 1,
          mediationTime: dayjs(),
          mediationDuration: 30,
          result: 0,
          isDraft: 0,
        });
      }
    }
  }, [open, initialData, form]);

  const loadTemplates = async () => {
    if (!templateDrawerOpen) return;
    setTemplatesLoading(true);
    try {
      const [catRes, listRes] = await Promise.all([
        mediationTemplateService.getCategories(),
        mediationTemplateService.getRecommend(caseId),
      ]);
      setTemplateCategories(catRes.data || []);
      setTemplateList(listRes.data || []);
    } catch (e: any) {
      message.error(e.message || '加载模板失败');
    } finally {
      setTemplatesLoading(false);
    }
  };

  useEffect(() => {
    if (templateDrawerOpen) {
      loadTemplates();
    }
  }, [templateDrawerOpen, caseId]);

  const handleApplyTemplate = async (template: MediationTemplate) => {
    setApplyingTemplate(true);
    try {
      const res = await mediationTemplateService.apply(String(template.id), { caseId });
      const data: ApplyMediationTemplateResponse = res.data || (res as any);

      form.setFieldsValue({
        recordType: data.recordType,
        mediationTime: dayjs(data.mediationTime),
        mediationPlace: data.mediationPlace,
        mediationDuration: data.mediationDuration,
        participants: data.participantNames
          ? data.participantNames.split(',').filter(Boolean)
          : [],
        processContent: data.processContent,
        disputeFocus: data.disputeFocus,
        mediationOpinion: data.mediationOpinion,
        agreementContent: data.agreementContent,
        nextStep: data.nextStep,
        templateId: String(data.templateId),
        templateName: data.templateName,
      });

      setSelectedTemplate(template);
      setTemplateDrawerOpen(false);
      message.success(`已套用模板「${template.templateName}」，请根据实际情况微调后保存`);
    } catch (e: any) {
      message.error(e.message || '套用模板失败');
    } finally {
      setApplyingTemplate(false);
    }
  };

  const handleSubmit = async (isDraft: boolean) => {
    try {
      const values = await form.validateFields();
      setSubmitting(true);

      const submitData: MediationRecordFormData = {
        ...values,
        mediationTime: values.mediationTime ? values.mediationTime.format('YYYY-MM-DD HH:mm:ss') : undefined,
        isDraft: isDraft ? 1 : 0,
        id: initialData?.id,
      } as any;

      await onSubmit(submitData, isDraft);
      message.success(isDraft ? '草稿已保存' : '调解记录已保存');
      onClose();
    } catch (e: any) {
      if (e.errorFields) {
        message.warning('请检查表单填写');
        return;
      }
      message.error(e.message || '保存失败');
    } finally {
      setSubmitting(false);
    }
  };

  const filteredTemplates = templateCategory
    ? templateList.filter((t) => t.category === templateCategory)
    : templateList;

  return (
    <>
      <Drawer
        title={
          <Space>
            <FileTextOutlined style={{ color: '#1677ff' }} />
            <span>{initialData ? '编辑调解记录' : '录入调解记录'}</span>
            {selectedTemplate && (
              <Tag color="blue" icon={<CheckCircleTwoTone />}>
                已套用模板：{selectedTemplate.templateName}
              </Tag>
            )}
          </Space>
        }
        width={1100}
        open={open}
        onClose={onClose}
        destroyOnClose
        extra={
          <Space>
            <Button
              icon={<FileTextOutlined />}
              onClick={() => setTemplateDrawerOpen(true)}
              loading={applyingTemplate}
            >
              选择模板
            </Button>
            <Button onClick={onClose}>取消</Button>
            <Button icon={<SaveOutlined />} onClick={() => handleSubmit(true)} loading={submitting}>
              保存草稿
            </Button>
            <Button
              type="primary"
              icon={<SaveFilled />}
              onClick={() => handleSubmit(false)}
              loading={submitting}
            >
              提交记录
            </Button>
          </Space>
        }
      >
        {initialData?.isDraft && (
          <Alert
            type="warning"
            showIcon
            style={{ marginBottom: 16 }}
            message="当前为草稿记录"
            description="请检查内容后点击「提交记录」转为正式记录，或继续调整后保存草稿。"
          />
        )}

        <Form<MediationRecordFormData>
          form={form}
          layout="vertical"
          initialValues={{ recordType: 1, result: 0, mediationDuration: 30, isDraft: 0 }}
        >
          <Row gutter={16}>
            <Col span={8}>
              <Form.Item
                label="记录类型"
                name="recordType"
                rules={[{ required: true, message: '请选择记录类型' }]}
              >
                <Radio.Group options={recordTypeOptions} />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                label="调解时间"
                name="mediationTime"
                rules={[{ required: true, message: '请选择调解时间' }]}
              >
                <DatePicker showTime style={{ width: '100%' }} format="YYYY-MM-DD HH:mm" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                label="调解时长（分钟）"
                name="mediationDuration"
                rules={[{ required: true, message: '请输入调解时长' }]}
              >
                <InputNumber min={1} max={480} style={{ width: '100%' }} addonAfter="分钟" />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item label="调解地点" name="mediationPlace">
                <Input placeholder="如：社区调解室、线上视频等" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="参与人员" name="participants">
                <Select
                  mode="tags"
                  placeholder="输入参与人姓名后回车添加"
                  style={{ width: '100%' }}
                  tokenSeparators={[',', '，']}
                />
              </Form.Item>
            </Col>
          </Row>

          <Divider plain orientation="left">
            <Space>
              <EditOutlined />
              <Text strong>调解内容</Text>
            </Space>
          </Divider>

          <Form.Item
            label={
              <Space>
                <span>调解过程记录</span>
                <Tooltip title="记录调解的完整过程，包括沟通、事实核查、情绪疏导、方案协商等环节">
                  <BulbOutlined style={{ color: '#faad14' }} />
                </Tooltip>
              </Space>
            }
            name="processContent"
            rules={[{ required: true, message: '请填写调解过程记录' }]}
          >
            <TextArea
              rows={8}
              placeholder="请填写调解的完整过程，包括沟通情况、事实核查、双方诉求、协商过程等..."
            />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item label="争议焦点" name="disputeFocus">
                <TextArea
                  rows={4}
                  placeholder="请描述双方争议的核心问题..."
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="调解意见" name="mediationOpinion">
                <TextArea
                  rows={4}
                  placeholder="请描述调解员对纠纷的分析和调解意见..."
                />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item label="协议内容" name="agreementContent">
            <TextArea
              rows={4}
              placeholder="如达成协议，请填写具体的协议条款；如未达成，可记录未达成原因..."
            />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="调解结果"
                name="result"
                rules={[{ required: true, message: '请选择调解结果' }]}
              >
                <Radio.Group options={resultOptions} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="下一步计划" name="nextStep">
                <Input placeholder="如：继续沟通、补充材料、转介法援等" />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item name="templateId" hidden>
            <Input />
          </Form.Item>
          <Form.Item name="templateName" hidden>
            <Input />
          </Form.Item>
        </Form>
      </Drawer>

      <Drawer
        title={
          <Space>
            <FileTextOutlined />
            <span>选择调解记录模板</span>
          </Space>
        }
        width={720}
        open={templateDrawerOpen}
        onClose={() => setTemplateDrawerOpen(false)}
        destroyOnClose
      >
        <Alert
          type="info"
          showIcon
          style={{ marginBottom: 16 }}
          message={
            <Space>
              <span>已根据案件纠纷类型为您推荐以下模板，选择后可一键套用标准格式，再根据实际情况微调即可。</span>
            </Space>
          }
        />

        <Space wrap style={{ marginBottom: 16 }}>
          <Tag
            color={templateCategory === '' ? 'blue' : 'default'}
            style={{ cursor: 'pointer', padding: '4px 12px', fontSize: 13 }}
            onClick={() => setTemplateCategory('')}
          >
            全部模板
          </Tag>
          {templateCategories.map((cat) => (
            <Tag
              key={cat.code}
              color={templateCategory === cat.code ? 'blue' : 'default'}
              style={{ cursor: 'pointer', padding: '4px 12px', fontSize: 13 }}
              onClick={() => setTemplateCategory(cat.code)}
            >
              {cat.name}（{cat.count}）
            </Tag>
          ))}
        </Space>

        <Spin spinning={templatesLoading}>
          {filteredTemplates.length > 0 ? (
            <Row gutter={[12, 12]}>
              {filteredTemplates.map((tpl) => (
                <Col xs={24} sm={12} md={12} lg={12} xl={12} key={tpl.id}>
                  <Card
                    size="small"
                    hoverable
                    style={{
                      height: '100%',
                      transition: 'all 0.2s',
                      borderColor: selectedTemplate?.id === tpl.id ? '#1677ff' : '#f0f0f0',
                      borderWidth: selectedTemplate?.id === tpl.id ? 2 : 1,
                    }}
                    bodyStyle={{ padding: 14 }}
                  >
                    <Space align="start" style={{ width: '100%' }}>
                      <div style={{ flex: 1, minWidth: 0 }}>
                        <Space style={{ marginBottom: 6 }} wrap>
                          <Text strong style={{ fontSize: 15 }}>{tpl.templateName}</Text>
                          {tpl.isSystem && <Tag color="geekblue">系统内置</Tag>}
                          {tpl.categoryName && <Tag color="purple">{tpl.categoryName}</Tag>}
                        </Space>
                        <Space size={12} style={{ marginBottom: 8 }} wrap>
                          <Text type="secondary" style={{ fontSize: 12 }}>
                            推荐时长：{tpl.defaultDuration}分钟
                          </Text>
                          <Text type="secondary" style={{ fontSize: 12 }}>
                            使用 {tpl.useCount} 次
                          </Text>
                        </Space>
                        {tpl.tips && (
                          <Alert
                            type="info"
                            style={{ marginBottom: 8 }}
                            showIcon={false}
                            message={<Text style={{ fontSize: 12 }}>{tpl.tips.split('\n')[0]}</Text>}
                          />
                        )}
                      </div>
                    </Space>
                    <div style={{ textAlign: 'right', marginTop: 8 }}>
                      <Button
                        type="primary"
                        size="small"
                        icon={<PlusOutlined />}
                        onClick={() => handleApplyTemplate(tpl)}
                        loading={applyingTemplate && selectedTemplate?.id === tpl.id}
                      >
                        套用模板
                      </Button>
                    </div>
                  </Card>
                </Col>
              ))}
            </Row>
          ) : (
            !templatesLoading && <Empty description="暂无匹配的模板" />
          )}
        </Spin>
      </Drawer>
    </>
  );
};

export default MediationRecordEditor;
