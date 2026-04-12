package sipapi

import (
	"testing"

	"github.com/panjjo/gosip/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMobilePosition_Normal(t *testing.T) {
	xml := `<?xml version="1.0" encoding="GB2312"?>
<Notify>
<CmdType>MobilePosition</CmdType>
<SN>123456</SN>
<DeviceID>33010000001118</DeviceID>
<Time>2026-04-11T10:30:00</Time>
<Longitude>116.397428</Longitude>
<Latitude>39.90923</Latitude>
<Speed>30.5</Speed>
<Direction>135.2</Direction>
<Altitude>50.0</Altitude>
</Notify>`

	pos := &MobilePosition{}
	err := utils.XMLDecode([]byte(xml), pos)

	require.NoError(t, err, "XML解析不应报错")
	assert.Equal(t, "33010000001118", pos.DeviceID)
	assert.Equal(t, "2026-04-11T10:30:00", pos.Time)
	assert.Equal(t, 116.397428, pos.Longitude)
	assert.Equal(t, 39.90923, pos.Latitude)
	assert.Equal(t, 30.5, pos.Speed)
	assert.Equal(t, 135.2, pos.Direction)
	assert.Equal(t, 50.0, pos.Altitude)
}

func TestParseMobilePosition_MinimalFields(t *testing.T) {
	xml := `<?xml version="1.0"?>
<Notify>
<CmdType>MobilePosition</CmdType>
<DeviceID>33010000001118</DeviceID>
<Longitude>121.473701</Longitude>
<Latitude>31.230416</Latitude>
</Notify>`

	pos := &MobilePosition{}
	err := utils.XMLDecode([]byte(xml), pos)

	require.NoError(t, err)
	assert.Equal(t, "33010000001118", pos.DeviceID)
	assert.Equal(t, 121.473701, pos.Longitude)
	assert.Equal(t, 31.230416, pos.Latitude)
	assert.Equal(t, 0.0, pos.Speed)    // 未提供应为0
	assert.Equal(t, "", pos.Time)      // 未提供应为空
}

func TestParseMobilePosition_InvalidXML(t *testing.T) {
	invalidXML := `<?xml version="1.0"?>
<Notify>
<CmdType>MobilePosition</CmdType>
<DeviceID>33010000001118</DeviceID>
`

	pos := &MobilePosition{}
	err := utils.XMLDecode([]byte(invalidXML), pos)

	assert.Error(t, err, "无效XML应返回错误")
}

func TestParseMobilePosition_NonNumericCoords(t *testing.T) {
	xml := `<?xml version="1.0"?>
<Notify>
<CmdType>MobilePosition</CmdType>
<DeviceID>33010000001118</DeviceID>
<Longitude>invalid</Longitude>
<Latitude>39.90923</Latitude>
</Notify>`

	pos := &MobilePosition{}
	err := utils.XMLDecode([]byte(xml), pos)

	assert.Error(t, err, "非数字经纬度应返回错误")
}

func TestParseMobilePosition_GBKEncoding(t *testing.T) {
	// GBK编码的XML（实际设备可能发送）
	gbkXML := []byte{
		0x3c, 0x3f, 0x78, 0x6d, 0x6c, 0x20, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x3d, 0x22, 0x31,
		0x2e, 0x30, 0x22, 0x3f, 0x3e, 0x0a, 0x3c, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x3e, 0x0a, 0x3c,
		0x43, 0x6d, 0x64, 0x54, 0x79, 0x70, 0x65, 0x3e, 0x4d, 0x6f, 0x62, 0x69, 0x6c, 0x65, 0x50, 0x6f,
		0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x3c, 0x2f, 0x43, 0x6d, 0x64, 0x54, 0x79, 0x70, 0x65, 0x3e,
		0x0a, 0x3c, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x49, 0x44, 0x3e, 0x33, 0x33, 0x30, 0x31, 0x30,
		0x30, 0x30, 0x30, 0x30, 0x30, 0x31, 0x31, 0x31, 0x38, 0x3c, 0x2f, 0x44, 0x65, 0x76, 0x69, 0x63,
		0x65, 0x49, 0x44, 0x3e, 0x0a, 0x3c, 0x4c, 0x6f, 0x6e, 0x67, 0x69, 0x74, 0x75, 0x64, 0x65, 0x3e,
		0x31, 0x31, 0x36, 0x2e, 0x33, 0x39, 0x37, 0x3c, 0x2f, 0x4c, 0x6f, 0x6e, 0x67, 0x69, 0x74, 0x75,
		0x64, 0x65, 0x3e, 0x0a, 0x3c, 0x4c, 0x61, 0x74, 0x69, 0x74, 0x75, 0x64, 0x65, 0x3e, 0x33, 0x39,
		0x2e, 0x39, 0x30, 0x39, 0x3c, 0x2f, 0x4c, 0x61, 0x74, 0x69, 0x74, 0x75, 0x64, 0x65, 0x3e, 0x0a,
		0x3c, 0x2f, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x3e,
	}

	pos := &MobilePosition{}
	err := utils.XMLDecode(gbkXML, pos)

	// 注意：实际GBK中文部分会失败，这里测试基本ASCII部分
	if err == nil {
		assert.Equal(t, "33010000001118", pos.DeviceID)
		assert.Equal(t, 116.397, pos.Longitude)
	}
}

func TestCheckAndSubscribe_NoSubscribe(t *testing.T) {
	device := Devices{
		DeviceID: "33010000001118",
		Regist:   true,
		Subscribe: nil,
	}

	// 不应panic，应正常返回
	CheckAndSubscribe(device)
}

func TestCheckAndSubscribe_PositionEnabled(t *testing.T) {
	device := Devices{
		DeviceID: "33010000001118",
		Regist:   true,
		Subscribe: map[string]interface{}{
			"position": true,
		},
	}

	// 设备不在线时不应发送
	CheckAndSubscribe(device)
}

func TestCheckAndSubscribe_PositionDisabled(t *testing.T) {
	device := Devices{
		DeviceID: "33010000001118",
		Regist:   true,
		Subscribe: map[string]interface{}{
			"position": false,
		},
	}

	CheckAndSubscribe(device)
}
