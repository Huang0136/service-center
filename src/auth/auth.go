// 认证和授权
package auth

import (
	"log"

	"beans"

	"nosql/my.redis"
)

// 身份认证
func Authentication(userId, token string) (userInfo beans.User, pass bool, msg string) {
	// 从reids中获取
	userInfo, err := my_redis.GetUserInfoByUserId(userId)
	if err != nil {
		pass = false
		msg = "认证失败，redis缓存查询失败"
		return
	}

	if userInfo.Id != 0 && userInfo.Token == token {
		pass = true
		return
	}

	log.Println(userInfo)
	msg = "用户未登录或令牌失效"
	return
}

// 授权
func Authorization(resourceRoles []string, userRoles []string) bool {
	if len(resourceRoles) == 0 { // 资源未授予任何角色
		return false
	}

	if len(userRoles) == 0 { // 用户没有任何角色
		return false
	}

	for resRole := range resourceRoles {
		for userRole := range userRoles {
			if resRole == userRole { // 具有权限
				return true
				break
			}
		}
	}

	return false
}
