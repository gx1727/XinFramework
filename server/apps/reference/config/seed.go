// Package config 通用配置 - __template__ 启动期 seed 自检
package config

import (
	"context"
	"fmt"
	"log"

	"gx1727.com/xin/framework/pkg/db"
)

// TemplateTenantCode 与 apps/boot/tenant/first_install.go 保持一致
const TemplateTenantCode = "__template__"

// EnsureTemplateSeeded 启动期自检：如果 __template__ 租户下没有 config_groups，
// 自动补种 4 个预置分组 + 19 个预置项 + 1 个菜单 + 5 个资源。
//
// 目的：解决"已部署过老 framework.sql 的库"在新 framework.sql 加了 config seed
// 之后，_schema_migrations 跳过 framework.sql 导致 seed 不跑的问题。
//
// 幂等：所有 INSERT 用 ON CONFLICT DO NOTHING，重复执行无副作用。
// Bypass RLS：__template__ 的 tenant_id 不为 0，必须用 RunInPlatformTx 才能写入。
func EnsureTemplateSeeded(ctx context.Context) error {
	return db.RunInPlatformTx(ctx, db.Get(), func(ctx context.Context) error {
		q, err := db.GetQuerier(ctx)
		if err != nil {
			return err
		}

		// 1) 查 __template__ 租户 id
		var templateID uint
		if err := q.QueryRow(ctx,
			`SELECT id FROM tenants WHERE code = $1 AND is_deleted = FALSE LIMIT 1`,
			TemplateTenantCode,
		).Scan(&templateID); err != nil {
			return fmt.Errorf("lookup __template__ tenant: %w", err)
		}
		if templateID == 0 {
			// __template__ 还没建（极少发生，可能是 framework.sql 未跑）；跳过
			return nil
		}

		// 2) 检查 __template__ 是否已有 config_groups
		var n int
		if err := q.QueryRow(ctx,
			`SELECT COUNT(*) FROM config_groups WHERE tenant_id = $1 AND is_deleted = FALSE`,
			templateID,
		).Scan(&n); err != nil {
			return fmt.Errorf("count config_groups: %w", err)
		}
		if n > 0 {
			// 已有数据，跳过 seed
			return nil
		}

		// 3) 幂等 seed 4 个预置分组
		groups := []struct {
			code, name, desc, icon string
			sort                   int
			isPublic               bool
		}{
			{"site", "站点信息", "站点名称、Logo、版权等公开信息", "GlobeIcon", 1, true},
			{"security", "安全策略", "密码强度、会话超时等安全相关配置", "ShieldIcon", 2, false},
			{"email", "邮件服务", "SMTP 邮件服务配置", "MailIcon", 3, false},
			{"feature_flag", "功能开关", "系统级功能启用/禁用开关", "ToggleLeftIcon", 4, false},
		}
		for _, g := range groups {
			if _, err := q.Exec(ctx, `
				INSERT INTO config_groups
					(tenant_id, code, name, description, icon, sort, is_system, is_public, status)
				VALUES ($1, $2, $3, $4, $5, $6, TRUE, $7, 1)
				ON CONFLICT (tenant_id, code) WHERE is_deleted = FALSE DO NOTHING`,
				templateID, g.code, g.name, g.desc, g.icon, g.sort, g.isPublic,
			); err != nil {
				return fmt.Errorf("seed group %s: %w", g.code, err)
			}
		}

		// 4) 幂等 seed site items（7 项）
		if err := seedSiteItems(ctx, q, templateID); err != nil {
			return err
		}
		// 5) 幂等 seed security items（5 项）
		if err := seedSecurityItems(ctx, q, templateID); err != nil {
			return err
		}
		// 6) 幂等 seed email items（6 项）
		if err := seedEmailItems(ctx, q, templateID); err != nil {
			return err
		}
		// 7) 幂等 seed feature_flag items（2 项）
		if err := seedFeatureFlagItems(ctx, q, templateID); err != nil {
			return err
		}

		// 8) 菜单：系统管理 → 配置管理
		// 注意：parent_id 必须是 system 菜单在 __template__ 里的实际 id（不是 default 里的 5）
		// ancestors 留空，下面 UPDATE 重建
		if _, err := q.Exec(ctx, `
			INSERT INTO menus (tenant_id, code, name, subtitle, url, path, icon, sort, parent_id, ancestors, visible, enabled)
			SELECT $1, 'config', '配置管理', '系统配置项管理', '', '/settings', 'SettingsIcon', 0,
			       (SELECT id FROM menus WHERE code = 'system' AND tenant_id = $1 AND is_deleted = FALSE),
			       '', TRUE, TRUE
			ON CONFLICT (tenant_id, code) WHERE is_deleted = FALSE DO NOTHING`,
			templateID,
		); err != nil {
			return fmt.Errorf("seed config menu: %w", err)
		}

		// 重建 config menu 的 ancestors（与 framework.sql 2c 段保持一致）
		if _, err := q.Exec(ctx, `
			UPDATE menus SET ancestors = parent_id::text
			WHERE tenant_id = $1 AND code = 'config' AND parent_id > 0 AND is_deleted = FALSE`,
			templateID,
		); err != nil {
			return fmt.Errorf("update config menu ancestors: %w", err)
		}

		// 9) 资源：config:list/get/create/update/delete
		resources := []struct {
			code, name, action, desc string
			sort                     int
		}{
			{"config:list", "查询配置", "list", "查询配置分组与项", 1},
			{"config:get", "查看配置", "get", "查看分组/项详情", 2},
			{"config:create", "创建配置", "create", "新建分组或项", 3},
			{"config:update", "更新配置", "update", "更新分组或项", 4},
			{"config:delete", "删除配置", "delete", "删除分组或项", 5},
		}
		for _, r := range resources {
			if _, err := q.Exec(ctx, `
				INSERT INTO resources (tenant_id, menu_id, code, name, action, description, sort, status)
				SELECT $1,
				       (SELECT id FROM menus WHERE code = 'config' AND tenant_id = $1 AND is_deleted = FALSE),
				       $2, $3, $4, $5, $6, 1
				ON CONFLICT (tenant_id, code) WHERE is_deleted = FALSE DO NOTHING`,
				templateID, r.code, r.name, r.action, r.desc, r.sort,
			); err != nil {
				return fmt.Errorf("seed resource %s: %w", r.code, err)
			}
		}

		// 10) 序列号兜底
		if _, err := q.Exec(ctx, `SELECT setval('config_groups_id_seq', GREATEST(
			(SELECT COALESCE(MAX(id), 0) FROM config_groups),
			$1 * 1000), true)`, templateID); err != nil {
			return fmt.Errorf("setval config_groups_id_seq: %w", err)
		}
		if _, err := q.Exec(ctx, `SELECT setval('config_items_id_seq', GREATEST(
			(SELECT COALESCE(MAX(id), 0) FROM config_items),
			$1 * 1000), true)`, templateID); err != nil {
			return fmt.Errorf("setval config_items_id_seq: %w", err)
		}

		return nil
	})
}

