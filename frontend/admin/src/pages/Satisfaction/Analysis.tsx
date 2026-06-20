import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Statistic, Progress, Tag, Table, DatePicker, Select, Spin } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import {
  SmileOutlined,
  MehOutlined,
  FrownOutlined,
  BarChartOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import {
  satisfactionService,
  SatisfactionSentimentStats,
  IssueTypeMap,
} from '../../services/satisfaction';

const { RangePicker } = DatePicker;

const SatisfactionAnalysis: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [stats, setStats] = useState<SatisfactionSentimentStats | null>(null);
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(null);

  const fetchStats = async () => {
    setLoading(true);
    try {
      const params: any = {};
      if (dateRange) {
        params.startDate = dateRange[0].format('YYYY-MM-DD');
        params.endDate = dateRange[1].format('YYYY-MM-DD');
      }
      const response = await satisfactionService.getStats(params);
      setStats(response);
    } catch (error) {
      console.error('获取满意度分析统计失败', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStats();
  }, []);

  const issueColumns: ColumnsType<any> = [
    {
      title: '问题类型',
      dataIndex: 'issue_type',
      key: 'issue_type',
      render: (type) => IssueTypeMap[type] || type,
    },
    {
      title: '数量',
      dataIndex: 'count',
      key: 'count',
      sorter: (a, b) => a.count - b.count,
    },
  ];

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">
          <BarChartOutlined className="mr-2" />
          满意度情感分析
        </h1>
        <RangePicker
          value={dateRange}
          onChange={(dates) => {
            setDateRange(dates as any);
          }}
          style={{ width: 300 }}
        />
      </div>

      <Spin spinning={loading}>
        <Row gutter={16} className="mb-6">
          <Col span={6}>
            <Card>
              <Statistic
                title="已分析评价"
                value={stats?.totalAnalyzed || 0}
                suffix="条"
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="正面评价"
                value={stats?.positiveCount || 0}
                prefix={<SmileOutlined style={{ color: '#52c41a' }} />}
                suffix={
                  <span className="text-sm text-gray-500 ml-1">
                    ({stats?.positiveRate?.toFixed(1) || 0}%)
                  </span>
                }
                valueStyle={{ color: '#52c41a' }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="中性评价"
                value={stats?.neutralCount || 0}
                prefix={<MehOutlined style={{ color: '#faad14' }} />}
                suffix={
                  <span className="text-sm text-gray-500 ml-1">
                    ({stats?.neutralRate?.toFixed(1) || 0}%)
                  </span>
                }
                valueStyle={{ color: '#faad14' }}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="负面评价"
                value={stats?.negativeCount || 0}
                prefix={<FrownOutlined style={{ color: '#ff4d4f' }} />}
                suffix={
                  <span className="text-sm text-gray-500 ml-1">
                    ({stats?.negativeRate?.toFixed(1) || 0}%)
                  </span>
                }
                valueStyle={{ color: '#ff4d4f' }}
              />
            </Card>
          </Col>
        </Row>

        <Row gutter={16}>
          <Col span={12}>
            <Card title="情感分布">
              {stats && stats.totalAnalyzed > 0 ? (
                <div className="flex justify-center items-center" style={{ height: 250 }}>
                  <Progress
                    type="dashboard"
                    percent={Math.round(((stats.positiveRate || 0) / 100) * 100)}
                    format={() => `${(stats.positiveRate || 0).toFixed(1)}%`}
                    strokeColor="#52c41a"
                    trailColor="#f0f0f0"
                    size={200}
                  />
                </div>
              ) : (
                <div className="text-center text-gray-400 py-12">暂无数据</div>
              )}
              {stats && stats.totalAnalyzed > 0 && (
                <div className="flex justify-around mt-4">
                  <div className="text-center">
                    <Tag color="success">正面</Tag>
                    <div>{stats.positiveCount}</div>
                  </div>
                  <div className="text-center">
                    <Tag color="warning">中性</Tag>
                    <div>{stats.neutralCount}</div>
                  </div>
                  <div className="text-center">
                    <Tag color="error">负面</Tag>
                    <div>{stats.negativeCount}</div>
                  </div>
                </div>
              )}
            </Card>
          </Col>
          <Col span={12}>
            <Card title="负面评价问题类型分布">
              <Table
                columns={issueColumns}
                dataSource={stats?.issueTypeStats || []}
                rowKey="issue_type"
                pagination={false}
                size="small"
              />
            </Card>
          </Col>
        </Row>
      </Spin>
    </div>
  );
};

export default SatisfactionAnalysis;
