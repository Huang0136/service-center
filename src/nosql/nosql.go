package nosql

import (
	"beans"
)

// 资源操作接口
type ResourceOperate interface {
	GetResourceByMethodAndUrl(method, url string) (beans.Resource, error)
	SaveResource(r beans.Resource) error
}

// NoSQL用户信息操作
type UserInfoOper interface {
	GetUserInfoByUserId(id string) (beans.User, error)
	SaveOrUpdateUserInfo(u beans.User) error
	DeleteUserInfo(id string) error
}
