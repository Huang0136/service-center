package beans

const RES = "RES_"

// 资源
type Resource struct {
	Method    string   // 方法类型
	URL       string   // URL
	Name      string   // 名称
	NeedLogin bool     // 是否等
	Desc      string   // 备注
	RoleIds   []string // 角色
}
