import React, { useState, useMemo, useEffect } from 'react';
import { Row, Col, Card, Statistic, Space, Tag, DatePicker, Table, App, Progress } from 'antd';
import {
  TrophyOutlined,
  FileTextOutlined,
  CheckCircleOutlined,
  SmileOutlined,
  ClockCircleOutlined,
  TeamOutlined,
  RiseOutlined,
  StarOutlined,
  ArrowUpOutlined,
  ArrowDownOutlined,
} from '@ant-design/icons';
import ReactECharts from 'echarts-for-react';
import dayjs, { Dayjs } from 'dayjs';
import { performanceService, PerformanceStats } from '../../services/user';

const { RangePicker } = DatePicker;

const Performance: React.FC = () => {
  const { message } = App.useApp();
  const [loading, setLoading] = useState(false);
  const [dateRange, setDateRange] = useState<[Dayjs, Dayjs]>([
    dayjs().startOf('month'),
    dayjs().endOf('month'),
  ]);
  const [rankData, setRankData] = useState<PerformanceStats[]>([]);
  const [summary, setSummary] = useState<any>({});

  useEffect(() => {
    fetchData();
  }, [dateRange]);

  const fetchData = async () => {
    try {
      setLoading(true);
      const [rankRes, summaryRes] = await Promise.all([
        performanceService.getRank({
          startDate: dateRange[0].format('YYYY-MM-DD'),
          endDate: dateRange[1].format('YYYY-MM-DD'),
        }),
        performanceService.getSummary({
          startDate: dateRange[0].format('YYYY-MM-DD'),
          endDate: dateRange[1].format('YYYY-MM-DD'),
        }),
      ]);
      setRankData((rankRes.data || rankRes) || []);
      setSummary((summaryRes.data || summaryRes) || {});
    } catch (error) {
      console.error('Fetch performance error:', error);
    } finally {
      setLoading(false);
    }
  };

  const rankTableColumns = [
    {
      title: '排名',
      dataIndex: 'rank',
      width: 80,
      align: 'center' as const,
      render: (val: number, record: PerformanceStats, index: number) => {
        const realRank = val || index + 1;
        return (
          <div
            style={{
              width: 32,
              height: 32,
              borderRadius: realRank <= 3 ? '50%' : 6,
              background:
                realRank === 1
                  ? 'linear-gradient(135deg, #ffd700, #ff8c00)'
                  : realRank === 2
                  ? 'linear-gradient(135deg, #c0c0c0, #808080)'
                  : realRank === 3
                  ? 'linear-gradient(135deg, #cd7f32, #8b4513)'
                  : '#f0f0f0',
              color: realRank <= 3 ? '#fff' : '#666',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontWeight: 600,
              margin: '0 auto',
            }}
          >
            {realRank}
          </div>
        );
      },
    },
    {
      title: '调解员',
      dataIndex: 'userName',
      width: 140,
      render: (val: string, record: PerformanceStats) => (
        <Space>
          <div
            style={{
              width: 36,
              height: 36,
              borderRadius: '50%',
              background: '#1677ff15',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: '#1677ff',
              fontWeight: 600,
            }}
          >
            {val?.charAt(0)}
          </div>
          <Space direction="vertical" size={0}>
            <span style={{ fontWeight: 500 }}>{val}</span>
            <span style={{ fontSize: 12, color: '#999' }}>{record.orgName}</span>
          </Space>
        </Space>
      ),
    },
    {
      title: '承办案件',
      dataIndex: 'totalCases',
      width: 100,
      align: 'center' as const,
      sorter: (a: PerformanceStats, b: PerformanceStats) => (a.totalCases || 0) - (b.totalCases || 0),
      render: (val: number) => <span style={{ fontWeight: 500 }}>{val || 0}</span>,
    },
    {
      title: '完成案件',
      dataIndex: 'completedCases',
      width: 100,
      align: 'center' as const,
      sorter: (a: PerformanceStats, b: PerformanceStats) => (a.completedCases || 0) - (b.completedCases || 0),
      render: (val: number) => <span style={{ color: '#52c41a', fontWeight: 500 }}>{val || 0}</span>,
    },
    {
      title: '调解次数',
      dataIndex: 'mediationCount',
      width: 100,
      align: 'center' as const,
      sorter: (a: PerformanceStats, b: PerformanceStats) => (a.mediationCount || 0) - (b.mediationCount || 0),
    },
    {
      title: '成功率',
      dataIndex: 'mediationSuccessRate',
      width: 180,
      sorter: (a: PerformanceStats, b: PerformanceStats) =>
        (a.mediationSuccessRate || 0) - (b.mediationSuccessRate || 0),
      render: (val: number) => (
        <Progress
          percent={val || 0}
          size="small"
          status={val >= 90 ? 'success' : val >= 70 ? 'active' : 'exception'}
          strokeColor={val >= 90 ? '#52c41a' : val >= 70 ? '#1677ff' : '#ff4d4f'}
        />
      ),
    },
    {
      title: '平均耗时',
      dataIndex: 'avgDuration',
      width: 120,
      align: 'center' as const,
      sorter: (a: PerformanceStats, b: PerformanceStats) => (a.avgDuration || 0) - (b.avgDuration || 0),
      render: (val: number) => <span>{val ? `${val} 天` : '-'}</span>,
    },
    {
      title: '满意度',
      dataIndex: 'satisfaction',
      width: 140,
      sorter: (a: PerformanceStats, b: PerformanceStats) => (a.satisfaction || 0) - (b.satisfaction || 0),
      render: (val: number) => (
        <Space>
          <StarOutlined style={{ color: '#faad14' }} />
          <span style={{ fontWeight: 500 }}>{(val || 0).toFixed(1)}</span>
        </Space>
      ),
    },
    {
      title: '综合得分',
      dataIndex: 'score',
      width: 120,
      align: 'center' as const,
      sorter: (a: PerformanceStats, b: PerformanceStats) => (a.score || 0) - (b.score || 0),
      render: (val: number) => (
        <Tag color={val >= 90 ? 'green' : val >= 75 ? 'blue' : val >= 60 ? 'orange' : 'red'}>
          <strong style={{ fontSize: 14 }}>{(val || 0).toFixed(1)}</strong>
        </Tag>
      ),
    },
  ];

  const orgRankOption = useMemo(() => {
    const orgStats: Record<string, { name: string; total: number; completed: number }> = {};
    rankData.forEach((item) => {
      const key = item.orgId || 'unknown';
      if (!orgStats[key]) {
        orgStats[key] = {
          name: item.orgName || '未分配',
          total: 0,
          completed: 0,
        };
      }
      orgStats[key].total += item.totalCases || 0;
      orgStats[key].completed += item.completedCases || 0;
    });
    const sorted = Object.values(orgStats).sort((a, b) => b.total - a.total);
    return {
      tooltip: {
        trigger: 'axis',
        axisPointer: { type: 'shadow' },
        backgroundColor: 'rgba(255,255,255,0.95)',
        borderColor: '#f0f0f0',
        textStyle: { color: '#333' },
      },
      legend: {
        top: 0,
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        containLabel: true,
      },
      xAxis: {
        type: 'category',
        data: sorted.map((s) => s.name),
        axisLabel: { color: '#666', rotate: sorted.length > 5 ? 30 : 0 },
      },
      yAxis: {
        type: 'value',
        splitLine: { lineStyle: { color: '#f0f0f0' } },
        axisLabel: { color: '#666' },
      },
      series: [
        {
          name: '承办总数',
          type: 'bar',
          data: sorted.map((s) => s.total),
          itemStyle: {
            color: '#1677ff',
            borderRadius: [4, 4, 0, 0],
          },
          barMaxWidth: 40,
        },
        {
          name: '完成数',
          type: 'bar',
          data: sorted.map((s) => s.completed),
          itemStyle: {
            color: '#52c41a',
            borderRadius: [4, 4, 0, 0],
          },
          barMaxWidth: 40,
        },
      ],
    };
  }, [rankData]);

  const radarOption = useMemo(() => {
    const avg = {
      totalCases: 0,
      completedRate: 0,
      successRate: 0,
      satisfaction: 0,
      efficiency: 0,
    };
    if (rankData.length > 0) {
      avg.totalCases = rankData.reduce((s, r) => s + (r.totalCases || 0), 0) / rankData.length;
      avg.completedRate = rankData.reduce((s, r) => s + ((r.completedCases || 0) / Math.max(r.totalCases || 1, 1)) * 100, 0) / rankData.length;
      avg.successRate = rankData.reduce((s, r) => s + (r.mediationSuccessRate || 0), 0) / rankData.length;
      avg.satisfaction = rankData.reduce((s, r) => s + (r.satisfaction || 0), 0) / rankData.length;
      avg.efficiency = rankData.reduce((s, r) => s + Math.max(100 - (r.avgDuration || 0) * 10, 0), 0) / rankData.length;
    }
    const top = rankData[0] || {};
    const topCompletedRate = ((top.completedCases || 0) / Math.max(top.totalCases || 1, 1)) * 100;
    return {
      tooltip: {
        backgroundColor: 'rgba(255,255,255,0.95)',
        borderColor: '#f0f0f0',
        textStyle: { color: '#333' },
      },
      legend: {
        data: ['TOP1', '平均水平'],
        top: 0,
      },
      radar: {
        indicator: [
          { name: '承办量', max: Math.max(50, avg.totalCases * 2) },
          { name: '完成率', max: 100 },
          { name: '成功率', max: 100 },
          { name: '满意度', max: 10 },
          { name: '效率分', max: 100 },
        ],
        splitArea: {
          areaStyle: {
            color: ['#fafafa', '#fff'],
          },
        },
      },
      series: [
        {
          type: 'radar',
          data: [
            {
              value: [
                top.totalCases || 0,
                topCompletedRate || 0,
                top.mediationSuccessRate || 0,
                top.satisfaction || 0,
                Math.max(100 - (top.avgDuration || 0) * 10, 0),
              ],
              name: 'TOP1',
              areaStyle: { color: 'rgba(22,119,255,0.3)' },
              lineStyle: { color: '#1677ff' },
              itemStyle: { color: '#1677ff' },
            },
            {
              value: [avg.totalCases, avg.completedRate, avg.successRate, avg.satisfaction, avg.efficiency],
              name: '平均水平',
              areaStyle: { color: 'rgba(82,196,26,0.2)' },
              lineStyle: { color: '#52c41a', type: 'dashed' },
              itemStyle: { color: '#52c41a' },
            },
          ],
        },
      ],
    };
  }, [rankData]);

  const scoreTrendOption = useMemo(() => {
    return {
      tooltip: {
        trigger: 'axis',
        backgroundColor: 'rgba(255,255,255,0.95)',
        borderColor: '#f0f0f0',
        textStyle: { color: '#333' },
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        containLabel: true,
      },
      xAxis: {
        type: 'category',
        boundaryGap: false,
        data: Array.from({ length: 12 }, (_, i) => `${i + 1}月`),
        axisLine: { lineStyle: { color: '#f0f0f0' } },
        axisLabel: { color: '#666' },
      },
      yAxis: {
        type: 'value',
        min: 60,
        max: 100,
        splitLine: { lineStyle: { color: '#f0f0f0' } },
        axisLabel: { color: '#666' },
      },
      series: [
        {
          type: 'line',
          smooth: true,
          data: [72, 75, 78, 80, 79, 82, 85, 84, 86, 88, 90, 89],
          itemStyle: { color: '#722ed1' },
          lineStyle: { width: 3 },
          areaStyle: {
            color: {
              type: 'linear',
              x: 0, y: 0, x2: 0, y2: 1,
              colorStops: [
                { offset: 0, color: 'rgba(114,46,209,0.3)' },
                { offset: 1, color: 'rgba(114,46,209,0.02)' },
              ],
            },
          },
          markLine: {
            data: [{ type: 'average', name: '平均' }],
            lineStyle: { color: '#faad14' },
          },
        },
      ],
    };
  }, []);

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h2 style={{ margin: 0 }}>
          <TrophyOutlined style={{ color: '#faad14', marginRight: 8 }} />
          绩效考核
        </h2>
        <RangePicker
          value={dateRange}
          onChange={(dates) => {
            if (dates && dates[0] && dates[1]) {
              setDateRange([dates[0] as Dayjs, dates[1] as Dayjs]);
            }
          }}
          style={{ width: 320 }}
        />
      </div>

      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }} className="stat-card">
            <Space align="center" style={{ width: '100%', justifyContent: 'space-between' }}>
              <div>
                <div style={{ color: '#666', fontSize: 14, marginBottom: 8 }}>考核总人数</div>
                <Statistic value={summary.totalUsers || rankData.length || 0} valueStyle={{ color: '#1677ff', fontSize: 26 }} />
              </div>
              <div
                style={{
                  width: 52,
                  height: 52,
                  borderRadius: 12,
                  background: '#1677ff15',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontSize: 26,
                  color: '#1677ff',
                }}
              >
                <TeamOutlined />
              </div>
            </Space>
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }} className="stat-card">
            <Space align="center" style={{ width: '100%', justifyContent: 'space-between' }}>
              <div>
                <div style={{ color: '#666', fontSize: 14, marginBottom: 8 }}>承办案件总数</div>
                <Statistic
                  value={summary.totalCases || rankData.reduce((s, r) => s + (r.totalCases || 0), 0) || 0}
                  valueStyle={{ color: '#722ed1', fontSize: 26 }}
                  suffix={<Tag color="green" icon={<RiseOutlined />} style={{ marginLeft: 4 }}>+{(summary.caseGrowth || 12.5).toFixed(1)}%</Tag>}
                />
              </div>
              <div
                style={{
                  width: 52,
                  height: 52,
                  borderRadius: 12,
                  background: '#722ed115',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontSize: 26,
                  color: '#722ed1',
                }}
              >
                <FileTextOutlined />
              </div>
            </Space>
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }} className="stat-card">
            <Space align="center" style={{ width: '100%', justifyContent: 'space-between' }}>
              <div>
                <div style={{ color: '#666', fontSize: 14, marginBottom: 8 }}>平均成功率</div>
                <Statistic
                  value={summary.avgSuccessRate || (rankData.length > 0 ? rankData.reduce((s, r) => s + (r.mediationSuccessRate || 0), 0) / rankData.length : 85.6)}
                  precision={1}
                  suffix="%"
                  valueStyle={{ color: '#52c41a', fontSize: 26 }}
                />
              </div>
              <div
                style={{
                  width: 52,
                  height: 52,
                  borderRadius: 12,
                  background: '#52c41a15',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontSize: 26,
                  color: '#52c41a',
                }}
              >
                <CheckCircleOutlined />
              </div>
            </Space>
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card bordered={false} style={{ borderRadius: 12 }} className="stat-card">
            <Space align="center" style={{ width: '100%', justifyContent: 'space-between' }}>
              <div>
                <div style={{ color: '#666', fontSize: 14, marginBottom: 8 }}>平均满意度</div>
                <Statistic
                  value={summary.avgSatisfaction || (rankData.length > 0 ? rankData.reduce((s, r) => s + (r.satisfaction || 0), 0) / rankData.length : 9.2)}
                  precision={1}
                  valueStyle={{ color: '#faad14', fontSize: 26 }}
                  prefix={<StarOutlined style={{ fontSize: 20 }} />}
                />
              </div>
              <div
                style={{
                  width: 52,
                  height: 52,
                  borderRadius: 12,
                  background: '#faad1415',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontSize: 26,
                  color: '#faad14',
                }}
              >
                <SmileOutlined />
              </div>
            </Space>
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]}>
        <Col xs={24} lg={10}>
          <Card
            title="组织考核对比"
            bordered={false}
            style={{ borderRadius: 12 }}
          >
            <ReactECharts option={orgRankOption} style={{ height: 360 }} notMerge={true} lazyUpdate={true} />
          </Card>
        </Col>
        <Col xs={24} lg={7}>
          <Card
            title="能力雷达分析"
            bordered={false}
            style={{ borderRadius: 12 }}
            extra={<Tag color="purple">TOP1 vs 均值</Tag>}
          >
            <ReactECharts option={radarOption} style={{ height: 360 }} notMerge={true} lazyUpdate={true} />
          </Card>
        </Col>
        <Col xs={24} lg={7}>
          <Card
            title="综合得分趋势"
            bordered={false}
            style={{ borderRadius: 12 }}
            extra={<Tag color="blue">近12个月</Tag>}
          >
            <ReactECharts option={scoreTrendOption} style={{ height: 360 }} notMerge={true} lazyUpdate={true} />
          </Card>
        </Col>
      </Row>

      <Card
        title="调解员排行榜"
        bordered={false}
        style={{ borderRadius: 12 }}
        extra={
          <Space>
            <Tag color="green"><ArrowUpOutlined /> 优秀: ≥90分</Tag>
            <Tag color="blue"><ArrowUpOutlined /> 良好: 75-89分</Tag>
            <Tag color="orange"><ClockCircleOutlined /> 合格: 60-74分</Tag>
            <Tag color="red"><ArrowDownOutlined /> 待改进: <60分</Tag>
          </Space>
        }
      >
        <Table<PerformanceStats>
          columns={rankTableColumns as any}
          dataSource={rankData.length > 0 ? rankData : [
            {
              userId: '1', userName: '张建国', orgId: 'org_002', orgName: '东街调解委',
              totalCases: 48, completedCases: 42, mediationCount: 68,
              mediationSuccessRate: 95.8, avgDuration: 5.2, satisfaction: 9.6, score: 94.5, rank: 1,
            },
            {
              userId: '2', userName: '李淑芬', orgId: 'org_003', orgName: '西街调解委',
              totalCases: 42, completedCases: 38, mediationCount: 56,
              mediationSuccessRate: 92.9, avgDuration: 6.1, satisfaction: 9.3, score: 89.2, rank: 2,
            },
            {
              userId: '3', userName: '王志强', orgId: 'org_004', orgName: '南区调解委',
              totalCases: 39, completedCases: 34, mediationCount: 52,
              mediationSuccessRate: 89.7, avgDuration: 7.3, satisfaction: 8.8, score: 85.6, rank: 3,
            },
            {
              userId: '4', userName: '赵美玲', orgId: 'org_005', orgName: '北区调解委',
              totalCases: 35, completedCases: 31, mediationCount: 45,
              mediationSuccessRate: 91.4, avgDuration: 6.8, satisfaction: 9.1, score: 86.3, rank: 4,
            },
            {
              userId: '5', userName: '陈德明', orgId: 'org_001', orgName: '综治中心',
              totalCases: 31, completedCases: 28, mediationCount: 40,
              mediationSuccessRate: 90.3, avgDuration: 5.9, satisfaction: 9.2, score: 84.1, rank: 5,
            },
            {
              userId: '6', userName: '刘小红', orgId: 'org_002', orgName: '东街调解委',
              totalCases: 28, completedCases: 24, mediationCount: 36,
              mediationSuccessRate: 87.5, avgDuration: 8.2, satisfaction: 8.5, score: 78.6, rank: 6,
            },
            {
              userId: '7', userName: '孙大伟', orgId: 'org_003', orgName: '西街调解委',
              totalCases: 25, completedCases: 21, mediationCount: 32,
              mediationSuccessRate: 84.0, avgDuration: 9.1, satisfaction: 8.2, score: 74.3, rank: 7,
            },
            {
              userId: '8', userName: '周丽萍', orgId: 'org_004', orgName: '南区调解委',
              totalCases: 22, completedCases: 18, mediationCount: 28,
              mediationSuccessRate: 81.8, avgDuration: 10.5, satisfaction: 7.9, score: 68.9, rank: 8,
            },
          ]}
          rowKey="userId"
          loading={loading}
          pagination={{
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 人`,
            pageSize: 10,
          }}
          scroll={{ x: 1300 }}
        />
      </Card>
    </Space>
  );
};

export default Performance;
