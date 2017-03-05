// 常量
package constants

import (
	"bufio"
	"io"
	"log"
	"net/rpc"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// 配置信息
var Configs map[string]string = make(map[string]string)

// rpc client
var RpcClient rpc.Client

// 请求url
var RequestUri map[string]string

// 监听完成
var MonitorEtcdChan chan (bool) = make(chan bool)

// 服务节点和服务接口
var ServerNodeInfoList []ServerNode = make([]ServerNode, 0)

// 服务节点RPC client
var RpcClients map[string]string

var GetPreciseHandlers map[string]Service = make(map[string]Service) // GET类型精确处理器
var GetRegexHandlers map[string]Service = make(map[string]Service)   // GET类型模糊处理器

var PostPreciseHandlers map[string]Service = make(map[string]Service) // POST类型精确处理器
var PostRegexHandlers map[string]Service = make(map[string]Service)   // POST类型模糊处理器

var PutPreciseHandlers map[string]Service = make(map[string]Service) // PUT类型精确处理器
var PutRegexHandlers map[string]Service = make(map[string]Service)   // PUT类型模糊处理器

var DeletePreciseHandlers map[string]Service = make(map[string]Service) // DELETE类型精确处理器
var DeleteRegexHandlers map[string]Service = make(map[string]Service)   // DELETE类型模糊处理器

// 服务节点信息
type ServerNode struct {
	Name         string    // 名称
	IP           string    // IP
	Port         int       // 端口
	RegisterTime time.Time // 注册时间
	Desc         string    // 备注
	DealTime     int       // 处理次数
	FailTime     int       // 失败次数
}

type Service struct {
	Id      string
	Name    string         // 接口名称
	Url     string         // 接口URL
	Type    string         // URL 方法类型
	Regexp  *regexp.Regexp // 匹配的正则表达式
	Servers []ServerNode   // 服务节点列表
}

func InitConstants() {
	// 读取配置文件
	readConfigProperties()

	// rpc连接到节点
	go initRpc()

	// 转换URL
	initRequestUri()
}

// 转换服务节点信息
func TranServerInfo(opType, key, val string) {
	arrayKey := strings.Split(key, ",")   // servers,10.8.26.129,9090,GET,true,/user/{id}
	arrayValue := strings.Split(val, ",") // service_name=获取用户信息,method=GetUser,remark=根据USER_ID获取用户信息,in_params=,out_params=

	ip := arrayKey[1]
	port := arrayKey[2]
	intPort, _ := strconv.Atoi(port)
	//	nodeAddr := ip + ":" + port

	// 服务节点
	if len(arrayKey) == 3 {
		switch opType {
		case "PUT":
			v0 := arrayValue[0]
			serverName := strings.Split(v0, "=")[1]

			serverNode := ServerNode{Name: serverName, IP: ip, Port: intPort}
			ServerNodeInfoList = append(ServerNodeInfoList, serverNode)
			break
		case "DELETE":
			for i, sss := range ServerNodeInfoList {
				if sss.IP == ip && sss.Port == intPort {
					ServerNodeInfoList = append(ServerNodeInfoList[:(i-1)], ServerNodeInfoList[(i+1):]...)
					break
				}
			}
			break
		}
		return
	}

	// 服务接口
	methodType := arrayKey[3]
	serviceUrl := arrayKey[5]
	regexp1 := arrayKey[4]

	serviceName := strings.Split(arrayValue[0], "=")[0]

	var sExists Service
	switch methodType {
	case "GET":
		if regexp1 == "true" {
			sExists = GetPreciseHandlers[serviceUrl]
		} else {
			sExists = GetRegexHandlers[serviceUrl]
		}
		break
	case "POST":
		if regexp1 == "true" {
			sExists = PostPreciseHandlers[serviceUrl]
		} else {
			sExists = PostRegexHandlers[serviceUrl]
		}
		break
	case "PUT":
		if regexp1 == "true" {
			sExists = PutPreciseHandlers[serviceUrl]
		} else {
			sExists = PutRegexHandlers[serviceUrl]
		}
		break
	case "DELETE":
		if regexp1 == "true" {
			sExists = DeletePreciseHandlers[serviceUrl]
		} else {
			sExists = DeleteRegexHandlers[serviceUrl]
		}
		break
	}

	var ssnn ServerNode
	var index int
	for ii, snTemp := range ServerNodeInfoList {
		if snTemp.IP == ip && snTemp.Port == intPort {
			ssnn = snTemp
			index = ii
			break
		}
	}

	var service Service
	var bExists bool = true
	if sExists.Name == "" { // 未存在
		bExists = false
		if regexp1 == "true" {
			service = Service{Id: "", Name: serviceName, Url: serviceUrl, Type: methodType, Regexp: urlToRegexp(serviceUrl)}
		} else if regexp1 == "false" {
			service = Service{Id: "", Name: serviceName, Url: serviceUrl, Type: methodType}
		}
		service.Servers = make([]ServerNode, 0)
		service.Servers = append(service.Servers, ssnn)
	}

	switch opType { // 新增还是删除
	case "PUT":
		if bExists {
			switch methodType {
			case "GET":
				if regexp1 == "false" {
					serviceUrlValue := GetPreciseHandlers[serviceUrl]
					suvServers := append(serviceUrlValue.Servers, ssnn)
					serviceUrlValue.Servers = suvServers
					GetPreciseHandlers[serviceUrl] = serviceUrlValue
				} else {
					serviceUrlValue := GetRegexHandlers[serviceUrl]
					suvServers := append(serviceUrlValue.Servers, ssnn)
					serviceUrlValue.Servers = suvServers
					GetRegexHandlers[serviceUrl] = serviceUrlValue
				}
				break
			case "PUT":
				if regexp1 == "false" {
					serviceUrlValue := PutPreciseHandlers[serviceUrl]
					suvServers := append(serviceUrlValue.Servers, ssnn)
					serviceUrlValue.Servers = suvServers
					PutPreciseHandlers[serviceUrl] = serviceUrlValue
				} else {
					serviceUrlValue := PutRegexHandlers[serviceUrl]
					suvServers := append(serviceUrlValue.Servers, ssnn)
					serviceUrlValue.Servers = suvServers
					PutRegexHandlers[serviceUrl] = serviceUrlValue
				}
				break
			case "POST":
				if regexp1 == "false" {
					serviceUrlValue := PostPreciseHandlers[serviceUrl]
					suvServers := append(serviceUrlValue.Servers, ssnn)
					serviceUrlValue.Servers = suvServers
					PostPreciseHandlers[serviceUrl] = serviceUrlValue
				} else {
					serviceUrlValue := PostRegexHandlers[serviceUrl]
					suvServers := append(serviceUrlValue.Servers, ssnn)
					serviceUrlValue.Servers = suvServers
					PostRegexHandlers[serviceUrl] = serviceUrlValue
				}
				break
			case "DELETE":
				if regexp1 == "false" {
					serviceUrlValue := DeletePreciseHandlers[serviceUrl]
					suvServers := append(serviceUrlValue.Servers, ssnn)
					serviceUrlValue.Servers = suvServers
					DeletePreciseHandlers[serviceUrl] = serviceUrlValue
				} else {
					serviceUrlValue := DeleteRegexHandlers[serviceUrl]
					suvServers := append(serviceUrlValue.Servers, ssnn)
					serviceUrlValue.Servers = suvServers
					DeleteRegexHandlers[serviceUrl] = serviceUrlValue
				}
				break
			}
		} else {
			switch methodType {
			case "GET":
				if regexp1 == "false" {
					GetPreciseHandlers[serviceUrl] = service
				} else {
					GetRegexHandlers[serviceUrl] = service
				}
				break
			case "PUT":
				if regexp1 == "false" {
					PutPreciseHandlers[serviceUrl] = service
				} else {
					PutRegexHandlers[serviceUrl] = service
				}
				break
			case "POST":
				if regexp1 == "false" {
					PostPreciseHandlers[serviceUrl] = service
				} else {
					PostRegexHandlers[serviceUrl] = service
				}
				break
			case "DELETE":
				if regexp1 == "false" {
					DeletePreciseHandlers[serviceUrl] = service
				} else {
					DeleteRegexHandlers[serviceUrl] = service
				}
				break
			}
		}

		break
	case "DELETE":
		var onlyOneServer bool = false
		if len(service.Servers) == 1 {
			onlyOneServer = true
		}

		if onlyOneServer {
			switch methodType {
			case "GET":
				if regexp1 == "false" {
					delete(GetPreciseHandlers, serviceUrl)
				} else {
					delete(GetRegexHandlers, serviceUrl)
				}
				break
			case "PUT":
				if regexp1 == "false" {
					delete(PutPreciseHandlers, serviceUrl)
				} else {
					delete(PutRegexHandlers, serviceUrl)
				}
				break
			case "POST":
				if regexp1 == "false" {
					delete(PostPreciseHandlers, serviceUrl)
				} else {
					delete(PostRegexHandlers, serviceUrl)
				}
				break
			case "DELETE":
				if regexp1 == "false" {
					delete(DeletePreciseHandlers, serviceUrl)
				} else {
					delete(DeleteRegexHandlers, serviceUrl)
				}
				break
			}
		} else {
			switch methodType {
			case "GET":
				if regexp1 == "false" {
					serviceUrlValue := GetPreciseHandlers[serviceUrl]
					suvServers := append(serviceUrlValue.Servers[:index-1], serviceUrlValue.Servers[index+1:]...)
					serviceUrlValue.Servers = suvServers
					GetPreciseHandlers[serviceUrl] = serviceUrlValue
				} else {
					serviceUrlValue := GetRegexHandlers[serviceUrl]
					suvServers := append(serviceUrlValue.Servers[:index-1], serviceUrlValue.Servers[index+1:]...)
					serviceUrlValue.Servers = suvServers
					GetRegexHandlers[serviceUrl] = serviceUrlValue
				}
				break
			case "PUT":
				if regexp1 == "false" {
					serviceUrlValue := PutPreciseHandlers[serviceUrl]
					suvServers := append(serviceUrlValue.Servers[:index-1], serviceUrlValue.Servers[index+1:]...)
					serviceUrlValue.Servers = suvServers
					PutPreciseHandlers[serviceUrl] = serviceUrlValue
				} else {
					serviceUrlValue := PutRegexHandlers[serviceUrl]
					suvServers := append(serviceUrlValue.Servers[:index-1], serviceUrlValue.Servers[index+1:]...)
					serviceUrlValue.Servers = suvServers
					PutRegexHandlers[serviceUrl] = serviceUrlValue
				}
				break
			case "POST":
				if regexp1 == "false" {
					serviceUrlValue := PostPreciseHandlers[serviceUrl]
					suvServers := append(serviceUrlValue.Servers[:index-1], serviceUrlValue.Servers[index+1:]...)
					serviceUrlValue.Servers = suvServers
					PostPreciseHandlers[serviceUrl] = serviceUrlValue
				} else {
					serviceUrlValue := PostRegexHandlers[serviceUrl]
					suvServers := append(serviceUrlValue.Servers[:index-1], serviceUrlValue.Servers[index+1:]...)
					serviceUrlValue.Servers = suvServers
					PostRegexHandlers[serviceUrl] = serviceUrlValue
				}
				break
			case "DELETE":
				if regexp1 == "false" {
					serviceUrlValue := DeletePreciseHandlers[serviceUrl]
					suvServers := append(serviceUrlValue.Servers[:index-1], serviceUrlValue.Servers[index+1:]...)
					serviceUrlValue.Servers = suvServers
					DeletePreciseHandlers[serviceUrl] = serviceUrlValue
				} else {
					serviceUrlValue := DeleteRegexHandlers[serviceUrl]
					suvServers := append(serviceUrlValue.Servers[:index-1], serviceUrlValue.Servers[index+1:]...)
					serviceUrlValue.Servers = suvServers
					DeleteRegexHandlers[serviceUrl] = serviceUrlValue
				}
				break
			}
		}

		break
	}
	log.Println("转换服务节点信息成功...")
}

// RPC连接到节点
func initRpc() {
	<-MonitorEtcdChan
	var rpcClient *rpc.Client
	var err error

connToServerNode:
	{
		rpcClient, err = rpc.DialHTTP("tcp", "127.0.0.1:9090")
	}

	if err != nil {
		log.Printf("rpc连接失败,将在%d秒后重试连接到节点。  %s \n", 2, err)
		time.Sleep(2 * time.Second)
		goto connToServerNode
	}

	log.Println("rpc:", rpcClient)
	RpcClient = *rpcClient
}

// request uri
func initRequestUri() {
	RequestUri = make(map[string]string)

	RequestUri["/user"] = "GetUserByUserId"
	RequestUri["/add/user"] = "Add"

}

// 读取系统配置信息
func readConfigProperties() {
	log.Println("正在读取系统配置文件server-center.properties...")

	file, err := os.Open("./config/server-center.properties")
	if err != nil {
		log.Fatalln("读取系统配置文件失败,", err)
	}

	reader := bufio.NewReader(file)
	for {
		b, _, err := reader.ReadLine()
		if err != nil || err == io.EOF {
			break
		}

		line := string(b)
		if !strings.HasPrefix(line, "#") && strings.Contains(line, "=") {
			array := strings.Split(line, "=")
			Configs[strings.TrimSpace(array[0])] = strings.TrimSpace(array[1])
		}
	}
	log.Println("读取系统配置文件完成")
}

// 将url转换成正则表达式
func urlToRegexp(url string) *regexp.Regexp {
	urlArray := strings.Split(url, "/") // /user/{id}  /user/0136
	for i, urlTemp := range urlArray {
		if strings.HasPrefix(urlTemp, "{") && strings.HasSuffix(urlTemp, "}") {
			urlArray[i] = "([^/]+)"
		}
	}

	strRex := strings.Join(urlArray, "/")
	rex, err := regexp.Compile(strRex)
	if err != nil {
		log.Fatal("正则表达式转换失败,", err)
	}
	return rex
}
