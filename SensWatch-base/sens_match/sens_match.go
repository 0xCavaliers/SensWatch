package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// AddressName 处理地址和人名识别
type AddressName struct {
	addressPattern    *regexp.Regexp
	personNamePattern *regexp.Regexp
}

// NewAddressName 创建新的 AddressName 实例
func NewAddressName() *AddressName {
	return &AddressName{
		addressPattern:    regexp.MustCompile(`(ns|nsf)`),
		personNamePattern: regexp.MustCompile(`nr`),
	}
}

// convertToUTF8 将可能的GBK/GB2312编码转换为UTF-8
func convertToUTF8(content []byte) (string, error) {
	// 尝试检测是否为GBK编码
	reader := transform.NewReader(strings.NewReader(string(content)), simplifiedchinese.GBK.NewDecoder())
	utf8Content, err := ioutil.ReadAll(reader)
	if err != nil {
		return string(content), err // 如果转换失败，返回原始内容
	}
	return string(utf8Content), nil
}

// CheckChineseAddress 检测中文地址和姓名
func (s *SensMatch) CheckChineseAddress(content string) []string {
	// 创建临时文件
	tmpFile, err := ioutil.TempFile("", "content_*.txt")
	if err != nil {
		fmt.Printf("创建临时文件失败: %v\n", err)
		return nil
	}
	defer os.Remove(tmpFile.Name())

	// 将内容转换为UTF-8编码
	utf8Content, err := convertToUTF8([]byte(content))
	if err != nil {
		fmt.Printf("转换编码失败: %v\n", err)
		return nil
	}

	// 写入内容到临时文件
	if _, err := tmpFile.WriteString(utf8Content); err != nil {
		fmt.Printf("写入临时文件失败: %v\n", err)
		return nil
	}
	if err := tmpFile.Close(); err != nil {
		fmt.Printf("关闭临时文件失败: %v\n", err)
		return nil
	}

	// 获取当前可执行文件所在目录
	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("获取可执行文件路径失败: %v\n", err)
		return nil
	}
	exeDir := filepath.Dir(exePath)

	// 构建Python脚本的完整路径
	scriptPath := filepath.Join(exeDir, "..", "lib", "address_name.py")
	scriptPath = filepath.Clean(scriptPath)

	// 检查Python脚本是否存在
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		fmt.Printf("Python脚本不存在: %s\n", scriptPath)
		return nil
	}

	// 执行Python脚本
	cmd := exec.Command("python", scriptPath, tmpFile.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("执行Python脚本失败: %v\n输出: %s\n", err, string(output))
		return nil
	}

	// 解析JSON输出
	var result struct {
		Error     string   `json:"error,omitempty"`
		Addresses []string `json:"addresses"`
		Names     []string `json:"names"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		fmt.Printf("解析Python脚本输出失败: %v\n输出: %s\n", err, string(output))
		return nil
	}

	// 检查是否有错误
	if result.Error != "" {
		fmt.Printf("Python脚本报告错误: %s\n", result.Error)
		return nil
	}

	// 合并地址和姓名
	var matches []string
	matches = append(matches, result.Addresses...)
	matches = append(matches, result.Names...)
	return matches
}

// SensMatch 处理敏感信息匹配
type SensMatch struct {
	addressNameChecker *AddressName
}

// NewSensMatch 创建新的 SensMatch 实例
func NewSensMatch() *SensMatch {
	return &SensMatch{
		addressNameChecker: NewAddressName(),
	}
}

// CheckSecret 检查电话号码
func (s *SensMatch) CheckSecret(value string) []string {
	phonePattern := `1(?:3\d|4[5-9]|5[0-35-9]|6[5-6]|7[0-8]|8\d|9[189])\d{8}`
	re := regexp.MustCompile(phonePattern)
	matches := re.FindAllString(value, -1)
	if len(matches) > 0 {
		return matches
	}
	return nil
}

// CheckIP 检查 IPv4 地址
func (s *SensMatch) CheckIP(value string) []string {
	ipPattern := `(?:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)\.){3}(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)`
	re := regexp.MustCompile(ipPattern)
	matches := re.FindAllString(value, -1)
	if len(matches) > 0 {
		return matches
	}
	return nil
}

// CheckMAC 检查 MAC 地址
func (s *SensMatch) CheckMAC(value string) []string {
	macPattern := `(?:(?:(?:[a-f0-9A-F]{2}:){5})|(?:(?:[a-f0-9A-F]{2}-){5}))[a-f0-9A-F]{2}`
	re := regexp.MustCompile(macPattern)
	matches := re.FindAllString(value, -1)
	if len(matches) > 0 {
		return matches
	}
	return nil
}

// CheckIPv6 检查 IPv6 地址
func (s *SensMatch) CheckIPv6(text string) []string {
	ipv6Pattern := `([a-fA-F0-9:]{2,39})`
	re := regexp.MustCompile(ipv6Pattern)
	matches := re.FindAllString(text, -1)

	validIPv6 := []string{}
	for _, match := range matches {
		if ip := net.ParseIP(match); ip != nil && ip.To4() == nil {
			validIPv6 = append(validIPv6, match)
		}
	}

	if len(validIPv6) > 0 {
		return validIPv6
	}
	return nil
}

// IsValidBankCard 使用 Luhn 算法检查银行卡号是否有效
func (s *SensMatch) IsValidBankCard(cardNum string) bool {
	total := 0
	cardNumLength := len(cardNum)
	for i := 1; i <= cardNumLength; i++ {
		digit := int(cardNum[cardNumLength-i] - '0')
		if i%2 == 0 {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		total += digit
	}
	return total%10 == 0
}

// CheckBankCard 检查有效的银行卡号
func (s *SensMatch) CheckBankCard(text string) []string {
	re := regexp.MustCompile(`\d{16,19}`)
	matches := re.FindAllString(text, -1)

	validCards := []string{}
	for _, card := range matches {
		if s.IsValidBankCard(card) {
			validCards = append(validCards, card)
		}
	}

	if len(validCards) > 0 {
		return validCards
	}
	return nil
}

// CheckEmail 检查电子邮箱地址
func (s *SensMatch) CheckEmail(text string) []string {
	emailPattern := `([A-Za-z0-9_\-\.])+\@([A-Za-z0-9_\-\.])+\.([A-Za-z]{2,4})`
	re := regexp.MustCompile(emailPattern)
	matches := re.FindAllString(text, -1)
	if len(matches) > 0 {
		return matches
	}
	return nil
}

// CheckPassport 检查护照号码
func (s *SensMatch) CheckPassport(value string) []string {
	// 修复正则表达式语法
	pattern := `1[45][0-9]{7}|([PpSs]\d{7})|([SsGg]\d{8})|([GgTtSsLlQqDdAaFf]\d{8})`
	re := regexp.MustCompile(pattern)
	matches := re.FindAllString(value, -1)
	if len(matches) > 0 {
		return matches
	}
	return nil
}

// CheckIDNumber 检查中国身份证号
func (s *SensMatch) CheckIDNumber(value string) []string {
	idPattern := `(?:^|[^0-9])([1-9]\d{5}(?:18|19|[23]\d)\d{2}(?:0[1-9]|1[0-2])(?:[0-2][1-9]|10|20|30|31)\d{3}[0-9Xx]|[1-9]\d{5}\d{2}(?:0[1-9]|1[0-2])(?:[0-2][1-9]|10|20|30|31)\d{2})(?:$|[^0-9])`
	re := regexp.MustCompile(idPattern)
	matches := re.FindAllStringSubmatch(value, -1)

	validMatches := []string{}
	for _, match := range matches {
		if len(match) > 1 {
			validMatches = append(validMatches, match[1])
		}
	}

	if len(validMatches) > 0 {
		return validMatches
	}
	return nil
}

// CheckGender 检查性别信息
func (s *SensMatch) CheckGender(value string) []string {
	genderPattern := `(男|male|女|female)`
	re := regexp.MustCompile(genderPattern)
	matches := re.FindAllString(value, -1)
	if len(matches) > 0 {
		return matches
	}
	return nil
}

// CheckNational 检查民族信息
func (s *SensMatch) CheckNational(value string) []string {
	nationalPattern := `(汉族|满族|蒙古族|回族|藏族|维吾尔族|苗族|彝族|壮族|布依族|侗族|瑶族|白族|土家族|哈尼族|哈萨克族|傣族|黎族|傈僳族|佤族|畲族|高山族|拉祜族|水族|东乡族|纳西族|景颇族|柯尔克孜族|土族|达斡尔族|仫佬族|羌族|布朗族|撒拉族|毛南族|仡佬族|锡伯族|阿昌族|普米族|朝鲜族|塔吉克族|怒族|乌孜别克族|俄罗斯族|鄂温克族|德昂族|保安族|裕固族|京族|塔塔尔族|独龙族|鄂伦春族|赫哲族|门巴族|珞巴族|基诺族|汉|满|蒙古|回|藏|维吾尔|苗|彝|壮|布依|侗|瑶|白|土家|哈尼|哈萨克|傣|黎|傈僳|佤|畲|高山|拉祜|水|东乡|纳西|景颇|柯尔克孜|土|达斡尔|仫佬|羌|布朗|撒拉|毛南|仡佬|锡伯|阿昌|普米|朝鲜|塔吉克|怒|乌孜别克|俄罗斯|鄂温克|德昂|保安|裕固|京|塔塔尔|独龙|鄂伦春|赫哲|门巴|珞巴|基诺)`
	re := regexp.MustCompile(nationalPattern)
	matches := re.FindAllString(value, -1)
	if len(matches) > 0 {
		return matches
	}
	return nil
}

// CheckCarNum 检查中国车牌号
func (s *SensMatch) CheckCarNum(value string) []string {
	carnumPattern := `[京津沪渝冀豫云辽黑湘皖鲁新苏浙赣鄂桂甘晋蒙陕吉闽贵粤青藏川宁琼使领A-Z]{1}[A-Z]{1}[A-Z0-9]{4}[A-Z0-9挂学警港澳]{1}`
	re := regexp.MustCompile(carnumPattern)
	matches := re.FindAllString(value, -1)

	validMatches := []string{}
	for _, match := range matches {
		if len(match) >= 7 {
			validMatches = append(validMatches, match)
		}
	}

	if len(validMatches) > 0 {
		return validMatches
	}
	return nil
}

// CheckTelephone 检查电话号码
func (s *SensMatch) CheckTelephone(value string) []string {
	telephonePattern := `(0[0-9]{2,3}\-)?([2-9][0-9]{6,7})+(\-[0-9]{1,4})?`
	re := regexp.MustCompile(telephonePattern)
	matches := re.FindAllString(value, -1)

	validMatches := []string{}
	for _, match := range matches {
		if len(match) >= 7 && len(match) <= 12 {
			validMatches = append(validMatches, match)
		}
	}

	if len(validMatches) > 0 {
		return validMatches
	}
	return nil
}

// CheckOfficer 检查军官证号码
func (s *SensMatch) CheckOfficer(value string) []string {
	officerPattern := `[^\x00-\x7F]字第[0-9a-zA-Z]{4,8}号?`
	re := regexp.MustCompile(officerPattern)
	matches := re.FindAllString(value, -1)
	if len(matches) > 0 {
		return matches
	}
	return nil
}

// CheckHMPass 检查港澳通行证号码
func (s *SensMatch) CheckHMPass(value string) []string {
	hmPassPattern := `[HMhm][0-9]{8,10}`
	re := regexp.MustCompile(hmPassPattern)
	matches := re.FindAllString(value, -1)
	if len(matches) > 0 {
		return matches
	}
	return nil
}

// CheckJDBC 检查 JDBC 连接字符串
func (s *SensMatch) CheckJDBC(value string) []string {
	jdbcPattern := `jdbc:(?:mysql://(?:\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}|[\w.-]+)(?::\d+)?/[\w-]+(?:\?[\w=&%-]+)?|oracle:thin:@(?:\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}|[\w.-]+)(?::\d+)?:[\w]+|(?:microsoft:)?sqlserver://(?:\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}|[\w.-]+)(?::\d+)?(?:;[\w=%-]+)*)`
	re := regexp.MustCompile(jdbcPattern)
	matches := re.FindAllString(value, -1)

	validMatches := []string{}
	for _, match := range matches {
		if strings.HasSuffix(match, "?") || strings.HasSuffix(match, "/") || strings.HasSuffix(match, ";") || strings.HasSuffix(match, ":") || unicode.IsLetter(rune(match[len(match)-1])) || unicode.IsNumber(rune(match[len(match)-1])) {
			validMatches = append(validMatches, match)
		}
	}

	if len(validMatches) > 0 {
		return validMatches
	}
	return nil
}

