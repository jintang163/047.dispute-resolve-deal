import React, { useState, useEffect } from 'react';
import {
  Form,
  Input,
  InputNumber,
  DatePicker,
  Select,
  Button,
  Space,
  Card,
  Row,
  Col,
  Divider,
  message,
  Modal,
  Tag,
  Spin,
  Tooltip,
  App,
  List,
} from 'antd';
import {
  RobotOutlined,
  FileTextOutlined,
  CheckCircleOutlined,
  BulbOutlined,
  DownloadOutlined,
  DeleteOutlined,
  PlusOutlined,
  MinusCircleOutlined,
  SafetyCertificateOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import ReactMarkdown from 'react-markdown';
import { protocolService, GenerateProtocolParams, MediationProtocol } from '../../services/protocol';
import { disputeService } from '../../services/dispute';

const { TextArea } = Input;
const { Option } = Select;

interface ProtocolGeneratorProps {
  caseId: string | number;
  onClose?: () => void;
  onAdopted?: (protocol: any) => void;
}

const ProtocolGenerator: React.FC<ProtocolGeneratorProps> = ({ caseId, onClose, onAdopted }) => {
  const { message } = App.useApp();
  const [form] = Form.useForm<GenerateProtocolParams>();
  const [generating, setGenerating] = useState(false);
  const [adoptingId, setAdoptingId] = useState<string | null>(null);
  const [generatedResult, setGeneratedResult] = useState<any>(null);
  const [protocolList, setProtocolList] = useState<MediationProtocol[]>([]);
  const [caseInfo, setCaseInfo] = useState<any>(null);
  const [otherTermsList, setOtherTermsList] = useState<string[]>(['']);

  useEffect(() => {
    fetchProtocolList();
    fetchCaseInfo();
  }, [caseId]);

  const fetchCaseInfo = async () => {
    try {
      const res: any = await disputeService.getDetail(String(caseId));
      const data = res.data || res;
      setCaseInfo(data);
      if (data) {
        form.setFieldsValue({
          caseNo: data.caseNo,
          caseTitle: data.title,
          disputeType: data.typeName || data.type,
          performanceDate: dayjs().add(7, 'day'),
          signDate: dayjs(),
          protocolYear: dayjs().year(),
          compensationType: '赔偿金',
          paymentMethod: '银行转账',
        });
        if (data.applicantName) {
          form.setFieldsValue({ partyAName: data.applicantName });
        }
        if (data.respondentName) {
          form.setFieldsValue({ partyBName: data.respondentName });
        }
        if (data.applicantPhone) {
          form.setFieldsValue({ partyAPhone: data.applicantPhone });
        }
        if (data.respondentPhone) {
          form.setFieldsValue({ partyBPhone: data.respondentPhone });
        }
      }
    } catch {}
  };

  const fetchProtocolList = async () => {
    try {
      const res: any = await protocolService.list(caseId);
      setProtocolList(res.data || res || []);
    } catch {}
  };

  const handleGenerate = async () => {
    try {
      const values = await form.validateFields();
      setGenerating(true);

      const params: GenerateProtocolParams = {
        ...values,
        caseId: Number(caseId),
        performanceDate: values.performanceDate
          ? dayjs(values.performanceDate).format('YYYY-MM-DD')
          : '',
        signDate: values.signDate ? dayjs(values.signDate).format('YYYY-MM-DD') : '',
        otherTerms: otherTermsList.filter((t) => t.trim()),
      };

      const resp: any = await protocolService.generate(caseId, params);
      const data = resp.data || resp;
      setGeneratedResult(data);
      message.success('协议生成成功');
      fetchProtocolList();
    } catch (err: any) {
      message.error(err?.message || '协议生成失败');
    } finally {
      setGenerating(false);
    }
  };

  const handleAdopt = async (protocol: MediationProtocol) => {
    Modal.confirm({
      title: '确认采用该协议？',
      content: '采用后协议内容将同步至案件调解协议字段，确定继续吗？',
      okText: '确认采用',
      cancelText: '取消',
      okButtonProps: { type: 'primary' },
      onOk: async () => {
        try {
          setAdoptingId(protocol.id);
          await protocolService.adopt(caseId, protocol.id);
          message.success('协议已成功采用');
          fetchProtocolList();
          if (onAdopted) {
            onAdopted(protocol);
          }
        } catch (err: any) {
          message.error(err?.message || '采用失败');
        } finally {
          setAdoptingId(null);
        }
      },
    });
  };

  const addOtherTerm = () => {
    setOtherTermsList([...otherTermsList, '']);
  };

  const updateOtherTerm = (idx: number, val: string) => {
    const next = [...otherTermsList];
    next[idx] = val;
    setOtherTermsList(next);
  };

  const removeOtherTerm = (idx: number) => {
    if (otherTermsList.length <= 1) {
      setOtherTermsList(['']);
      return;
    }
    const next = otherTermsList.filter((_, i) => i !== idx);
    setOtherTermsList(next);
  };

  const handleDownload = (content: string, filename: string) => {
    const blob = new Blob([content], { type: 'text/markdown;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${filename || '人民调解协议书'}.md`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const renderPreview = () => {
    if (!generatedResult && protocolList.length === 0) {
      return (
        <div style={{ padding: 40, textAlign: 'center', color: '#999' }}>
          <FileTextOutlined style={{ fontSize: 48, marginBottom: 12 }} />
          <p>填写左侧关键要素后点击「AI生成协议」，或选择下方已有协议进行预览</p>
        </div>
      );
    }

    const showContent = generatedResult?.content || protocolList[0]?.content;
    const showProtocol = generatedResult || protocolList[0];
    const legalBasis = generatedResult?.legalBasis || (showProtocol as any)?.legalBasis;

    return (
      <div>
        {generatedResult && (
          <div style={{ marginBottom: 12 }}>
            <Tag color="green" icon={<RobotOutlined />}>
              本次生成：{generatedResult.protocolNo}
            </Tag>
            <span style={{ marginLeft: 8, color: '#999', fontSize: 12 }}>
              生成时间：{generatedResult.generatedAt}
            </span>
          </div>
        )}

        {legalBasis && legalBasis.length > 0 && (
          <Card size="small" style={{ marginBottom: 12 }} type="inner">
            <div style={{ fontSize: 12, color: '#666', marginBottom: 6 }}>
              <SafetyCertificateOutlined /> 引用法律依据
            </div>
            {legalBasis.map((lb: string, i: number) => (
              <Tag key={i} color="blue" style={{ marginBottom: 4 }}>
                {lb}
              </Tag>
            ))}
          </Card>
        )}

        <div
          style={{
            background: '#fff',
            padding: 20,
            border: '1px solid #eee',
            borderRadius: 6,
            maxHeight: 600,
            overflowY: 'auto',
          }}
        >
          <div style={{ maxWidth: 800, margin: '0 auto' }}>
            <ReactMarkdown
              components={{
                h1: ({ children }) => (
                  <h1 style={{ textAlign: 'center', fontSize: 22, marginBottom: 20 }}>
                    {children}
                  </h1>
                ),
                h2: ({ children }) => (
                  <h2 style={{ fontSize: 16, marginTop: 20, borderBottom: '1px solid #eee', paddingBottom: 6 }}>
                    {children}
                  </h2>
                ),
                h3: ({ children }) => (
                  <h3 style={{ fontSize: 14, marginTop: 14 }}>{children}</h3>
                ),
                p: ({ children }) => (
                  <p style={{ lineHeight: 1.8, margin: '6px 0', fontSize: 14 }}>{children}</p>
                ),
                ul: ({ children }) => <ul style={{ paddingLeft: 20 }}>{children}</ul>,
                ol: ({ children }) => <ol style={{ paddingLeft: 20 }}>{children}</ol>,
                hr: () => <hr style={{ margin: '16px 0', borderColor: '#f0f0f0' }} />,
                strong: ({ children }) => <strong>{children}</strong>,
              }}
            >
              {showContent}
            </ReactMarkdown>
          </div>
        </div>

        {generatedResult && (
          <Space style={{ marginTop: 16 }}>
            <Button type="primary" icon={<CheckCircleOutlined />} onClick={() => handleAdopt(generatedResult)}>
              采用本协议
            </Button>
            <Button
              icon={<DownloadOutlined />}
              onClick={() => handleDownload(generatedResult.content, generatedResult.protocolNo)}
            >
              下载Markdown
            </Button>
          </Space>
        )}
      </div>
    );
  };

  return (
    <div style={{ minHeight: 600 }}>
      <Row gutter={24}>
        <Col span={10}>
          <Card
            title={
              <Space>
                <BulbOutlined style={{ color: '#faad14' }} />
                <span>关键要素（AI自动生成完整协议）</span>
              </Space>
            }
            extra={
              <Tooltip title="只需填写关键信息，AI自动生成包含法律依据和权利义务条款的完整协议书">
                <Tag color="purple">AI 辅助</Tag>
              </Tooltip>
            }
          >
            <Form form={form} layout="vertical">
              <Divider orientation="left" plain style={{ fontSize: 13 }}>
                双方当事人
              </Divider>
              <Row gutter={12}>
                <Col span={12}>
                  <Form.Item
                    label="甲方（申请人）姓名"
                    name="partyAName"
                    rules={[{ required: true, message: '请输入甲方姓名' }]}
                  >
                    <Input placeholder="例如：张三" />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item label="甲方性别" name="partyAGender">
                    <Select placeholder="选择性别">
                      <Option value="男">男</Option>
                      <Option value="女">女</Option>
                    </Select>
                  </Form.Item>
                </Col>
              </Row>
              <Row gutter={12}>
                <Col span={14}>
                  <Form.Item label="甲方身份证号" name="partyAIDCard">
                    <Input placeholder="选填，建议填写" />
                  </Form.Item>
                </Col>
                <Col span={10}>
                  <Form.Item label="甲方电话" name="partyAPhone">
                    <Input placeholder="联系电话" />
                  </Form.Item>
                </Col>
              </Row>
              <Form.Item label="甲方住址" name="partyAAddress">
                <Input placeholder="选填" />
              </Form.Item>

              <Divider style={{ margin: '8px 0' }} />

              <Row gutter={12}>
                <Col span={12}>
                  <Form.Item
                    label="乙方（被申请人）姓名"
                    name="partyBName"
                    rules={[{ required: true, message: '请输入乙方姓名' }]}
                  >
                    <Input placeholder="例如：李四" />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item label="乙方性别" name="partyBGender">
                    <Select placeholder="选择性别">
                      <Option value="男">男</Option>
                      <Option value="女">女</Option>
                    </Select>
                  </Form.Item>
                </Col>
              </Row>
              <Row gutter={12}>
                <Col span={14}>
                  <Form.Item label="乙方身份证号" name="partyBIDCard">
                    <Input placeholder="选填" />
                  </Form.Item>
                </Col>
                <Col span={10}>
                  <Form.Item label="乙方电话" name="partyBPhone">
                    <Input placeholder="联系电话" />
                  </Form.Item>
                </Col>
              </Row>
              <Form.Item label="乙方住址" name="partyBAddress">
                <Input placeholder="选填" />
              </Form.Item>

              <Divider orientation="left" plain style={{ fontSize: 13 }}>
                纠纷与责任
              </Divider>

              <Form.Item
                label="纠纷简要情况"
                name="disputeSummary"
                rules={[{ required: true, message: '请简要描述纠纷情况' }]}
              >
                <TextArea rows={3} placeholder="描述纠纷发生时间、地点、原因、经过、双方主张等" />
              </Form.Item>

              <Row gutter={12}>
                <Col span={12}>
                  <Form.Item label="主要责任方" name="liabilityParty">
                    <Select placeholder="选择主要责任方">
                      <Option value={form.getFieldValue('partyAName') || '甲方'}>甲方</Option>
                      <Option value={form.getFieldValue('partyBName') || '乙方'}>乙方</Option>
                      <Option value="双方">双方均有责任</Option>
                    </Select>
                  </Form.Item>
                </Col>
                <Col span={6}>
                  <Form.Item label="甲方责任%" name="liabilityRatioA" initialValue={50}>
                    <InputNumber min={0} max={100} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
                <Col span={6}>
                  <Form.Item label="乙方责任%" name="liabilityRatioB" initialValue={50}>
                    <InputNumber min={0} max={100} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
              </Row>

              <Form.Item label="责任划分理由" name="liabilityReason">
                <TextArea rows={2} placeholder="说明责任划分的理由" />
              </Form.Item>

              <Divider orientation="left" plain style={{ fontSize: 13 }}>
                赔偿与履行
              </Divider>

              <Row gutter={12}>
                <Col span={12}>
                  <Form.Item
                    label="赔偿/补偿金额（元）"
                    name="compensationAmount"
                    rules={[{ required: true, message: '请输入金额' }]}
                  >
                    <InputNumber min={0} style={{ width: '100%' }} placeholder="例如：5000" />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item label="款项类型" name="compensationType">
                    <Select>
                      <Option value="赔偿金">赔偿金</Option>
                      <Option value="补偿金">补偿金</Option>
                      <Option value="医疗费">医疗费</Option>
                      <Option value="误工费">误工费</Option>
                      <Option value="财产损失赔偿">财产损失赔偿</Option>
                      <Option value="违约金">违约金</Option>
                      <Option value="其他">其他</Option>
                    </Select>
                  </Form.Item>
                </Col>
              </Row>

              <Row gutter={12}>
                <Col span={12}>
                  <Form.Item label="支付方式" name="paymentMethod">
                    <Select>
                      <Option value="银行转账">银行转账</Option>
                      <Option value="现金">现金</Option>
                      <Option value="微信支付">微信支付</Option>
                      <Option value="支付宝">支付宝</Option>
                      <Option value="分期支付">分期支付</Option>
                    </Select>
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    label="履行期限（截止日）"
                    name="performanceDate"
                    rules={[{ required: true, message: '请选择履行期限' }]}
                  >
                    <DatePicker style={{ width: '100%' }} placeholder="选择日期" />
                  </Form.Item>
                </Col>
              </Row>

              <Form.Item label="支付账户/信息" name="paymentAccount">
                <Input placeholder="收款方账号/卡号等，选填" />
              </Form.Item>

              <Divider orientation="left" plain style={{ fontSize: 13 }}>
                其他约定（可选）
              </Divider>

              <Form.Item label="额外履行事项">
                {otherTermsList.map((term, idx) => (
                  <div key={idx} style={{ display: 'flex', gap: 8, marginBottom: 8 }}>
                    <Input
                      placeholder={`第${idx + 1}项约定内容`}
                      value={term}
                      onChange={(e) => updateOtherTerm(idx, e.target.value)}
                      style={{ flex: 1 }}
                    />
                    <Button
                      type="text"
                      danger
                      icon={<MinusCircleOutlined />}
                      onClick={() => removeOtherTerm(idx)}
                    />
                  </div>
                ))}
                <Button type="dashed" block icon={<PlusOutlined />} onClick={addOtherTerm}>
                  添加其他约定
                </Button>
              </Form.Item>

              <Form.Item label="违约责任（可选，AI自动生成）" name="breachClause">
                <TextArea rows={2} placeholder="留空将由AI自动生成规范的违约条款" />
              </Form.Item>

              <Divider orientation="left" plain style={{ fontSize: 13 }}>
                签署与编号
              </Divider>

              <Row gutter={12}>
                <Col span={12}>
                  <Form.Item label="签订地点" name="signPlace">
                    <Input placeholder="例如：XX街道人民调解委员会" />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item label="签订日期" name="signDate">
                    <DatePicker style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
              </Row>

              <Row gutter={12}>
                <Col span={8}>
                  <Form.Item label="编号前缀" name="regionPrefix">
                    <Input placeholder="例如：京、沪" />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item label="年份" name="protocolYear">
                    <InputNumber style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item label="序号（自动）" name="protocolSeq">
                    <InputNumber style={{ width: '100%' }} placeholder="留空自动递增" />
                  </Form.Item>
                </Col>
              </Row>

              <Form.Item>
                <Space direction="vertical" style={{ width: '100%' }}>
                  <Button
                    type="primary"
                    size="large"
                    block
                    icon={<RobotOutlined />}
                    loading={generating}
                    onClick={handleGenerate}
                  >
                    {generating ? 'AI正在生成协议，请稍候...' : '✨ AI生成调解协议'}
                  </Button>
                  <div style={{ textAlign: 'center', fontSize: 12, color: '#999' }}>
                    基于人民调解法、民法典等相关法规生成规范协议
                  </div>
                </Space>
              </Form.Item>
            </Form>
          </Card>
        </Col>

        <Col span={14}>
          <Card
            title={
              <Space>
                <FileTextOutlined />
                <span>协议预览</span>
                {generating && <Spin size="small" />}
              </Space>
            }
            extra={
              generatedResult && (
                <Tag color="green" icon={<CheckCircleOutlined />}>
                  生成成功
                </Tag>
              )
            }
          >
            {renderPreview()}

            {protocolList.length > 0 && (
              <div style={{ marginTop: 24 }}>
                <Divider orientation="left" plain style={{ fontSize: 13 }}>
                  历史协议（共{protocolList.length}份）
                </Divider>
                <List
                  size="small"
                  dataSource={protocolList}
                  renderItem={(item) => (
                    <List.Item
                      key={item.id}
                      actions={[
                        <Button
                          type="link"
                          size="small"
                          icon={<FileTextOutlined />}
                          onClick={() => {
                            setGeneratedResult({
                              ...item,
                              protocolNo: item.protocolNo,
                              generatedAt: item.aiGeneratedAt || item.createdAt,
                            });
                          }}
                        >
                          预览
                        </Button>,
                        <Button
                          type="link"
                          size="small"
                          icon={<DownloadOutlined />}
                          onClick={() => handleDownload(item.content, item.protocolNo)}
                        >
                          下载
                        </Button>,
                        item.isAdopted ? (
                          <Tag color="green" icon={<CheckCircleOutlined />}>
                            已采用
                          </Tag>
                        ) : (
                          <Button
                          size="small"
                          type="primary"
                            icon={<CheckCircleOutlined />}
                            loading={adoptingId === item.id}
                            onClick={() => handleAdopt(item)}
                          >
                            采用
                          </Button>
                        ),
                      ]}
                    >
                      <List.Item.Meta
                        title={
                          <Space>
                            <span>{item.title || item.protocolNo}</span>
                            {item.isAIGenerated ? (
                              <Tag color="purple" icon={<RobotOutlined />}>
                                AI生成
                              </Tag>
                            ) : (
                              <Tag>人工上传</Tag>
                            )}
                            {item.isSigned ? (
                              <Tag color="green">已签署</Tag>
                            ) : (
                              <Tag>待签署</Tag>
                            )}
                          </Space>
                        }
                        description={
                          <span style={{ fontSize: 12, color: '#999' }}>
                            {item.partyAName} ↔ {item.partyBName} · 生成时间：
                            {item.aiGeneratedAt || item.createdAt}
                          </span>
                        }
                      />
                    </List.Item>
                  )}
                />
              </div>
            )}
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default ProtocolGenerator;
