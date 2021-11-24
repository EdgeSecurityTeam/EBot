//Author:shihuang@Edge Team
package main

import (
	"bufio"
	"fmt"
	"github.com/robfig/cron"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)


//主函数
func main() {
	Ebot()
	cron2 := cron.New() //创建一个cron实例
	//执行定时任务（每天10点执行一次）
	err:= cron2.AddFunc("0 0 10 * * ?", Ebot)
	if err!=nil{
		fmt.Println(err)
	}

	//启动/关闭
	cron2.Start()
	defer cron2.Stop()
	select {
	//查询语句，保持程序运行，在这里等同于for{}
	}
}

func Ebot() {
	d := getdaily()
	x := strformat(d)
	sendboot(x)
}

type Config struct {
	Key    string
}


func GetConfig() Config {
	//创建一个空的结构体,将本地文件读取的信息放入
	c := &Config{}
	//创建一个结构体变量的反射
	cr := reflect.ValueOf(c).Elem()
	//打开文件io流
	f, err := os.Open("config.ini")
	if err != nil {
		//log.Fatal(err)
		log.Println("[Error] configuration file error!!!")
		os.Exit(1)
	}
	defer func() {
		if err = f.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	//我们要逐行读取文件内容
	s := bufio.NewScanner(f)
	for s.Scan() {
		//以=分割,前面为key,后面为value
		var str = s.Text()
		var index = strings.Index(str, "=")
		var key = str[0:index]
		var value = str[index+1:]
		//通过反射将字段设置进去
		cr.FieldByName(key).Set(reflect.ValueOf(value))
	}
	err = s.Err()
	if err != nil {
		log.Println(err)
	}
	//返回Config结构体变量
	return *c
}

//发送消息
func sendboot(smg string){
	client := &http.Client{}
	now  := time.Now()
	Week := map[string]string{"1":"星期一","2":"星期二","3":"星期三","4":"星期四","5":"星期五","6":"星期六","7":"星期日"}
	sendtime := fmt.Sprintf("%d-%d-%d %d:%d:%d  %s",now.Year(),now.Month(),now.Day(),now.Hour(),now.Minute(),now.Second(),Week[strconv.Itoa(int(now.Weekday()))])
	title := "# 棱角社区攻防日报推送\\n" + sendtime + "\\n\\n\\n"
	text := title + smg
	textb := `{"msgtype": "markdown","markdown": {"content": "`+text+`"}}`
	var data = strings.NewReader(textb)
	c := GetConfig()
	req, err := http.NewRequest("POST", "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key="+c.Key, data)
	if err != nil {
		log.Println(err)
	}
	req.Header.Set("content-type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	log.Println(fmt.Sprintf("%s\n", bodyText))
}

//格式化消息
func strformat(s [][]string) string{
	body := ""
	for _, i := range s {
		text := "**标题：**" + i[1] + "\\n**地址：**[" + i[0] + "](" + i[0] + ")\\n**标签：**" +i[2]+ "\\n-----------------------\\n"
		body = body + text
	}
	return body
}



//抓取每天的日报
func getdaily() [][]string {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://forum.ywhack.com/forumdisplay.php?fid=59&orderby=lastpost&filter=86400", nil)
	if err != nil {
		log.Println(err)
	}
	req.Header.Set("cache-control", "max-age=0")
	req.Header.Set("upgrade-insecure-requests", "1")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.162 Safari/537.36")
	req.Header.Set("sec-fetch-dest", "document")
	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("sec-fetch-site", "none")
	req.Header.Set("sec-fetch-mode", "navigate")
	req.Header.Set("sec-fetch-user", "?1")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9")
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	src := string(bodyText)

	//去除STYLE
	re, _ := regexp.Compile("\\<style[\\S\\s]+?\\</style\\>")
	src = re.ReplaceAllString(src, "")

	//去除SCRIPT
	re, _ = regexp.Compile("\\<script[\\S\\s]+?\\</script\\>")
	src = re.ReplaceAllString(src, "")

	//去除连续的换行符
	re, _ = regexp.Compile("\\s{2,}")
	src = re.ReplaceAllString(src, "")

	re = regexp.MustCompile(`<a href="(.*?)" target="_blank">(.*?)</a>.*?<img src="" style="vertical-align: top;margin-top: 2px;"></div><small class="card-subtitle text-muted">.*?<span class="badge badge-tag">(.*?)</span></small>`)
	res := re.FindAllStringSubmatch(strings.TrimSpace(src),-1)
	var result [][]string
	for _, s := range res {
		result = append(result, s[1:])
	}
	return result
}
