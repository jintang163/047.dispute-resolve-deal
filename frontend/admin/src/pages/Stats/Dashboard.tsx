import React, { useEffect, useMemo, useState } from 'react';
import { Row, Col, Card, Statistic, Space, Tag, DatePicker, Spin } from 'antd';
import {
  FileTextOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  RiseOutlined,
  FallOutlined,
  TeamOutlined,
  TrophyOutlined,
} from '@ant-design/icons';
import ReactECharts from 'echarts-for-react';
import dayjs, { Dayjs } from 'dayjs';

const { RangePicker } = DatePicker;

interface StatCardProps {
  title: string;
  value: number | string;
  prefix?: React.ReactNode;
  suffix?: string;
  trend?: number;
  color?: string;
}

const StatCard: React.FC<StatCardProps> = ({ title, value, prefix, suffix, trend, color }) => {
  return (
    <Card className="stat-card" bordered={false} style={{ borderRadius: 12 }}>
      <Space align="center" style={{ width: '100%', justifyContent: 'space-between' }}>
        <div>
          <div style={{ color: '#666', fontSize: 14, marginBottom: 8 }}>{title}</div>
          <Statistic
            value={value}
            suffix={suffix}
            valueStyle={{ color: color || '#1677ff', fontSize: 28, fontWeight: 600 }}
          />
          {trend !== undefined && (
            <div style={{ marginTop: 8, fontSize: 12 }}>
              {trend >= 0 ? (
                <Tag color="green" icon={<RiseOutlined />}>
                  较上周 {trend.toFixed(1)}%
                </Tag>
              ) : (
                <Tag color="red" icon={<FallOutlined />}>
                  较上周 {Math.abs(trend).toFixed(1)}%
                </Tag>
              )}
            </div>
          )}
        </div>
        <div
          style={{
            width: 56,
            height: 56,
            borderRadius: 12,
            background: `${color || '#1677ff'}15`,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontSize: 28,
            color: color || '#1677ff',
          }}
        >
          {prefix}
        </div>
      </Space>
    </Card>
  );
};

