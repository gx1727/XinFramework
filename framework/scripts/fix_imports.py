import os

FRAMEWORK_DIR = r"d:\work\xin\XinFramework\framework"

def main():
    modules = ["user", "role", "auth", "tenant", "organization", "menu", "resource", "cms", "asset", "dict", "weixin", "permission"]
    for mod in modules:
        path = os.path.join(FRAMEWORK_DIR, "internal", "module", mod, "module.go")
        if not os.path.exists(path): continue
        
        with open(path, "r", encoding="utf-8") as f:
            content = f.read()
        
        # fix double imports
        content = content.replace("import import (", "import (")
        
        if "config.Get" in content and "pkg/config" not in content:
            content = content.replace('import (', 'import (\n\t"gx1727.com/xin/framework/pkg/config"\n')
        if "db.Get" in content and "pkg/db" not in content:
            content = content.replace('import (', 'import (\n\t"gx1727.com/xin/framework/pkg/db"\n')
        if "session.Get" in content and "pkg/session" not in content:
            content = content.replace('import (', 'import (\n\t"gx1727.com/xin/framework/pkg/session"\n')
        if "service.GetPermissionService" in content and "internal/service" not in content:
            content = content.replace('import (', 'import (\n\t"gx1727.com/xin/framework/internal/service"\n')
            
        with open(path, "w", encoding="utf-8") as f:
            f.write(content)

if __name__ == "__main__":
    main()
