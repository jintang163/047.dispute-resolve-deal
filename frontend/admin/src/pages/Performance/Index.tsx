import React, { useState, useMemo, useEffect, useCallback } from 'react';
import {
  Row, Col, Card, Statistic, Space, Tag, DatePicker, Table, App, Progress,
  Modal, Form, InputNumber, Input, Select, Button, Tooltip, Divider, Drawer, Descriptions,
  Switch, Badge, Popover, List, Empty, Spin, Alert,
} from 'antd';
import {
  TrophyOutlined, FileTextOutlined, CheckCircleOutlined, SmileOutlined,
  ClockCircleOutlined, TeamOutlined, RiseOutlined, StarOutlined,
  ArrowUpOutlined, ArrowDownOutlined, BellOutlined, SettingOutlined,
  DownloadOutlined, EditOutlined, EyeOutlined, ReloadOutlined,
  CalculatorOutlined, CheckOutlined, DatabaseOutlined, DashboardOutlined,
} from '@ant-design/icons';
import ReactECharts from 'echarts-for-react';
import dayjs, { Dayjs } from 'dayjs';
import { performanceService, PerformanceStats, IndicatorConfig, PerformanceInterview } from '../../services/user';
import { notificationService, Notification } from '../../services/notification';
import { useUserStore } from '../../stores/user';

const { MonthPicker } = DatePicker;

const levelColorMap: Record<string, string> = {
  S: '#52c41a', A: '#1677ff', B: '#faad14', C: '#fa8c16', D: '#ff4d4f',
};

const ChangeTag: React.FC<{ value: number; suffix?: string }> = ({ value, suffix = '%' }) => {
  if (value === 0) return <Tag>-</Tag>;
  return (
    <Tag color={value > 0 ? '#f6ffed' : '#fff2f0'} style={{ margin: 0 }}>
      <span style={{ color: value > 0 ? '#52c41a' : '#ff4d4f', fontSize: 12 }}>
        {value > 0 ? <ArrowUpOutlined /> : <ArrowDownOutlined />} {Math.abs(value).toFixed(1)}{suffix}
      </span>
    </Tag>
  );
};

