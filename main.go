/*
此包用于提取游戏项目中的包含中文的文件，并将其中的中文提取出来，以便可以放入数据库中
*/

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// 定义全局常量
const (
	CONFIGFILENAME = "config.ini"
	OUTPUTFILENAME = "output.txt"
)

// 定义全局变量
var (
	//用于存放配置文件里面的属性
	TargetPath     string
	TargetFileList []string

	//定义中文的正则表达式匹配模式
	zh_line_pattern = regexp.MustCompile("^[^/#]*\".*[\\p{Han}]+.*\"")
	zh_pattern      = regexp.MustCompile("\".*[\\p{Han}]+.*\"")
)

// 初始化方法，用于检查配置是否正确
func init() {
	//读取配置文件（一次性读取整个文件，则使用ioutil）
	bytes, err := ioutil.ReadFile(CONFIGFILENAME)
	if err != nil {
		panic(err)
	}

	//使用json反序列化
	content := make(map[string]string)
	if err = json.Unmarshal(bytes, &content); err != nil {
		panic(err)
	}

	//解析参数
	//单独定义ok变量，是为了避免在与全局变量合用时，将全局变量覆盖了
	var ok bool
	TargetPath, ok = content["TargetPath"]
	if !ok || len(TargetPath) == 0 {
		panic("不存在名为TargetPath的配置或配置为空")
	}

	targetFile, ok := content["TargetFile"]
	if !ok || len(targetFile) == 0 {
		panic("不存在名为TargetFile的配置或配置为空")
	}

	TargetFileList = strings.Split(strings.Replace(targetFile, " ", "", 100), ",")
}

// 应用程序主入口
func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("There are some errors:", err)
		}
	}()

	//获取目标文件列表（完整路径）
	files := getTargetFileList()

	//判断是否有目标文件
	if len(files) == 0 {
		panic("找不到指定的文件，请检查配置")
	}

	//读取文件内容
	chinese_list := extractChinese(files)

	//将内容写到输出文件
	writeToFile(chinese_list)

	//结束
	fmt.Println("提取完成，TotalCount:", len(chinese_list))
}

// 获取目标文件列表（完整路径）
// 返回值：目标文件列表（完整路径）
func getTargetFileList() []string {
	files := make([]string, 0, 5)

	//遍历目录，获取所有文件列表
	filepath.Walk(TargetPath, func(filename string, fi os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}

		//忽略目录
		if fi.IsDir() {
			return nil
		}

		//选择指定文件
		for _, value := range TargetFileList {
			if value == fi.Name() {
				files = append(files, filename)
			}
		}

		return nil
	})

	return files
}

// 读取每一个文件的内容
// filename:文件的绝对路径
// 返回值：提取出的中文列表
func readEachFile(filename string) []string {
	zhList := make([]string, 0, 100)

	//打开文件
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	//读取文件
	buf := bufio.NewReader(file)
	for {
		//按行读取
		line, _, err := buf.ReadLine()
		if err == io.EOF {
			break
		}

		//将byte[]转换为string
		lineStr := string(line)

		//判断是否匹配正则表达式
		if zh_line_pattern.MatchString(lineStr) {
			//提取其中的所有符合条件的项
			zhList_tmp := zh_pattern.FindAllString(lineStr, 5)
			zhList = append(zhList, zhList_tmp...)
		}
	}

	return zhList
}

// 提取文件中的中文
// targetFileList:目标文件列表（完整路径）
// 返回值：中文列表
func extractChinese(targetFileList []string) []string {
	chinese_list := make([]string, 0, 100)

	//用于临时保存已经添加过的列表，以避免重复
	chinese_map := make(map[string]bool)

	//遍历所有文件
	for _, targetFile := range targetFileList {
		zhList := readEachFile(targetFile)
		for _, value := range zhList {
			if _, ok := chinese_map[value]; !ok {
				chinese_list = append(chinese_list, value)
				chinese_map[value] = true
			}
		}
	}

	return chinese_list
}

// 将中文写入到文件中
// chinese_list:欲写入的中文
// 返回值：无
func writeToFile(chinese_list []string) {
	outfile, err := os.Create(OUTPUTFILENAME)
	if err != nil {
		panic(err)
	}
	defer outfile.Close()

	//遍历并写入文件
	for _, value := range chinese_list {
		outfile.WriteString(value + "\n")
	}
}
