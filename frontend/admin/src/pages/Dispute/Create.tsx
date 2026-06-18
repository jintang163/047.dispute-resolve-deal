import React, { useState } from 'react';
import { Card, Row, Col, App } from 'antd';
import { ArrowLeftOutlined, SaveOutlined } from '@ant-design/icons';
import {
  ProForm,
  ProFormText,
  ProFormTextArea,
  ProFormSelect,
  ProFormGroup,
  ProFormDigit,
  FooterToolbar,
  ProCard,
  ModalForm,
  DrawerForm,
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

const DisputeCreate: React.FC = () => {
  const navigate = useNavigate();
  const { message } = App.useApp();
  const [loading, setLoading] = useState(false);

  const onFinish = async (values: any) => {
    try {
      setLoading(true);
      await disputeService.create(values);
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
                fieldProps={{ size: 'large' }}
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
                placeholder="请详细描述案件情况"
                fieldProps={{
                  size: 'large',
                  rows: 4,
                  showCount: true,
                  maxLength: 2000,
                }}
                rules={[{ max: 2000, message: '描述长度不超过2000个字符' }]}
              />
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
