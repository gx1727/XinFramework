import os
import re

FRAMEWORK_DIR = r"d:\work\xin\XinFramework\framework"

def main():
    # we need to replace `func Module(app *boot.App) plugin.Module`
    # with `func Module() plugin.Module`
    # and `app.DB` with `db.Get()`
    # and remove `gx1727.com/xin/framework/internal/core/boot` from imports
    # and add `gx1727.com/xin/framework/pkg/db`
    
    modules = ["user", "role", "auth", "tenant", "organization", "menu", "resource", "cms", "asset", "dict", "weixin"]
    for mod in modules:
        path = os.path.join(FRAMEWORK_DIR, "internal", "module", mod, "module.go")
        if not os.path.exists(path): continue
        
        with open(path, "r", encoding="utf-8") as f:
            content = f.read()
            
        content = content.replace("func Module(app *boot.App)", "func Module()")
        content = content.replace("app.DB", "db.Get()")
        content = content.replace("app.SessionMgr", "session.Get()")
        content = content.replace("app.PermService", "service.GetPermissionService()")
        content = content.replace("app.Config", "config.Get()")
        
        content = content.replace('"gx1727.com/xin/framework/internal/core/boot"\n', "")
        if "db.Get" in content and "pkg/db" not in content:
            content = content.replace('"gx1727.com/xin/framework/pkg/plugin"', '"gx1727.com/xin/framework/pkg/plugin"\n\t"gx1727.com/xin/framework/pkg/db"')
        
        # for auth module, it uses SessionMgr and PermService, we need to import them
        if mod == "auth":
            content = content.replace("app.SessionMgr", "session.Get()")
            if "pkg/session" not in content:
                content = content.replace('"gx1727.com/xin/framework/pkg/plugin"', '"gx1727.com/xin/framework/pkg/plugin"\n\t"gx1727.com/xin/framework/pkg/session"')
        
        # for permission module, it uses PermService
        if mod == "permission":
            content = content.replace("app.PermService", "service.GetPermissionService()")
            if "internal/service" not in content:
                content = content.replace('"gx1727.com/xin/framework/pkg/plugin"', '"gx1727.com/xin/framework/pkg/plugin"\n\t"gx1727.com/xin/framework/internal/service"')
                
        # for weixin module
        if mod == "weixin":
            content = content.replace("app.SessionMgr", "session.Get()")
            if "pkg/session" not in content:
                content = content.replace('"gx1727.com/xin/framework/pkg/plugin"', '"gx1727.com/xin/framework/pkg/plugin"\n\t"gx1727.com/xin/framework/pkg/session"')
            
        with open(path, "w", encoding="utf-8") as f:
            f.write(content)
            
    # Also update framework.go
    framework_go = os.path.join(FRAMEWORK_DIR, "framework.go")
    with open(framework_go, "r", encoding="utf-8") as f:
        content = f.read()
    content = content.replace("assetModule.Module(app)", "assetModule.Module()")
    content = content.replace("authModule.Module(app)", "authModule.Module()")
    content = content.replace("userModule.Module(app)", "userModule.Module()")
    content = content.replace("menuModule.Module(app)", "menuModule.Module()")
    content = content.replace("dictModule.Module(app)", "dictModule.Module()")
    content = content.replace("roleModule.Module(app)", "roleModule.Module()")
    content = content.replace("resourceModule.Module(app)", "resourceModule.Module()")
    content = content.replace("orgModule.Module(app)", "orgModule.Module()")
    content = content.replace("permModule.Module(app)", "permModule.Module()")
    content = content.replace("weixinModule.Module(app)", "weixinModule.Module()")
    
    with open(framework_go, "w", encoding="utf-8") as f:
        f.write(content)

if __name__ == "__main__":
    main()
