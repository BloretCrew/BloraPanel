package model

// RoleTemplate expands into explicit resource grants. Editing a template never
// silently changes existing grants; the administrator reviews each application.
type RoleTemplate struct {
	ID       string   `json:"roleId"`
	Name     string   `json:"name"`
	Actions  []string `json:"actions"`
	NodeOnly bool     `json:"nodeOnly"`
	Builtin  bool     `json:"builtin"`
	Revision int64    `json:"revision"`
}

func BuiltinRoles() []RoleTemplate {
	return []RoleTemplate{
		{ID: "reader", Name: "只读成员", Actions: []string{"instance.read", "file.read", "terminal.read"}, Builtin: true, Revision: 1},
		{ID: "operator", Name: "实例操作员", Actions: []string{"instance.read", "instance.start", "instance.stop", "instance.restart", "file.read", "file.write", "terminal.read", "terminal.input"}, Builtin: true, Revision: 1},
		{ID: "node-admin", Name: "节点管理员", Actions: []string{"node.read", "instance.read", "instance.create", "instance.configure", "instance.start", "instance.stop", "instance.restart", "instance.kill", "file.read", "file.write", "terminal.read", "terminal.input", "host.manage", "backup.create", "backup.restore", "schedule.create"}, NodeOnly: true, Builtin: true, Revision: 1},
	}
}

type UserPatch struct {
	Name     *string `json:"name,omitempty"`
	Admin    *bool   `json:"admin,omitempty"`
	Disabled *bool   `json:"disabled,omitempty"`
	Revision int64   `json:"revision"`
}
