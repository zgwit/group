package main

import (
	"classify/api"
	"classify/config"
	_ "classify/docs"
	"classify/types"
	"embed"
	"encoding/json"
	"github.com/zgwit/iot-master/v3/model"
	"github.com/zgwit/iot-master/v3/pkg/banner"
	"github.com/zgwit/iot-master/v3/pkg/build"
	"github.com/zgwit/iot-master/v3/pkg/db"
	"github.com/zgwit/iot-master/v3/pkg/log"
	"github.com/zgwit/iot-master/v3/pkg/mqtt"
	"github.com/zgwit/iot-master/v3/pkg/web"
	"net/http"
)

//go:embed all:app/classify
var wwwFiles embed.FS

// @title 设备分类数据接口文档
// @version 1.0 版本
// @description API文档
// @BasePath /app/classify/api/
// @query.collection.format multi
func main() {
	banner.Print("iot-master-plugin:classify")
	build.Print()

	config.Load()

	err := log.Open()
	if err != nil {
		log.Fatal(err)
	}

	//加载数据库
	err = db.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//同步表结构
	err = db.Engine.Sync2(
		new(types.Device), new(types.DeviceType), new(types.DeviceArea), new(types.DeviceGroup),
	)
	if err != nil {
		log.Fatal(err)
	}

	//MQTT总线
	err = mqtt.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer mqtt.Close()

	//注册应用
	payload, _ := json.Marshal(model.App{
		Id:   "classify",
		Name: "设备分类",
		Entries: []model.AppEntry{{
			Path: "app/classify/area",
			Name: "区域",
		}, {
			Path: "app/classify/group",
			Name: "分组",
		}, {
			Path: "app/classify/type",
			Name: "类型",
		}},
		Type:    "tcp",
		Address: "http://localhost" + web.GetOptions().Addr,
	})
	_ = mqtt.Publish("master/register", payload, false, 0)

	app := web.CreateEngine()

	//注册前端接口
	api.RegisterRoutes(app.Group("/app/classify/api"))

	//注册接口文档
	web.RegisterSwaggerDocs(app.Group("/app/classify"))

	//前端静态文件
	app.RegisterFS(http.FS(wwwFiles), "", "app/classify/index.html")

	//监听HTTP
	app.Serve()
}