-- =====================================================
-- 矛盾纠纷热力图 - 空间索引升级脚本
-- 适用于: TiDB 7.x (兼容 MySQL 空间函数)
-- 创建时间: 2026-06-20
-- =====================================================

USE dispute_resolve;

-- =====================================================
-- 1. 纠纷案件表: 添加空间字段与索引
-- =====================================================

-- 1.1 添加空间几何字段 (存储案发点坐标)
ALTER TABLE dispute_case
    ADD COLUMN geom POINT SRID 4326 DEFAULT NULL COMMENT '案发地空间坐标(POINT, WGS84)'
    AFTER latitude;

-- 1.2 为已有经纬度数据填充 geom 字段
UPDATE dispute_case
SET geom = ST_SRID(POINT(longitude, latitude), 4326)
WHERE longitude IS NOT NULL
  AND latitude IS NOT NULL
  AND longitude != 0
  AND latitude != 0
  AND geom IS NULL;

-- 1.3 创建空间索引 (TiDB 支持 SPATIAL INDEX)
-- 注: TiDB 空间索引要求 SRID 必须指定且列非 GENERATED
ALTER TABLE dispute_case
    ADD SPATIAL INDEX idx_spatial_geom (geom);

-- 1.4 创建 BEFORE INSERT 触发器自动维护 geom 字段
DELIMITER //
CREATE TRIGGER trg_dispute_case_set_geom
BEFORE INSERT ON dispute_case
FOR EACH ROW
BEGIN
    IF NEW.longitude IS NOT NULL AND NEW.latitude IS NOT NULL
       AND NEW.longitude != 0 AND NEW.latitude != 0
       AND NEW.geom IS NULL THEN
        SET NEW.geom = ST_SRID(POINT(NEW.longitude, NEW.latitude), 4326);
    END IF;
END//
DELIMITER ;

-- 1.5 创建 BEFORE UPDATE 触发器自动维护 geom 字段
DELIMITER //
CREATE TRIGGER trg_dispute_case_update_geom
BEFORE UPDATE ON dispute_case
FOR EACH ROW
BEGIN
    IF (NEW.longitude != OLD.longitude OR NEW.latitude != OLD.latitude)
       AND NEW.longitude IS NOT NULL AND NEW.latitude IS NOT NULL
       AND NEW.longitude != 0 AND NEW.latitude != 0 THEN
        SET NEW.geom = ST_SRID(POINT(NEW.longitude, NEW.latitude), 4326);
    END IF;
END//
DELIMITER ;

-- =====================================================
-- 2. 组织架构表: 添加空间字段与索引 (用于社区/街道中心点)
-- =====================================================

ALTER TABLE sys_organization
    ADD COLUMN geom POINT SRID 4326 DEFAULT NULL COMMENT '组织中心空间坐标(POINT, WGS84)'
    AFTER latitude;

UPDATE sys_organization
SET geom = ST_SRID(POINT(longitude, latitude), 4326)
WHERE longitude IS NOT NULL
  AND latitude IS NOT NULL
  AND longitude != 0
  AND latitude != 0
  AND geom IS NULL;

ALTER TABLE sys_organization
    ADD SPATIAL INDEX idx_spatial_geom (geom);

-- =====================================================
-- 3. 区域统计表: 重构支持空间网格/边界聚合
-- =====================================================

ALTER TABLE stats_case_area
    ADD COLUMN geom POLYGON SRID 4326 DEFAULT NULL COMMENT '区域空间边界(POLYGON, WGS84)'
    AFTER latitude;

ALTER TABLE stats_case_area
    ADD COLUMN center_geom POINT SRID 4326 DEFAULT NULL COMMENT '区域中心点空间坐标(POINT, WGS84)'
    AFTER geom;

ALTER TABLE stats_case_area
    ADD SPATIAL INDEX idx_spatial_geom (geom),
    ADD SPATIAL INDEX idx_spatial_center (center_geom);

UPDATE stats_case_area
SET center_geom = ST_SRID(POINT(longitude, latitude), 4326)
WHERE longitude IS NOT NULL
  AND latitude IS NOT NULL
  AND longitude != 0
  AND latitude != 0
  AND center_geom IS NULL;

-- =====================================================
-- 4. 创建空间网格临时统计表 (用于热力图快速聚合)
--    通过网格编码 geohash/网格ID 快速聚类
-- =====================================================

CREATE TABLE IF NOT EXISTS stats_spatial_grid (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
    grid_id VARCHAR(32) NOT NULL COMMENT '网格编码(geohash/自定义)',
    grid_level TINYINT NOT NULL DEFAULT 6 COMMENT '网格级别 1-12 (越大越细)',
    sw_lng DECIMAL(12,8) NOT NULL COMMENT '西南角经度',
    sw_lat DECIMAL(12,8) NOT NULL COMMENT '西南角纬度',
    ne_lng DECIMAL(12,8) NOT NULL COMMENT '东北角经度',
    ne_lat DECIMAL(12,8) NOT NULL COMMENT '东北角纬度',
    center_lng DECIMAL(12,8) NOT NULL COMMENT '中心经度',
    center_lat DECIMAL(12,8) NOT NULL COMMENT '中心纬度',
    center_geom POINT SRID 4326 NOT NULL COMMENT '中心点空间坐标',
    grid_geom POLYGON SRID 4326 NOT NULL COMMENT '网格边界多边形',
    total_count INT NOT NULL DEFAULT 0 COMMENT '案件总数',
    pending_count INT NOT NULL DEFAULT 0 COMMENT '待处理数',
    processing_count INT NOT NULL DEFAULT 0 COMMENT '处理中数',
    completed_count INT NOT NULL DEFAULT 0 COMMENT '已完成数',
    success_count INT NOT NULL DEFAULT 0 COMMENT '调解成功数',
    urgent_count INT NOT NULL DEFAULT 0 COMMENT '紧急案件数',
    stat_date DATE NOT NULL COMMENT '统计日期',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_grid_date (grid_id, grid_level, stat_date),
    SPATIAL INDEX idx_spatial_center (center_geom),
    SPATIAL INDEX idx_spatial_grid (grid_geom),
    INDEX idx_stat_date (stat_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='空间网格统计表(热力图加速)';