// CheckOrganization 检查组织机构代码
func (s *SensMatch) CheckOrganization(value string) []string {
	organizationPattern := `^[\dA-Z]{8}[X\d]$`
	re := regexp.MustCompile(organizationPattern)

	matches := regexp.MustCompile(`(?:^|[^0-9A-Za-z-])([A-Za-z0-9-]{9})(?:$|[^0-9A-Za-z-])`).FindAllStringSubmatch(strings.ToUpper(value), -1)
	validMatches := []string{}

	for _, match := range matches {
		if len(match) > 1 {
			orgStr := regexp.MustCompile(`[^A-Z0-9]`).ReplaceAllString(match[1], "")
			if re.MatchString(orgStr) {
				verifyCode := []int{3, 7, 9, 10, 5, 8, 4, 2}
				sum := 0
				for i := 0; i < 8; i++ {
					if unicode.IsLetter(rune(orgStr[i])) {
						sum += (int(orgStr[i]) - 55) * verifyCode[i]
					} else {
						sum += int(orgStr[i]-'0') * verifyCode[i]
					}
				}
				verify := 11 - sum%11
				if verify == 10 {
					verify = 'X'
				} else if verify == 11 {
					verify = '0'
				} else {
					verify += '0'
				}
				if string(verify) == string(orgStr[8]) {
					validMatches = append(validMatches, orgStr)
				}
			}
		}
	}

	if len(validMatches) > 0 {
		return validMatches
	}
	return nil
}

