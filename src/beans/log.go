package beans

import (
	"time"
)

// 请求Log实体
type RequestLog struct {
	Id                string    // UUID
	RequestIP         string    // 访问者IP
	OpChannel         string    // 访问者渠道，如：网站、微信公众号、手机APP等
	AccessUser        string    // 访问用户ID
	ServiceCode       string    // 访问接口号
	InParams          string    // 请求参数
	ScReceiveTime     time.Time // 服务中心接收到请求时间
	ScReturnTime      time.Time // 服务中心处理完成返回时间
	CallNodeCount     int64     // 调用节点次数
	DealResult        string    // 调用完成返回结果
	RequestDealStatus int64     //  请求处理结果状态
}

// NodeDealLog实体
type NodeDealLog struct {
	ScLogId           string    // 服务中心日志ID
	NodeLogId         string    // 节点日志ID
	NodeName          string    // 节点名称
	NodeAddr          string    // 节点地址
	Status            int       // 执行结果状态
	NodeDealBeginTime time.Time // 节点开始调用时间
	NodeDealOverTime  time.Time // 时间处理完成时间
}

// SqlExecLog实体
type SqlExecLog struct {
	NodeLogId string // 节点日志ID
	SqlCode   string // SQL代码
	ExecSql   string // 执行的详细SQL
	ExecTime  int64  // 耗时，毫秒
}
