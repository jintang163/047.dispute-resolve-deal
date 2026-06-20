import React, { useEffect, useRef, useState, useCallback } from 'react';
import { Card, Select, DatePicker, Button, Space, Tag, Spin, Slider, message, Tooltip, Badge } from 'antd';
import {
  PlayCircleOutlined,
  PauseCircleOutlined,
  CameraOutlined,
  ReloadOutlined,
  EnvironmentOutlined,
  FireOutlined,
  ClockCircleOutlined,
} from '@ant-design/icons';
import AMapLoader from '@amap/amap-jsapi-loader';
import html2canvas from 'html2canvas';
import dayjs, { Dayjs } from 'dayjs';
import { disputeService, HeatmapPoint, HeatmapTimelineDay, TopCommunity, HeatmapQueryParams } from '../../services/dispute';

const { RangePicker } = DatePicker;

const AMAP_KEY = 'YOUR_AMAP_KEY';
const AMAP_SECURITY_CODE = 'YOUR_AMAP_SECURITY_CODE';
const CENTER_LNG = 116.397428;
const CENTER_LAT = 39.90923;
const DEFAULT_ZOOM = 12;

const HeatMap: React.FC = () => {
  const mapContainerRef = useRef<HTMLDivElement>(null);
  const heatmapPageRef = useRef<HTMLDivElement>(null);
  const mapRef = useRef<any>(null);
  const heatmapLayerRef = useRef<any>(null);
  const topMarkersRef = useRef<any[]>([]);

  const [loading, setLoading] = useState(false);
  const [dateRange, setDateRange] = useState<[Dayjs, Dayjs]>([dayjs().subtract(7, 'day'), dayjs()]);
  const [selectedType, setSelectedType] = useState<number | undefined>(undefined);
  const [disputeTypes, setDisputeTypes] = useState<{ id: string; name: string; code: string }[]>([]);
  const [timelineData, setTimelineData] = useState<HeatmapTimelineDay[]>([]);
  const [topCommunities, setTopCommunities] = useState<TopCommunity[]>([]);
  const [currentDayIndex, setCurrentDayIndex] = useState(0);
  const [isPlaying, setIsPlaying] = useState(false);
  const [animationSpeed, setAnimationSpeed] = useState(1);
  const animationTimerRef = useRef<any>(null);
  const [totalPoints, setTotalPoints] = useState(0);

  useEffect(() => {
    fetchDisputeTypes();
    return () => {
      if (animationTimerRef.current) {
        clearInterval(animationTimerRef.current);
      }
      if (heatmapLayerRef.current) {
        heatmapLayerRef.current.setMap(null);
      }
      topMarkersRef.current.forEach((m) => m.setMap(null));
      topMarkersRef.current = [];
    };
  }, []);

  useEffect(() => {
    if (timelineData.length > 0 && currentDayIndex < timelineData.length) {
      updateHeatmapLayer(timelineData[currentDayIndex].items || []);
    }
  }, [currentDayIndex, timelineData]);

  useEffect(() => {
    if (isPlaying && timelineData.length > 0) {
      animationTimerRef.current = setInterval(() => {
        setCurrentDayIndex((prev) => {
          if (prev >= timelineData.length - 1) {
            return 0;
          }
          return prev + 1;
        });
      }, 2000 / animationSpeed);
    } else {
      if (animationTimerRef.current) {
        clearInterval(animationTimerRef.current);
        animationTimerRef.current = null;
      }
    }
    return () => {
      if (animationTimerRef.current) {
        clearInterval(animationTimerRef.current);
      }
    };
  }, [isPlaying, animationSpeed, timelineData.length]);

  const fetchDisputeTypes = async () => {
    try {
      const res = await disputeService.getTypes();
      const data = (res as any)?.data || res;
      setDisputeTypes(data || []);
    } catch {}
  };

  const initMap = useCallback(() => {
    AMapLoader.load({
      key: AMAP_KEY,
      version: '2.0',
      plugins: ['AMap.HeatMap'],
    }).then((AMap: any) => {
      (window as any)._AMapSecurityConfig = {
        securityJsCode: AMAP_SECURITY_CODE,
      };

      const map = new AMap.Map(mapContainerRef.current, {
        zoom: DEFAULT_ZOOM,
        center: [CENTER_LNG, CENTER_LAT],
        mapStyle: 'amap://styles/dark',
        resizeEnable: true,
      });

      mapRef.current = map;
      fetchAllData();
    }).catch((e: Error) => {
      message.error('地图加载失败，请检查高德地图API Key配置');
      console.error('AMap load error:', e);
    });
  }, []);

  useEffect(() => {
    if (mapContainerRef.current) {
      initMap();
    }
  }, [initMap]);

  const buildQueryParams = useCallback((): HeatmapQueryParams => {
    const params: HeatmapQueryParams = {};
    if (dateRange[0] && dateRange[1]) {
      params.startTime = dateRange[0].format('YYYY-MM-DD 00:00:00');
      params.endTime = dateRange[1].format('YYYY-MM-DD 23:59:59');
    }
    if (selectedType) {
      params.typeId = selectedType;
    }
    return params;
  }, [dateRange, selectedType]);

  const fetchAllData = useCallback(async () => {
    setLoading(true);
    try {
      const params = buildQueryParams();
      const [timelineRes, topRes] = await Promise.all([
        disputeService.getHeatmapTimeline(params),
        disputeService.getTopCommunities({ ...params, limit: 5 }),
      ]);

      const timelineData = (timelineRes as any)?.data || timelineRes || [];
      const topData = (topRes as any)?.data || topRes || [];

      setTimelineData(timelineData);
      setTopCommunities(topData);
      setCurrentDayIndex(timelineData.length - 1);

      const total = timelineData.reduce((sum: number, d: HeatmapTimelineDay) => sum + d.count, 0);
      setTotalPoints(total);

      if (timelineData.length > 0) {
        const lastDay = timelineData[timelineData.length - 1];
        updateHeatmapLayer(lastDay.items || []);
      }

      updateTopMarkers(topData);
    } catch (err) {
      message.error('数据加载失败');
      console.error('Fetch heatmap data error:', err);
    } finally {
      setLoading(false);
    }
  }, [buildQueryParams]);

  const updateHeatmapLayer = useCallback((points: HeatmapPoint[]) => {
    if (!mapRef.current) return;

    if (heatmapLayerRef.current) {
      heatmapLayerRef.current.setMap(null);
      heatmapLayerRef.current = null;
    }

    const AMap = (window as any).AMap;
    if (!AMap) return;

    const heatmapData = points
      .filter((p) => p.latitude && p.longitude && p.latitude !== 0 && p.longitude !== 0)
      .map((p) => ({
        lng: p.longitude,
        lat: p.latitude,
        count: p.count || 1,
      }));

    if (heatmapData.length === 0) return;

    const heatmap = new AMap.HeatMap(mapRef.current, {
      radius: 30,
      opacity: [0, 0.9],
      gradient: {
        0.2: '#1a9850',
        0.4: '#91cf60',
        0.6: '#fee08b',
        0.8: '#fc8d59',
        1.0: '#d73027',
      },
    });

    heatmap.setDataSet({
      data: heatmapData,
      max: Math.max(...heatmapData.map((d: any) => d.count), 10),
    });

    heatmapLayerRef.current = heatmap;
  }, []);

  const updateTopMarkers = useCallback((communities: TopCommunity[]) => {
    topMarkersRef.current.forEach((m) => m.setMap(null));
    topMarkersRef.current = [];

    if (!mapRef.current || communities.length === 0) return;

    const AMap = (window as any).AMap;
    if (!AMap) return;

    communities.forEach((c) => {
      if (!c.longitude || !c.latitude) return;

      const markerContent = document.createElement('div');
      markerContent.style.cssText = `
        position: relative;
        width: 80px;
        height: 80px;
        display: flex;
        align-items: center;
        justify-content: center;
      `;

      const ring = document.createElement('div');
      ring.style.cssText = `
        position: absolute;
        width: 80px;
        height: 80px;
        border: 3px solid #ff4d4f;
        border-radius: 50%;
        animation: pulse-ring 2s ease-out infinite;
        box-shadow: 0 0 20px rgba(255, 77, 79, 0.5);
      `;
      markerContent.appendChild(ring);

      const innerRing = document.createElement('div');
      innerRing.style.cssText = `
        position: absolute;
        width: 56px;
        height: 56px;
        border: 2px solid rgba(255, 77, 79, 0.6);
        border-radius: 50%;
        animation: pulse-ring 2s ease-out infinite 0.5s;
      `;
      markerContent.appendChild(innerRing);

      const label = document.createElement('div');
      label.style.cssText = `
        position: absolute;
        top: -32px;
        left: 50%;
        transform: translateX(-50%);
        background: linear-gradient(135deg, #ff4d4f, #cf1322);
        color: #fff;
        padding: 4px 10px;
        border-radius: 6px;
        font-size: 12px;
        font-weight: 600;
        white-space: nowrap;
        box-shadow: 0 2px 8px rgba(255, 77, 79, 0.4);
        z-index: 10;
      `;
      label.textContent = `TOP${c.rank} ${c.org_name}`;
      markerContent.appendChild(label);

      const countLabel = document.createElement('div');
      countLabel.style.cssText = `
        position: absolute;
        bottom: -24px;
        left: 50%;
        transform: translateX(-50%);
        background: rgba(0, 0, 0, 0.75);
        color: #ff4d4f;
        padding: 2px 8px;
        border-radius: 4px;
        font-size: 11px;
        font-weight: 700;
        white-space: nowrap;
      `;
      countLabel.textContent = `${c.case_count}件`;
      markerContent.appendChild(countLabel);

      const marker = new AMap.Marker({
        position: [c.longitude, c.latitude],
        content: markerContent,
        offset: new AMap.Pixel(-40, -40),
        zIndex: 200,
      });

      marker.setMap(mapRef.current);
      topMarkersRef.current.push(marker);
    });
  }, []);

  const handleSearch = () => {
    setIsPlaying(false);
    setCurrentDayIndex(0);
    fetchAllData();
  };

  const handleScreenshot = async () => {
    if (!heatmapPageRef.current) return;
    try {
      message.loading({ content: '正在生成截图...', key: 'screenshot', duration: 0 });
      const canvas = await html2canvas(heatmapPageRef.current, {
        useCORS: true,
        allowTaint: true,
        backgroundColor: '#1a1a2e',
        scale: 2,
      });
      const link = document.createElement('a');
      link.download = `矛盾纠纷热力图_${dayjs().format('YYYYMMDD_HHmmss')}.png`;
      link.href = canvas.toDataURL('image/png');
      link.click();
      message.success({ content: '截图已保存', key: 'screenshot' });
    } catch (err) {
      message.error({ content: '截图生成失败', key: 'screenshot' });
      console.error('Screenshot error:', err);
    }
  };

  const toggleAnimation = () => {
    if (timelineData.length === 0) {
      message.warning('暂无时间线数据');
      return;
    }
    setIsPlaying((prev) => !prev);
  };

  const handleDaySliderChange = (value: number) => {
    setIsPlaying(false);
    setCurrentDayIndex(value);
  };

  const currentDay = timelineData[currentDayIndex];

  return (
    <div ref={heatmapPageRef} style={{ height: '100%', display: 'flex', flexDirection: 'column', gap: 12 }}>
      <style>{`
        @keyframes pulse-ring {
          0% { transform: scale(0.8); opacity: 1; }
          100% { transform: scale(1.4); opacity: 0; }
        }
        .heatmap-controls {
          background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
          border: 1px solid rgba(255,255,255,0.08);
          border-radius: 12px;
          padding: 16px 20px;
        }
        .heatmap-controls .ant-select,
        .heatmap-controls .ant-picker {
          background: rgba(255,255,255,0.06) !important;
          border-color: rgba(255,255,255,0.15) !important;
          color: #e0e0e0 !important;
          border-radius: 8px !important;
        }
        .heatmap-controls .ant-select-selection-item,
        .heatmap-controls .ant-picker-input > input {
          color: #e0e0e0 !important;
        }
        .heatmap-controls .ant-select-arrow {
          color: rgba(255,255,255,0.4) !important;
        }
        .heatmap-controls .ant-picker-suffix {
          color: rgba(255,255,255,0.4) !important;
        }
        .heatmap-controls .ant-btn {
          border-radius: 8px !important;
          font-weight: 500 !important;
        }
        .animation-panel {
          background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
          border: 1px solid rgba(255,255,255,0.08);
          border-radius: 12px;
          padding: 12px 20px;
        }
        .animation-panel .ant-slider-track {
          background: linear-gradient(90deg, #1677ff, #722ed1) !important;
        }
        .animation-panel .ant-slider-handle {
          border-color: #722ed1 !important;
        }
        .stat-badge {
          background: rgba(22,119,255,0.15);
          border: 1px solid rgba(22,119,255,0.3);
          border-radius: 8px;
          padding: 4px 12px;
          font-size: 13px;
          color: #91caff;
          display: inline-flex;
          align-items: center;
          gap: 6px;
        }
        .top5-legend {
          position: absolute;
          top: 12px;
          right: 12px;
          z-index: 200;
          background: rgba(26,26,46,0.92);
          backdrop-filter: blur(8px);
          border: 1px solid rgba(255,77,79,0.25);
          border-radius: 10px;
          padding: 12px 16px;
          min-width: 180px;
        }
        .top5-legend-title {
          color: #ff4d4f;
          font-size: 13px;
          font-weight: 600;
          margin-bottom: 8px;
          display: flex;
          align-items: center;
          gap: 6px;
        }
        .top5-legend-item {
          color: #e0e0e0;
          font-size: 12px;
          padding: 4px 0;
          display: flex;
          align-items: center;
          justify-content: space-between;
          border-bottom: 1px solid rgba(255,255,255,0.05);
        }
        .top5-legend-item:last-child {
          border-bottom: none;
        }
      `}</style>

      <div className="heatmap-controls">
        <Space size={12} wrap style={{ width: '100%', justifyContent: 'space-between' }}>
          <Space size={12} wrap>
            <RangePicker
              value={dateRange}
              onChange={(dates) => {
                if (dates && dates[0] && dates[1]) {
                  setDateRange([dates[0] as Dayjs, dates[1] as Dayjs]);
                }
              }}
              style={{ width: 300 }}
              placeholder={['开始日期', '结束日期']}
            />
            <Select
              placeholder="纠纷类型"
              allowClear
              style={{ width: 160 }}
              value={selectedType}
              onChange={(v) => setSelectedType(v)}
              options={disputeTypes.map((t) => ({ label: t.name, value: parseInt(t.id) }))}
            />
            <Button type="primary" icon={<ReloadOutlined />} onClick={handleSearch} loading={loading}>
              查询
            </Button>
          </Space>
          <Space size={8}>
            <span className="stat-badge">
              <EnvironmentOutlined />
              {totalPoints} 个案件
            </span>
            <Tooltip title="一键截图保存热力图">
              <Button
                type="primary"
                danger
                icon={<CameraOutlined />}
                onClick={handleScreenshot}
                style={{
                  background: 'linear-gradient(135deg, #ff4d4f, #cf1322)',
                  border: 'none',
                }}
              >
                截图保存
              </Button>
            </Tooltip>
          </Space>
        </Space>
      </div>

      <div className="animation-panel">
        <Space size={16} style={{ width: '100%' }} align="center">
          <Button
            type={isPlaying ? 'default' : 'primary'}
            shape="circle"
            icon={isPlaying ? <PauseCircleOutlined /> : <PlayCircleOutlined />}
            onClick={toggleAnimation}
            size="large"
            style={{
              background: isPlaying ? 'rgba(255,255,255,0.08)' : 'linear-gradient(135deg, #1677ff, #722ed1)',
              border: 'none',
              color: '#fff',
              boxShadow: '0 2px 8px rgba(22,119,255,0.3)',
            }}
          />
          <div style={{ flex: 1 }}>
            <Slider
              min={0}
              max={Math.max(timelineData.length - 1, 0)}
              value={currentDayIndex}
              onChange={handleDaySliderChange}
              marks={timelineData.reduce((acc: Record<number, string>, d: HeatmapTimelineDay, i: number) => {
                if (timelineData.length <= 7 || i % Math.ceil(timelineData.length / 7) === 0 || i === timelineData.length - 1) {
                  acc[i] = dayjs(d.date).format('MM/DD');
                }
                return acc;
              }, {})}
              tooltip={{
                formatter: (v) => {
                  if (v !== undefined && timelineData[v]) {
                    return `${timelineData[v].date} (${timelineData[v].count}件)`;
                  }
                  return '';
                },
              }}
            />
          </div>
          <Space size={4} align="center">
            <span style={{ color: 'rgba(255,255,255,0.5)', fontSize: 12 }}>速度</span>
            <Select
              value={animationSpeed}
              onChange={setAnimationSpeed}
              size="small"
              style={{ width: 72 }}
              options={[
                { label: '0.5x', value: 0.5 },
                { label: '1x', value: 1 },
                { label: '2x', value: 2 },
                { label: '4x', value: 4 },
              ]}
            />
          </Space>
          {currentDay && (
            <Tag
              color="blue"
              style={{
                margin: 0,
                background: 'rgba(22,119,255,0.15)',
                border: '1px solid rgba(22,119,255,0.3)',
                color: '#91caff',
                borderRadius: 6,
              }}
            >
              <ClockCircleOutlined style={{ marginRight: 4 }} />
              {currentDay.date} | {currentDay.count}件纠纷
            </Tag>
          )}
        </Space>
      </div>

      <div style={{ flex: 1, position: 'relative', borderRadius: 12, overflow: 'hidden', border: '1px solid rgba(255,255,255,0.08)' }}>
        <Spin spinning={loading} tip="数据加载中..." style={{ position: 'absolute', top: '50%', left: '50%', zIndex: 300 }}>
          <div ref={mapContainerRef} style={{ width: '100%', height: '100%', minHeight: 500 }} />
        </Spin>

        {topCommunities.length > 0 && (
          <div className="top5-legend">
            <div className="top5-legend-title">
              <FireOutlined />
              高发区域 TOP5
            </div>
            {topCommunities.map((c) => (
              <div key={c.org_id} className="top5-legend-item">
                <span>
                  <Badge
                    count={c.rank}
                    style={{
                      backgroundColor: c.rank <= 3 ? '#ff4d4f' : '#faad14',
                      marginRight: 8,
                      fontSize: 10,
                      minWidth: 18,
                      height: 18,
                      lineHeight: '18px',
                    }}
                  />
                  {c.org_name}
                </span>
                <span style={{ color: '#ff4d4f', fontWeight: 600 }}>{c.case_count}件</span>
              </div>
            ))}
          </div>
        )}

        <div
          style={{
            position: 'absolute',
            bottom: 12,
            left: 12,
            zIndex: 200,
            background: 'rgba(26,26,46,0.92)',
            backdropFilter: 'blur(8px)',
            border: '1px solid rgba(255,255,255,0.08)',
            borderRadius: 10,
            padding: '10px 16px',
          }}
        >
          <div style={{ fontSize: 12, color: 'rgba(255,255,255,0.6)', marginBottom: 6 }}>纠纷密度</div>
          <div
            style={{
              width: 180,
              height: 12,
              borderRadius: 6,
              background: 'linear-gradient(90deg, #1a9850, #91cf60, #fee08b, #fc8d59, #d73027)',
            }}
          />
          <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: 4 }}>
            <span style={{ fontSize: 11, color: 'rgba(255,255,255,0.4)' }}>低</span>
            <span style={{ fontSize: 11, color: 'rgba(255,255,255,0.4)' }}>高</span>
          </div>
        </div>
      </div>
    </div>
  );
};

export default HeatMap;