// CheckBusiness 检查工商注册号
func (s *SensMatch) CheckBusiness(value string) []string {
	businessPattern := `\d{15}`
	re := regexp.MustCompile(businessPattern)
	matches := re.FindAllString(value, -1)

	validMatches := []string{}
	for _, match := range matches {
		verifyCode := 10
		for i := 0; i < 14; i++ {
			verifyCode = (((verifyCode%11 + int(match[i]-'0')) % 10) * 2) % 11
		}
		verifyCode = (11 - (verifyCode % 10)) % 10
		if string(verifyCode+'0') == string(match[14]) {
			validMatches = append(validMatches, match)
		}
	}

	if len(validMatches) > 0 {
		return validMatches
	}
	return nil
}

// CheckCredit 检查统一社会信用代码
func (s *SensMatch) CheckCredit(value string) []string {
	creditPattern := `^(1[129]|5[1239]|9[123]|Y1)\d{6}[\dA-Z]{8}[X\d][\dA-Z]$`
	re := regexp.MustCompile(creditPattern)

	matches := regexp.MustCompile(`(?:^|[^0-9A-Za-z])([1-9Y][0-9A-Za-z]{17})(?:$|[^0-9A-Za-z])`).FindAllStringSubmatch(strings.ToUpper(value), -1)
	validMatches := []string{}

	strToNum := map[rune]int{
		'A': 10, 'B': 11, 'C': 12, 'D': 13, 'E': 14, 'F': 15, 'G': 16, 'H': 17,
		'J': 18, 'K': 19, 'L': 20, 'M': 21, 'N': 22, 'P': 23, 'Q': 24, 'R': 25,
		'T': 26, 'U': 27, 'W': 28, 'X': 29, 'Y': 30,
	}

	numToStr := map[int]rune{
		10: 'A', 11: 'B', 12: 'C', 13: 'D', 14: 'E', 15: 'F', 16: 'G', 17: 'H',
		18: 'J', 19: 'K', 20: 'L', 21: 'M', 22: 'N', 23: 'P', 24: 'Q', 25: 'R',
		26: 'T', 27: 'U', 28: 'W', 29: 'X', 30: 'Y',
	}

	verifyWeights := []int{1, 3, 9, 27, 19, 26, 16, 17, 20, 29, 25, 13, 8, 24, 10, 30, 28}

	for _, match := range matches {
		if len(match) > 1 {
			creditStr := match[1]
			if len(creditStr) != 18 {
				continue
			}

			if re.MatchString(creditStr) {
				sum := 0
				for i := 0; i < 17; i++ {
					if unicode.IsLetter(rune(creditStr[i])) {
						sum += strToNum[rune(creditStr[i])] * verifyWeights[i]
					} else {
						sum += int(creditStr[i]-'0') * verifyWeights[i]
					}
				}
				verify := 31 - sum%31
				var verifyChar rune
				if verify > 9 {
					verifyChar = numToStr[verify]
				} else {
					verifyChar = rune(verify + '0')
				}
				if verifyChar == rune(creditStr[17]) {
					validMatches = append(validMatches, creditStr)
				}
			}
		}
	}

	if len(validMatches) > 0 {
		return validMatches
	}
	return nil
}

// RunAllChecks 运行所有敏感字段检查
func (s *SensMatch) RunAllChecks(value string) map[string][]string {
	results := map[string][]string{
		"phone":        s.CheckSecret(value),
		"ip":           s.CheckIP(value),
		"mac":          s.CheckMAC(value),
		"ipv6":         s.CheckIPv6(value),
		"bank_card":    s.CheckBankCard(value),
		"email":        s.CheckEmail(value),
		"passport":     s.CheckPassport(value),
		"id_number":    s.CheckIDNumber(value),
		"gender":       s.CheckGender(value),
		"national":     s.CheckNational(value),
		"carnum":       s.CheckCarNum(value),
		"telephone":    s.CheckTelephone(value),
		"officer":      s.CheckOfficer(value),
		"HM_pass":      s.CheckHMPass(value),
		"jdbc":         s.CheckJDBC(value),
		"organization": s.CheckOrganization(value),
		"business":     s.CheckBusiness(value),
		"credit":       s.CheckCredit(value),
		//"address_name": s.CheckChineseAddress(value),
	}

	// 过滤掉空结果
	filteredResults := make(map[string][]string)
	for k, v := range results {
		if v != nil {
			filteredResults[k] = v
		}
	}

	return filteredResults
}
