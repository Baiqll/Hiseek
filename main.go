package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"time"
	"os/exec"
	"regexp"

	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-resty/resty/v2"
)


var url_set []string
var domain_set []string
var client resty.Client
var is_proxy bool

func main() {

	var banner = `
 _______ __                     __    
|   |   |__|.-----.-----.-----.|  |--.
|       |  ||__ --|  -__|  -__||    < 
|___|___|__||_____|_____|_____||__|__|
                                   v 1.0.1

Search from "https://web.archive.org/cdx/search/cdx" matching urls containing a specific name
example: Hiseek -d example.com -s jump,proxy ...
    `
	fmt.Println(string(banner))

	now:=time.Now().Format("2006-01-02 15:04:05")
	fmt.Println("[*] Starting search @ ",now)

	var domain string
	var search string
	var exclude string
	var world_dict string
	var repeat string
	var out_domain_path string
	var proxy string
	var online string

	// &domain 就是接收命令行中输入 -d 后面的参数值，其他同理
	flag.StringVar(&domain, "d", "", "域名")
	flag.StringVar(&search, "s", "", "查询匹配字符 （可同时匹配多个字符,使用 ',' 隔开）")
	flag.StringVar(&exclude, "e", "", "排除匹配字符 （可同时匹配多个字符,使用 ',' 隔开）")
	flag.StringVar(&world_dict, "w", "", "子域名字典")
	flag.StringVar(&repeat, "re", "true", "是否去除重复path  (true,false)")
	flag.StringVar(&out_domain_path, "od", "", "导出域名字典")
	flag.StringVar(&proxy, "proxy", "", "使用代理")
	flag.StringVar(&online, "online", "false", "检查是否在线-域名是否解析正常 (true,false)")

	// 解析命令行参数写入注册的flag里
	flag.Parse()

	// 使用proxy 代理
	if proxy != "" {

		is_proxy = true

		_, parseErr := url.Parse(proxy)
		if parseErr != nil {
			return
		}
		client = *resty.New()
		client.SetProxy(proxy)
		client.SetTimeout(0)

	}

	fmt.Println("[*] matching result:")

	// 如果管道有参数传递
	if has_stdin() {

		s := bufio.NewScanner(os.Stdin)

		for s.Scan() {

			if online == "true" {
				if IsOnline(s.Text()) {
					search_web_archive(s.Text(), search, exclude, repeat, out_domain_path)
					continue
				}
			}

			search_web_archive(s.Text(), search, exclude, repeat, out_domain_path)
		}

	}

	// 使用子域名字典
	if world_dict != "" {
		for _, sub_domain := range get_word_dict_list(world_dict) {

			// 字典可使用子域名字典、子域名名称字典
			// 字典可以是 test.example.com 也可以是 test

			// 如果有指定域名
			if domain != "" {
				sub_domain = strings.Replace(strings.Replace(sub_domain, "\n", "", -1), "."+domain, "", -1) + "." + domain
			}

			// 如果sub_domain不是合法的域名（）
			if is_domain, _ := regexp.MatchString(`[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+\.?`, sub_domain); !is_domain {
				continue
			}

			fmt.Println(sub_domain)
			if online == "true" {
				if IsOnline(sub_domain) {
					search_web_archive(sub_domain, search, exclude, repeat, out_domain_path)
					continue
				}
			}

			search_web_archive(sub_domain, search, exclude, repeat, out_domain_path)
		}

	} else {
		// 查询是否包含
		if online == "true" {
			if IsOnline(domain) {
				search_web_archive(domain, search, exclude, repeat, out_domain_path)
				return
			}
		}

		search_web_archive(domain, search, exclude, repeat, out_domain_path)

	}

	// 导出域名字典
	if out_domain_path != "" {
		save_domain(out_domain_path, domain)
	}

}

