package config

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
	Server       ServerConfig       `mapstructure:"server"`
	Database     DatabaseConfig     `mapstructure:"database"`
	Redis        RedisConfig        `mapstructure:"redis"`
	Elasticsearch ESConfig           `mapstructure:"elasticsearch"`
	RocketMQ     RocketMQConfig     `mapstructure:"rocketmq"`
	JWT          JWTConfig          `mapstructure:"jwt"`
	MinIO        MinIOConfig        `mapstructure:"minio"`
	DeepSeek     DeepSeekConfig     `mapstructure:"deepseek"`
	Milvus       MilvusConfig       `mapstructure:"milvus"`
	Flowable     FlowableConfig     `mapstructure:"flowable"`
	Court        CourtConfig        `mapstructure:"court"`
	TRTC         TRTCConfig         `mapstructure:"trtc"`
	FaDaDa       FaDaDaConfig       `mapstructure:"fadada"`
	Blockchain   BlockchainConfig   `mapstructure:"blockchain"`
	AliyunVoice  AliyunVoiceConfig  `mapstructure:"aliyun-voice"`
	Services     ServicesConfig     `mapstructure:"services"`
	ServicePorts ServicePortsConfig `mapstructure:"service-ports"`
	Amap         AmapConfig         `mapstructure:"amap"`
	Spatial      SpatialConfig      `mapstructure:"spatial"`
}

type ServicesConfig struct {
	UserService         string `mapstructure:"user-service"`
	DisputeService      string `mapstructure:"dispute-service"`
	WorkflowService     string `mapstructure:"workflow-service"`
	AIService           string `mapstructure:"ai-service"`
	NotificationService string `mapstructure:"notification-service"`
}

type ServicePortsConfig struct {
	User         int `mapstructure:"user"`
	Dispute      int `mapstructure:"dispute"`
	Workflow     int `mapstructure:"workflow"`
	AI           int `mapstructure:"ai"`
	Notification int `mapstructure:"notification"`
}

var GlobalConfig *Config

func LoadConfig(path string) error {
	var err error
	once.Do(func() {
		viper.SetConfigFile(path)
		viper.SetConfigType("yaml")
		viper.AutomaticEnv()

		if err = viper.ReadInConfig(); err != nil {
			err = fmt.Errorf("read config failed: %v", err)
			return
		}

		GlobalConfig = &Config{}
		if err = viper.Unmarshal(GlobalConfig); err != nil {
			err = fmt.Errorf("unmarshal config failed: %v", err)
			return
		}
		config = GlobalConfig
	})
	return err
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
	Name string `mapstructure:"name"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	Charset  string `mapstructure:"charset"`
	MaxIdle  int    `mapstructure:"max_idle"`
	MaxOpen  int    `mapstructure:"max_open"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
	Cluster  bool   `mapstructure:"cluster"`
	Addrs    []string `mapstructure:"addrs"`
}

type ESConfig struct {
	Addresses []string `mapstructure:"addresses"`
	Username  string   `mapstructure:"username"`
	Password  string   `mapstructure:"password"`
}

type RocketMQConfig struct {
	NameServer []string `mapstructure:"nameserver"`
	GroupName  string   `mapstructure:"group_name"`
	RetryTimes int      `mapstructure:"retry_times"`
}

type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	ExpireTime int64  `mapstructure:"expire_time"`
	Issuer     string `mapstructure:"issuer"`
}

type MinIOConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	UseSSL    bool   `mapstructure:"use_ssl"`
	Bucket    string `mapstructure:"bucket"`
}

type DeepSeekConfig struct {
	APIKey     string `mapstructure:"api_key"`
	BaseURL    string `mapstructure:"base_url"`
	Model      string `mapstructure:"model"`
	MaxTokens  int    `mapstructure:"max_tokens"`
	Temperature float64 `mapstructure:"temperature"`
}

type MilvusConfig struct {
	Address     string `mapstructure:"address"`
	Username    string `mapstructure:"username"`
	Password    string `mapstructure:"password"`
	Collection  string `mapstructure:"collection"`
	Dimension   int    `mapstructure:"dimension"`
}

type FlowableConfig struct {
	BaseURL  string `mapstructure:"base_url"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type CourtConfig struct {
	APIEndpoint string `mapstructure:"api_endpoint"`
	APIAppID    string `mapstructure:"api_app_id"`
	APISecret   string `mapstructure:"api_secret"`
	APIPublicKey string `mapstructure:"api_public_key"`
}

type TRTCConfig struct {
	SdkAppID           uint32 `mapstructure:"sdk_app_id"`
	SecretKey          string `mapstructure:"secret_key"`
	AdminUserID        string `mapstructure:"admin_user_id"`
	RecordCallbackURL  string `mapstructure:"record_callback_url"`
	RecordStoragePath  string `mapstructure:"record_storage_path"`
	RecordSegmentSec   int    `mapstructure:"record_segment_sec"`
	MaxQueueSize       int    `mapstructure:"max_queue_size"`
	QueueCheckInterval int    `mapstructure:"queue_check_interval"`
}

type FaDaDaConfig struct {
	APIDomain    string `mapstructure:"api_domain"`
	AppID        string `mapstructure:"app_id"`
	AppSecret    string `mapstructure:"app_secret"`
	CustomerID   string `mapstructure:"customer_id"`
	CertPath     string `mapstructure:"cert_path"`
	AutoSeal     bool   `mapstructure:"auto_seal"`
	CrossPageSeal bool  `mapstructure:"cross_page_seal"`
	NotifyURL    string `mapstructure:"notify_url"`
}

type BlockchainConfig struct {
	APIEndpoint    string `mapstructure:"api_endpoint"`
	AppCode        string `mapstructure:"app_code"`
	AppKey         string `mapstructure:"app_key"`
	AppSecret      string `mapstructure:"app_secret"`
	ChainName      string `mapstructure:"chain_name"`
	ContractAddr   string `mapstructure:"contract_addr"`
	CertTemplateID string `mapstructure:"cert_template_id"`
	QRCodeBaseURL  string `mapstructure:"qr_code_base_url"`
}

var (
	config *Config
	once   sync.Once
)

func GetConfig() *Config {
	return config
}

type AmapConfig struct {
	WebKey         string `mapstructure:"web_key"`
	WebServiceKey  string `mapstructure:"web_service_key"`
	SecurityCode   string `mapstructure:"security_code"`
	DefaultCity    string `mapstructure:"default_city"`
	DefaultLng     string `mapstructure:"default_lng"`
	DefaultLat     string `mapstructure:"default_lat"`
	DefaultZoom    int    `mapstructure:"default_zoom"`
}

type SpatialConfig struct {
	GridLevel          int     `mapstructure:"grid_level"`
	ClusterRadiusMeters float64 `mapstructure:"cluster_radius_meters"`
	UseGeohashPrefix   int     `mapstructure:"use_geohash_prefix"`
	UseSpatialIndex    bool    `mapstructure:"use_spatial_index"`
}

type AliyunVoiceConfig struct {
	AccessKeyID     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
	RegionID        string `mapstructure:"region_id"`
	CallerShowNumber string `mapstructure:"caller_show_number"`
	TtsCode         string `mapstructure:"tts_code"`
	AsrVocabID      string `mapstructure:"asr_vocab_id"`
	CallbackURL     string `mapstructure:"callback_url"`
	Endpoint        string `mapstructure:"endpoint"`
}
