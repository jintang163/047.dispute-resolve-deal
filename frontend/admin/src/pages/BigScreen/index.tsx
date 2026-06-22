import React, { useEffect, useRef, useState, useCallback } from 'react';
import * as echarts from 'echarts';
import html2canvas from 'html2canvas';
import { getToken } from '../../utils/auth';
import request from '../../utils/request';
import './index.css';

interface OverviewData {
  totalCases: number;
  pendingCases: number;
  mediatingCases: number;
  closedCases: number;
  successCount: number;
  successRate: string;
  avgDays: string;
  timeoutCount: number;
  avgSatisfaction: string;
  todayNew: number;
  isWarning: boolean;
}

interface BigScreenData {
  overview: OverviewData;
  trendData: Array<{ date: string; count: number }>;
  typeStats: Array<{ type_name: string; count: number }>;
  mediatorRanking: Array<{
    rank: number;
    real_name: string;
    org_name: string;
    total_cases: number;
    closed_cases: number;
    success_cases: number;
    success_rate: string;
    is_warning: boolean;
  }>;
  orgStats: Array<{
    id: number;
    org_name: string;
    total_cases: number;
    closed_cases: number;
    success_cases: number;
    success_rate: string;
    is_warning: boolean;
  }>;
  heatmapData: Array<{
    day: string;
    hour: string;
    value: number;
    dayIndex: number;
    hourIndex: number;
  }>;
  updateTime: string;
}

