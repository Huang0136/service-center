// 认证和授权
package auth

// 身份认证
func Authentication(userName, pwd string) (pass bool, msg string) {
	if userName == pwd {
		pass = true
		return
	}

	msg = "用户名和密码不正确"
	return
}

// 授权
func Authorization() {
}
