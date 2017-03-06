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
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"golang.org/x/net/context"
)

// 配置信息
var Configs map[string]string = make(map[string]string)

// rpc client
var RpcClient rpc.Client

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

// 服务接口
type Service struct {
	Id                string
	Name              string         // 接口名称
	Url               string         // 接口URL
	Type              string         // URL 方法类型
	Regexp            *regexp.Regexp // 匹配的正则表达式
	ContentType       string         // 返回结果类型
	Servers           []ServerNode   // 服务节点列表
	HandlerMethodName string         // 节点处理的方法
}

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

var EtcdLock sync.Mutex

func InitConstants() {
	// 读取配置文件
	readConfigProperties()
	initDB()

	// 读取Etcd的服务信息
	readEtcdInfo()

	// rpc连接到节点
	go initRpc()

	// 转换URL
	initRequestUri()
}

// 转换监听到获取获取到的服务信息（节点、服务接口）
func TranEtcdServerInfo(opType, key, val string) {
	EtcdLock.Lock()
	defer EtcdLock.Unlock()

	arrayKey := strings.Split(key, ",")   // key:servers,10.8.26.129,9090,GET,true,/user/{id}
	arrayValue := strings.Split(val, ",") // value:service_name=获取用户信息,method=GetUser,remark=根据USER_ID获取用户信息,in_params=,out_params=

	ip := arrayKey[1]
	port := arrayKey[2]
	intPort, _ := strconv.Atoi(port)

	// 处理服务节点
	if len(arrayKey) == 3 {
		switch opType {
		case "PUT":
			v0 := arrayValue[0]
			serverName := strings.Split(v0, "=")[1] // 节点名称

			serverNode := ServerNode{Name: serverName, IP: ip, Port: intPort}
			ServerNodeInfoList = append(ServerNodeInfoList, serverNode)
			break
		case "DELETE":
			if len(ServerNodeInfoList) == 1 { // 只有一个节点,则删除
				ServerNodeInfoList = make([]ServerNode, 0)
			} else { // 多个节点
				for i, sss := range ServerNodeInfoList {
					if sss.IP == ip && sss.Port == intPort {
						// {1,2}
						var lefeI int
						if i == 0 {
							lefeI = 0
						} else {
							lefeI = i
						}

						var rightI int
						if i == len(ServerNodeInfoList) {
							rightI = i
						} else {
							rightI = i + 1
						}

						ServerNodeInfoList = append(ServerNodeInfoList[:lefeI], ServerNodeInfoList[rightI:]...)
						break
					}
				}
			}
			break
		}
		log.Println("转换服务节点成功,", ServerNodeInfoList)
		return
	}

	// 处理服务接口
	methodType := arrayKey[3] // type
	serviceUrl := arrayKey[5] // url
	regexp1 := arrayKey[4]    // regexp

	serviceName := strings.Split(arrayValue[0], "=")[0]

	var sExists Service
	switch methodType {
	case "GET":
		if regexp1 == "false" {
			sExists = GetPreciseHandlers[serviceUrl]
		} else {
			sExists = GetRegexHandlers[serviceUrl]
		}
		break
	case "POST":
		if regexp1 == "false" {
			sExists = PostPreciseHandlers[serviceUrl]
		} else {
			sExists = PostRegexHandlers[serviceUrl]
		}
		break
	case "PUT":
		if regexp1 == "false" {
			sExists = PutPreciseHandlers[serviceUrl]
		} else {
			sExists = PutRegexHandlers[serviceUrl]
		}
		break
	case "DELETE":
		if regexp1 == "false" {
			sExists = DeletePreciseHandlers[serviceUrl]
		} else {
			sExists = DeleteRegexHandlers[serviceUrl]
		}
		break
	}

	ssnn := ServerNode{IP: ip, Port: intPort, RegisterTime: time.Now()}

	var afterService Service
	switch opType {
	case "PUT": // 新增
		if sExists.Type != "" { // 某个服务接口已有节点
			switch methodType {
			case "GET":
				if regexp1 == "false" {
					serviceUrlValue := GetPreciseHandlers[serviceUrl]
					suvServers := append(serviceUrlValue.Servers, ssnn)
					serviceUrlValue.Servers = suvServers
					GetPreciseHandlers[serviceUrl] = serviceUrlValue
					afterService = serviceUrlValue
				} else {
					serviceUrlValue := GetRegexHandlers[serviceUrl]
					suvServers := append(serviceUrlValue.Servers, ssnn)
					serviceUrlValue.Servers = suvServers
					GetRegexHandlers[serviceUrl] = serviceUrlValue
					afterService = serviceUrlValue
				}
				break
			case "PUT":
				if regexp1 == "false" {
					serviceUrlValue := PutPreciseHandlers[serviceUrl]
					suvServers := append(serviceUrlValue.Servers, ssnn)
					serviceUrlValue.Servers = suvServers
					PutPreciseHandlers[serviceUrl] = serviceUrlValue
					afterService = serviceUrlValue
				} else {
					serviceUrlValue := PutRegexHandlers[serviceUrl]
					suvServers := append(serviceUrlValue.Servers, ssnn)
					serviceUrlValue.Servers = suvServers
					PutRegexHandlers[serviceUrl] = serviceUrlValue
					afterService = serviceUrlValue
				}
				break
			case "POST":
				if regexp1 == "false" {
					serviceUrlValue := PostPreciseHandlers[serviceUrl]
					suvServers := append(serviceUrlValue.Servers, ssnn)
					serviceUrlValue.Servers = suvServers
					PostPreciseHandlers[serviceUrl] = serviceUrlValue
					afterService = serviceUrlValue
				} else {
					serviceUrlValue := PostRegexHandlers[serviceUrl]
					suvServers := append(serviceUrlValue.Servers, ssnn)
					serviceUrlValue.Servers = suvServers
					PostRegexHandlers[serviceUrl] = serviceUrlValue
					afterService = serviceUrlValue
				}
				break
			case "DELETE":
				if regexp1 == "false" {
					serviceUrlValue := DeletePreciseHandlers[serviceUrl]
					suvServers := append(serviceUrlValue.Servers, ssnn)
					serviceUrlValue.Servers = suvServers
					DeletePreciseHandlers[serviceUrl] = serviceUrlValue
					afterService = serviceUrlValue
				} else {
					serviceUrlValue := DeleteRegexHandlers[serviceUrl]
					suvServers := append(serviceUrlValue.Servers, ssnn)
					serviceUrlValue.Servers = suvServers
					DeleteRegexHandlers[serviceUrl] = serviceUrlValue
					afterService = serviceUrlValue
				}
				break
			}
		} else { // 某个服务接口没有节点
			var service Service // 服务接口
			handlerMethodName := strings.Split(arrayValue[2], "=")[1]
			contentType := strings.Split(arrayValue[3], "=")[1]

			if regexp1 == "true" {
				service = Service{Id: "", Name: serviceName, Url: serviceUrl, Type: methodType, Regexp: urlToRegexp(serviceUrl), HandlerMethodName: handlerMethodName, ContentType: contentType}
			} else if regexp1 == "false" {
				service = Service{Id: "", Name: serviceName, Url: serviceUrl, Type: methodType, HandlerMethodName: handlerMethodName, ContentType: contentType}
			}
			service.Servers = make([]ServerNode, 0)
			service.Servers = append(service.Servers, ssnn)

			switch methodType {
			case "GET":
				if regexp1 == "false" {
					GetPreciseHandlers[serviceUrl] = service
					afterService = GetPreciseHandlers[serviceUrl]
				} else {
					GetRegexHandlers[serviceUrl] = service
					afterService = GetRegexHandlers[serviceUrl]
				}
				break
			case "PUT":
				if regexp1 == "false" {
					PutPreciseHandlers[serviceUrl] = service
					afterService = PutPreciseHandlers[serviceUrl]
				} else {
					PutRegexHandlers[serviceUrl] = service
					afterService = PutRegexHandlers[serviceUrl]
				}
				break
			case "POST":
				if regexp1 == "false" {
					PostPreciseHandlers[serviceUrl] = service
					afterService = PostPreciseHandlers[serviceUrl]
				} else {
					PostRegexHandlers[serviceUrl] = service
					afterService = PostRegexHandlers[serviceUrl]
				}
				break
			case "DELETE":
				if regexp1 == "false" {
					DeletePreciseHandlers[serviceUrl] = service
					afterService = DeletePreciseHandlers[serviceUrl]
				} else {
					DeleteRegexHandlers[serviceUrl] = service
					afterService = DeleteRegexHandlers[serviceUrl]
				}
				break
			}
		}
		break
	case "DELETE": // 删除
		if len(sExists.Servers) == 1 { // 某个接口只有一个服务节点
			switch methodType {
			case "GET":
				if regexp1 == "false" {
					delete(GetPreciseHandlers, serviceUrl)
					afterService = GetPreciseHandlers[serviceUrl]
				} else {
					delete(GetRegexHandlers, serviceUrl)
					afterService = GetRegexHandlers[serviceUrl]
				}
				break
			case "PUT":
				if regexp1 == "false" {
					delete(PutPreciseHandlers, serviceUrl)
					afterService = PutPreciseHandlers[serviceUrl]
				} else {
					delete(PutRegexHandlers, serviceUrl)
					afterService = PutRegexHandlers[serviceUrl]
				}
				break
			case "POST":
				if regexp1 == "false" {
					delete(PostPreciseHandlers, serviceUrl)
					afterService = PostPreciseHandlers[serviceUrl]
				} else {
					delete(PostRegexHandlers, serviceUrl)
					afterService = PostRegexHandlers[serviceUrl]
				}
				break
			case "DELETE":
				if regexp1 == "false" {
					delete(DeletePreciseHandlers, serviceUrl)
					afterService = DeletePreciseHandlers[serviceUrl]
				} else {
					delete(DeleteRegexHandlers, serviceUrl)
					afterService = DeleteRegexHandlers[serviceUrl]
				}
				break
			}
		} else { // 某个接口有多个服务节点
			var iii int
			for i, s1 := range sExists.Servers {
				if s1.IP == ip && s1.Port == intPort {
					iii = i
					break
				}
			}

			var lefeII int
			if iii == 0 {
				lefeII = 0
			} else {
				lefeII = iii
			}

			var rightII int
			if iii == len(sExists.Servers) {
				rightII = iii
			} else {
				rightII = iii + 1
			}

			switch methodType {
			case "GET":
				if regexp1 == "false" {
					serviceUrlValue := GetPreciseHandlers[serviceUrl]
					suvServers := append(serviceUrlValue.Servers[:lefeII], serviceUrlValue.Servers[rightII:]...)
					serviceUrlValue.Servers = suvServers
					GetPreciseHandlers[serviceUrl] = serviceUrlValue
					afterService = serviceUrlValue
				} else {
					serviceUrlValue := GetRegexHandlers[serviceUrl]
					suvServers := append(serviceUrlValue.Servers[:lefeII], serviceUrlValue.Servers[rightII:]...)
					serviceUrlValue.Servers = suvServers
					GetRegexHandlers[serviceUrl] = serviceUrlValue
					afterService = serviceUrlValue
				}
				break
			case "PUT":
				if regexp1 == "false" {
					serviceUrlValue := PutPreciseHandlers[serviceUrl]
					suvServers := append(serviceUrlValue.Servers[:lefeII], serviceUrlValue.Servers[rightII:]...)
					serviceUrlValue.Servers = suvServers
					PutPreciseHandlers[serviceUrl] = serviceUrlValue
					afterService = serviceUrlValue
				} else {
					serviceUrlValue := PutRegexHandlers[serviceUrl]
					suvServers := append(serviceUrlValue.Servers[:lefeII], serviceUrlValue.Servers[rightII:]...)
					serviceUrlValue.Servers = suvServers
					PutRegexHandlers[serviceUrl] = serviceUrlValue
					afterService = serviceUrlValue
				}
				break
			case "POST":
				if regexp1 == "false" {
					serviceUrlValue := PostPreciseHandlers[serviceUrl]
					suvServers := append(serviceUrlValue.Servers[:lefeII], serviceUrlValue.Servers[rightII:]...)
					serviceUrlValue.Servers = suvServers
					PostPreciseHandlers[serviceUrl] = serviceUrlValue
					afterService = serviceUrlValue
				} else {
					serviceUrlValue := PostRegexHandlers[serviceUrl]
					suvServers := append(serviceUrlValue.Servers[:lefeII], serviceUrlValue.Servers[rightII:]...)
					serviceUrlValue.Servers = suvServers
					PostRegexHandlers[serviceUrl] = serviceUrlValue
					afterService = serviceUrlValue
				}
				break
			case "DELETE":
				if regexp1 == "false" {
					serviceUrlValue := DeletePreciseHandlers[serviceUrl]
					suvServers := append(serviceUrlValue.Servers[:lefeII], serviceUrlValue.Servers[rightII:]...)
					serviceUrlValue.Servers = suvServers
					DeletePreciseHandlers[serviceUrl] = serviceUrlValue
					afterService = serviceUrlValue
				} else {
					serviceUrlValue := DeleteRegexHandlers[serviceUrl]
					suvServers := append(serviceUrlValue.Servers[:lefeII], serviceUrlValue.Servers[rightII:]...)
					serviceUrlValue.Servers = suvServers
					DeleteRegexHandlers[serviceUrl] = serviceUrlValue
					afterService = serviceUrlValue
				}
				break
			}
		}
		break
	}
	log.Println("转换服务接口成功,", afterService)
}

// RPC连接到节点
func initRpc() {
	<-MonitorEtcdChan

	/*
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
	*/
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
	rex, err := regexp.Compile("^" + strRex + "$")
	if err != nil {
		log.Fatal("正则表达式转换失败,", err)
	}
	return rex
}

// 读取Etcd的节点配置，服务接口配置信息
func readEtcdInfo() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{Configs["etcd.url"]},
		DialTimeout: 1 * time.Minute,
	})
	if err != nil {
		log.Fatalln("建立etcd连接失败,", err)
	}

	resp, err := cli.Get(context.TODO(), "servers", clientv3.WithPrefix())
	if err != nil {
		log.Fatalln("获取Etcd配置信息失败,", err)
	}
	log.Println("读取Etcd配置...")

	for _, kv := range resp.Kvs {
		log.Printf("获取Etcd服务信息,key:%s,value:%s", string(kv.Key), string(kv.Value))
		TranEtcdServerInfo("PUT", string(kv.Key), string(kv.Value))
	}
	log.Println("读取Etcd配置完成")

}
