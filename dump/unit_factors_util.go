package dump

// 将带单位的字符串转换为字节数

import (
	"fmt"
	"strconv"
	"strings"
)

// 单位到字节的转换因子
var unitFactors = map[string]float64{
	"b": 1,       // Byte
	"kb": 1024,   // KiloByte
	"mb": 1024*1024, // MegaByte
	"gb": 1024*1024*1024, // GigaByte
	"tb": 1024*1024*1024*1024, // TeraByte
}

// 将带单位的字符串转换为字节数
func ParseUnitToBytes(sizeStr string) (int64, error) {
	// 去除字符串中的空格，并将其转换为小写，以便统一处理
	sizeStr = strings.ToLower(strings.TrimSpace(sizeStr))
	// 查找最后一个非数字字符，即单位部分
	lastNonDigit := -1
	for i, c := range sizeStr {
		if c < '0' || c > '9' {
			lastNonDigit = i
			break
		}
	}
	if lastNonDigit == -1 {
		return 0, fmt.Errorf("no valid unit found in the string")
	}
	// 提取数值和单位部分
	valueStr := sizeStr[:lastNonDigit]
	unit := sizeStr[lastNonDigit:]
	// 将数值部分转换为float64类型
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0, err
	}
	// 获取单位对应的因子
	factor, ok := unitFactors[unit]
	if !ok {
		return 0, fmt.Errorf("unknown unit: %s", unit)
	}
	// 计算结果并返回，注意转换为int64类型以适应字节数表示法
	return int64(value * factor), nil
}

// 使用方法，测试
func test_ParseUnitToBytes() {
	// 测试函数
	sizeStr := "2MB" // 可以是 "1kb", "3gb", 等任何带单位的字符串
	bytes, err := ParseUnitToBytes(sizeStr)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("%s is %d bytes\n", sizeStr, bytes)
	}
}