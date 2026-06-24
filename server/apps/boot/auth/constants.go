package auth

// 业务魔数集中放在这里,避免在 service / handler 里散落字面量。
// 这些值与数据库 enum/状态字段对齐,如果改 DB 必须同步改这里。

// StatusActive 表示账号 / 用户在系统中处于"启用"状态。
// 业务上 1 = active,0 = disabled(与 accounts.status / users.status 列对齐)。
// 故意不写具体类型(untyped const),让 Go 在使用点自动转成调用方的
// 数值类型 —— 不同代码路径里 accountStatus 分别是 int8 / int16,
// 用 typed const 就要在每个调用点显式转换,反而更乱。
const StatusActive = 1

// RoleCodeUser 是默认兜底角色 code,用于注册新用户或用户没有任何角色时,
// 在 user_roles 写一条占位记录,让前端能拿到一个稳定的标识。
// 数据库种子: sys_roles.code='user'。
const RoleCodeUser = "user"

// RoleCodePlatform 是平台超级管理员的占位角色,出现在 PlatformLogin 响应中,
// 便于前端识别当前会话是平台域还是租户域。
const RoleCodePlatform = "_platform"
