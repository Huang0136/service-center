package my_redis

import (
	"beans"
	"constants"
	"fmt"
	"log"
	"strconv"
	"strings"
)

// Redis操作资源
type RedisResouceOperate int

// 根据method,url获取资源
func (o *RedisResouceOperate) GetResourceByMethodAndUrl(method, url string) (r beans.Resource, err error) {
	rs := constants.RedisDB.HGetAll(beans.RES + method + "_" + url)
	if rs.Err() != nil {
		log.Println("根据method、url查询缓存中的资源失败", rs.Err())
		return
	}
	fmt.Println("redis数据:", rs)

	val := rs.Val()
	if val["roleIds"] == "" {
		return
	}

	roleIds := strings.Split(val["roleIds"], ",")
	var needLogin bool = false
	if val["needLogin"] != "0" {
		needLogin = true
	}

	r = beans.Resource{Method: method, URL: url, Name: val["name"], Desc: val["desc"], NeedLogin: needLogin, RoleIds: roleIds}
	return
}

// 保存资源
func (o *RedisResouceOperate) SaveResource(r beans.Resource) error {
	return nil
}

// 用户信息
func GetUserInfoByUserId(userId string) (u beans.User, err error) {
	ss := constants.RedisDB.HGetAll(beans.USER + userId)

	if ss.Err() != nil {
		log.Println("查询redis用户信息失败", ss.Err())
		err = ss.Err()
		return
	}

	rs, err := ss.Result()
	if err != nil {
		log.Println("获取reids用户信息失败", err)
		return
	}

	u = beans.User{}
	u.Id, _ = strconv.Atoi(userId)
	u.Token = rs["token"]
	u.UserName = rs["user_name"]
	u.Status, _ = strconv.Atoi(rs["status"])

	u.RoleIds = strings.Split(rs["roles"], ",")
	return
}
