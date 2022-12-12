package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// 定义结构体Movie，映射数据表movies
type Movie struct {
	// 定义数据表字段
	gorm.Model
	Name        string `gorm:"type:varchar(50)"`
	Star        string `gorm:"type:varchar(50)"`
	Releasetime string `gorm:"type:varchar(50)"`
	Score       string `gorm:"type:varchar(50)"`
}

func get_data(offset string) string {
	urls := "https://maoyan.com/board/4?offset=" + offset
	fmt.Print(urls)
	// 定义请求对象NewRequest
	req, _ := http.NewRequest("GET", urls, nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.81 Safari/537.36")
	transport := &http.Transport{}
	// 在Client设置参数Transport即可实现代理IP
	client := &http.Client{Transport: transport}
	// 发送HTTP请求
	resp, _ := client.Do(req)
	// 获取网站响应内容
	body, _ := ioutil.ReadAll(resp.Body)
	// 网页响应内容转码
	result := string(body)
	// 设置延时，请求过快会引发反爬
	time.Sleep(5 * time.Second)
	return result
}

func clean_data(data string) []map[string]string {
	// 使用goquery解析HTML代码
	dom, _ := goquery.NewDocumentFromReader(strings.NewReader(data))
	// 定义变量result和info
	result := []map[string]string{}
	var info map[string]string
	// 遍历网页所有电影信息
	selection := dom.Find(".board-item-content")
	selection.Each(func(i int, selection *goquery.Selection) {
		// 记录每部电影信息，每存储一部电影必须清空集合
		info = map[string]string{}
		name := selection.Find(".name").Text()
		star := selection.Find(".star").Text()
		releasetime := selection.Find(".releasetime").Text()
		score := selection.Find(".score").Text()
		info["name"] = strings.TrimSpace(name)
		info["star"] = strings.TrimSpace(star)
		info["releasetime"] = strings.TrimSpace(releasetime)
		info["score"] = strings.TrimSpace(score)
		// 将电影信息写入切片
		result = append(result, info)
	})
	return result
}

func save_data(data []map[string]string) {
	// 连接数据库
	dsn := `root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local`
	db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	sqlDB, _ := db.DB()
	// 关闭数据库，释放资源
	defer sqlDB.Close()
	// 执行数据迁移
	db.AutoMigrate(&Movie{})
	// 遍历变量data，获取每部电影信息
	for _, k := range data {
		fmt.Printf("当前数据：%v\n", k)
		// 查找电影是否已在数据库
		var m []Movie
		db.Where("name = ?", k["name"]).First(&m)
		// len(m)等于0说明数据不存在数据库
		if len(m) == 0 {
			// 新增数据
			m1 := Movie{Name: k["name"], Star: k["star"],
				Releasetime: k["releasetime"], Score: k["score"]}
			db.Create(&m1)
		} else {
			// 更新数据
			db.Where("name = ?", k["name"]).
				Find(&m).Update("score", k["score"])
		}
	}
}

func main() {
	// 遍历10次，每次遍历代表不同页的网页信息
	for i := 0; i < 10; i++ {
		// 函数调用
		// 调用次序：发送HTTP请求->清洗数据->数据入库
		webData := get_data(strconv.Itoa(i * 10))
		cleanData := clean_data(webData)
		save_data(cleanData)
	}
}
