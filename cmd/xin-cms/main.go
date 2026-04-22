package main

import (
	framework "gx1727.com/xin/framework"
	cms "gx1727.com/xin/module/cms"
)

func main() {
	framework.RegisterModule(cms.Module())
	framework.Run()
}
