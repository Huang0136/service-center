package beans

// 群组
type Group struct {
	Id      int      // ID
	Name    string   // 名称
	Desc    string   // 备注
	RoleIds []string // 角色
}
