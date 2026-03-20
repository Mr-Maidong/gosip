package utils

import (
	"bytes"
	"encoding/base64"
	"image/color"
	"image/png"

	"github.com/issue9/identicon/v2"
	"github.com/sirupsen/logrus"
)

// GenerateAvatar 生成随机马赛克头像
// seed: 种子字符串，用于生成唯一的头像（通常使用用户名或用户 ID）
// size: 头像尺寸（宽度和高度），建议 200-400
// 返回 base64 编码的 PNG 图片字符串
func GenerateAvatar(seed string, size int) string {
	if seed == "" {
		seed = RandString(16)
	}

	// 使用 Style2（马赛克风格）创建实例
	icon := identicon.S2(size)

	// 生成头像图片
	img := icon.Make([]byte(seed))

	// 将图片编码为 PNG 格式
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		logrus.Errorln("Encode avatar image error:", err)
		return ""
	}

	// 将 PNG 数据转换为 base64 编码
	pngData := buf.Bytes()
	base64Str := base64.StdEncoding.EncodeToString(pngData)

	logrus.Debugln("Generated avatar for", seed, "size:", len(base64Str), "bytes")
	return base64Str
}

// GenerateAvatarWithColor 生成带背景色的随机马赛克头像
// seed: 种子字符串
// size: 头像尺寸
// bgColor: 背景颜色
// fgColor: 前景颜色
// 返回 base64 编码的 PNG 图片字符串
func GenerateAvatarWithColor(seed string, size int, bgColor, fgColor color.Color) string {
	if seed == "" {
		seed = RandString(16)
	}

	// 使用 Style2（马赛克风格）创建实例
	icon := identicon.New(identicon.Style2, size, bgColor, fgColor)

	// 生成头像图片
	img := icon.Make([]byte(seed))

	// 将图片编码为 PNG 格式
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return ""
	}

	// 将 PNG 数据转换为 base64 编码
	pngData := buf.Bytes()
	base64Str := base64.StdEncoding.EncodeToString(pngData)

	return base64Str
}

// GetAvatarDataURI 获取头像的 Data URI 格式
// base64Str: base64 编码的 PNG 图片
// 返回 data:image/png;base64,xxx 格式的字符串
func GetAvatarDataURI(base64Str string) string {
	if base64Str == "" {
		return ""
	}
	return "data:image/png;base64," + base64Str
}
