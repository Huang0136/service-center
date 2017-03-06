package beans

const USER = "USER_"

// 用户
type User struct {
	Id       int
	UserName string   // 用户名
	Status   int      // 状态
	GroupIds []string // 群组
	RoleIds  []string // 角色
	Token    string   // 回话票据
}
