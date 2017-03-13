// 控制器
package handlers

import (
	"auth"
	"beans"
	"bytes"
	"constants"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/rpc"
	"nosql"
	"os"
	"service"
	"strconv"
	"strings"
	"time"

	"utils"

	myLog "service/log"

	"nosql/my.redis"
)

// 静态资源处理
func dealStaticResource(uri string, w http.ResponseWriter) bool {
	// 图标
	if uri == "/favicon.ico" {
		f, err := os.Open("./static/commons/favicon.ico")
		defer f.Close()
		if err != nil {
			log.Panicln("读取favicon.jpg失败", err)
		}
		b, err := ioutil.ReadAll(f)
		if err != nil {
			log.Panicln("读取字节数组失败", err)
		}
		w.Header().Set("Content-Type", "image/x-icon")
		w.Write(b)
		log.Println("返回favicon.ico")
		return true
	}

	// 其他静态资源
	if strings.HasPrefix(uri, "/static/admin/") {
		staticByte, err := ioutil.ReadFile("." + uri)
		if err != nil {
			log.Println("读取静态资源失败，", err)
			w.Write([]byte("没有url:" + uri + " 的静态资源"))
			return true
		}

		if strings.HasSuffix(uri, ".css") {
			w.Header().Set("Content-Type", "text/css")
		} else if strings.HasSuffix(uri, ".js") {
			w.Header().Set("Content-Type", "text/javascript")
		} else if strings.HasSuffix(uri, ".html") {
			w.Header().Set("Content-Type", "text/html;charset=utf-8")
		} else if strings.HasSuffix(uri, ".png") {
			w.Header().Set("Content-Type", "image/png")
		} else if strings.HasSuffix(uri, ".gif") {
			w.Header().Set("Content-Type", "image/gif")
		} else if strings.HasSuffix(uri, ".ico") {
			w.Header().Set("Content-Type", "image/x-icon")
		}

		w.Write(staticByte)
		log.Println("静态资源,", uri)
		return true
	}
	return false
}

// 获取访问者IP
func getAccessIp(r *http.Request) string {
	ip := r.Header.Get("x-forwarded-for")
	if ip == "" || ip == "unknown" {
		ip = r.Header.Get("Proxy-Client-IP")
	}
	if ip == "" || ip == "unknown" {
		ip = r.Header.Get("WL-Proxy-Client-IP")
	}
	if ip == "" || ip == "unknown" {
		ip = r.RemoteAddr
	}
	return ip
}

