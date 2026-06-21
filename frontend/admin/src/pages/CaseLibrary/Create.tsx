import React from 'react';
import { useNavigate } from 'react-router-dom';
import { Form, Input, Select, InputNumber, Button, Card, Space, App } from 'antd';
import { caseLibraryService } from '../../services/caseLibrary';

const { TextArea } = Input;

const CaseLibraryCreate: React.FC = () => {
  const navigate = useNavigate();
  const { message } = App.useApp();
  const [form] = Form.useForm();

  const handleSubmit = async (values: any) => {
    try {
      await caseLibraryService.create(values);
      message.success('案例创建成功');
      navigate('/case-library');
    } catch {
      message.error('创建失败');
    }
  };

  return (
    <div style={{ maxWidth: 900, margin: '0 auto' }}>
      <Card title="新增典型案例（脱敏录入）">
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{ difficultyLevel: 2, isSuccess: 1 }}
        >
          <Form.Item name="title" label="案例标题" rules={[{ required: true, message: '请输入案例标题' }]}>
            <Input placeholder="请输入脱敏后的案例标题" maxLength={200} />
          </Form.Item>

          <Space style={{ width: '100%' }} size="large">
            <Form.Item name="disputeType" label="纠纷类型" style={{ width: 300 }}>
              <Input placeholder="如：邻里纠纷、劳资纠纷" maxLength={128} />
            </Form.Item>
            <Form.Item name="difficultyLevel" label="难度等级" style={{ width: 200 }}>
              <Select options={[
                { value: 1, label: '简单' },
                { value: 2, label: '一般' },
                { value: 3, label: '复杂' },
                { value: 4, label: '疑难' },
              ]} />
            </Form.Item>
            <Form.Item name="isSuccess" label="调解结果" style={{ width: 200 }}>
              <Select options={[
                { value: 1, label: '调解成功' },
                { value: 0, label: '调解未成' },
              ]} />
            </Form.Item>
          </Space>

          <Form.Item name="description" label="案例描述（脱敏）">
            <TextArea rows={4} placeholder="请输入脱敏后的案例描述，注意隐去当事人真实姓名、身份证号等敏感信息" maxLength={5000} showCount />
          </Form.Item>

          <Form.Item name="mediationTactics" label="调解话术/策略">
            <TextArea rows={4} placeholder="记录调解过程中使用的话术和策略，支持一键引用" maxLength={5000} showCount />
          </Form.Item>

          <Form.Item name="keyPoints" label="调解要点/关键经验">
            <TextArea rows={4} placeholder="记录调解的关键经验和要点" maxLength={5000} showCount />
          </Form.Item>

          <Form.Item name="resultSummary" label="调解结果摘要">
            <TextArea rows={3} placeholder="请输入调解结果摘要" maxLength={2000} showCount />
          </Form.Item>

          <Space style={{ width: '100%' }} size="large">
            <Form.Item name="mediatorName" label="调解员（脱敏）">
              <Input placeholder="脱敏后的调解员姓名" maxLength={64} />
            </Form.Item>
            <Form.Item name="orgName" label="调解组织（脱敏）">
              <Input placeholder="脱敏后的组织名称" maxLength={128} />
            </Form.Item>
          </Space>

          <Form.Item name="keywords" label="关键词标签">
            <Input placeholder="多个关键词用逗号分隔，如：噪音扰民,邻里,物业" maxLength={500} />
          </Form.Item>

          <Form.Item name="tags" label="分类标签">
            <Input placeholder="分类标签，如：典型案例,调解成功" maxLength={500} />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">创建案例</Button>
              <Button onClick={() => navigate('/case-library')}>取消</Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default CaseLibraryCreate;
