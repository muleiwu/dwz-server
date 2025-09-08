package interfaces

// 安装模块，应该支持安装检查，安装提交
type Installed interface {
	SetInstalled(installed bool)
	IsInstalled() bool
	Install(fun func() error) error
}
