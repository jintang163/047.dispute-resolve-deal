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
  const trendChartLargeRef = useRef<HTMLDivElement>(null);
  const typeChartRef = useRef<HTMLDivElement>(null);
  const rankChartRef = useRef<HTMLDivElement>(null);
  const rankChartLargeRef = useRef<HTMLDivElement>(null);
  const heatmapChartRef = useRef<HTMLDivElement>(null);
  const heatmapChartLargeRef = useRef<HTMLDivElement>(null);
  const orgChartRef = useRef<HTMLDivElement>(null);
  const orgChartLargeRef = useRef<HTMLDivElement>(null);

  const [data, setData] = useState<BigScreenData | null>(null);
  const [carouseIndex, setCarouseIndex] = useState(0);
  const [isCarouselPaused, setIsCarouselPaused] = useState(false);
  const [currentTime, setCurrentTime] = useState(new Date());
  const [wsConnected, setWsConnected] = useState(false);
  const [countdown, setCountdown] = useState(30);
  const wsRef = useRef<WebSocket | null>(null);

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

  const uploadScreenshot = useCallback(async (canvas: HTMLCanvasElement) => {
    try {
      const base64 = canvas.toDataURL('image/png');
      const blob = await (await fetch(base64)).blob();
      const formData = new FormData();
      formData.append('file', blob, `bigscreen_${new Date().toISOString().slice(0, 10)}.png`);
      formData.append('type', 'bigscreen_daily');

      const token = getToken();
      const response = await fetch('/api/v1/stats/bigscreen/screenshot', {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: formData,
      });
      const result = await response.json();
      console.log('Screenshot uploaded:', result);
    } catch (error) {
      console.error('Upload screenshot failed:', error);
    }
  }, []);

  const takeScreenshot = useCallback(async (upload = false) => {
    if (!screenRef.current) return null;
    try {
      const canvas = await html2canvas(screenRef.current, {
        backgroundColor: '#001529',
        scale: 2,
        useCORS: true,
        allowTaint: true,
      });

      if (upload) {
        uploadScreenshot(canvas);
      } else {
        const link = document.createElement('a');
        link.download = `数据驾驶舱_${new Date().toLocaleDateString()}.png`;
        link.href = canvas.toDataURL('image/png');
        link.click();
      }
      return canvas;
    } catch (error) {
      console.error('Screenshot failed:', error);
      return null;
    }
  }, [uploadScreenshot]);

  const initWebSocket = useCallback(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/api/v1/ws/bigscreen?token=${getToken()}`;

    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      setWsConnected(true);
      console.log('BigScreen WebSocket connected');
    };

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        if (msg.type === 'bigscreen_update' && msg.data) {
          setData((prev) => (prev ? { ...prev, ...msg.data } : msg.data));
        } else if (msg.type === 'screenshot_request') {
          console.log('Received screenshot request, taking screenshot...');
          setTimeout(() => takeScreenshot(true), 1000);
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
  }, [takeScreenshot]);

  const getTrendChartOption = (data: BigScreenData, large = false) => {
    const dates = data.trendData.map((item) => item.date?.slice(5) || '');
    const counts = data.trendData.map((item) => item.count || 0);

    return {
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
        top: large ? '10%' : '15%',
        containLabel: true,
      },
      xAxis: {
        type: 'category',
        data: dates,
        axisLine: { lineStyle: { color: '#00d4ff33' } },
        axisLabel: { color: '#8ec5fc', fontSize: large ? 12 : 10 },
        axisTick: { show: false },
      },
      yAxis: {
        type: 'value',
        splitLine: { lineStyle: { color: '#00d4ff22' } },
        axisLine: { show: false },
        axisLabel: { color: '#8ec5fc', fontSize: large ? 12 : 10 },
        axisTick: { show: false },
      },
      series: [
        {
          name: '新增案件',
          type: 'line',
          smooth: true,
          data: counts,
          lineStyle: { color: '#00d4ff', width: large ? 3 : 2 },
          areaStyle: {
            color: {
              type: 'linear',
              x: 0,
              y: 0,
              x2: 0,
              y2: 1,
              colorStops: [
                { offset: 0, color: 'rgba(0, 212, 255, 0.4)' },
                { offset: 1, color: 'rgba(0, 212, 255, 0.02)' },
              ],
            },
          },
          itemStyle: { color: '#00d4ff' },
          symbol: 'circle',
          symbolSize: large ? 6 : 4,
        },
      ],
    } as echarts.EChartsOption;
  };

  const getTypeChartOption = (data: BigScreenData) => {
    const names = data.typeStats.map((item) => item.type_name || '');
    const counts = data.typeStats.map((item) => item.count || 0);

    return {
      backgroundColor: 'transparent',
      tooltip: {
        trigger: 'item',
        backgroundColor: 'rgba(0, 20, 40, 0.9)',
        borderColor: '#00d4ff',
        textStyle: { color: '#fff' },
        formatter: '{b}: {c} ({d}%)',
      },
      legend: {
        orient: 'vertical',
        right: '5%',
        top: 'center',
        textStyle: { color: '#8ec5fc', fontSize: 11 },
        itemWidth: 10,
        itemHeight: 10,
      },
      series: [
        {
          name: '案件类型',
          type: 'pie',
          radius: ['40%', '65%'],
          center: ['35%', '50%'],
          avoidLabelOverlap: false,
          itemStyle: {
            borderRadius: 4,
            borderColor: '#001529',
            borderWidth: 2,
          },
          label: { show: false },
          emphasis: {
            label: { show: true, color: '#fff', fontSize: 12, fontWeight: 'bold' },
          },
          data: counts.map((value, index) => ({
            value,
            name: names[index],
            itemStyle: {
              color: [
                '#00d4ff',
                '#00ff88',
                '#ffd93d',
                '#ff6b6b',
                '#c56cf0',
                '#48dbfb',
                '#ff9ff3',
              ][index % 7],
            },
          })),
        },
      ],
    } as echarts.EChartsOption;
  };

  const getRankChartOption = (data: BigScreenData, large = false) => {
    const ranking = large ? data.mediatorRanking : data.mediatorRanking.slice(0, 8);
    const names = ranking.map((item) => item.real_name || '').reverse();
    const counts = ranking.map((item) => item.total_cases || 0).reverse();

    return {
      backgroundColor: 'transparent',
      tooltip: {
        trigger: 'axis',
        backgroundColor: 'rgba(0, 20, 40, 0.9)',
        borderColor: '#00d4ff',
        textStyle: { color: '#fff' },
        axisPointer: { type: 'shadow' },
        formatter: '{b}: {c}件',
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
        axisLabel: { color: '#8ec5fc', fontSize: large ? 12 : 10 },
        axisTick: { show: false },
      },
      yAxis: {
        type: 'category',
        data: names,
        axisLine: { lineStyle: { color: '#00d4ff33' } },
        axisLabel: { color: '#8ec5fc', fontSize: large ? 12 : 10 },
        axisTick: { show: false },
      },
      series: [
        {
          name: '案件数',
          type: 'bar',
          barWidth: '60%',
          data: counts,
          itemStyle: {
            color: {
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
          label: {
            show: true,
            position: 'right',
            color: '#8ec5fc',
            fontSize: large ? 12 : 10,
          },
        },
      ],
    } as echarts.EChartsOption;
  };

  const getHeatmapChartOption = (data: BigScreenData, large = false) => {
    const days = ['周一', '周二', '周三', '周四', '周五', '周六', '周日'];
    const hours = ['0时', '3时', '6时', '9时', '12时', '15时', '18时', '21时'];

    const heatData: [number, number, number][] = data.heatmapData.map((item) => [
      item.dayIndex ?? 0,
      item.hourIndex ?? 0,
      item.value ?? 0,
    ]);

    return {
      backgroundColor: 'transparent',
      tooltip: {
        position: 'top',
        backgroundColor: 'rgba(0, 20, 40, 0.9)',
        borderColor: '#00d4ff',
        textStyle: { color: '#fff' },
        formatter: (params: any) => {
          const [dayIdx, hourIdx, value] = params.data;
          return `${days[dayIdx]} ${hours[hourIdx]}: ${value}件`;
        },
      },
      grid: {
        left: '8%',
        right: '8%',
        top: '10%',
        bottom: large ? '20%' : '15%',
      },
      xAxis: {
        type: 'category',
        data: days,
        splitArea: { show: true },
        axisLine: { lineStyle: { color: '#00d4ff33' } },
        axisLabel: { color: '#8ec5fc', fontSize: large ? 13 : 11 },
        axisTick: { show: false },
      },
      yAxis: {
        type: 'category',
        data: hours,
        splitArea: { show: true },
        axisLine: { lineStyle: { color: '#00d4ff33' } },
        axisLabel: { color: '#8ec5fc', fontSize: large ? 13 : 11 },
        axisTick: { show: false },
      },
      visualMap: {
        min: 0,
        max: 15,
        calculable: true,
        orient: 'horizontal',
        left: 'center',
        bottom: '0%',
        textStyle: { color: '#8ec5fc', fontSize: large ? 12 : 10 },
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
    } as echarts.EChartsOption;
  };

  const getOrgChartOption = (data: BigScreenData, large = false) => {
    const names = data.orgStats.map((item) => item.org_name || '').reverse();
    const successRates = data.orgStats.map((item) => parseFloat(item.success_rate || '0')).reverse();
    const warnings = data.orgStats.map((item) => item.is_warning).reverse();

    return {
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
        axisLabel: { color: '#8ec5fc', fontSize: large ? 12 : 10, formatter: '{value}%' },
        axisTick: { show: false },
      },
      yAxis: {
        type: 'category',
        data: names,
        axisLine: { lineStyle: { color: '#00d4ff33' } },
        axisLabel: { color: '#8ec5fc', fontSize: large ? 12 : 10 },
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
            fontSize: large ? 12 : 10,
            formatter: '{c}%',
          },
          markLine: {
            silent: true,
            lineStyle: { color: '#ff4757', type: 'dashed', width: 2 },
            data: [
              {
                xAxis: 50,
                label: { show: true, position: 'end', color: '#ff4757', formatter: '50%警戒线' },
              },
            ],
          },
        },
      ],
    } as echarts.EChartsOption;
  };

  useEffect(() => {
    fetchData();
    const cleanup = initWebSocket();
    return cleanup;
  }, [fetchData, initWebSocket]);

  useEffect(() => {
    if (data) {
      if (trendChartRef.current) {
        const chart = echarts.init(trendChartRef.current);
        chart.setOption(getTrendChartOption(data, false));
      }
      if (typeChartRef.current) {
        const chart = echarts.init(typeChartRef.current);
        chart.setOption(getTypeChartOption(data));
      }
      if (rankChartRef.current) {
        const chart = echarts.init(rankChartRef.current);
        chart.setOption(getRankChartOption(data, false));
      }
      if (heatmapChartRef.current) {
        const chart = echarts.init(heatmapChartRef.current);
        chart.setOption(getHeatmapChartOption(data, false));
      }
      if (orgChartRef.current) {
        const chart = echarts.init(orgChartRef.current);
        chart.setOption(getOrgChartOption(data, false));
      }

      if (carouseIndex === 1 && trendChartLargeRef.current) {
        const chart = echarts.getInstanceByDom(trendChartLargeRef.current) || echarts.init(trendChartLargeRef.current);
        chart.setOption(getTrendChartOption(data, true));
      }
      if (carouseIndex === 2 && heatmapChartLargeRef.current) {
        const chart = echarts.getInstanceByDom(heatmapChartLargeRef.current) || echarts.init(heatmapChartLargeRef.current);
        chart.setOption(getHeatmapChartOption(data, true));
      }
      if (carouseIndex === 3 && rankChartLargeRef.current) {
        const chart = echarts.getInstanceByDom(rankChartLargeRef.current) || echarts.init(rankChartLargeRef.current);
        chart.setOption(getRankChartOption(data, true));
      }
      if (carouseIndex === 4 && orgChartLargeRef.current) {
        const chart = echarts.getInstanceByDom(orgChartLargeRef.current) || echarts.init(orgChartLargeRef.current);
        chart.setOption(getOrgChartOption(data, true));
      }
    }
  }, [data, carouseIndex]);

  useEffect(() => {
    const timer = setInterval(() => setCurrentTime(new Date()), 1000);
    return () => clearInterval(timer);
  }, []);

  useEffect(() => {
    setCountdown(30);
    if (!isCarouselPaused) {
      const timer = setInterval(() => {
        setCountdown((prev) => {
          if (prev <= 1) {
            setCarouseIndex((prevIdx) => (prevIdx + 1) % carouselViews.length);
            return 30;
          }
          return prev - 1;
        });
      }, 1000);
      return () => clearInterval(timer);
    }
  }, [isCarouselPaused, carouselViews.length]);

  useEffect(() => {
    const handleResize = () => {
      const refs = [
        trendChartRef,
        typeChartRef,
        rankChartRef,
        heatmapChartRef,
        orgChartRef,
        trendChartLargeRef,
        heatmapChartLargeRef,
        rankChartLargeRef,
        orgChartLargeRef,
      ];
      refs.forEach((ref) => {
        if (ref.current) {
          const chart = echarts.getInstanceByDom(ref.current);
          chart?.resize();
        }
      });
    };
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  const handleScreenshot = () => {
    takeScreenshot(false);
  };

  const handleTabClick = (index: number) => {
    setCarouseIndex(index);
    setIsCarouselPaused(true);
    setCountdown(30);
    setTimeout(() => setIsCarouselPaused(false), 10000);
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

  const renderOverview = () => (
    <>
      <div className={`stats-overview ${data?.overview.isWarning ? 'warning' : ''}`}>
        <div className="stat-card">
          <div className="stat-label">总案件数</div>
          <div className="stat-value">{data?.overview.totalCases}</div>
          <div className="stat-suffix">件</div>
        </div>
        <div className="stat-card">
          <div className="stat-label">今日新增</div>
          <div className="stat-value highlight-blue">{data?.overview.todayNew}</div>
          <div className="stat-suffix">件</div>
        </div>
        <div className="stat-card">
          <div className="stat-label">处理中</div>
          <div className="stat-value highlight-yellow">{data?.overview.mediatingCases}</div>
          <div className="stat-suffix">件</div>
        </div>
        <div className="stat-card">
          <div className="stat-label">已结案</div>
          <div className="stat-value highlight-green">{data?.overview.closedCases}</div>
          <div className="stat-suffix">件</div>
        </div>
        <div className={`stat-card ${data?.overview.isWarning ? 'warning-card' : ''}`}>
          <div className="stat-label">调解成功率</div>
          <div className={`stat-value ${data?.overview.isWarning ? 'warning-value' : 'highlight-purple'}`}>
            {data?.overview.successRate}%
          </div>
          {data?.overview.isWarning && <div className="warning-badge">异常</div>}
        </div>
        <div className="stat-card">
          <div className="stat-label">平均办结时长</div>
          <div className="stat-value highlight-cyan">{data?.overview.avgDays}</div>
          <div className="stat-suffix">天</div>
        </div>
        <div className="stat-card warning-card">
          <div className="stat-label">超时案件</div>
          <div className={`stat-value ${(data?.overview.timeoutCount ?? 0) > 0 ? 'warning-value' : ''}`}>
            {data?.overview.timeoutCount}
          </div>
          <div className="stat-suffix">件</div>
        </div>
        <div className="stat-card">
          <div className="stat-label">满意度</div>
          <div className="stat-value highlight-green">{data?.overview.avgSatisfaction}</div>
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
    </>
  );

  const renderTrendView = () => (
    <div className="carousel-view trend-view">
      <div className="chart-panel full-panel">
        <div className="panel-title large">
          <span className="title-bar"></span>
          纠纷趋势分析
          <span className="panel-tag">近30天新增案件走势</span>
        </div>
        <div ref={trendChartLargeRef} className="chart-content large"></div>
      </div>
      <div className="view-side-stats">
        <div className="side-stat-card">
          <div className="side-stat-label">月新增案件</div>
          <div className="side-stat-value">
            {data?.trendData.reduce((sum, item) => sum + (item.count || 0), 0)}
          </div>
          <div className="side-stat-suffix">件</div>
        </div>
        <div className="side-stat-card">
          <div className="side-stat-label">日均新增</div>
          <div className="side-stat-value">
            {data ? Math.round(data.trendData.reduce((sum, item) => sum + (item.count || 0), 0) / 30) : 0}
          </div>
          <div className="side-stat-suffix">件</div>
        </div>
        <div className="side-stat-card">
          <div className="side-stat-label">峰值日期</div>
          <div className="side-stat-value">
            {data?.trendData.reduce((max, item) => ((item.count || 0) > (max.count || 0) ? item : max)).date?.slice(5)}
          </div>
          <div className="side-stat-suffix">件</div>
        </div>
      </div>
    </div>
  );

  const renderHeatmapView = () => (
    <div className="carousel-view heatmap-view">
      <div className="chart-panel full-panel">
        <div className="panel-title large">
          <span className="title-bar"></span>
          案件时间分布热力图
          <span className="panel-tag">周一至周日 × 24小时分布</span>
        </div>
        <div ref={heatmapChartLargeRef} className="chart-content large"></div>
      </div>
      <div className="view-side-stats">
        <div className="side-stat-card">
          <div className="side-stat-label">高峰时段</div>
          <div className="side-stat-value">9-12点</div>
          <div className="side-stat-suffix">案件最集中</div>
        </div>
        <div className="side-stat-card">
          <div className="side-stat-label">高峰日</div>
          <div className="side-stat-value">周三</div>
          <div className="side-stat-suffix">周内峰值</div>
        </div>
        <div className="side-stat-card">
          <div className="side-stat-label">低谷时段</div>
          <div className="side-stat-value">0-6点</div>
          <div className="side-stat-suffix">案件最少</div>
        </div>
      </div>
    </div>
  );

  const renderRankingView = () => (
    <div className="carousel-view ranking-view">
      <div className="chart-panel full-panel">
        <div className="panel-title large">
          <span className="title-bar"></span>
          调解员工作量排行
          <span className="panel-tag">TOP 10 调解员</span>
        </div>
        <div ref={rankChartLargeRef} className="chart-content large"></div>
      </div>
      <div className="view-side-stats">
        <div className="side-stat-card">
          <div className="side-stat-label">办案冠军</div>
          <div className="side-stat-value highlight-gold">
            {data?.mediatorRanking[0]?.real_name || '-'}
          </div>
          <div className="side-stat-suffix">{data?.mediatorRanking[0]?.total_cases || 0}件</div>
        </div>
        <div className="side-stat-card">
          <div className="side-stat-label">平均办案量</div>
          <div className="side-stat-value">
            {data
              ? Math.round(
                  data.mediatorRanking.reduce((sum, item) => sum + (item.total_cases || 0), 0) /
                    (data.mediatorRanking.length || 1),
                )
              : 0}
          </div>
          <div className="side-stat-suffix">件/人</div>
        </div>
        <div className="side-stat-card">
          <div className="side-stat-label">参与调解人数</div>
          <div className="side-stat-value highlight-blue">
            {data?.mediatorRanking.length || 0}
          </div>
          <div className="side-stat-suffix">人</div>
        </div>
      </div>
    </div>
  );

  const renderOrgView = () => (
    <div className="carousel-view org-view">
      <div className="chart-panel full-panel">
        <div className="panel-title large">
          <span className="title-bar"></span>
          组织调解成功率
          <span className="panel-tag warning-tag">低于50%告警</span>
        </div>
        <div ref={orgChartLargeRef} className="chart-content large"></div>
      </div>
      <div className="view-side-stats">
        <div className="side-stat-card">
          <div className="side-stat-label">最高成功率</div>
          <div className="side-stat-value highlight-green">
            {data?.orgStats.reduce((max, item) => {
              const rate = parseFloat(item.success_rate || '0');
              return rate > parseFloat(max.success_rate || '0') ? item : max;
            }).success_rate}
            %
          </div>
          <div className="side-stat-suffix">
            {data?.orgStats.reduce((max, item) => {
              const rate = parseFloat(item.success_rate || '0');
              return rate > parseFloat(max.success_rate || '0') ? item : max;
            }).org_name}
          </div>
        </div>
        <div className="side-stat-card">
          <div className="side-stat-label">参与组织数</div>
          <div className="side-stat-value highlight-cyan">{data?.orgStats.length || 0}</div>
          <div className="side-stat-suffix">个组织</div>
        </div>
        <div className={`side-stat-card ${(data?.orgStats.filter((o) => o.is_warning).length ?? 0) > 0 ? 'warning' : ''}`}>
          <div className="side-stat-label">告警组织数</div>
          <div className="side-stat-value">
            {data?.orgStats.filter((o) => o.is_warning).length || 0}
          </div>
          <div className="side-stat-suffix">需关注</div>
        </div>
      </div>
    </div>
  );

  const renderCarouselContent = () => {
    switch (carouseIndex) {
      case 0:
        return renderOverview();
      case 1:
        return renderTrendView();
      case 2:
        return renderHeatmapView();
      case 3:
        return renderRankingView();
      case 4:
        return renderOrgView();
      default:
        return renderOverview();
    }
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
            onClick={() => handleTabClick(index)}
          >
            {view.label}
          </div>
        ))}
      </div>

      {data ? (
        <div className="carousel-content">{renderCarouselContent()}</div>
      ) : (
        <div className="loading-container">
          <div className="loading-spinner"></div>
          <div className="loading-text">数据加载中...</div>
        </div>
      )}

      {data && (
        <div className="bigscreen-footer">
          <span>数据更新时间：{data.updateTime}</span>
          <span className="footer-divider">|</span>
          <span>
            轮播模式：{isCarouselPaused ? '已暂停（10秒后恢复）' : `${countdown}秒后切换`}
          </span>
        </div>
      )}
    </div>
  );
};

export default BigScreen;