// seedSiteItems seed site 分组下的 7 个项
func seedSiteItems(ctx context.Context, q db.Querier, tenantID uint) error {
	items := []struct {
		key, val, defaultVal, typ, label, desc string
		sort                                   int
		isPublic                               bool
	}{
		{"site_name", `"XinFramework"`, `"XinFramework"`, "string", "站点名称", "显示在页面标题、登录页等位置", 1, true},
		{"site_logo", `""`, `""`, "image", "站点 Logo", "建议 PNG/SVG，背景透明", 2, true},
		{"site_favicon", `""`, `""`, "image", "Favicon", "浏览器标签图标", 3, true},
		{"site_copyright", `""`, `""`, "string", "版权信息", "页面底部显示", 4, true},
		{"site_icp", `""`, `""`, "string", "ICP 备案号", "中国大陆站点必填", 5, true},
		{"site_locale_default", `"zh-CN"`, `"zh-CN"`, "select", "默认语言", "zh-CN / en-US", 6, true},
		{"login_background", `""`, `""`, "image", "登录页背景", "登录页右侧大图", 7, true},
	}
	for _, it := range items {
		if _, err := q.Exec(ctx, `
			INSERT INTO config_items
			    (tenant_id, group_id, key, value, default_value, type, label, description, sort, is_public, is_system, status)
			SELECT $1,
			       (SELECT id FROM config_groups WHERE code = 'site' AND tenant_id = $1 AND is_deleted = FALSE),
			       $2, $3::jsonb, $4::jsonb, $5, $6, $7, $8, $9, TRUE, 1
			ON CONFLICT (tenant_id, group_id, key) WHERE is_deleted = FALSE DO NOTHING`,
			tenantID, it.key, it.val, it.defaultVal, it.typ, it.label, it.desc, it.sort, it.isPublic,
		); err != nil {
			return fmt.Errorf("seed site item %s: %w", it.key, err)
		}
	}
	return nil
}

