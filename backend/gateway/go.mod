module github.com/dispute-resolve/gateway

go 1.21

require (
	github.com/dispute-resolve/common v0.0.0
	github.com/dispute-resolve/user-service v0.0.0
	github.com/dispute-resolve/dispute-service v0.0.0
	github.com/dispute-resolve/workflow-service v0.0.0
	github.com/dispute-resolve/ai-service v0.0.0
	github.com/dispute-resolve/notification-service v0.0.0
	github.com/cloudwego/hertz v0.7.2
	github.com/cloudwego/hertz/pkg/app v0.7.2
	github.com/cloudwego/hertz/pkg/network/netpoll v0.7.2
	github.com/cloudwego/hertz/pkg/protocol/consts v0.7.2
	github.com/cloudwego/kitex v0.9.0
	github.com/bytedance/sonic v1.9.1
	github.com/google/uuid v1.3.0
	github.com/redis/go-redis/v9 v9.3.0
	github.com/spf13/viper v1.17.0
	go.uber.org/zap v1.26.0
	golang.org/x/crypto v0.14.0
	gorm.io/driver/mysql v1.5.2
	gorm.io/gorm v1.25.5
)

replace (
	github.com/dispute-resolve/common => ../common
	github.com/dispute-resolve/user-service => ../services/user-service
	github.com/dispute-resolve/dispute-service => ../services/dispute-service
	github.com/dispute-resolve/workflow-service => ../services/workflow-service
	github.com/dispute-resolve/ai-service => ../services/ai-service
	github.com/dispute-resolve/notification-service => ../services/notification-service
)