const Dashboard: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [dateRange, setDateRange] = useState<[Dayjs, Dayjs]>([dayjs().subtract(30, 'day'), dayjs()]);

  const trendOption = useMemo(
    () => ({
      tooltip: {
        trigger: 'axis',
        backgroundColor: 'rgba(255,255,255,0.95)',
        borderColor: '#f0f0f0',
        textStyle: { color: '#333' },
      },
      legend: {
        data: ['新增案件', '已结案', '调解成功'],
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
        boundaryGap: false,
        data: Array.from({ length: 15 }, (_, i) =>
          dayjs().subtract(14 - i, 'day').format('MM-DD'),
        ),
        axisLine: { lineStyle: { color: '#f0f0f0' } },
        axisLabel: { color: '#666' },
      },
      yAxis: {
        type: 'value',
        splitLine: { lineStyle: { color: '#f0f0f0' } },
        axisLabel: { color: '#666' },
      },
      series: [
        {
          name: '新增案件',
          type: 'line',
          smooth: true,
          symbol: 'circle',
          symbolSize: 6,
          data: [12, 15, 9, 18, 22, 16, 14, 20, 25, 19, 23, 17, 21, 28, 24],
          itemStyle: { color: '#1677ff' },
          lineStyle: { width: 3 },
          areaStyle: {
            color: {
              type: 'linear',
              x: 0, y: 0, x2: 0, y2: 1,
              colorStops: [
                { offset: 0, color: 'rgba(22,119,255,0.3)' },
                { offset: 1, color: 'rgba(22,119,255,0.02)' },
              ],
            },
          },
        },
        {
          name: '已结案',
          type: 'line',
          smooth: true,
          symbol: 'circle',
          symbolSize: 6,
          data: [8, 10, 7, 14, 18, 13, 11, 16, 20, 15, 19, 14, 17, 22, 20],
          itemStyle: { color: '#52c41a' },
          lineStyle: { width: 3 },
          areaStyle: {
            color: {
              type: 'linear',
              x: 0, y: 0, x2: 0, y2: 1,
              colorStops: [
                { offset: 0, color: 'rgba(82,196,26,0.3)' },
                { offset: 1, color: 'rgba(82,196,26,0.02)' },
              ],
            },
          },
        },
        {
          name: '调解成功',
          type: 'line',
          smooth: true,
          symbol: 'circle',
          symbolSize: 6,
          data: [6, 8, 5, 11, 15, 10, 9, 13, 17, 12, 16, 11, 14, 19, 17],
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
        },
      ],
    }),
    [dateRange],
  );

  const pieOption = useMemo(
    () => ({
      tooltip: {
        trigger: 'item',
        backgroundColor: 'rgba(255,255,255,0.95)',
        borderColor: '#f0f0f0',
        textStyle: { color: '#333' },
        formatter: '{b}: {c} ({d}%)',
      },
      legend: {
        orient: 'vertical',
        right: '5%',
        top: 'center',
      },
      series: [
        {
          name: '案件类型分布',
          type: 'pie',
          radius: ['45%', '70%'],
          center: ['35%', '50%'],
          avoidLabelOverlap: false,
          itemStyle: {
            borderRadius: 8,
            borderColor: '#fff',
            borderWidth: 2,
          },
          label: {
            show: false,
            position: 'center',
          },
          emphasis: {
            label: {
              show: true,
              fontSize: 16,
              fontWeight: 'bold',
              formatter: '{b}\n{c}件',
            },
          },
          labelLine: {
            show: false,
          },
          data: [
            { value: 320, name: '民事纠纷', itemStyle: { color: '#1677ff' } },
            { value: 186, name: '劳动争议', itemStyle: { color: '#52c41a' } },
            { value: 154, name: '家庭纠纷', itemStyle: { color: '#faad14' } },
            { value: 128, name: '邻里纠纷', itemStyle: { color: '#722ed1' } },
            { value: 98, name: '合同纠纷', itemStyle: { color: '#eb2f96' } },
            { value: 76, name: '物业纠纷', itemStyle: { color: '#13c2c2' } },
            { value: 54, name: '其他纠纷', itemStyle: { color: '#8c8c8c' } },
          ],
        },
      ],
    }),
    [dateRange],
  );

  const barOption = useMemo(
    () => ({
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
        data: ['待分配', '分配中', '调解中', '待审批', '审批通过', '已完成', '已结案'],
        axisLine: { lineStyle: { color: '#f0f0f0' } },
        axisLabel: { color: '#666', interval: 0 },
      },
      yAxis: {
        type: 'value',
        splitLine: { lineStyle: { color: '#f0f0f0' } },
        axisLabel: { color: '#666' },
      },
      series: [
        {
          name: '本月',
          type: 'bar',
          barWidth: '30%',
          data: [28, 16, 42, 18, 156, 89, 245],
          itemStyle: {
            color: '#1677ff',
            borderRadius: [4, 4, 0, 0],
          },
        },
        {
          name: '上月',
          type: 'bar',
          barWidth: '30%',
          data: [24, 18, 38, 15, 142, 76, 228],
          itemStyle: {
            color: '#91caff',
            borderRadius: [4, 4, 0, 0],
          },
        },
      ],
    }),
    [dateRange],
  );

  const heatmapOption = useMemo(
    () => {
      const weeks = ['周一', '周二', '周三', '周四', '周五', '周六', '周日'];
      const hours = ['00', '02', '04', '06', '08', '10', '12', '14', '16', '18', '20', '22'];
      const data: number[][] = [];
      for (let i = 0; i < weeks.length; i++) {
        for (let j = 0; j < hours.length; j++) {
          const baseVal = j >= 4 && j <= 10 ? Math.floor(Math.random() * 10 + 5) : Math.floor(Math.random() * 5);
          data.push([j, i, baseVal]);
        }
      }
      return {
        tooltip: {
          position: 'top',
          backgroundColor: 'rgba(255,255,255,0.95)',
          borderColor: '#f0f0f0',
          textStyle: { color: '#333' },
          formatter: (params: any) => `${weeks[params.value[1]]} ${hours[params.value[0]]}:00 - 案件量: ${params.value[2]}`,
        },
        grid: {
          left: '8%',
          right: '10%',
          top: '8%',
          bottom: '15%',
        },
        xAxis: {
          type: 'category',
          data: hours,
          splitArea: { show: true },
          axisLine: { lineStyle: { color: '#f0f0f0' } },
          axisLabel: { color: '#666' },
        },
        yAxis: {
          type: 'category',
          data: weeks,
          splitArea: { show: true },
          axisLine: { lineStyle: { color: '#f0f0f0' } },
          axisLabel: { color: '#666' },
        },
        visualMap: {
          min: 0,
          max: 15,
          calculable: true,
          orient: 'horizontal',
          left: 'center',
          bottom: '0%',
          inRange: {
            color: ['#f0f7ff', '#91caff', '#1677ff', '#003a8c'],
          },
        },
        series: [
          {
            name: '案件分布',
            type: 'heatmap',
            data,
            label: { show: false },
            emphasis: {
              itemStyle: {
                shadowBlur: 10,
                shadowColor: 'rgba(0, 0, 0, 0.3)',
              },
            },
          },
        ],
      };
    },
    [dateRange],
  );

  return (
    <Spin spinning={loading}>
      <Space direction="vertical" size={16} style={{ width: '100%' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <h2 style={{ margin: 0 }}>数据概览</h2>
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
            <StatCard
              title="总案件数"
              value={1286}
              prefix={<FileTextOutlined />}
              trend={8.5}
              color="#1677ff"
            />
          </Col>
          <Col xs={24} sm={12} md={6}>
            <StatCard
              title="已结案"
              value={867}
              prefix={<CheckCircleOutlined />}
              trend={12.3}
              color="#52c41a"
            />
          </Col>
          <Col xs={24} sm={12} md={6}>
            <StatCard
              title="处理中"
              value={245}
              prefix={<ClockCircleOutlined />}
              trend={-3.2}
              color="#faad14"
            />
          </Col>
          <Col xs={24} sm={12} md={6}>
            <StatCard
              title="调解成功率"
              value={89.6}
              suffix="%"
              prefix={<TrophyOutlined />}
              trend={2.1}
              color="#722ed1"
            />
          </Col>
        </Row>

        <Row gutter={[16, 16]}>
          <Col xs={24} lg={16}>
            <Card
              title="案件趋势分析"
              bordered={false}
              style={{ borderRadius: 12 }}
              extra={<Tag color="blue">近15天</Tag>}
            >
              <ReactECharts option={trendOption} style={{ height: 360 }} notMerge={true} lazyUpdate={true} />
            </Card>
          </Col>
          <Col xs={24} lg={8}>
            <Card title="案件类型分布" bordered={false} style={{ borderRadius: 12 }}>
              <ReactECharts option={pieOption} style={{ height: 360 }} notMerge={true} lazyUpdate={true} />
            </Card>
          </Col>
        </Row>

        <Row gutter={[16, 16]}>
          <Col xs={24} lg={12}>
            <Card title="案件状态统计" bordered={false} style={{ borderRadius: 12 }}>
              <ReactECharts option={barOption} style={{ height: 360 }} notMerge={true} lazyUpdate={true} />
            </Card>
          </Col>
          <Col xs={24} lg={12}>
            <Card
              title="案件时间分布热力图"
              bordered={false}
              style={{ borderRadius: 12 }}
              extra={<Tag color="purple">按周×时段</Tag>}
            >
              <ReactECharts option={heatmapOption} style={{ height: 360 }} notMerge={true} lazyUpdate={true} />
            </Card>
          </Col>
        </Row>

        <Row gutter={[16, 16]}>
          <Col xs={24} lg={12}>
            <Card
              title="调解员排行 TOP 5"
              bordered={false}
              style={{ borderRadius: 12 }}
              extra={<Tag color="green">本月</Tag>}
            >
              <div style={{ padding: '8px 0' }}>
                {[
                  { rank: 1, name: '张建国', cases: 48, success: 95.8, org: '东街调解委' },
                  { rank: 2, name: '李淑芬', cases: 42, success: 92.9, org: '西街调解委' },
                  { rank: 3, name: '王志强', cases: 39, success: 89.7, org: '南区调解委' },
                  { rank: 4, name: '赵美玲', cases: 35, success: 91.4, org: '北区调解委' },
                  { rank: 5, name: '陈德明', cases: 31, success: 90.3, org: '综治中心' },
                ].map((item, index) => (
                  <div
                    key={index}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      padding: '12px 0',
                      borderBottom: index < 4 ? '1px solid #f0f0f0' : 'none',
                    }}
                  >
                    <div
                      style={{
                        width: 28,
                        height: 28,
                        borderRadius: item.rank <= 3 ? '50%' : 6,
                        background: item.rank === 1 ? '#faad14' : item.rank === 2 ? '#bfbfbf' : item.rank === 3 ? '#d48806' : '#f0f0f0',
                        color: item.rank <= 3 ? '#fff' : '#666',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        fontWeight: 600,
                        marginRight: 12,
                      }}
                    >
                      {item.rank}
                    </div>
                    <div style={{ flex: 1 }}>
                      <div style={{ fontWeight: 500, marginBottom: 4 }}>{item.name}</div>
                      <div style={{ fontSize: 12, color: '#999' }}>
                        {item.org}
                      </div>
                    </div>
                    <div style={{ textAlign: 'right' }}>
                      <div style={{ fontWeight: 600, color: '#1677ff' }}>{item.cases} 件</div>
                      <div style={{ fontSize: 12, color: '#52c41a' }}>
                        成功率 {item.success}%
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </Card>
          </Col>
          <Col xs={24} lg={12}>
            <Card
              title="组织案件统计"
              bordered={false}
              style={{ borderRadius: 12 }}
              extra={<Tag color="orange">本月</Tag>}
            >
              <ReactECharts
                option={{
                  tooltip: {
                    trigger: 'axis',
                    axisPointer: { type: 'shadow' },
                  },
                  grid: {
                    left: '3%',
                    right: '4%',
                    bottom: '3%',
                    containLabel: true,
                  },
                  xAxis: {
                    type: 'value',
                    splitLine: { lineStyle: { color: '#f0f0f0' } },
                  },
                  yAxis: {
                    type: 'category',
                    data: ['综治中心', '东街街道', '西街街道', '南区调解委', '北区调解委', '园区调委会'],
                    axisLine: { lineStyle: { color: '#f0f0f0' } },
                  },
                  series: [
                    {
                      type: 'bar',
                      data: [286, 198, 176, 245, 212, 169],
                      itemStyle: {
                        color: {
                          type: 'linear',
                          x: 0, y: 0, x2: 1, y2: 0,
                          colorStops: [
                            { offset: 0, color: '#91caff' },
                            { offset: 1, color: '#1677ff' },
                          ],
                        },
                        borderRadius: [0, 4, 4, 0],
                      },
                      label: {
                        show: true,
                        position: 'right',
                        formatter: '{c} 件',
                        color: '#666',
                      },
                    },
                  ],
                }}
                style={{ height: 340 }}
                notMerge={true}
                lazyUpdate={true}
              />
            </Card>
          </Col>
        </Row>
      </Space>
    </Spin>
  );
};

export default Dashboard;