// seedSecurityItems seed security 分组下的 5 个项
func seedSecurityItems(ctx context.Context, q db.Querier, tenantID uint) error {
	items := []struct {
		key, val, defaultVal, typ, label, desc, validation string
		sort                                               int
	}{
		{"password_min_length", `8`, `8`, "number", "密码最小长度", "新建/修改密码时校验", `{"min":6,"max":32,"required":true}`, 1},
		{"password_complexity", `"standard"`, `"standard"`, "select", "密码复杂度", "low/standard/strong", `[{"label":"低(纯字母数字)","value":"low"},{"label":"标准(字母+数字)","value":"standard"},{"label":"强(字母+数字+符号)","value":"strong"}]`, 2},
		{"session_timeout_min", `30`, `30`, "number", "会话超时(分钟)", "空闲超过此时间强制下线", `{"min":5,"max":1440,"required":true}`, 3},
		{"max_login_attempts", `5`, `5`, "number", "最大登录失败次数", "超过后锁定账户", `{"min":1,"max":20,"required":true}`, 4},
		{"lock_duration_min", `5`, `5`, "number", "锁定时长(分钟)", "失败次数超限后的锁定时长", `{"min":1,"max":1440,"required":true}`, 5},
	}
	for _, it := range items {
		if _, err := q.Exec(ctx, `
			INSERT INTO config_items
			    (tenant_id, group_id, key, value, default_value, type, label, description, validation, sort, is_public, is_system, status)
			SELECT $1,
			       (SELECT id FROM config_groups WHERE code = 'security' AND tenant_id = $1 AND is_deleted = FALSE),
			       $2, $3::jsonb, $4::jsonb, $5, $6, $7, $8::jsonb, $9, FALSE, TRUE, 1
			ON CONFLICT (tenant_id, group_id, key) WHERE is_deleted = FALSE DO NOTHING`,
			tenantID, it.key, it.val, it.defaultVal, it.typ, it.label, it.desc, it.validation, it.sort,
		); err != nil {
			return fmt.Errorf("seed security item %s: %w", it.key, err)
		}
	}
	return nil
}