// 处理所有的请求
func GenericHandler(w http.ResponseWriter, r *http.Request) {
	tRequestBegin := time.Now() // 请求处理开始时间
	uri := r.URL.Path           // 请求URI
	methodType := r.Method      // 请求方法类型，GET/POST/PUT/DELETE
	opChannel := r.Header.Get("OP_CHANNEL")
	reqIp := getAccessIp(r)       // 访问者IP
	logUUID := utils.CreateUUID() // 日志UUID

	// 请求日志记录
	reqLog := beans.RequestLog{Id: logUUID, RequestIP: reqIp, OpChannel: opChannel, ScReceiveTime: tRequestBegin}

	// 静态请求
	if dealStaticResource(uri, w) {
		reqLog.InParams = uri
		reqLog.ScReturnTime = time.Now()
		go myLog.SaveReqLog(reqLog)
		fmt.Println(reqLog)
		return
	}

	// 查找匹配的处理器
	s := matchHandler(methodType, uri)
	if s.Url == "" {
		w.WriteHeader(http.StatusNotFound)

		var msg bytes.Buffer
		msg.WriteString("没有注册有Method:")
		msg.WriteString(methodType)
		msg.WriteString(",URL:")
		msg.WriteString(uri)
		msg.WriteString("的服务接口!")

		w.Write([]byte(msg.String()))
		return
	}

	// 根据 method、url获取服务接口的配置信息
	tRedisBegin := time.Now()
	var resOper nosql.ResourceOperate
	resOper = new(my_redis.RedisResouceOperate)
	resource, err := resOper.GetResourceByMethodAndUrl(methodType, s.Regexp.String())
	if err != nil {
		log.Panicln("nosql操作失败,", err)
	}
	tRedisEnd := time.Now()

	userId := r.Header.Get("user_id")
	token := r.Header.Get("token")
	if resource.NeedLogin { // 需要认证
		// 认证
		userInfo, pass, msg := auth.Authentication(userId, token)
		if !pass {
			w.WriteHeader(http.StatusUnauthorized) // 未认证
			w.Write([]byte(msg))
			fmt.Println(uri, " ,", userId, " ,", token, " ,", msg)
			return
		}

		// 授权
		if !auth.Authorization(resource.RoleIds, userInfo.RoleIds) {
			w.WriteHeader(http.StatusNotAcceptable) // 没有相应权限
			w.Write([]byte("用户没有相应的权限"))
			return
		}

	}

	// 将http请求数据封装成rpc的参数
	rpcReq := service.Req{}
	rpcReq.Params = make(map[string]interface{})

	r.ParseForm()
	formValues := r.Form
	for key, _ := range formValues {
		v := r.FormValue(key)
		rpcReq.Params[key] = v
		//		fmt.Println("key:", key, ",value:", v, ",values:", values)
	}
	rpcReq.Params["METHOD_NAME"] = s.HandlerMethodName // 节点处理方法
	if userId == "" {
		userId = "test"
	}
	rpcReq.Params["USER_ID"] = userId

	// rpc返回结果
	rpcResp := new(service.Resp)
	rpcResp.Params = make(map[string]interface{})

	// 查找服务节点
	rpcServer := chonseServerNode(s.Servers)

	tRpcBegin := time.Now()
	// rpc调用服务节点
	addr1 := rpcServer.IP + ":" + strconv.Itoa(rpcServer.Port)
	rpcClient, err := rpc.DialHTTP("tcp", addr1)
	if err != nil {
		log.Panicf("创建rpc连接失败,%s\n", err)
	}
	defer rpcClient.Close()
	//	fmt.Println("创建rpc连接成功,", rpcClient)

	err = rpcClient.Call("ServiceNode.RpcCallHandler", rpcReq, rpcResp)
	if err != nil {
		// TODO 失败重试

		log.Printf("rpc请求失败,%s\n", err)
		w.Write([]byte("请求节点失败," + err.Error()))
		return
	}
	tRpcEnd := time.Now()
	//	log.Println("业务节点返回数据:", rpcResp)

	// 针对业务数据
	switch s.ContentType {
	case "json":
		strJson := rpcResp.Params["RESULT"].(string)
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(strJson))
		break

	case "file":
		bFile := rpcResp.Params["RESULT"].([]byte)
		fileName := r.FormValue("file_name")
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("content-disposition", "attachment;filename=\""+fileName+"\"")
		w.Write(bFile)

		break
	case "static": // 静态资源，如html、js、css文件等
		bStatic := rpcResp.Params["RESULT"].([]byte)
		w.Header().Set("Content-Type", "text/html")
		w.Write(bStatic)

		break
	}

	var dByte bytes.Buffer
	for kResp, vResp := range rpcResp.Params {
		if kResp != "RESULT" {
			dByte.WriteString(kResp)
			dByte.WriteString(":")
			dByte.WriteString(vResp.(string))
			dByte.WriteString(",")
		}
	}

	dStr := strings.TrimRight(dByte.String(), ",")
	fmt.Println(dStr)

	reqLog.ScReturnTime = time.Now()

	log.Printf("服务中心总耗时:%f,rpc请求耗时:%f,Redis耗时:%f\n", reqLog.ScReturnTime.Sub(reqLog.ScReceiveTime).Seconds(), tRpcEnd.Sub(tRpcBegin).Seconds(), tRedisEnd.Sub(tRedisBegin).Seconds())
}

// 截取时间戳
// 9 + 10
// 19 -8  11 19-10
// 18-10 = 8
// 1479445690 896 353 600
func subUnixNano(t int64) (t1, t2 int64) {
	sa := []rune(strconv.Itoa(int(t)))

	t11, _ := strconv.Atoi(string(sa[:(len(sa) - 9)]))
	t21, _ := strconv.Atoi(string(sa[(len(sa) - 1 - 9):]))

	t1 = int64(t11)
	t2 = int64(t21)
	return
}

// rpc 调用
func rpcCall(client rpc.Client, serviceMethod string, req1 interface{}, resp1 interface{}) (err error) {
	err = client.Call(serviceMethod, req1, resp1)
	return
}

// 根据method、url查找注册在Etcd中的处理方法
func matchHandler(methodType, uri string) (s constants.Service) {
	var preciseMap map[string]constants.Service // 精确匹配
	var regexMap map[string]constants.Service   // 模糊匹配

	switch methodType {
	case "GET":
		preciseMap = constants.GetPreciseHandlers
		regexMap = constants.GetRegexHandlers
		break
	case "POST":
		preciseMap = constants.PostPreciseHandlers
		regexMap = constants.PostRegexHandlers
		break
	case "PUT":
		preciseMap = constants.PutPreciseHandlers
		regexMap = constants.PutRegexHandlers
		break
	case "DELETE":
		preciseMap = constants.DeletePreciseHandlers
		regexMap = constants.DeleteRegexHandlers
		break
	default:
		log.Panicln("未识别的http请求类型")
	}

	// 先精确匹配
	preciseHandler := preciseMap[uri]
	if preciseHandler.Id != "" {
		return preciseHandler
	}

	// 精确匹配未找到，则到模糊匹配中查找
	for _, v := range regexMap {
		if v.Regexp.MatchString(uri) {
			return v
		}
	}

	return
}

// 处理常用的请求
func commonRequest() {

}

// 节点选择随机数
var r *rand.Rand = rand.New(rand.NewSource(20161118))

// 选择节点
func chonseServerNode(s []constants.ServerNode) (sn constants.ServerNode) {
	nodes := len(s)
	if nodes == 1 {
		sn = s[0]
	} else {
		sn = s[r.Intn(nodes)]
	}

	return
}
