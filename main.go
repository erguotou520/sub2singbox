package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"erguotou520/sub2singbox/convert"
	"erguotou520/sub2singbox/httputils"
	"erguotou520/sub2singbox/model/clash"
)

const toolConfigPath = "sub.json"

// 定义toolConfig结构体
type toolConfig struct {
	Urls                 []string `json:"urls"`
	SbConfigTemplatePath string   `json:"sbConfigTemplatePath"`
	SbConfigPath         string   `json:"sbConfigPath"`
	Include              string   `json:"include"`
	Exclude              string   `json:"exclude"`
	Insecure             bool     `json:"insecure"`
	Ignore               bool     `json:"ignore"`
}

var (
	url                  string
	sbConfigTemplatePath string
	sbConfigPath         string
	include              string
	exclude              string
	insecure             bool
	ignore               bool
)

func init() {
	flag.StringVar(&url, "url", "", "订阅地址，多个链接使用 | 分割")
	flag.StringVar(&sbConfigTemplatePath, "t", "config-template.json", "sing-box 配置模板文件路径")
	flag.StringVar(&sbConfigPath, "o", "config.json", "输出的 sing-box 配置文件路径")
	flag.StringVar(&include, "include", "", "选择的节点")
	flag.StringVar(&exclude, "exclude", "网站|地址|剩余|时间|过期|到期|有效|官网", "排除的节点")
	flag.BoolVar(&insecure, "insecure", true, "所有节点不验证证书")
	flag.BoolVar(&ignore, "ignore", true, "忽略无法转换的节点")
	flag.Parse()
}

func main() {
	urls := make([]string, 0)
	if strings.Contains(url, "|") {
		urls = strings.Split(url, "|")
	} else {
		if url != "" {
			urls = append(urls, url)
		}
	}
	c := clash.Clash{}
	var singList []map[string]any
	if len(urls) > 0 {
		// 更新配置文件
		toolConfig := toolConfig{
			Urls: urls,
		}
		if sbConfigTemplatePath != "" {
			toolConfig.SbConfigTemplatePath = sbConfigTemplatePath
		}
		if sbConfigPath != "" {
			toolConfig.SbConfigPath = sbConfigPath
		}
		if include != "" {
			toolConfig.Include = include
		}
		if exclude != "" {
			toolConfig.Exclude = exclude
		}
		toolConfig.Insecure = insecure
		toolConfig.Ignore = ignore
		data, err := json.MarshalIndent(toolConfig, "", "  ")
		if err != nil {
			panic(err)
		}
		err = os.WriteFile(toolConfigPath, data, 0644)
		if err != nil {
			panic(err)
		}
	} else {
		file, err := os.ReadFile(toolConfigPath)
		if err != nil {
			log.Panicln("请使用 -url 指定订阅地址，多个地址使用 | 分割（没有空格）")
		}
		// 转成结构体
		var toolConfig toolConfig
		err = json.Unmarshal(file, &toolConfig)
		if err != nil {
			panic(err)
		}
		urls = toolConfig.Urls
		sbConfigTemplatePath = toolConfig.SbConfigTemplatePath
		sbConfigPath = toolConfig.SbConfigPath
		include = toolConfig.Include
		exclude = toolConfig.Exclude
		insecure = toolConfig.Insecure
		ignore = toolConfig.Ignore
	}
	if len(urls) == 0 {
		log.Panicln("未读取到订阅地址")
	}
	var err error
	templateB, err := os.ReadFile(sbConfigTemplatePath)
	if err != nil {
		log.Panicln("模板文件读取失败", err)
	}
	if templateB == nil {
		log.Panicln("请使用 -t 指定配置模板文件路径")
	}
	c, singList, tags, err := httputils.GetAny(context.TODO(), &http.Client{Timeout: 10 * time.Second}, urls, false, include, exclude)
	if err != nil {
		panic(err)
	}

	if insecure {
		convert.ToInsecure(&c)
	}

	// s, err := convert.Clash2sing(c)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	outb, err := convert.Template(templateB, singList, tags)
	if err != nil {
		panic(err)
	}
	os.WriteFile(sbConfigPath, outb, 0644)
}