const Performance: React.FC = () => {
  const { message, modal } = App.useApp();
  const userInfo = useUserStore((s) => s.userInfo);
  const isAdmin = userInfo?.role === '4' || userInfo?.role === '1';
  const isLeader = userInfo?.role === '2';
  const isMediator = userInfo?.role === '3';

  const [selectedMonth, setSelectedMonth] = useState<Dayjs>(dayjs());
  const [loading, setLoading] = useState(false);
  const [dashboardData, setDashboardData] = useState<any>({ summary: {}, mediators: [] });
  const [comparisonData, setComparisonData] = useState<any>({ current: {}, previous: {}, comparison: {}, trend: [] });
  const [indicatorConfig, setIndicatorConfig] = useState<{ indicators: IndicatorConfig[]; totalWeight: number }>({ indicators: [], totalWeight: 0 });
  const [interviewList, setInterviewList] = useState<any[]>([]);
  const [interviewTotal, setInterviewTotal] = useState(0);
  const [interviewPage, setInterviewPage] = useState(1);

  const [weightModalOpen, setWeightModalOpen] = useState(false);
  const [weightForm] = Form.useForm();
  const [autoRecalculate, setAutoRecalculate] = useState(true);
  const [interviewModalOpen, setInterviewModalOpen] = useState(false);
  const [interviewForm] = Form.useForm();
  const [interviewDetailOpen, setInterviewDetailOpen] = useState(false);
  const [interviewDetail, setInterviewDetail] = useState<any>(null);
  const [selectedMediator, setSelectedMediator] = useState<PerformanceStats | null>(null);

  const [batchCalculating, setBatchCalculating] = useState(false);
  const [unreadCount, setUnreadCount] = useState(0);
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [confirmModalOpen, setConfirmModalOpen] = useState(false);
  const [confirmForm] = Form.useForm();
  const [pendingInterview, setPendingInterview] = useState<any>(null);
  const [notificationPopoverOpen, setNotificationPopoverOpen] = useState(false);

  const fetchNotifications = useCallback(async () => {
    try {
      const res = await notificationService.getMyNotifications({ page: 1, pageSize: 10, isRead: false });
      const data = (res as any)?.data || res || {};
      setNotifications(data?.list || []);
      setUnreadCount(data?.extra?.unreadCount || 0);
    } catch (error) {
      console.error('Fetch notifications error:', error);
    }
  }, []);

  const fetchUnreadCount = useCallback(async () => {
    try {
      const res = await notificationService.getUnreadCount();
      const data = (res as any)?.data || res || {};
      setUnreadCount(data?.total || 0);
    } catch (error) {
      console.error('Fetch unread count error:', error);
    }
  }, []);

  const fetchDashboard = useCallback(async () => {
    try {
      setLoading(true);
      const y = selectedMonth.year();
      const m = selectedMonth.month() + 1;
      const [dashRes, compRes] = await Promise.all([
        performanceService.getDashboard({ year: y, month: m }),
        performanceService.getMonthComparison({ year: y, month: m }),
      ]);
      const dash = (dashRes as any)?.data || dashRes || {};
      const comp = (compRes as any)?.data || compRes || {};
      setDashboardData(dash);
      setComparisonData(comp);
    } catch (error) {
      console.error('Fetch dashboard error:', error);
    } finally {
      setLoading(false);
    }
  }, [selectedMonth]);

  const fetchIndicators = async () => {
    try {
      const res = await performanceService.getIndicatorConfig();
      const data = (res as any)?.data || res || {};
      setIndicatorConfig(data);
    } catch (error) {
      console.error('Fetch indicators error:', error);
    }
  };

  const fetchInterviews = useCallback(async (page = 1) => {
    try {
      const y = selectedMonth.year();
      const m = selectedMonth.month() + 1;
      const res = await performanceService.getInterviewList({
        page,
        pageSize: 5,
        periodValue: `${y}-${String(m).padStart(2, '0')}`,
      });
      const data = (res as any)?.data || res || {};
      setInterviewList(data?.list || []);
      setInterviewTotal(data?.total || 0);
      setInterviewPage(page);
    } catch (error) {
      console.error('Fetch interviews error:', error);
    }
  }, [selectedMonth]);

  useEffect(() => {
    fetchDashboard();
    fetchIndicators();
    fetchInterviews(1);
    fetchUnreadCount();

    const interval = setInterval(fetchUnreadCount, 30000);
    return () => clearInterval(interval);
  }, [fetchDashboard, fetchInterviews, fetchUnreadCount]);

  const handleBatchCalculate = async () => {
    try {
      setBatchCalculating(true);
      const y = selectedMonth.year();
      const m = selectedMonth.month() + 1;
      await performanceService.batchCalculateScore({ year: y, month: m });
      message.success('批量计算完成');
      fetchDashboard();
      fetchInterviews(1);
    } catch (error) {
      message.error('批量计算失败');
    } finally {
      setBatchCalculating(false);
    }
  };

  const handleConfirmInterview = async (record: any) => {
    setPendingInterview(record);
    confirmForm.resetFields();
    setConfirmModalOpen(true);
  };

  const submitConfirmInterview = async () => {
    try {
      const values = await confirmForm.validateFields();
      await performanceService.confirmInterview(pendingInterview.id, values.mediatorComment);
      message.success('确认成功');
      setConfirmModalOpen(false);
      fetchInterviews(1);
      fetchDashboard();
    } catch (error) {
      message.error('确认失败');
    }
  };

  const summary = dashboardData.summary || {};
  const comparison = comparisonData.comparison || {};
  const trend = comparisonData.trend || [];

  const summaryCards = [
    { title: '考核人数', value: summary.mediatorCount || 0, icon: <TeamOutlined />, color: '#1677ff', bg: '#1677ff15' },
    { title: '受理数', value: summary.totalCases || 0, icon: <FileTextOutlined />, color: '#722ed1', bg: '#722ed115', change: comparison.caseCountChange },
    { title: '成功率', value: summary.avgSuccessRate || 0, suffix: '%', icon: <CheckCircleOutlined />, color: '#52c41a', bg: '#52c41a15', change: comparison.successRateChange, precision: 1 },
    { title: '平均时长', value: summary.avgDays || 0, suffix: '天', icon: <ClockCircleOutlined />, color: '#fa8c16', bg: '#fa8c1615', change: comparison.avgDaysChange, precision: 1 },
    { title: '满意度', value: summary.avgSatisfaction || 0, suffix: '/5', icon: <SmileOutlined />, color: '#faad14', bg: '#faad1415', change: comparison.avgSatisfactionChange, precision: 1 },
    { title: '被催办次数', value: summary.totalUrge || 0, icon: <BellOutlined />, color: '#ff4d4f', bg: '#ff4d4f15', change: comparison.urgeCountChange },
  ];

  const trendOption = useMemo(() => {
    const months = trend.map((t: any) => `${t.month || t.MONTH}月`);
    const scoreData = trend.map((t: any) => t.total_score ?? t.total_score ?? 0);
    const caseData = trend.map((t: any) => t.case_count ?? t.CASE_COUNT ?? 0);
    return {
      tooltip: { trigger: 'axis', backgroundColor: 'rgba(255,255,255,0.95)', borderColor: '#f0f0f0', textStyle: { color: '#333' } },
      legend: { data: ['综合得分', '受理数'], top: 0 },
      grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
      xAxis: { type: 'category', data: months.length > 0 ? months : Array.from({ length: 12 }, (_, i) => `${i + 1}月`), axisLabel: { color: '#666' } },
      yAxis: [
        { type: 'value', name: '得分', min: 50, max: 100, splitLine: { lineStyle: { color: '#f0f0f0' } }, axisLabel: { color: '#666' } },
        { type: 'value', name: '受理数', splitLine: { show: false }, axisLabel: { color: '#666' } },
      ],
      series: [
        {
          name: '综合得分', type: 'line', smooth: true, data: scoreData.length > 0 ? scoreData : Array(12).fill(null),
          itemStyle: { color: '#722ed1' }, lineStyle: { width: 3 },
          areaStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: 'rgba(114,46,209,0.3)' }, { offset: 1, color: 'rgba(114,46,209,0.02)' }] } },
        },
        {
          name: '受理数', type: 'bar', yAxisIndex: 1, data: caseData.length > 0 ? caseData : Array(12).fill(null),
          itemStyle: { color: '#1677ff', borderRadius: [4, 4, 0, 0] }, barMaxWidth: 30,
        },
      ],
    };
  }, [trend]);

  const radarOption = useMemo(() => {
    const mediators = dashboardData.mediators || [];
    const top = mediators[0] || ({} as PerformanceStats);
    const avg = {
      caseCount: 0, successRate: 0, avgDays: 0, satisfaction: 0, urgeCount: 0, count: 0,
    };
    mediators.forEach((m: PerformanceStats) => {
      avg.caseCount += (m.caseCount || m.totalCases || 0);
      avg.successRate += (m.successRate || m.mediationSuccessRate || 0);
      avg.avgDays += (m.avgDays || m.avgDuration || 0);
      avg.satisfaction += (m.avgSatisfaction || m.satisfaction || 0);
      avg.urgeCount += (m.urgeCount || 0);
      avg.count++;
    });
    if (avg.count > 0) {
      avg.caseCount /= avg.count;
      avg.successRate /= avg.count;
      avg.avgDays /= avg.count;
      avg.satisfaction /= avg.count;
      avg.urgeCount /= avg.count;
    }
    return {
      tooltip: { backgroundColor: 'rgba(255,255,255,0.95)', borderColor: '#f0f0f0', textStyle: { color: '#333' } },
      legend: { data: ['TOP1', '平均水平'], top: 0 },
      radar: {
        indicator: [
          { name: '受理量', max: 50 },
          { name: '成功率(%)', max: 100 },
          { name: '效率(逆向)', max: 100 },
          { name: '满意度', max: 5 },
          { name: '无催办(逆向)', max: 100 },
        ],
        splitArea: { areaStyle: { color: ['#fafafa', '#fff'] } },
      },
      series: [{
        type: 'radar',
        data: [
          {
            value: [
              top.caseCount || top.totalCases || 0,
              top.successRate || top.mediationSuccessRate || 0,
              Math.max(0, 100 - (top.avgDays || top.avgDuration || 0) * 5),
              top.avgSatisfaction || top.satisfaction || 0,
              Math.max(0, 100 - (top.urgeCount || 0) * 20),
            ],
            name: 'TOP1',
            areaStyle: { color: 'rgba(22,119,255,0.3)' },
            lineStyle: { color: '#1677ff' },
            itemStyle: { color: '#1677ff' },
          },
          {
            value: [
              avg.caseCount, avg.successRate,
              Math.max(0, 100 - avg.avgDays * 5),
              avg.satisfaction,
              Math.max(0, 100 - avg.urgeCount * 20),
            ],
            name: '平均水平',
            areaStyle: { color: 'rgba(82,196,26,0.2)' },
            lineStyle: { color: '#52c41a', type: 'dashed' },
            itemStyle: { color: '#52c41a' },
          },
        ],
      }],
    };
  }, [dashboardData]);

  const indicatorBarOption = useMemo(() => {
    const indicators = indicatorConfig.indicators || [];
    if (indicators.length === 0) return {};
    return {
      tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' }, backgroundColor: 'rgba(255,255,255,0.95)', borderColor: '#f0f0f0', textStyle: { color: '#333' } },
      grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
      xAxis: { type: 'category', data: indicators.map((i) => i.indicator_name), axisLabel: { color: '#666' } },
      yAxis: { type: 'value', name: '权重(%)', max: 50, splitLine: { lineStyle: { color: '#f0f0f0' } }, axisLabel: { color: '#666' } },
      series: [{
        type: 'bar',
        data: indicators.map((i) => Math.round(i.weight * 100)),
        itemStyle: {
          color: (params: any) => {
            const colors = ['#1677ff', '#52c41a', '#fa8c16', '#faad14', '#ff4d4f'];
            return colors[params.dataIndex % colors.length];
          },
          borderRadius: [4, 4, 0, 0],
        },
        barMaxWidth: 50,
        label: { show: true, position: 'top', formatter: '{c}%' },
      }],
    };
  }, [indicatorConfig]);

  const rankTableColumns = [
    {
      title: '排名', dataIndex: 'rank', width: 65, align: 'center' as const,
      render: (val: number, _: any, index: number) => {
        const realRank = val || index + 1;
        return (
          <div style={{ width: 30, height: 30, borderRadius: realRank <= 3 ? '50%' : 6, background: realRank === 1 ? 'linear-gradient(135deg, #ffd700, #ff8c00)' : realRank === 2 ? 'linear-gradient(135deg, #c0c0c0, #808080)' : realRank === 3 ? 'linear-gradient(135deg, #cd7f32, #8b4513)' : '#f0f0f0', color: realRank <= 3 ? '#fff' : '#666', display: 'flex', alignItems: 'center', justifyContent: 'center', fontWeight: 600, margin: '0 auto', fontSize: 13 }}>
            {realRank}
          </div>
        );
      },
    },
    {
      title: '调解员', dataIndex: 'user_name', width: 130,
      render: (val: string, record: any) => (
        <Space>
          <div style={{ width: 34, height: 34, borderRadius: '50%', background: '#1677ff15', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#1677ff', fontWeight: 600 }}>
            {val?.charAt(0)}
          </div>
          <Space direction="vertical" size={0}>
            <span style={{ fontWeight: 500 }}>{val}</span>
            <span style={{ fontSize: 12, color: '#999' }}>{record.org_name}</span>
          </Space>
        </Space>
      ),
    },
    { title: '受理数', dataIndex: 'case_count', width: 80, align: 'center' as const, sorter: (a: any, b: any) => (a.case_count || 0) - (b.case_count || 0) },
    { title: '办结率(%)', dataIndex: 'close_rate', width: 100, align: 'center' as const, render: (v: number) => <span>{(v || 0).toFixed(1)}%</span> },
    {
      title: '成功率', dataIndex: 'success_rate', width: 150,
      render: (v: number) => <Progress percent={v || 0} size="small" status={v >= 90 ? 'success' : v >= 70 ? 'active' : 'exception'} strokeColor={v >= 90 ? '#52c41a' : v >= 70 ? '#1677ff' : '#ff4d4f'} />,
    },
    { title: '平均天数', dataIndex: 'avg_days', width: 90, align: 'center' as const, render: (v: number) => <span>{(v || 0).toFixed(1)}</span> },
    { title: '满意度', dataIndex: 'avg_satisfaction', width: 90, align: 'center' as const, render: (v: number) => <Space><StarOutlined style={{ color: '#faad14' }} />{(v || 0).toFixed(1)}</Space> },
    { title: '催办', dataIndex: 'urge_count', width: 70, align: 'center' as const, render: (v: number) => <Tag color={v > 3 ? 'red' : v > 0 ? 'orange' : 'green'}>{v || 0}</Tag> },
    {
      title: '得分', dataIndex: 'total_score', width: 90, align: 'center' as const,
      render: (v: number) => <Tag color={v >= 90 ? 'green' : v >= 75 ? 'blue' : v >= 60 ? 'orange' : 'red'}><strong>{(v || 0).toFixed(1)}</strong></Tag>,
    },
    {
      title: '等级', dataIndex: 'level', width: 65, align: 'center' as const,
      render: (v: string) => <Tag style={{ background: levelColorMap[v] || '#d9d9d9', color: '#fff', border: 'none', fontWeight: 700 }}>{v || '-'}</Tag>,
    },
    {
      title: '操作', width: 80, align: 'center' as const,
      render: (_: any, record: any) => isAdmin ? (
        <Tooltip title="创建面谈记录">
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => { setSelectedMediator(record); setInterviewModalOpen(true); }} />
        </Tooltip>
      ) : null,
    },
  ];

  const handleExportExcel = async () => {
    try {
      const y = selectedMonth.year();
      const m = selectedMonth.month() + 1;
      const res = await performanceService.exportExcel({ year: y, month: m });
      const blob = new Blob([res as any], { type: 'text/csv;charset=utf-8' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `performance_${y}_${String(m).padStart(2, '0')}.csv`;
      link.click();
      window.URL.revokeObjectURL(url);
      message.success('导出成功');
    } catch (error) {
      message.error('导出失败');
    }
  };

  const handleWeightUpdate = async () => {
    try {
      const values = weightForm.getFieldsValue();
      const indicators = indicatorConfig.indicators.map((ind) => ({
        id: ind.id,
        weight: values[`weight_${ind.id}`],
      }));

      if (autoRecalculate) {
        modal.confirm({
          title: '确认更新权重并重新计算',
          content: `是否更新权重并立即重新计算 ${selectedMonth.format('YYYY年MM月')} 所有调解员的绩效得分？`,
          okText: '确认更新并计算',
          cancelText: '仅更新权重',
          onOk: async () => {
            await performanceService.updateIndicatorConfig(indicators, true);
            message.success('权重更新成功，已触发重新计算');
            setWeightModalOpen(false);
            fetchIndicators();
            setTimeout(fetchDashboard, 500);
            setTimeout(() => fetchInterviews(1), 800);
          },
          onCancel: async () => {
            await performanceService.updateIndicatorConfig(indicators, false);
            message.success('权重更新成功，后续计算将使用新权重');
            setWeightModalOpen(false);
            fetchIndicators();
          },
        });
      } else {
        await performanceService.updateIndicatorConfig(indicators, false);
        message.success('权重更新成功');
        setWeightModalOpen(false);
        fetchIndicators();
      }
    } catch (error) {
      message.error('权重更新失败');
    }
  };

  const handleCreateInterview = async () => {
    try {
      const values = await interviewForm.validateFields();
      await performanceService.createInterview({
        ...values,
        userId: selectedMediator?.user_id || selectedMediator?.userId,
        userName: selectedMediator?.user_name || selectedMediator?.userName,
        totalScore: selectedMediator?.total_score || selectedMediator?.score,
        level: selectedMediator?.level,
      });
      message.success('面谈记录创建成功');
      setInterviewModalOpen(false);
      interviewForm.resetFields();
      fetchInterviews(1);
    } catch (error) {
      message.error('创建面谈记录失败');
    }
  };

  const handleViewInterview = async (id: number) => {
    try {
      const res = await performanceService.getInterviewDetail(id);
      const data = (res as any)?.data || res || {};
      setInterviewDetail(data);
      setInterviewDetailOpen(true);
    } catch (error) {
      message.error('获取面谈记录失败');
    }
  };

  const openWeightModal = () => {
    const formValues: Record<string, number> = {};
    indicatorConfig.indicators.forEach((ind) => {
      formValues[`weight_${ind.id}`] = ind.weight;
    });
    weightForm.setFieldsValue(formValues);
    setWeightModalOpen(true);
  };

  const interviewColumns = [
    { title: '编号', dataIndex: 'interview_no', width: 140 },
    { title: '调解员', dataIndex: 'user_name', width: 90 },
    { title: '类型', dataIndex: 'interview_type_name', width: 90 },
    { title: '面谈人', dataIndex: 'interviewer_name', width: 90 },
    { title: '面谈时间', dataIndex: 'interview_time', width: 150 },
    {
      title: '状态', dataIndex: 'status_name', width: 80,
      render: (v: string) => <Tag color={v === '已确认' ? 'green' : v === '待确认' ? 'orange' : 'default'}>{v}</Tag>,
    },
    {
      title: '操作', width: 140,
      render: (_: any, record: any) => (
        <Space size={4}>
          <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => handleViewInterview(record.id)} />
          {isMediator && record.status === 1 && (
            <Button type="primary" size="small" icon={<CheckOutlined />} onClick={() => handleConfirmInterview(record)}>
              确认
            </Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Space>
          <h2 style={{ margin: 0 }}>
            <TrophyOutlined style={{ color: '#faad14', marginRight: 8 }} />
            调解员绩效考核看板
          </h2>
          <Tag icon={<DatabaseOutlined />} color={summary.dataSource === 'realtime' ? 'orange' : 'green'}>
            {summary.dataSource === 'realtime' ? '实时计算' : '快照数据'}
          </Tag>
        </Space>
        <Space>
          <Popover
            open={notificationPopoverOpen}
            onOpenChange={(open) => setNotificationPopoverOpen(open)}
            placement="bottomRight"
            content={
              <div style={{ width: 320, maxHeight: 400, overflowY: 'auto' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
                  <strong>消息通知</strong>
                  {unreadCount > 0 && (
                    <Button type="link" size="small" onClick={async () => {
                      await notificationService.markAllAsRead();
                      fetchUnreadCount();
                      fetchNotifications();
                    }}>全部已读</Button>
                  )}
                </div>
                <Divider style={{ margin: '8px 0' }} />
                {notifications.length === 0 ? (
                  <Empty description="暂无新消息" image={Empty.PRESENTED_IMAGE_SIMPLE} />
                ) : (
                  <List
                    dataSource={notifications}
                    size="small"
                    renderItem={(item) => (
                      <List.Item
                        style={{ cursor: 'pointer', padding: '8px 0', borderBottom: '1px solid #f0f0f0' }}
                        onClick={async () => {
                          await notificationService.markAsRead(item.id);
                          fetchUnreadCount();
                          setNotificationPopoverOpen(false);
                        }}
                      >
                        <List.Item.Meta
                          title={<span style={{ fontSize: 13 }}>{item.title}</span>}
                          description={
                            <span style={{ fontSize: 12, color: '#999' }}>
                              {item.content?.substring(0, 50)}...
                            </span>
                          }
                        />
                      </List.Item>
                    )}
                  />
                )}
              </div>
            }
            trigger="click"
          >
            <Badge count={unreadCount} size="small">
              <Button icon={<BellOutlined />} shape="circle" />
            </Badge>
          </Popover>
          <MonthPicker
            value={selectedMonth}
            onChange={(date) => { if (date) setSelectedMonth(date); }}
            style={{ width: 180 }}
            allowClear={false}
          />
          {(isAdmin || isLeader) && (
            <Button
              icon={<CalculatorOutlined />}
              loading={batchCalculating}
              onClick={handleBatchCalculate}
              type="primary"
            >
              批量计算
            </Button>
          )}
          {isAdmin && (
            <Button icon={<SettingOutlined />} onClick={openWeightModal}>考核权重</Button>
          )}
          <Button icon={<ReloadOutlined />} onClick={fetchDashboard} loading={loading}>刷新</Button>
          <Button icon={<DownloadOutlined />} onClick={handleExportExcel}>导出Excel</Button>
        </Space>
      </div>

      <Row gutter={[12, 12]}>
        {summaryCards.map((card, idx) => (
          <Col xs={12} sm={8} md={4} key={idx}>
            <Card bordered={false} style={{ borderRadius: 12 }} bodyStyle={{ padding: '16px 20px' }}>
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <div>
                  <div style={{ color: '#999', fontSize: 13, marginBottom: 6 }}>{card.title}</div>
                  <div style={{ display: 'flex', alignItems: 'baseline', gap: 4 }}>
                    <Statistic
                      value={card.value}
                      precision={card.precision}
                      suffix={card.suffix}
                      valueStyle={{ color: card.color, fontSize: 22, fontWeight: 700 }}
                    />
                    {card.change !== undefined && <ChangeTag value={card.change} />}
                  </div>
                </div>
                <div style={{ width: 46, height: 46, borderRadius: 12, background: card.bg, display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 22, color: card.color }}>
                  {card.icon}
                </div>
              </div>
            </Card>
          </Col>
        ))}
      </Row>

      <Row gutter={[16, 16]}>
        <Col xs={24} lg={10}>
          <Card title="绩效趋势(本月vs上月)" bordered={false} style={{ borderRadius: 12 }}
            extra={<Tag color="purple">{selectedMonth.format('YYYY年')}</Tag>}
          >
            <ReactECharts option={trendOption} style={{ height: 340 }} notMerge lazyUpdate />
          </Card>
        </Col>
        <Col xs={24} lg={7}>
          <Card title="能力雷达分析" bordered={false} style={{ borderRadius: 12 }}
            extra={<Tag color="purple">TOP1 vs 均值</Tag>}
          >
            <ReactECharts option={radarOption} style={{ height: 340 }} notMerge lazyUpdate />
          </Card>
        </Col>
        <Col xs={24} lg={7}>
          <Card title="考核权重分布" bordered={false} style={{ borderRadius: 12 }}
            extra={isAdmin ? <Button type="link" size="small" onClick={openWeightModal}>调整权重</Button> : null}
          >
            <ReactECharts option={indicatorBarOption} style={{ height: 340 }} notMerge lazyUpdate />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]}>
        <Col xs={24} lg={16}>
          <Card title="调解员绩效排行" bordered={false} style={{ borderRadius: 12 }}
            extra={
              <Space>
                <Tag color="green"><ArrowUpOutlined /> 优秀: ≥90</Tag>
                <Tag color="blue">良好: 75-89</Tag>
                <Tag color="orange">合格: 60-74</Tag>
                <Tag color="red"><ArrowDownOutlined /> 待改进: &lt;60</Tag>
              </Space>
            }
          >
            <Table
              columns={rankTableColumns as any}
              dataSource={dashboardData.mediators || []}
              rowKey={(r: any) => r.user_id || r.userId || Math.random()}
              loading={loading}
              pagination={{ showSizeChanger: true, showTotal: (t) => `共 ${t} 人`, pageSize: 10 }}
              scroll={{ x: 1100 }}
              size="middle"
            />
          </Card>
        </Col>
        <Col xs={24} lg={8}>
          <Card title="绩效面谈记录" bordered={false} style={{ borderRadius: 12 }}
            extra={<Tag>{selectedMonth.format('YYYY-MM')}</Tag>}
          >
            <Table
              columns={interviewColumns}
              dataSource={interviewList}
              rowKey={(r: any) => r.id}
              pagination={{
                current: interviewPage,
                total: interviewTotal,
                pageSize: 5,
                onChange: fetchInterviews,
                size: 'small',
              }}
              size="small"
            />
          </Card>
        </Col>
      </Row>

      <Modal
        title="自定义考核权重"
        open={weightModalOpen}
        onOk={handleWeightUpdate}
        onCancel={() => setWeightModalOpen(false)}
        width={560}
        okText="保存"
      >
        <div style={{ marginBottom: 12, color: '#666', fontSize: 13 }}>
          权重总和必须为 100%，当前用于计算调解员综合得分
        </div>
        <Form form={weightForm} layout="vertical">
          {indicatorConfig.indicators.map((ind) => (
            <Form.Item key={ind.id} label={`${ind.indicator_name} (${ind.indicator_code})`} name={`weight_${ind.id}`}
              rules={[{ required: true, message: '请输入权重' }]}
              extra={ind.description}
            >
              <InputNumber min={0} max={1} step={0.05} style={{ width: '100%' }}
                formatter={(v) => `${((v || 0) as number) * 100}%`}
                parser={(v) => (parseFloat((v || '0').replace('%', '')) / 100) as 0}
              />
            </Form.Item>
          ))}
        </Form>
        <div style={{ marginTop: 16, padding: 12, background: '#f5f5f5', borderRadius: 8 }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <Space>
              <RiseOutlined style={{ color: '#faad14' }} />
              <span style={{ color: '#666', fontSize: 13 }}>自动重新计算本月绩效得分</span>
            </Space>
            <Switch checked={autoRecalculate} onChange={setAutoRecalculate} />
          </div>
          <div style={{ color: '#999', fontSize: 12, marginTop: 4 }}>
            开启后保存权重时将立即重新计算所有调解员的绩效数据
          </div>
        </div>
      </Modal>

      <Modal
        title={`创建绩效面谈 - ${selectedMediator?.user_name || selectedMediator?.userName || ''}`}
        open={interviewModalOpen}
        onOk={handleCreateInterview}
        onCancel={() => { setInterviewModalOpen(false); interviewForm.resetFields(); }}
        width={640}
        okText="创建"
      >
        <Form form={interviewForm} layout="vertical">
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item label="面谈类型" name="interviewType" rules={[{ required: true, message: '请选择面谈类型' }]}>
                <Select options={[
                  { value: 1, label: '绩效反馈' },
                  { value: 2, label: '改进计划' },
                  { value: 3, label: '表彰面谈' },
                  { value: 4, label: '预警面谈' },
                ]} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="面谈时间" name="interviewTime" rules={[{ required: true, message: '请选择时间' }]}>
                <Input type="datetime-local" />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item label="面谈地点" name="interviewPlace">
            <Input placeholder="请输入面谈地点" />
          </Form.Item>
          <Form.Item label="周期" name="periodValue" initialValue={selectedMonth.format('YYYY-MM')}>
            <Input disabled />
          </Form.Item>
          <Form.Item name="periodType" initialValue={1} hidden><Input /></Form.Item>
          <Form.Item label="工作亮点" name="strengths">
            <Input.TextArea rows={2} placeholder="调解员在本考核期的工作亮点" />
          </Form.Item>
          <Form.Item label="待改进方面" name="weaknesses">
            <Input.TextArea rows={2} placeholder="需要改进的方面" />
          </Form.Item>
          <Form.Item label="改进计划" name="improvementPlan">
            <Input.TextArea rows={2} placeholder="下一周期的改进计划" />
          </Form.Item>
          <Form.Item label="下期目标" name="targetNextPeriod">
            <Input.TextArea rows={2} placeholder="下一考核期的目标" />
          </Form.Item>
        </Form>
      </Modal>

      <Drawer
        title="面谈记录详情"
        open={interviewDetailOpen}
        onClose={() => setInterviewDetailOpen(false)}
        width={520}
        extra={
          isMediator && interviewDetail?.status === 1 ? (
            <Button type="primary" icon={<CheckOutlined />} onClick={() => handleConfirmInterview(interviewDetail)}>
              确认面谈
            </Button>
          ) : null
        }
      >
        {interviewDetail && (
          <Descriptions column={1} bordered size="small">
            <Descriptions.Item label="面谈编号">{interviewDetail.interview_no}</Descriptions.Item>
            <Descriptions.Item label="调解员">{interviewDetail.user_name}</Descriptions.Item>
            <Descriptions.Item label="面谈人">{interviewDetail.interviewer_name}</Descriptions.Item>
            <Descriptions.Item label="面谈类型">{interviewDetail.interview_type_name}</Descriptions.Item>
            <Descriptions.Item label="面谈时间">{interviewDetail.interview_time}</Descriptions.Item>
            <Descriptions.Item label="面谈地点">{interviewDetail.interview_place}</Descriptions.Item>
            <Descriptions.Item label="考核得分">{interviewDetail.total_score}</Descriptions.Item>
            <Descriptions.Item label="考核等级">
              <Tag style={{ background: levelColorMap[interviewDetail.level] || '#d9d9d9', color: '#fff', border: 'none' }}>{interviewDetail.level}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="状态">
              <Tag color={interviewDetail.status === 2 ? 'green' : 'orange'}>{interviewDetail.status_name}</Tag>
            </Descriptions.Item>
            <Divider />
            <Descriptions.Item label="工作亮点">{interviewDetail.strengths || '-'}</Descriptions.Item>
            <Descriptions.Item label="待改进方面">{interviewDetail.weaknesses || '-'}</Descriptions.Item>
            <Descriptions.Item label="改进计划">{interviewDetail.improvement_plan || '-'}</Descriptions.Item>
            <Descriptions.Item label="下期目标">{interviewDetail.target_next_period || '-'}</Descriptions.Item>
            <Descriptions.Item label="调解员反馈">{interviewDetail.mediator_comment || '-'}</Descriptions.Item>
          </Descriptions>
        )}
      </Drawer>

      <Modal
        title={
          <Space>
            <CheckOutlined style={{ color: '#52c41a' }} />
            确认绩效面谈
          </Space>
        }
        open={confirmModalOpen}
        onOk={submitConfirmInterview}
        onCancel={() => setConfirmModalOpen(false)}
        width={560}
        okText="确认提交"
        cancelText="取消"
      >
        {pendingInterview && (
          <Space direction="vertical" size={12} style={{ width: '100%' }}>
            <Alert
              type="info"
              showIcon
              message={`您正在确认 ${pendingInterview.period_value} 的${pendingInterview.interview_type_name}记录`}
              description={`面谈编号: ${pendingInterview.interview_no}，综合得分: ${pendingInterview.total_score}分(${pendingInterview.level})`}
            />
            <Descriptions column={2} size="small" bordered>
              <Descriptions.Item label="面谈人">{pendingInterview.interviewer_name}</Descriptions.Item>
              <Descriptions.Item label="面谈时间">{pendingInterview.interview_time}</Descriptions.Item>
              <Descriptions.Item label="工作亮点" span={2}>{pendingInterview.strengths || '-'}</Descriptions.Item>
              <Descriptions.Item label="待改进" span={2}>{pendingInterview.weaknesses || '-'}</Descriptions.Item>
              <Descriptions.Item label="改进计划" span={2}>{pendingInterview.improvement_plan || '-'}</Descriptions.Item>
              <Descriptions.Item label="下期目标" span={2}>{pendingInterview.target_next_period || '-'}</Descriptions.Item>
            </Descriptions>
            <Form form={confirmForm} layout="vertical">
              <Form.Item
                label="您的意见和反馈（必填）"
                name="mediatorComment"
                rules={[{ required: true, message: '请填写您的意见和反馈' }]}
              >
                <Input.TextArea
                  rows={4}
                  placeholder="请填写您对本次绩效面谈的意见、建议或确认信息..."
                />
              </Form.Item>
            </Form>
            <div style={{ color: '#999', fontSize: 12 }}>
              <CheckOutlined style={{ color: '#52c41a', marginRight: 4 }} />
              确认后，您将无法再修改反馈内容
            </div>
          </Space>
        )}
      </Modal>
    </Space>
  );
};

export default Performance;
