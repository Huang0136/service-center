package router

import (
	"sync"
	"time"
	"regexp"
)

// HTTP method类型
const (
	GET    string = "GET"
	POST   string = "POST"
	PUT    string = "PUT"
	DELETE string = "DELETE"
)

// 处理方法
type Handlers struct {
	GetHandlers    map[string][]Service // GET类型精确的处理器
	PostHandlers   map[string][]Service // POST类型精确的处理器
	PutHandlers    map[string][]Service // PUT类型精确的处理器
	DeleteHandlers map[string][]Service // DELETE类型精确的处理器

	GetRegexHandlers    map[string][]Service // GET类型模糊的处理器
	PostRegexHandlers   map[string][]Service // POST类型模糊的处理器
	PutRegexHandlers    map[string][]Service // PUT类型模糊的处理器
	DeleteRegexHandlers map[string][]Service // DELETE类型模糊的处理器

	Lock sync.Mutex // 锁
}

// 服务节点
type ServerNode struct {
	IP           string    // IP
	Port         int       // 端口
	Instance     string    // 节点实例名
	Desc         string    // 描述
	Remark       string    // 备注
	RegisterDate time.Time // 注册时间
	Enable       bool      // 是否可访问
}

// 服务接口
type Service struct {
	ServiceName string        // 接口名称
	MethodName  string        // 方法名称
	MethodType  string        // 方法类型
	InParams    string        // 入参
	OutParams   string        // 出参
	Node        ServerNode    // 所属节点
	RegexStr    regexp.Regexp // 正则表达式
}

var AllHandlers Handlers

// 服务节点列表
var serverNodeList []ServerNode

// 服务接口列表
var serviceList []Service

// 业务方法结构体
type BusinessHandle struct {
	IP         string // IP地址
	Port       int    // 端口号
	MethodName string // 方法名称
	Desc       string // 方法描述
	UseRegex   bool   // 是否精确匹配
}

// GET方法对应的业务处理方法
var getHandle map[string][]BusinessHandle

// POST方法对应的业务处理方法
var postHandle map[string][]BusinessHandle

// PUT方法对应的业务处理方法
var putHandle map[string][]BusinessHandle

// DELETE方法对应的业务处理方法
var deleteHandle map[string][]BusinessHandle

func init() {
	AllHandlers = Handlers{}
	AllHandlers.GetHandlers = make(map[string][]Service)
	AllHandlers.PostHandlers = make(map[string][]Service)
	AllHandlers.PutHandlers = make(map[string][]Service)
	AllHandlers.DeleteHandlers = make(map[string][]Service)
	AllHandlers.GetRegexHandlers = make(map[string][]Service)
	AllHandlers.PostRegexHandlers = make(map[string][]Service)
	AllHandlers.PutRegexHandlers = make(map[string][]Service)
	AllHandlers.DeleteRegexHandlers = make(map[string][]Service)

	serverNodeList = make([]ServerNode,0)
}

// 根据HTTP请求URL和MethodType匹配处理节点
func MatchHandleMethod(uri, methodType string) (node string, err error) {

	return
}
