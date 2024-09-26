package zero

type Instance struct {
	SelfID        int64    `json:"seld_id"`
	NickName      []string `json:"nickname"  yaml:"nickname"`            // 机器人名称
	CommandPrefix string   `json:"command_prefix" yaml:"command_prefix"` // 触发命令
	SuperUsers    []int64  `json:"super_users" yaml:"super_users"`       // 超级用户
}

func (c Config) GetNickName(selfID int64) []string {
	if instance, ok := c.InstanceMap[selfID]; ok {
		return instance.NickName
	}
	return c.NickName
}