// seedEmailItems seed email 分组下的 6 个项
func seedEmailItems(ctx context.Context, q db.Querier, tenantID uint) error {
	items := []struct {
		key, val, defaultVal, typ, label, desc string
		sort                                   int
		isReadonly                             bool
	}{
		{"smtp_host", `""`, `""`, "string", "SMTP 主机", "如 smtp.example.com", 1, false},
		{"smtp_port", `465`, `465`, "number", "SMTP 端口", "常用 25/465/587", 2, false},
		{"smtp_user", `""`, `""`, "string", "SMTP 用户", "通常为邮箱地址", 3, false},
		{"smtp_password", `""`, `""`, "password", "SMTP 密码", "授权码或登录密码", 4, true},
		{"smtp_from", `""`, `""`, "string", "发件人邮箱", "邮件 From 头", 5, false},
		{"smtp_use_tls", `true`, `true`, "boolean", "启用 TLS", "465 通常 TLS，587 STARTTLS", 6, false},
	}
	for _, it := range items {
		if _, err := q.Exec(ctx, `
			INSERT INTO config_items
			    (tenant_id, group_id, key, value, default_value, type, label, description, sort, is_public, is_readonly, is_system, status)
			SELECT $1,
			       (SELECT id FROM config_groups WHERE code = 'email' AND tenant_id = $1 AND is_deleted = FALSE),
			       $2, $3::jsonb, $4::jsonb, $5, $6, $7, $8, FALSE, $9, TRUE, 1
			ON CONFLICT (tenant_id, group_id, key) WHERE is_deleted = FALSE DO NOTHING`,
			tenantID, it.key, it.val, it.defaultVal, it.typ, it.label, it.desc, it.sort, it.isReadonly,
		); err != nil {
			return fmt.Errorf("seed email item %s: %w", it.key, err)
		}
	}
	return nil
}

// seedFeatureFlagItems seed feature_flag 分组下的 2 个项
func seedFeatureFlagItems(ctx context.Context, q db.Querier, tenantID uint) error {
	items := []struct {
		key, val, defaultVal, typ, label, desc string
		sort                                   int
	}{
		{"enable_registration", `true`, `true`, "boolean", "开放注册", "允许外部用户自助注册", 1},
		{"enable_audit_log", `true`, `true`, "boolean", "审计日志", "记录关键操作审计日志", 2},
	}
	for _, it := range items {
		if _, err := q.Exec(ctx, `
			INSERT INTO config_items
			    (tenant_id, group_id, key, value, default_value, type, label, description, sort, is_public, is_system, status)
			SELECT $1,
			       (SELECT id FROM config_groups WHERE code = 'feature_flag' AND tenant_id = $1 AND is_deleted = FALSE),
			       $2, $3::jsonb, $4::jsonb, $5, $6, $7, $8, FALSE, TRUE, 1
			ON CONFLICT (tenant_id, group_id, key) WHERE is_deleted = FALSE DO NOTHING`,
			tenantID, it.key, it.val, it.defaultVal, it.typ, it.label, it.desc, it.sort,
		); err != nil {
			return fmt.Errorf("seed feature_flag item %s: %w", it.key, err)
		}
	}
	return nil
}

// HealConfigMenuParent 启动期自愈：修复 config menu 的 parent_id。
//
// 历史 bug：老 framework.sql 写死 parent_id=5，但 __template__ 里 system 菜单
// 实际 id 已被 setval 推到几千，导致 config menu 成了孤儿（菜单树不显示）。
//
// 本函数在平台事务（bypass RLS）下扫描所有租户，把 config menu 的 parent_id
// 改成该租户 system 菜单的实际 id，并重建 ancestors。
//
// 幂等：parent_id 已正确时 no-op。
func HealConfigMenuParent(ctx context.Context) error {
	return db.RunInPlatformTx(ctx, db.Get(), func(ctx context.Context) error {
		q, err := db.GetQuerier(ctx)
		if err != nil {
			return err
		}

		// 修复所有租户里 config menu 的 parent_id
		// 用子查询拿同租户的 system 菜单 id，确保正确映射
		res, err := q.Exec(ctx, `
			UPDATE menus AS cfg
			SET parent_id = sys.id,
			    ancestors = sys.id::text
			FROM menus AS sys
			WHERE cfg.code = 'config'
			  AND cfg.is_deleted = FALSE
			  AND sys.code = 'system'
			  AND sys.is_deleted = FALSE
			  AND cfg.tenant_id = sys.tenant_id
			  AND cfg.parent_id != sys.id
			  AND sys.id IS NOT NULL`,
		)
		if err != nil {
			return fmt.Errorf("heal config menu parent_id: %w", err)
		}
		if n := res.RowsAffected(); n > 0 {
			log.Printf("[config] healed %d config menu(s) with wrong parent_id", n)
		}
		return nil
	})
}