const BigScreen: React.FC = () => {
  const screenRef = useRef<HTMLDivElement>(null);
  const trendChartRef = useRef<HTMLDivElement>(null);
  const typeChartRef = useRef<HTMLDivElement>(null);
  const rankChartRef = useRef<HTMLDivElement>(null);
  const heatmapChartRef = useRef<HTMLDivElement>(null);
  const orgChartRef = useRef<HTMLDivElement>(null);

  const [data, setData] = useState<BigScreenData | null>(null);
  const [carouseIndex, setCarouseIndex] = useState(0);
  const [isCarouselPaused, setIsCarouselPaused] = useState(false);
  const [currentTime, setCurrentTime] = useState(new Date());
  const [wsConnected, setWsConnected] = useState(false);

  const carouselViews = [
    { key: 'overview', label: '总览' },
    { key: 'trend', label: '纠纷趋势' },
    { key: 'heatmap', label: '热力图' },
    { key: 'ranking', label: '工作量排行' },
    { key: 'org', label: '组织统计' },
  ];

  const fetchData = useCallback(async () => {
    try {
      const res = await request.get('/stats/bigscreen');
      if (res.data?.code === 0) {
        setData(res.data.data);
      }
    } catch (error) {
      console.error('Fetch bigscreen data failed:', error);
    }
  }, []);

  const initWebSocket = useCallback(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/api/v1/ws/bigscreen?token=${getToken()}`;

    const ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      setWsConnected(true);
      console.log('BigScreen WebSocket connected');
    };

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        if (msg.type === 'bigscreen_update' && msg.data) {
          setData((prev) => (prev ? { ...prev, ...msg.data } : msg.data));
        }
      } catch (e) {
        console.error('Parse websocket message failed:', e);
      }
    };

    ws.onclose = () => {
      setWsConnected(false);
      console.log('BigScreen WebSocket disconnected, reconnecting...');
      setTimeout(initWebSocket, 5000);
    };

    ws.onerror = (error) => {
      console.error('BigScreen WebSocket error:', error);
    };

    return () => {
      ws.close();
    };
  }, []);

  const initTrendChart = useCallback(() => {
    if (!trendChartRef.current || !data) return;
    const chart = echarts.init(trendChartRef.current);
    const dates = data.trendData.map((item) => item.date?.slice(5) || '');
    const counts = data.trendData.map((item) => item.count || 0);

    const option: echarts.EChartsOption = {
      backgroundColor: 'transparent',
      tooltip: {
        trigger: 'axis',
        backgroundColor: 'rgba(0, 20, 40, 0.9)',
        borderColor: '#00d4ff',
        textStyle: { color: '#fff' },
        axisPointer: {
          type: 'shadow',
          shadowStyle: {
            color: 'rgba(0, 212, 255, 0.1)',
          },
        },
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        top: '15%',
        containLabel: true,
      },
      xAxis: {
        type: 'category',
        data: dates,
        axisLine: { lineStyle: { color: '#00d4ff33' } },
        axisLabel: { color: '#8ec5fc', fontSize: 11 },
        axisTick: { show: false },
      },
      yAxis: {
        type: 'value',
        splitLine: { lineStyle: { color: '#00d4ff22' } },
        axisLine: { show: false },
        axisLabel: { color: '#8ec5fc', fontSize: 11 },
        axisTick: { show: false },
      },
      series: [
        {
          name: '新增案件',
          type: 'line',
          smooth: true,
          symbol: 'circle',
          symbolSize: 8,
          data: counts,
          itemStyle: { color: '#00d4ff' },
          lineStyle: { width: 3, color: '#00d4ff' },
          areaStyle: {
            color: {
              type: 'linear',
              x: 0,
              y: 0,
              x2: 0,
              y2: 1,
              colorStops: [
                { offset: 0, color: 'rgba(0, 212, 255, 0.5)' },
                { offset: 1, color: 'rgba(0, 212, 255, 0.02)' },
              ],
            },
          },
        },
      ],
    };

    chart.setOption(option);
    return () => chart.dispose();
  }, [data]);

  const initTypeChart = useCallback(() => {
    if (!typeChartRef.current || !data) return;
    const chart = echarts.init(typeChartRef.current);
    const colors = ['#00d4ff', '#00ff88', '#ffcc00', '#ff6b6b', '#9b59b6', '#3498db', '#1abc9c', '#e67e22'];

    const option: echarts.EChartsOption = {
      backgroundColor: 'transparent',
      tooltip: {
        trigger: 'item',
        backgroundColor: 'rgba(0, 20, 40, 0.9)',
        borderColor: '#00d4ff',
        textStyle: { color: '#fff' },
        formatter: '{b}: {c}件 ({d}%)',
      },
      legend: {
        orient: 'vertical',
        right: '5%',
        top: 'center',
        textStyle: { color: '#8ec5fc', fontSize: 12 },
        itemWidth: 10,
        itemHeight: 10,
      },
      series: [
        {
          name: '案件类型',
          type: 'pie',
          radius: ['40%', '70%'],
          center: ['35%', '50%'],
          avoidLabelOverlap: false,
          itemStyle: {
            borderRadius: 6,
            borderColor: '#001529',
            borderWidth: 2,
          },
          label: { show: false },
          emphasis: {
            label: {
              show: true,
              fontSize: 14,
              fontWeight: 'bold',
              color: '#fff',
            },
          },
          labelLine: { show: false },
          data: data.typeStats.map((item, index) => ({
            value: item.count,
            name: item.type_name,
            itemStyle: { color: colors[index % colors.length] },
          })),
        },
      ],
    };

    chart.setOption(option);
    return () => chart.dispose();
  }, [data]);

  const initRankChart = useCallback(() => {
    if (!rankChartRef.current || !data) return;
    const chart = echarts.init(rankChartRef.current);

    const names = data.mediatorRanking.map((item) => item.real_name || '').reverse();
    const cases = data.mediatorRanking.map((item) => item.total_cases || 0).reverse();
    const warnings = data.mediatorRanking.map((item) => item.is_warning).reverse();

    const option: echarts.EChartsOption = {
      backgroundColor: 'transparent',
      tooltip: {
        trigger: 'axis',
        backgroundColor: 'rgba(0, 20, 40, 0.9)',
        borderColor: '#00d4ff',
        textStyle: { color: '#fff' },
        axisPointer: { type: 'shadow' },
      },
      grid: {
        left: '3%',
        right: '10%',
        bottom: '3%',
        top: '5%',
        containLabel: true,
      },
      xAxis: {
        type: 'value',
        splitLine: { lineStyle: { color: '#00d4ff22' } },
        axisLine: { show: false },
        axisLabel: { color: '#8ec5fc', fontSize: 11 },
        axisTick: { show: false },
      },
      yAxis: {
        type: 'category',
        data: names,
        axisLine: { lineStyle: { color: '#00d4ff33' } },
        axisLabel: { color: '#8ec5fc', fontSize: 12 },
        axisTick: { show: false },
      },
      series: [
        {
          name: '办案量',
          type: 'bar',
          barWidth: '60%',
          data: cases.map((value, index) => ({
            value,
            itemStyle: {
              color: warnings[index]
                ? {
                    type: 'linear',
                    x: 0,
                    y: 0,
                    x2: 1,
                    y2: 0,
                    colorStops: [
                      { offset: 0, color: '#ff4757' },
                      { offset: 1, color: '#ff6b81' },
                    ],
                  }
                : {
                    type: 'linear',
                    x: 0,
                    y: 0,
                    x2: 1,
                    y2: 0,
                    colorStops: [
                      { offset: 0, color: '#00d4ff' },
                      { offset: 1, color: '#00ff88' },
                    ],
                  },
              borderRadius: [0, 4, 4, 0],
            },
          })),
          label: {
            show: true,
            position: 'right',
            color: '#8ec5fc',
            fontSize: 12,
            formatter: '{c}件',
          },
        },
      ],
    };

    chart.setOption(option);
    return () => chart.dispose();
  }, [data]);

  const initHeatmapChart = useCallback(() => {
    if (!heatmapChartRef.current || !data) return;
    const chart = echarts.init(heatmapChartRef.current);

    const hours = ['00', '02', '04', '06', '08', '10', '12', '14', '16', '18', '20', '22'];
    const weeks = ['周一', '周二', '周三', '周四', '周五', '周六', '周日'];
    const heatData = data.heatmapData.map((item) => [item.hourIndex, item.dayIndex, item.value]);

    const option: echarts.EChartsOption = {
      backgroundColor: 'transparent',
      tooltip: {
        position: 'top',
        backgroundColor: 'rgba(0, 20, 40, 0.9)',
        borderColor: '#00d4ff',
        textStyle: { color: '#fff' },
        formatter: (params: any) => {
          const d = params.data;
          return `${weeks[d[1]]} ${hours[d[0]]}:00 - 案件量: ${d[2]}`;
        },
      },
      grid: {
        left: '8%',
        right: '12%',
        top: '8%',
        bottom: '18%',
      },
      xAxis: {
        type: 'category',
        data: hours,
        splitArea: { show: true, areaStyle: { color: ['#001529', '#001c38'] } },
        axisLine: { lineStyle: { color: '#00d4ff33' } },
        axisLabel: { color: '#8ec5fc', fontSize: 11 },
        axisTick: { show: false },
      },
      yAxis: {
        type: 'category',
        data: weeks,
        splitArea: { show: true, areaStyle: { color: ['#001529', '#001c38'] } },
        axisLine: { lineStyle: { color: '#00d4ff33' } },
        axisLabel: { color: '#8ec5fc', fontSize: 12 },
        axisTick: { show: false },
      },
      visualMap: {
        min: 0,
        max: 15,
        calculable: true,
        orient: 'horizontal',
        left: 'center',
        bottom: '0%',
        textStyle: { color: '#8ec5fc', fontSize: 11 },
        inRange: {
          color: ['#001a33', '#0066cc', '#00d4ff', '#00ff88', '#ffcc00', '#ff6b6b'],
        },
      },
      series: [
        {
          name: '案件分布',
          type: 'heatmap',
          data: heatData,
          label: { show: false },
          emphasis: {
            itemStyle: {
              shadowBlur: 10,
              shadowColor: 'rgba(0, 212, 255, 0.5)',
            },
          },
        },
      ],
    };

    chart.setOption(option);
    return () => chart.dispose();
  }, [data]);

  const initOrgChart = useCallback(() => {
    if (!orgChartRef.current || !data) return;
    const chart = echarts.init(orgChartRef.current);

    const names = data.orgStats.map((item) => item.org_name || '').reverse();
    const successRates = data.orgStats.map((item) => parseFloat(item.success_rate || '0')).reverse();
    const warnings = data.orgStats.map((item) => item.is_warning).reverse();

    const option: echarts.EChartsOption = {
      backgroundColor: 'transparent',
      tooltip: {
        trigger: 'axis',
        backgroundColor: 'rgba(0, 20, 40, 0.9)',
        borderColor: '#00d4ff',
        textStyle: { color: '#fff' },
        formatter: '{b}: {c}%',
        axisPointer: { type: 'shadow' },
      },
      grid: {
        left: '3%',
        right: '10%',
        bottom: '3%',
        top: '5%',
        containLabel: true,
      },
      xAxis: {
        type: 'value',
        max: 100,
        splitLine: { lineStyle: { color: '#00d4ff22' } },
        axisLine: { show: false },
        axisLabel: { color: '#8ec5fc', fontSize: 11, formatter: '{value}%' },
        axisTick: { show: false },
      },
      yAxis: {
        type: 'category',
        data: names,
        axisLine: { lineStyle: { color: '#00d4ff33' } },
        axisLabel: { color: '#8ec5fc', fontSize: 12 },
        axisTick: { show: false },
      },
      series: [
        {
          name: '成功率',
          type: 'bar',
          barWidth: '60%',
          data: successRates.map((value, index) => ({
            value: value.toFixed(1),
            itemStyle: {
              color: warnings[index]
                ? {
                    type: 'linear',
                    x: 0,
                    y: 0,
                    x2: 1,
                    y2: 0,
                    colorStops: [
                      { offset: 0, color: '#ff4757' },
                      { offset: 1, color: '#ff6b81' },
                    ],
                  }
                : {
                    type: 'linear',
                    x: 0,
                    y: 0,
                    x2: 1,
                    y2: 0,
                    colorStops: [
                      { offset: 0, color: '#00d4ff' },
                      { offset: 1, color: '#00ff88' },
                    ],
                  },
              borderRadius: [0, 4, 4, 0],
            },
          })),
          label: {
            show: true,
            position: 'right',
            color: '#8ec5fc',
            fontSize: 12,
            formatter: '{c}%',
          },
          markLine: {
            silent: true,
            lineStyle: { color: '#ff4757', type: 'dashed', width: 2 },
            data: [{ xAxis: 50, label: { show: true, position: 'end', color: '#ff4757', formatter: '50%警戒线' } }],
          },
        },
      ],
    };

    chart.setOption(option);
    return () => chart.dispose();
  }, [data]);

  useEffect(() => {
    fetchData();
    const cleanup = initWebSocket();
    return cleanup;
  }, [fetchData, initWebSocket]);

  useEffect(() => {
    if (data) {
      const disposers = [
        initTrendChart(),
        initTypeChart(),
        initRankChart(),
        initHeatmapChart(),
        initOrgChart(),
      ];
      return () => disposers.forEach((d) => d?.());
    }
  }, [data, initTrendChart, initTypeChart, initRankChart, initHeatmapChart, initOrgChart]);

  useEffect(() => {
    const timer = setInterval(() => setCurrentTime(new Date()), 1000);
    return () => clearInterval(timer);
  }, []);

  useEffect(() => {
    if (!isCarouselPaused) {
      const timer = setInterval(() => {
        setCarouseIndex((prev) => (prev + 1) % carouselViews.length);
      }, 30000);
      return () => clearInterval(timer);
    }
  }, [isCarouselPaused, carouselViews.length]);

  useEffect(() => {
    const handleResize = () => {
      const charts = [trendChartRef, typeChartRef, rankChartRef, heatmapChartRef, orgChartRef];
      charts.forEach((ref) => {
        if (ref.current) {
          const chart = echarts.getInstanceByDom(ref.current);
          chart?.resize();
        }
      });
    };
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  const handleScreenshot = async () => {
    if (!screenRef.current) return;
    try {
      const canvas = await html2canvas(screenRef.current, {
        backgroundColor: '#001529',
        scale: 2,
        useCORS: true,
      });
      const link = document.createElement('a');
      link.download = `数据驾驶舱_${new Date().toLocaleDateString()}.png`;
      link.href = canvas.toDataURL('image/png');
      link.click();
    } catch (error) {
      console.error('Screenshot failed:', error);
    }
  };

  const formatTime = (date: Date) => {
    return date.toLocaleTimeString('zh-CN', { hour12: false });
  };

  const formatDate = (date: Date) => {
    return date.toLocaleDateString('zh-CN', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      weekday: 'long',
    });
  };

  return (
    <div className="bigscreen-container" ref={screenRef}>
      <div className="bigscreen-header">
        <div className="header-left">
          <span className={`status-dot ${wsConnected ? 'connected' : 'disconnected'}`}></span>
          <span className="status-text">{wsConnected ? '实时数据' : '连接中...'}</span>
        </div>
        <h1 className="header-title">
          <span className="title-decoration left"></span>
          纠纷调解数据驾驶舱
          <span className="title-decoration right"></span>
        </h1>
        <div className="header-right">
          <div className="time-info">
            <div className="current-time">{formatTime(currentTime)}</div>
            <div className="current-date">{formatDate(currentTime)}</div>
          </div>
          <button className="screenshot-btn" onClick={handleScreenshot} title="截图保存">
            📷
          </button>
        </div>
      </div>

      <div className="carousel-tabs">
        {carouselViews.map((view, index) => (
          <div
            key={view.key}
            className={`carousel-tab ${carouseIndex === index ? 'active' : ''}`}
            onClick={() => {
              setCarouseIndex(index);
              setIsCarouselPaused(true);
              setTimeout(() => setIsCarouselPaused(false), 10000);
            }}
          >
            {view.label}
          </div>
        ))}
      </div>

      {data ? (
        <>
          <div className={`stats-overview ${data.overview.isWarning ? 'warning' : ''}`}>
            <div className="stat-card">
              <div className="stat-label">总案件数</div>
              <div className="stat-value">{data.overview.totalCases}</div>
              <div className="stat-suffix">件</div>
            </div>
            <div className="stat-card">
              <div className="stat-label">今日新增</div>
              <div className="stat-value highlight-blue">{data.overview.todayNew}</div>
              <div className="stat-suffix">件</div>
            </div>
            <div className="stat-card">
              <div className="stat-label">处理中</div>
              <div className="stat-value highlight-yellow">{data.overview.mediatingCases}</div>
              <div className="stat-suffix">件</div>
            </div>
            <div className="stat-card">
              <div className="stat-label">已结案</div>
              <div className="stat-value highlight-green">{data.overview.closedCases}</div>
              <div className="stat-suffix">件</div>
            </div>
            <div className={`stat-card ${data.overview.isWarning ? 'warning-card' : ''}`}>
              <div className="stat-label">调解成功率</div>
              <div className={`stat-value ${data.overview.isWarning ? 'warning-value' : 'highlight-purple'}`}>
                {data.overview.successRate}%
              </div>
              {data.overview.isWarning && <div className="warning-badge">异常</div>}
            </div>
            <div className="stat-card">
              <div className="stat-label">平均办结时长</div>
              <div className="stat-value highlight-cyan">{data.overview.avgDays}</div>
              <div className="stat-suffix">天</div>
            </div>
            <div className="stat-card warning-card">
              <div className="stat-label">超时案件</div>
              <div className={`stat-value ${data.overview.timeoutCount > 0 ? 'warning-value' : ''}`}>
                {data.overview.timeoutCount}
              </div>
              <div className="stat-suffix">件</div>
            </div>
            <div className="stat-card">
              <div className="stat-label">满意度</div>
              <div className="stat-value highlight-green">{data.overview.avgSatisfaction}</div>
              <div className="stat-suffix">分</div>
            </div>
          </div>

          <div className="charts-container">
            <div className="chart-panel panel-left">
              <div className="panel-title">
                <span className="title-bar"></span>
                纠纷趋势分析
                <span className="panel-tag">近30天</span>
              </div>
              <div ref={trendChartRef} className="chart-content"></div>
            </div>

            <div className="chart-panel panel-center">
              <div className="panel-title">
                <span className="title-bar"></span>
                案件时间分布热力图
                <span className="panel-tag">周×时段</span>
              </div>
              <div ref={heatmapChartRef} className="chart-content"></div>
            </div>

            <div className="chart-panel panel-right">
              <div className="panel-title">
                <span className="title-bar"></span>
                案件类型分布
              </div>
              <div ref={typeChartRef} className="chart-content"></div>
            </div>
          </div>

          <div className="charts-container bottom">
            <div className="chart-panel panel-left">
              <div className="panel-title">
                <span className="title-bar"></span>
                调解员工作量排行
                <span className="panel-tag">TOP 10</span>
              </div>
              <div ref={rankChartRef} className="chart-content"></div>
            </div>
            <div className="chart-panel panel-right">
              <div className="panel-title">
                <span className="title-bar"></span>
                组织调解成功率
                <span className="panel-tag warning-tag">低于50%告警</span>
              </div>
              <div ref={orgChartRef} className="chart-content"></div>
            </div>
          </div>

          <div className="bigscreen-footer">
            <span>数据更新时间：{data.updateTime}</span>
            <span className="footer-divider">|</span>
            <span>
              轮播模式：{isCarouselPaused ? '已暂停' : `${30 - (Math.floor((Date.now() / 1000) % 30))}秒后切换`}
            </span>
          </div>
        </>
      ) : (
        <div className="loading-container">
          <div className="loading-spinner"></div>
          <div className="loading-text">数据加载中...</div>
        </div>
      )}
    </div>
  );
};

export default BigScreen;