func search_web_archive(domain string, search string, exclude string, repeat string, out_domain_path string) string {

	// 替换逗号
	match_search := strings.Replace(search, ",", ")|(", -1)
	search_pattern := "(" + match_search + ")"

	match_exclude := strings.Replace(exclude, ",", ")|(", -1)
	exclude_pattern := "(" + match_exclude + ")"

	params := url.Values{}
	Url, _ := url.Parse("https://web.archive.org/cdx/search/cdx")

	if domain == "" {
		return "domain为空"
	}

	// 参数
	params.Set("url", "*."+domain+"/*")
	params.Set("output", "text")
	params.Set("fl", "original")
	params.Set("collapse", "urlkey")

	Url.RawQuery = params.Encode()
	urlPath := Url.String()

	resp, _ := http.Get(urlPath)
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	// fmt.Println(reflect.TypeOf(body))

	url_list := strings.Split(string(body), "\n")

	// 遍历返回内容
	for _, value := range url_list {
		// 判断是否包含特殊字段, 如果匹配的上就打印

		if is_matching, _ := regexp.MatchString(search_pattern, value); is_matching {

			path := strings.Split(strings.Replace(value, "https", "http", -1), "?")[0]

			if exclude != "" {
				// 排除含有特殊字符的 例如排除 www.domain.com

				if is_exclude, _ := regexp.MatchString(exclude_pattern, path); is_exclude {
					continue
				}

			}

			// 判断是否去重
			if repeat == "false" {
				if is_proxy {
					client.R().Get(value)
				}
				fmt.Println(value)
				continue
			}

			// 判断是否是新的path ,是就保存
			if ok := IsContainStr(url_set, path); !ok {
				url_set = append(url_set, path)
				if is_proxy {
					client.R().Get(value)
				}
				fmt.Println(value)
			}

		}
	}

	return "结束"
}

// 判断元素是否存在于数组中
func IsContainStr(items []string, item string) bool {
	for _, eachItem := range items {
		if eachItem == item {
			return true
		}
	}
	return false
}

// 判断网络是否正常访问
func IsOnline(domain string) bool {

	cmd := exec.Command("ping", domain, "-c", "1", "-W", "5")
	err := cmd.Run()

	if err != nil {
		return false
	}

	return true
}

// 解析url 获取域名信息
func get_domain(url_path string, domain string) (sub_domain string, err error) {

	r, err := url.Parse(url_path)

	if err != nil {
		// fmt.Println("url解析出错：",url_path)
		return "", err
	}

	return r.Hostname(), nil

}

// 子域名字典
func save_domain(filename string, domain string) {

	var dict_list string
	var sub_domain string
	var sub_domain_list []string

	for _, path := range url_set {

		// 收集子域名
		if new_domain, err := get_domain(path, domain); err == nil {

			// 子域名
			if ok := IsContainStr(sub_domain_list, new_domain); !ok {

				sub_domain_list = append(sub_domain_list, new_domain)
				sub_domain += new_domain + "\n"

			}

			/*
				子域名拆分
			*/

			new_dict_list := strings.Split(strings.Replace(new_domain, "."+domain, "", -1), ".")

			// new_domain_list := append(strings.Split(new_domain, "-"), domain)

			for _, new_name := range new_dict_list {

				if ok := IsContainStr(domain_set, new_name); !ok {

					domain_set = append(domain_set, new_name)
					dict_list += new_name + "\n"
				}

			}

		}

	}

	// if err := ioutil.WriteFile(filename, []byte(arrayToString(domain_set, "\n")), 0666); err != nil {
	if err := ioutil.WriteFile("sub_domain_"+filename, []byte(sub_domain), 0666); err != nil {
		fmt.Println("子域名保存失败")
	}

	if err := ioutil.WriteFile("dict_"+filename, []byte(dict_list), 0666); err != nil {
		fmt.Println("字典保存失败")
	}

}

// 字符串数组转字符串
func arrayToString(arr []string, add string) string {

	var result string

	for _, i := range arr { //遍历数组中所有元素追加成string

		result += i + add

	}
	return result
}

// 返回字典列表
func get_word_dict_list(file_path string) []string {

	f, err := os.Open(file_path)
	if err != nil {
		fmt.Println(err.Error())
	}
	//建立缓冲区，把文件内容放到缓冲区中
	buf := bufio.NewReader(f)
	var dict_list []string
	for {
		//遇到\n结束读取
		b, errR := buf.ReadBytes('\n')

		if errR == io.EOF {
			break
		}
		dict_list = append(dict_list, string(b))
	}

	return dict_list

}

// 获取 linux 管道传递的参数
func has_stdin() bool {

	// resList := make([]string, 0, 0)

	fileInfo, _ := os.Stdin.Stat()
	if (fileInfo.Mode() & os.ModeNamedPipe) != os.ModeNamedPipe {
		return false
	}

	return true
	// s := bufio.NewScanner(os.Stdin)

	// for s.Scan() {
	// 	resList = append(resList, s.Text())
	// }

	// return "管道有数据" ,true
}
