import React, { useEffect, useRef, useState, useCallback, useMemo } from 'react';
import {
  Select,
  DatePicker,
  Button,
  Space,
  Tag,
  Spin,
  Slider,
  message,
  Tooltip,
  Badge,
  Drawer,
  Table,
  Pagination,
  Empty,
  Alert,
  Typography,
  Row,
  Col,
  Divider,
} from 'antd';
import {
  PlayCircleOutlined,
  PauseCircleOutlined,
  CameraOutlined,
  ReloadOutlined,
  EnvironmentOutlined,
  FireOutlined,
  ClockCircleOutlined,
  SearchOutlined,
  CloseOutlined,
  FileTextOutlined,
  EyeOutlined,
  InfoCircleOutlined,
} from '@ant-design/icons';
import AMapLoader from '@amap/amap-jsapi-loader';
import html2canvas from 'html2canvas';
import dayjs, { Dayjs } from 'dayjs';
import {
  disputeService,
  HeatmapPoint,
  HeatmapTimelineDay,
  TopCommunity,
  HeatmapQueryParams,
  DrilldownCase,
  DrilldownParams,
  AmapConfig,
} from '../../services/dispute';

const { RangePicker } = DatePicker;
const { Title, Text } = Typography;

const FALLBACK_AMAP_KEY = 'YOUR_AMAP_KEY';
const FALLBACK_AMAP_SECURITY_CODE = 'YOUR_AMAP_SECURITY_CODE';

