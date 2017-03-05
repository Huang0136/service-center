// 控制器
package handlers

import (
	"auth"
	"bytes"
	"constants"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/rpc"
	"os"
	"service"
	"time"
)

// 处理所有的请求
func GenericHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.String()
	methodType := r.Method

	// 常用的请求
	if url == "/favicon.ico" {
		f, err := os.Open("./static/commons/favicon.ico")
		defer f.Close()
		if err != nil {
			log.Fatalln("读取favicon.jpg失败", err)
		}
		b, err := ioutil.ReadAll(f)
		if err != nil {
			log.Fatalln("读取字节数组失败", err)
		}
		w.Write(b)
		log.Println("返回favicon.ico")
		return
	}

	// 获取匹配的处理器
	s := matchHandler(methodType, url)
	if s.Url != "" {
		var msg bytes.Buffer
		msg.WriteString("找到Type:")
		msg.WriteString(methodType)
		msg.WriteString(",URL:")
		msg.WriteString(url)
		msg.WriteString(" 的处理器!")

		w.Write([]byte(msg.String()))
		return
	} else {
		var msg bytes.Buffer
		msg.WriteString("未找到Type:")
		msg.WriteString(methodType)
		msg.WriteString(",URL:")
		msg.WriteString(url)
		msg.WriteString(" 的处理器!")

		w.Write([]byte(msg.String()))
		return
	}

	// 获取该接口的信息(认证、授权配置等)
	// redis get

	// 认证
	userName := r.Header.Get("userName")
	pwd := r.Header.Get("pwd")
	pass, msg := auth.Authentication(userName, pwd)
	if !pass {
		w.Write([]byte(msg))
		fmt.Println(url, " ,", userName, " ,", pwd, " ,", msg)
		return
	}

	// 授权
	nowTime := time.Now().Format("2006-01-02 15:04:05.999999999")
	uri := r.URL.RequestURI()
	info := "GenericHandler,URI:" + uri + ",URL:" + url + " ,methodType:" + methodType + ", " + nowTime
	fmt.Println(info)

	r.ParseForm()
	formValues := r.Form

	// rpc 请求参数
	rpcReq := service.Req{}
	rpcReq.Params = make(map[string]interface{})

	for key, values := range formValues {
		fmt.Println("key:", key, ",values:", values)

		rpcReq.Params[key] = values

	}

	// 返回数据
	resp1 := new(service.Resp)

	matchUrl := false
	for k, v := range constants.RequestUri {
		if k == url {
			rpcReq.Params["method"] = v

			err := rpcCall(constants.RpcClient, "ServiceNode.RpcCallHandler", rpcReq, resp1)
			if err != nil {
				w.Write([]byte("请求成功:\n"))
				w.Write([]byte(resp1.Params["result"].(string)))
			} else {
				w.Write([]byte("rpc请求失败:\n"))
				w.Write([]byte(err.Error()))
			}
			matchUrl = true
			return
			break
		}
	}

	if !matchUrl {
		w.Write([]byte("无匹配的URL:\n"))
		w.Write([]byte(url))
		return
	}

	// 转发请求
	// rpc调用

	/*

		//
		rpcReq.Params["method"] = "lj_sb"

		if strings.Contains(url, "hgh") {
			client, err := rpc.DialHTTP("tcp", "192.168.1.101:9877")
			if err != nil {
				log.Fatalln("dialing:", err)
			}

			rpcResp := service.Resp{}
			//		rpcResp.Params = make(map[string]interface{})

			err = client.Call("ServiceTest.Addition", rpcReq, &rpcResp)
			if err != nil {
				log.Fatalln("rpc call error:", err)
			}

			w.Write([]byte(strconv.Itoa(rpcResp.Params["status"].(int)) + "\n"))
			w.Write([]byte(string(rpcResp.Params["resutl"].([]uint8)) + "\n"))
			w.Write([]byte(rpcResp.Params["time"].(string)))
		}
	*/

}

// rpc 调用
func rpcCall(client rpc.Client, serviceMethod string, req1 interface{}, resp1 interface{}) (err error) {
	err = client.Call(serviceMethod, req1, resp1)
	return
}

// 根据methodType、url获取匹配的处理方法
func matchHandler(methodType, url string) (s constants.Service) {
	var preciseMap map[string]constants.Service
	var regexMap map[string]constants.Service

	switch methodType {
	case "GET":
		preciseMap = constants.GetPreciseHandlers
		regexMap = constants.GetRegexHandlers
		break
	case "POST":
		break

	case "PUT":
		break

	case "DELETE":

		break

	default:
		log.Println("未识别的http请求类型")
	}

	preciseHandler := preciseMap[url]
	if preciseHandler.Id != "" {
		log.Println("找到匹配的服务接口")
		s = preciseHandler
		return
	}

	for k, v := range regexMap {
		fmt.Printf("url:%s,匹配url:%s,实际url:%s \n", v.Url, v.Regexp.String(), url)
		if v.Regexp.MatchString(url) {
			log.Printf("正则匹配到相应的处理方法,type:%s,url:%s \n", methodType, k)
			s = v
			return
		}
	}

	log.Printf("未找到methodType:%s,url:%s的处理方法!\n", methodType, url)
	return s
}

// 处理常用的请求
func commonRequest() {

}

// 选择节点
func chonseServerNode(s constants.Service) {
	for _, server := range s.Servers {
		fmt.Println(server.IP)
	}
}