const HeatMap: React.FC = () => {
  const mapContainerRef = useRef<HTMLDivElement>(null);
  const heatmapPageRef = useRef<HTMLDivElement>(null);
  const mapRef = useRef<any>(null);
  const heatmapLayerRef = useRef<any>(null);
  const topMarkersRef = useRef<any[]>([]);
  const rectangleRefs = useRef<any[]>([]);
  const clickMarkerRef = useRef<any>(null);

  const [loading, setLoading] = useState(false);
  const [configLoading, setConfigLoading] = useState(true);
  const [mapReady, setMapReady] = useState(false);
  const [amapCfg, setAmapCfg] = useState<AmapConfig | null>(null);

  const [dateRange, setDateRange] = useState<[Dayjs, Dayjs]>([
    dayjs().subtract(7, 'day'),
    dayjs(),
  ]);
  const [selectedType, setSelectedType] = useState<number | undefined>(undefined);
  const [disputeTypes, setDisputeTypes] = useState<
    { id: string; name: string; code: string }[]
  >([]);
  const [timelineData, setTimelineData] = useState<HeatmapTimelineDay[]>([]);
  const [topCommunities, setTopCommunities] = useState<TopCommunity[]>([]);
  const [currentDayIndex, setCurrentDayIndex] = useState(0);
  const [isPlaying, setIsPlaying] = useState(false);
  const [animationSpeed, setAnimationSpeed] = useState(1);
  const animationTimerRef = useRef<any>(null);
  const [totalPoints, setTotalPoints] = useState(0);

  const [drawerOpen, setDrawerOpen] = useState(false);
  const [drawerTitle, setDrawerTitle] = useState('区域案件列表');
  const [drawerLoading, setDrawerLoading] = useState(false);
  const [drilldownList, setDrilldownList] = useState<DrilldownCase[]>([]);
  const [drilldownTotal, setDrilldownTotal] = useState(0);
  const [drilldownPage, setDrilldownPage] = useState(1);
  const [drilldownPageSize, setDrilldownPageSize] = useState(10);
  const drilldownParamsRef = useRef<DrilldownParams>({});

  useEffect(() => {
    fetchDisputeTypes();
    fetchAmapConfig();
    return () => {
      if (animationTimerRef.current) clearInterval(animationTimerRef.current);
      if (heatmapLayerRef.current) heatmapLayerRef.current.setMap(null);
      topMarkersRef.current.forEach((m) => m.setMap(null));
      topMarkersRef.current = [];
      rectangleRefs.current.forEach((r) => r.setMap(null));
      rectangleRefs.current = [];
      if (clickMarkerRef.current) clickMarkerRef.current.setMap(null);
      if (mapRef.current) mapRef.current.destroy();
    };
  }, []);

  const fetchDisputeTypes = async () => {
    try {
      const res = await disputeService.getTypes();
      const data = (res as any)?.data || res;
      setDisputeTypes(data || []);
    } catch {}
  };

  const fetchAmapConfig = async () => {
    setConfigLoading(true);
    try {
      const res = await disputeService.getAmapConfig();
      const cfg = (res as any)?.data || res;
      if (cfg && cfg.web_key && cfg.web_key !== 'YOUR_AMAP_WEB_JS_KEY') {
        setAmapCfg(cfg);
      } else {
        setAmapCfg({
          web_key: FALLBACK_AMAP_KEY,
          security_code: FALLBACK_AMAP_SECURITY_CODE,
          default_city: cfg?.default_city || '北京市',
          default_lng: cfg?.default_lng || '116.397428',
          default_lat: cfg?.default_lat || '39.90923',
          default_zoom: cfg?.default_zoom || 12,
          cluster_radius: cfg?.cluster_radius || 500,
          grid_level: cfg?.grid_level || 6,
          use_spatial: cfg?.use_spatial || false,
        });
      }
    } catch {
      setAmapCfg({
        web_key: FALLBACK_AMAP_KEY,
        security_code: FALLBACK_AMAP_SECURITY_CODE,
        default_city: '北京市',
        default_lng: '116.397428',
        default_lat: '39.90923',
        default_zoom: 12,
        cluster_radius: 500,
        grid_level: 6,
        use_spatial: false,
      });
    } finally {
      setConfigLoading(false);
    }
  };

  useEffect(() => {
    if (!configLoading && amapCfg && mapContainerRef.current && !mapRef.current) {
      initMap();
    }
  }, [configLoading, amapCfg]);

  const initMap = useCallback(() => {
    if (!amapCfg) return;
    const securityCode =
      amapCfg.security_code && amapCfg.security_code !== 'YOUR_AMAP_SECURITY_CODE'
        ? amapCfg.security_code
        : FALLBACK_AMAP_SECURITY_CODE;
    (window as any)._AMapSecurityConfig = { securityJsCode: securityCode };

    const mapKey =
      amapCfg.web_key && amapCfg.web_key !== 'YOUR_AMAP_KEY' &&
      amapCfg.web_key !== 'YOUR_AMAP_WEB_JS_KEY'
        ? amapCfg.web_key
        : FALLBACK_AMAP_KEY;

    AMapLoader.load({
      key: mapKey,
      version: '2.0',
      plugins: ['AMap.HeatMap', 'AMap.Polygon', 'AMap.Rectangle', 'AMap.ToolBar'],
    })
      .then((AMap: any) => {
        const map = new AMap.Map(mapContainerRef.current, {
          zoom: amapCfg.default_zoom || 12,
          center: [
            parseFloat(amapCfg.default_lng) || 116.397428,
            parseFloat(amapCfg.default_lat) || 39.90923,
          ],
          mapStyle: 'amap://styles/dark',
          resizeEnable: true,
        });

        map.addControl(new AMap.ToolBar({ position: 'RB' }));

        map.on('click', (e: any) => {
          handleMapClick(e.lnglat.lng, e.lnglat.lat);
        });

        mapRef.current = map;
        setMapReady(true);
        fetchAllData();
      })
      .catch((e: Error) => {
        message.error('地图加载失败，请检查高德地图API Key配置');
        console.error('AMap load error:', e);
        setMapReady(true);
      });
  }, [amapCfg]);

  const buildQueryParams = useCallback((): HeatmapQueryParams => {
    const params: HeatmapQueryParams = { useSpatial: amapCfg?.use_spatial };
    if (dateRange[0] && dateRange[1]) {
      params.startTime = dateRange[0].format('YYYY-MM-DD 00:00:00');
      params.endTime = dateRange[1].format('YYYY-MM-DD 23:59:59');
    }
    if (selectedType) {
      params.typeId = selectedType;
    }
    return params;
  }, [dateRange, selectedType, amapCfg]);

  const fetchAllData = useCallback(async () => {
    setLoading(true);
    try {
      const params = buildQueryParams();
      const [timelineRes, topRes] = await Promise.all([
        disputeService.getHeatmapTimeline(params),
        disputeService.getTopCommunities({ ...params, limit: 5 }),
      ]);

      const timelineList = (timelineRes as any)?.data || timelineRes || [];
      const topList = (topRes as any)?.data || topRes || [];

      setTimelineData(timelineList);
      setTopCommunities(topList);
      setCurrentDayIndex(Math.max(0, timelineList.length - 1));

      const total = timelineList.reduce(
        (sum: number, d: HeatmapTimelineDay) => sum + (d.count || 0),
        0,
      );
      setTotalPoints(total);
      updateTopMarkers(topList);
    } catch (err) {
      message.error('数据加载失败');
      console.error('Fetch heatmap data error:', err);
    } finally {
      setLoading(false);
    }
  }, [buildQueryParams]);

  const cumulativePoints = useMemo(() => {
    const map = new Map<number, HeatmapPoint>();
    for (let i = 0; i <= Math.min(currentDayIndex, timelineData.length - 1); i++) {
      const day = timelineData[i];
      if (day && day.items) {
        for (const p of day.items) {
          if (p && typeof p.id === 'number') {
            map.set(p.id, p);
          }
        }
      }
    }
    return Array.from(map.values());
  }, [timelineData, currentDayIndex]);

  useEffect(() => {
    if (mapRef.current) {
      updateHeatmapLayer(cumulativePoints);
    }
  }, [cumulativePoints, mapReady]);

  const updateHeatmapLayer = useCallback((points: HeatmapPoint[]) => {
    if (!mapRef.current) return;
    if (heatmapLayerRef.current) {
      heatmapLayerRef.current.setMap(null);
      heatmapLayerRef.current = null;
    }
    const AMap = (window as any).AMap;
    if (!AMap) return;

    const heatmapData = points
      .filter(
        (p) =>
          p.latitude &&
          p.longitude &&
          p.latitude !== 0 &&
          p.longitude !== 0,
      )
      .map((p) => ({
        lng: p.longitude,
        lat: p.latitude,
        count: p.count || 1,
      }));

    if (heatmapData.length === 0) return;

    const heatmap = new AMap.HeatMap(mapRef.current, {
      radius: 32,
      opacity: [0.05, 0.92],
      gradient: {
        0.15: '#1a9850',
        0.35: '#91cf60',
        0.55: '#fee08b',
        0.75: '#fc8d59',
        1.0: '#d73027',
      },
    });

    const counts = heatmapData.map((d: any) => d.count || 1);
    heatmap.setDataSet({
      data: heatmapData,
      max: Math.max(...counts, 8),
    });

    heatmapLayerRef.current = heatmap;
  }, []);

  const updateTopMarkers = useCallback((communities: TopCommunity[]) => {
    topMarkersRef.current.forEach((m) => m.setMap(null));
    topMarkersRef.current = [];
    rectangleRefs.current.forEach((r) => r.setMap(null));
    rectangleRefs.current = [];

    if (!mapRef.current || communities.length === 0) return;
    const AMap = (window as any).AMap;
    if (!AMap) return;

    communities.forEach((c, idx) => {
      const lng =
        typeof c.longitude === 'number'
          ? c.longitude
          : parseFloat(String(c.longitude));
      const lat =
        typeof c.latitude === 'number'
          ? c.latitude
          : parseFloat(String(c.latitude));
      if (!lng || !lat) return;

      const radius = c.bbox && c.bbox.east && c.bbox.west
        ? Math.max((c.bbox.east - c.bbox.west) / 2 * 111320 * Math.cos(lat * Math.PI / 180), 300)
        : c.radius_meters || 500;

      if (c.bbox && c.bbox.east > c.bbox.west && c.bbox.north > c.bbox.south) {
        const rectangle = new AMap.Rectangle({
          bounds: new AMap.Bounds(
            [c.bbox.west, c.bbox.south],
            [c.bbox.east, c.bbox.north],
          ),
          strokeColor: '#ff4d4f',
          strokeWeight: 2,
          strokeOpacity: 0.9,
          strokeDasharray: [6, 4],
          fillColor: '#ff4d4f',
          fillOpacity: 0.06,
          zIndex: 100,
          cursor: 'pointer',
        });
        rectangle.on('click', () => handleClusterClick(c));
        rectangle.setMap(mapRef.current);
        rectangleRefs.current.push(rectangle);
      }

      const markerContent = document.createElement('div');
      markerContent.style.cssText = 'position:relative;width:100px;height:100px;display:flex;align-items:center;justify-content:center;';

      const ringCount = 3;
      for (let r = 0; r < ringCount; r++) {
        const ring = document.createElement('div');
        const size = 70 - r * 14;
        ring.style.cssText = `
          position:absolute;
          width:${size}px;height:${size}px;
          border:${2 + r}px solid rgba(255,77,79,${0.9 - r * 0.25});
          border-radius:50%;
          animation:pulse-ring 2s ease-out infinite;
          animation-delay:${r * 0.4}s;
          box-shadow:0 0 ${16 + r * 4}px rgba(255,77,79,${0.45 - r * 0.1});
        `;
        markerContent.appendChild(ring);
      }

      const rank = document.createElement('div');
      rank.style.cssText = `
        width:42px;height:42px;border-radius:50%;
        background:linear-gradient(135deg,#ff4d4f,#cf1322);
        color:#fff;display:flex;align-items:center;justify-content:center;
        font-weight:800;font-size:18px;
        box-shadow:0 4px 16px rgba(207,19,34,0.5);
        z-index:10;border:2px solid rgba(255,255,255,0.35);
      `;
      rank.textContent = `T${c.rank}`;
      markerContent.appendChild(rank);

      const label = document.createElement('div');
      label.style.cssText = `
        position:absolute;top:-40px;left:50%;transform:translateX(-50%);
        background:linear-gradient(135deg,#cf1322,#8b0000);
        color:#fff;padding:5px 12px;border-radius:8px;
        font-size:12px;font-weight:600;white-space:nowrap;
        box-shadow:0 4px 14px rgba(255,77,79,0.45);z-index:20;
        border:1px solid rgba(255,255,255,0.2);
      `;
      label.textContent = `TOP${c.rank} · ${c.cluster_name || c.org_name}`;
      markerContent.appendChild(label);

      const countLabel = document.createElement('div');
      countLabel.style.cssText = `
        position:absolute;bottom:-30px;left:50%;transform:translateX(-50%);
        background:rgba(0,0,0,0.78);
        color:#ff7875;padding:3px 10px;border-radius:6px;
        font-size:11px;font-weight:700;white-space:nowrap;z-index:15;
        border:1px solid rgba(255,77,79,0.3);
      `;
      countLabel.textContent = `${c.case_count} 件纠纷 · ${Math.round(radius)}m范围`;
      markerContent.appendChild(countLabel);

      const marker = new AMap.Marker({
        position: [lng, lat],
        content: markerContent,
        offset: new AMap.Pixel(-50, -50),
        zIndex: 200 + idx,
        cursor: 'pointer',
      });
      marker.on('click', () => handleClusterClick(c));
      marker.setMap(mapRef.current);
      topMarkersRef.current.push(marker);
    });
  }, []);

  const handleClusterClick = (c: TopCommunity) => {
    const lng =
      typeof c.longitude === 'number'
        ? c.longitude
        : parseFloat(String(c.longitude));
    const lat =
      typeof c.latitude === 'number'
        ? c.latitude
        : parseFloat(String(c.latitude));

    const params: DrilldownParams = {
      ...buildQueryParams(),
      centerLng: lng,
      centerLat: lat,
      radiusMeters: c.bbox
        ? Math.max(
            (c.bbox.east - c.bbox.west) / 2 * 111320 * Math.cos(lat * Math.PI / 180),
            300,
          )
        : c.radius_meters || 500,
    };
    if (c.bbox) {
      params.westLng = c.bbox.west;
      params.southLat = c.bbox.south;
      params.eastLng = c.bbox.east;
      params.northLat = c.bbox.north;
    }
    if (c.cluster_id) params.clusterId = c.cluster_id;

    setDrawerTitle(`TOP${c.rank} · ${c.cluster_name || c.org_name} · ${c.case_count}件`);
    openDrilldown(params);

    if (mapRef.current) {
      mapRef.current.setCenter([lng, lat]);
      mapRef.current.setZoom(Math.max(mapRef.current.getZoom(), 14));
    }
  };

  const handleMapClick = (lng: number, lat: number) => {
    const AMap = (window as any).AMap;
    if (clickMarkerRef.current) clickMarkerRef.current.setMap(null);
    if (AMap && mapRef.current) {
      const content = document.createElement('div');
      content.style.cssText = `
        width:16px;height:16px;border-radius:50%;
        background:rgba(22,119,255,0.9);
        border:2px solid #fff;
        box-shadow:0 0 0 4px rgba(22,119,255,0.3);
      `;
      const marker = new AMap.Marker({
        position: [lng, lat],
        offset: new AMap.Pixel(-8, -8),
        content,
        zIndex: 300,
      });
      marker.on('click', () => {
        marker.setMap(null);
        clickMarkerRef.current = null;
      });
      marker.setMap(mapRef.current);
      clickMarkerRef.current = marker;
    }

    const params: DrilldownParams = {
      ...buildQueryParams(),
      centerLng: lng,
      centerLat: lat,
      radiusMeters: 300,
    };
    setDrawerTitle(`点击位置附近案件 · ${lng.toFixed(5)},${lat.toFixed(5)}`);
    openDrilldown(params);
  };

  const openDrilldown = (params: DrilldownParams) => {
    drilldownParamsRef.current = params;
    setDrilldownPage(1);
    setDrawerOpen(true);
    loadDrilldown(1, drilldownPageSize, params);
  };

  const loadDrilldown = async (
    page: number,
    pageSize: number,
    params?: DrilldownParams,
  ) => {
    setDrawerLoading(true);
    try {
      const p: DrilldownParams = {
        ...(params || drilldownParamsRef.current),
        page,
        pageSize,
      };
      const res = await disputeService.getHeatmapDrilldown(p);
      const data = (res as any)?.data || res;
      setDrilldownList(data?.list || []);
      setDrilldownTotal(data?.total || 0);
      setDrilldownPage(page);
      setDrilldownPageSize(pageSize);
    } catch (e) {
      message.error('案件列表加载失败');
      console.error('Drilldown error:', e);
    } finally {
      setDrawerLoading(false);
    }
  };

  const handleCurrentBBoxDrilldown = () => {
    if (!mapRef.current) {
      message.warning('地图尚未加载');
      return;
    }
    const bounds = mapRef.current.getBounds();
    if (!bounds) {
      message.warning('无法获取可视区域');
      return;
    }
    const sw = bounds.getSouthWest();
    const ne = bounds.getNorthEast();
    const params: DrilldownParams = {
      ...buildQueryParams(),
      westLng: sw.lng,
      southLat: sw.lat,
      eastLng: ne.lng,
      northLat: ne.lat,
    };
    const center = bounds.getCenter();
    setDrawerTitle(
      `当前可视区域案件 · 中心 ${center.lng.toFixed(4)},${center.lat.toFixed(4)}`,
    );
    openDrilldown(params);
  };

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
        backgroundColor: '#0f0f1e',
        scale: 2,
        ignoreElements: (el) => el.tagName === 'CANVAS' && !el.classList.contains('amap-layer'),
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

  useEffect(() => {
    if (isPlaying && timelineData.length > 0) {
      animationTimerRef.current = setInterval(() => {
        setCurrentDayIndex((prev) => {
          if (prev >= timelineData.length - 1) return 0;
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
      if (animationTimerRef.current) clearInterval(animationTimerRef.current);
    };
  }, [isPlaying, animationSpeed, timelineData.length]);

  const handleDaySliderChange = (value: number) => {
    setIsPlaying(false);
    setCurrentDayIndex(value);
  };

  const cumulativeCount = useMemo(
    () =>
      timelineData
        .slice(0, Math.min(currentDayIndex, timelineData.length - 1) + 1)
        .reduce((s, d) => s + (d.count || 0), 0),
    [timelineData, currentDayIndex],
  );

  const currentDay = timelineData[currentDayIndex];

  const caseColumns = [
    {
      title: '案件编号',
      dataIndex: 'case_no',
      key: 'case_no',
      width: 140,
      render: (v: string, r: DrilldownCase) => (
        <Space>
          <FileTextOutlined style={{ color: '#1677ff' }} />
          <Text copyable={{ text: v }} style={{ fontSize: 12 }}>
            {v}
          </Text>
        </Space>
      ),
    },
    {
      title: '案件标题',
      dataIndex: 'title',
      key: 'title',
      ellipsis: true,
      render: (v: string, r: DrilldownCase) => (
        <Tooltip title={v}>
          <span>{v || '-'}</span>
        </Tooltip>
      ),
    },
    {
      title: '纠纷类型',
      dataIndex: 'type_name',
      key: 'type_name',
      width: 110,
      render: (v: string) =>
        v ? (
          <Tag color="blue" style={{ margin: 0 }}>
            {v}
          </Tag>
        ) : (
          '-'
        ),
    },
    {
      title: '申请人',
      dataIndex: 'applicant_name',
      key: 'applicant_name',
      width: 90,
      render: (v: string) => v || '-',
    },
    {
      title: '被申请人',
      dataIndex: 'respondent_name',
      key: 'respondent_name',
      width: 90,
      render: (v: string) => v || '-',
    },
    {
      title: '所属组织',
      dataIndex: 'org_name',
      key: 'org_name',
      width: 130,
      ellipsis: true,
      render: (v: string) =>
        v ? (
          <Tag color="purple" style={{ margin: 0 }}>
            {v}
          </Tag>
        ) : (
          '-'
        ),
    },
    {
      title: '状态',
      dataIndex: 'status_name',
      key: 'status_name',
      width: 90,
      render: (v: string, r: DrilldownCase) => {
        const color =
          r.status === 50
            ? 'success'
            : r.status >= 20 && r.status < 50
              ? 'processing'
              : r.status === 10
                ? 'warning'
                : 'default';
        return <Tag color={color as any}>{v || '-'}</Tag>;
      },
    },
    {
      title: '案发地址',
      dataIndex: 'event_address',
      key: 'event_address',
      width: 160,
      ellipsis: true,
      render: (v: string, r: DrilldownCase) => (
        <Tooltip title={v}>
          <Space size={4}>
            <EnvironmentOutlined style={{ color: '#52c41a' }} />
            <span style={{ fontSize: 12 }}>{v || '-'}</span>
          </Space>
        </Tooltip>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 160,
      render: (v: string) => v?.slice(0, 19) || '-',
    },
  ];

  return (
    <div
      ref={heatmapPageRef}
      style={{
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        gap: 12,
        padding: 4,
      }}
    >
      <style>{`
        @keyframes pulse-ring {
          0% { transform: scale(0.7); opacity: 1; }
          100% { transform: scale(1.6); opacity: 0; }
        }
        .heatmap-controls {
          background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
          border: 1px solid rgba(255,255,255,0.08);
          border-radius: 14px;
          padding: 16px 20px;
          box-shadow: 0 4px 20px rgba(0,0,0,0.3);
        }
        .heatmap-controls .ant-select,
        .heatmap-controls .ant-picker {
          background: rgba(255,255,255,0.06) !important;
          border-color: rgba(255,255,255,0.15) !important;
          color: #e0e0e0 !important;
          border-radius: 10px !important;
        }
        .heatmap-controls .ant-select-selector,
        .heatmap-controls .ant-picker {
          border-radius: 10px !important;
        }
        .heatmap-controls .ant-select-selection-item,
        .heatmap-controls .ant-picker-input > input {
          color: #e0e0e0 !important;
        }
        .heatmap-controls .ant-select-arrow,
        .heatmap-controls .ant-picker-suffix {
          color: rgba(255,255,255,0.4) !important;
        }
        .heatmap-controls .ant-btn {
          border-radius: 10px !important;
          font-weight: 500 !important;
        }
        .animation-panel {
          background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
          border: 1px solid rgba(255,255,255,0.08);
          border-radius: 14px;
          padding: 14px 20px;
          box-shadow: 0 4px 20px rgba(0,0,0,0.3);
        }
        .animation-panel .ant-slider-track {
          background: linear-gradient(90deg, #1677ff, #722ed1) !important;
        }
        .animation-panel .ant-slider-handle {
          border-color: #722ed1 !important;
        }
        .stat-badge {
          background: rgba(22,119,255,0.12);
          border: 1px solid rgba(22,119,255,0.3);
          border-radius: 10px;
          padding: 5px 14px;
          font-size: 13px;
          color: #91caff;
          display: inline-flex;
          align-items: center;
          gap: 6px;
          font-weight: 500;
        }
        .stat-badge.success {
          background: rgba(82,196,26,0.12);
          border-color: rgba(82,196,26,0.3);
          color: #95de64;
        }
        .stat-badge.danger {
          background: rgba(255,77,79,0.12);
          border-color: rgba(255,77,79,0.3);
          color: #ff7875;
        }
        .top5-legend {
          position: absolute;
          top: 14px;
          right: 14px;
          z-index: 200;
          background: rgba(26,26,46,0.92);
          backdrop-filter: blur(10px);
          border: 1px solid rgba(255,77,79,0.25);
          border-radius: 12px;
          padding: 14px 18px;
          min-width: 210px;
          box-shadow: 0 6px 20px rgba(0,0,0,0.4);
        }
        .top5-legend-title {
          color: #ff4d4f;
          font-size: 13px;
          font-weight: 700;
          margin-bottom: 10px;
          display: flex;
          align-items: center;
          gap: 6px;
          padding-bottom: 8px;
          border-bottom: 1px solid rgba(255,77,79,0.15);
        }
        .top5-legend-item {
          color: #e0e0e0;
          font-size: 12px;
          padding: 7px 0;
          display: flex;
          align-items: center;
          justify-content: space-between;
          border-bottom: 1px solid rgba(255,255,255,0.05);
          cursor: pointer;
          border-radius: 4px;
          padding-left: 6px;
          padding-right: 6px;
          transition: all 0.15s;
        }
        .top5-legend-item:hover {
          background: rgba(255,77,79,0.08);
        }
        .top5-legend-item:last-child {
          border-bottom: none;
        }
        .drill-btn {
          position: absolute;
          top: 14px;
          left: 14px;
          z-index: 200;
        }
      `}</style>

      {configLoading && (
        <Spin tip="加载地图配置...">
          <div style={{ height: 100 }} />
        </Spin>
      )}

      {!configLoading &&
        amapCfg &&
        amapCfg.web_key === FALLBACK_AMAP_KEY && (
          <Alert
            message={
              <Space>
                <InfoCircleOutlined />
                <span>
                  高德地图密钥未配置，当前使用占位符。请在后端配置文件
                  <Text code style={{ margin: '0 4px' }}>
                    config.yaml
                  </Text>
                  中设置 <Text code>amap.web_key</Text> 与{' '}
                  <Text code>amap.security_code</Text>
                </span>
              </Space>
            }
            type="warning"
            showIcon={false}
            style={{ borderRadius: 12 }}
          />
        )}

      <div className="heatmap-controls">
        <Space
          size={12}
          wrap
          style={{ width: '100%', justifyContent: 'space-between' }}
        >
          <Space size={12} wrap>
            <RangePicker
              value={dateRange}
              onChange={(dates) => {
                if (dates && dates[0] && dates[1]) {
                  setDateRange([dates[0] as Dayjs, dates[1] as Dayjs]);
                }
              }}
              style={{ width: 320 }}
              placeholder={['开始日期', '结束日期']}
              allowClear={false}
            />
            <Select
              placeholder="纠纷类型"
              allowClear
              style={{ width: 180 }}
              value={selectedType}
              onChange={(v) => setSelectedType(v)}
              options={disputeTypes.map((t) => ({
                label: t.name,
                value: parseInt(t.id),
              }))}
            />
            <Button
              type="primary"
              icon={<ReloadOutlined />}
              onClick={handleSearch}
              loading={loading}
            >
              查询刷新
            </Button>
          </Space>
          <Space size={10} wrap>
            <span className="stat-badge">
              <EnvironmentOutlined />
              {totalPoints} 总案件
            </span>
            <span className="stat-badge success">
              <ClockCircleOutlined />
              累计 {cumulativeCount} 件
            </span>
            <Tooltip title="查询当前可视区域内案件列表">
              <Button
                icon={<SearchOutlined />}
                onClick={handleCurrentBBoxDrilldown}
                style={{
                  background: 'rgba(82,196,26,0.15)',
                  borderColor: 'rgba(82,196,26,0.35)',
                  color: '#95de64',
                }}
              >
                区域下钻
              </Button>
            </Tooltip>
            <Tooltip title="一键截图保存热力图（含图例）">
              <Button
                type="primary"
                danger
                icon={<CameraOutlined />}
                onClick={handleScreenshot}
                style={{
                  background: 'linear-gradient(135deg, #ff4d4f, #cf1322)',
                  border: 'none',
                  fontWeight: 600,
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
          <Tooltip title={isPlaying ? '暂停播放' : '开始播放（从最早日期到当前）'}>
            <Button
              type={isPlaying ? 'default' : 'primary'}
              shape="circle"
              icon={isPlaying ? <PauseCircleOutlined /> : <PlayCircleOutlined />}
              onClick={toggleAnimation}
              size="large"
              style={{
                background: isPlaying
                  ? 'rgba(255,255,255,0.08)'
                  : 'linear-gradient(135deg, #1677ff, #722ed1)',
                border: 'none',
                color: '#fff',
                boxShadow: '0 4px 14px rgba(22,119,255,0.35)',
                width: 48,
                height: 48,
                fontSize: 22,
              }}
            />
          </Tooltip>
          <div style={{ flex: 1 }}>
            <Text style={{ color: 'rgba(255,255,255,0.55)', fontSize: 12 }}>
              <FireOutlined style={{ marginRight: 4 }} />
              纠纷累积扩散时间轴 · 拖动或播放查看从
              <Text strong style={{ color: '#91caff', margin: '0 4px' }}>
                {timelineData[0]?.date}
              </Text>
              至
              <Text strong style={{ color: '#ff7875', margin: '0 4px' }}>
                {timelineData[timelineData.length - 1]?.date}
              </Text>
              的案件累计
            </Text>
            <Slider
              min={0}
              max={Math.max(timelineData.length - 1, 0)}
              value={currentDayIndex}
              onChange={handleDaySliderChange}
              marks={timelineData.reduce(
                (acc: Record<number, string>, d: HeatmapTimelineDay, i: number) => {
                  if (
                    timelineData.length <= 10 ||
                    i % Math.max(1, Math.ceil(timelineData.length / 10)) === 0 ||
                    i === timelineData.length - 1
                  ) {
                    acc[i] = dayjs(d.date).format('MM/DD');
                  }
                  return acc;
                },
                {},
              )}
              tooltip={{
                formatter: (v) => {
                  if (v !== undefined && timelineData[v]) {
                    const sum = timelineData
                      .slice(0, v + 1)
                      .reduce((s, d) => s + (d.count || 0), 0);
                    return `${timelineData[v].date} · 当日 ${timelineData[v].count}件 · 累计 ${sum}件`;
                  }
                  return '';
                },
              }}
              style={{ marginTop: 8 }}
            />
          </div>
          <Space size={4} align="center">
            <span style={{ color: 'rgba(255,255,255,0.5)', fontSize: 12 }}>
              播放速度
            </span>
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
              color="geekblue"
              style={{
                margin: 0,
                background: 'linear-gradient(135deg, rgba(22,119,255,0.2), rgba(114,46,209,0.2))',
                border: '1px solid rgba(22,119,255,0.35)',
                color: '#91caff',
                borderRadius: 8,
                padding: '4px 12px',
                fontSize: 12,
                fontWeight: 600,
              }}
            >
              <ClockCircleOutlined style={{ marginRight: 5 }} />
              第 {currentDayIndex + 1}/{timelineData.length} 天 · {currentDay.date}
              <Divider type="vertical" style={{ borderColor: 'rgba(255,255,255,0.2)', margin: '0 8px' }} />
              当日 {currentDay.count} 件 · 累计 {cumulativeCount} 件
            </Tag>
          )}
        </Space>
      </div>

      <div
        style={{
          flex: 1,
          position: 'relative',
          borderRadius: 14,
          overflow: 'hidden',
          border: '1px solid rgba(255,255,255,0.08)',
          boxShadow: '0 8px 30px rgba(0,0,0,0.4)',
          minHeight: 520,
        }}
      >
        <Spin
          spinning={loading || configLoading}
          tip={configLoading ? '加载地图配置...' : '数据加载中...'}
          style={{
            position: 'absolute',
            top: '50%',
            left: '50%',
            transform: 'translate(-50%,-50%)',
            zIndex: 300,
          }}
        >
          <div style={{ width: 1, height: 1 }} />
        </Spin>

        <div
          ref={mapContainerRef}
          style={{ width: '100%', height: '100%', minHeight: 520 }}
        />

        <div className="drill-btn">
          <Tooltip title="点击地图任意位置下钻案件；点击红圈下钻TOP5区域案件">
            <Button
              type="primary"
              icon={<EyeOutlined />}
              style={{
                borderRadius: 10,
                background: 'rgba(22,119,255,0.9)',
                border: 'none',
                boxShadow: '0 4px 14px rgba(22,119,255,0.35)',
              }}
            >
              点击地图下钻
            </Button>
          </Tooltip>
        </div>

        {topCommunities.length > 0 && (
          <div className="top5-legend">
            <div className="top5-legend-title">
              <FireOutlined />
              高发区域 TOP5（聚类）
            </div>
            {topCommunities.map((c) => (
              <div
                key={c.cluster_id || c.org_id || c.rank}
                className="top5-legend-item"
                onClick={() => handleClusterClick(c)}
              >
                <span>
                  <Badge
                    count={`T${c.rank}`}
                    style={{
                      backgroundColor: c.rank <= 3 ? '#ff4d4f' : '#faad14',
                      marginRight: 8,
                      fontSize: 10,
                      minWidth: 26,
                      height: 18,
                      lineHeight: '18px',
                      padding: '0 5px',
                      fontWeight: 700,
                    }}
                  />
                  <span
                    style={{
                      maxWidth: 100,
                      display: 'inline-block',
                      overflow: 'hidden',
                      textOverflow: 'ellipsis',
                      verticalAlign: 'middle',
                    }}
                  >
                    {c.cluster_name || c.org_name}
                  </span>
                </span>
                <span style={{ color: '#ff4d4f', fontWeight: 700 }}>
                  {c.case_count}
                </span>
              </div>
            ))}
          </div>
        )}

        <div
          style={{
            position: 'absolute',
            bottom: 14,
            left: 14,
            zIndex: 200,
            background: 'rgba(26,26,46,0.92)',
            backdropFilter: 'blur(10px)',
            border: '1px solid rgba(255,255,255,0.08)',
            borderRadius: 12,
            padding: '12px 18px',
            boxShadow: '0 6px 20px rgba(0,0,0,0.4)',
          }}
        >
          <div
            style={{
              fontSize: 12,
              color: 'rgba(255,255,255,0.6)',
              marginBottom: 6,
              fontWeight: 600,
            }}
          >
            <FireOutlined style={{ color: '#ff7875', marginRight: 4 }} />
            纠纷密度分级
          </div>
          <div
            style={{
              width: 200,
              height: 14,
              borderRadius: 7,
              background:
                'linear-gradient(90deg, #1a9850, #91cf60, #fee08b, #fc8d59, #d73027)',
              boxShadow: 'inset 0 1px 4px rgba(0,0,0,0.2)',
            }}
          />
          <Row style={{ marginTop: 4 }}>
            <Col span={8}>
              <span
                style={{
                  fontSize: 11,
                  color: 'rgba(255,255,255,0.45)',
                }}
              >
                低
              </span>
            </Col>
            <Col span={8} style={{ textAlign: 'center' }}>
              <span style={{ fontSize: 11, color: 'rgba(255,255,255,0.45)' }}>
                中
              </span>
            </Col>
            <Col span={8} style={{ textAlign: 'right' }}>
              <span style={{ fontSize: 11, color: 'rgba(255,255,255,0.45)' }}>
                高
              </span>
            </Col>
          </Row>
          <div
            style={{
              marginTop: 8,
              paddingTop: 8,
              borderTop: '1px solid rgba(255,255,255,0.06)',
              fontSize: 11,
              color: 'rgba(255,255,255,0.4)',
            }}
          >
            {amapCfg?.use_spatial ? (
              <span>
                <Badge status="success" /> TiDB空间索引已启用
              </span>
            ) : (
              <span>
                <Badge status="default" /> 标准经纬度查询模式
              </span>
            )}
          </div>
        </div>
      </div>

      <Drawer
        title={
          <Space>
            <FileTextOutlined style={{ color: '#1677ff' }} />
            <Text strong>{drawerTitle}</Text>
            <Tag color="blue" style={{ margin: 0 }}>
              共 {drilldownTotal} 条
            </Tag>
          </Space>
        }
        placement="right"
        width={980}
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        extra={
          <Button
            icon={<CloseOutlined />}
            onClick={() => setDrawerOpen(false)}
            type="text"
          />
        }
      >
        <Alert
          message="支持空间下钻：点击红圈查看TOP5区域案件，点击地图任意点查看周边300m范围，或点击工具栏「区域下钻」查看当前可视范围全部案件。"
          type="info"
          showIcon
          style={{ marginBottom: 16, borderRadius: 10 }}
        />
        <Spin spinning={drawerLoading}>
          {drilldownList.length === 0 && !drawerLoading ? (
            <Empty
              description="该范围内暂无符合条件的案件"
              style={{ padding: '80px 0' }}
            />
          ) : (
            <>
              <Table
                rowKey="id"
                size="middle"
                columns={caseColumns}
                dataSource={drilldownList}
                pagination={false}
                scroll={{ x: 1100 }}
                onRow={(r) => ({
                  onClick: () => {
                    if (mapRef.current && r.longitude && r.latitude) {
                      mapRef.current.setCenter([r.longitude, r.latitude]);
                      mapRef.current.setZoom(16);
                    }
                  },
                  style: { cursor: 'pointer' },
                })}
              />
              <div
                style={{
                  marginTop: 16,
                  display: 'flex',
                  justifyContent: 'flex-end',
                }}
              >
                <Pagination
                  current={drilldownPage}
                  pageSize={drilldownPageSize}
                  total={drilldownTotal}
                  showSizeChanger
                  pageSizeOptions={['10', '20', '50', '100']}
                  showQuickJumper
                  showTotal={(t, range) =>
                    `第 ${range?.[0]}-${range?.[1]} 条 / 共 ${t} 条`
                  }
                  onChange={(p, ps) => loadDrilldown(p, ps)}
                />
              </div>
            </>
          )}
        </Spin>
      </Drawer>
    </div>
  );
};

export default HeatMap;
