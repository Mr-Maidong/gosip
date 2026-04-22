package db

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/panjjo/gorm"
	"github.com/panjjo/gosip/utils"
)

type DBModel struct {
	ID        uint   `json:"id" gorm:"primary_key"`
	CreatedAt int64  `json:"addtime" gorm:"column:addtime"`
	UpdatedAt int64  `json:"uptime" gorm:"column:uptime"`
	DeletedAt *int64 `json:"-" sql:"index" gorm:"column:deltime"`
}

type M map[string]interface{}

func (j M) Value() (driver.Value, error) {
	return utils.JSONEncode(&j), nil
}

func (j *M) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	return utils.JSONDecode(bytes, j)
}

// ==================== 设备位置历史 ====================

type DevicePosition struct {
	ID        uint    `json:"id" gorm:"primary_key"`
	DeviceID string  `json:"deviceid" gorm:"column:deviceid;type:varchar(64);not null;index"`
	Longitude float64 `json:"longitude" gorm:"column:longitude;type:decimal(10,6)"`
	Latitude float64 `json:"latitude" gorm:"column:latitude;type:decimal(10,6)"`
	GPSTime  int64   `json:"gpstime" gorm:"column:gpstime;type:bigint;not null"`
	Speed   float64 `json:"speed" gorm:"column:speed;type:float"`
	Direction float64`json:"direction" gorm:"column:direction;type:float"`
	Altitude float64`json:"altitude" gorm:"column:altitude;type:float"`
	CreatedAt int64  `json:"addtime" gorm:"column:addtime"`
}

func (DevicePosition) TableName() string { return "device_positions" }

// ==================== 设备事件 ====================

type DeviceEvent struct {
	ID        uint   `json:"id" gorm:"primary_key"`
	DeviceID string `json:"deviceid" gorm:"column:deviceid;type:varchar(64);not null;index"`
	EventType string `json:"eventtype" gorm:"column:eventtype;type:enum('ONLINE','OFFLINE');not null"`
	EventTime int64  `json:"eventtime" gorm:"column:eventtime;type:bigint;not null"`
	Source   string `json:"source" gorm:"column:source;type:varchar(64)"`
	Remark   string `json:"remark" gorm:"column:remark;type:varchar(255)"`
	CreatedAt int64  `json:"addtime" gorm:"column:addtime"`
}

func (DeviceEvent) TableName() string { return "device_events" }

// ==================== 设备告警 ====================

type DeviceAlarm struct {
	ID        uint   `json:"id" gorm:"primary_key"`
	DeviceID string `json:"deviceid" gorm:"column:deviceid;type:varchar(64);not null;index"`
	AlarmType string `json:"alarmtype" gorm:"column:alarmtype;type:varchar(64);not null"`
	AlarmLevel string `json:"alarmlevel" gorm:"column:alarmlevel;type:enum('NORMAL','WARNING','CRITICAL');default:'WARNING'"`
	AlarmMsg string `json:"alarmmsg" gorm:"column:alarmmsg;type:varchar(512)"`
	AlarmTime int64  `json:"alarmtime" gorm:"column:alarmtime;type:bigint;not null"`
	Handled int    `json:"handled" gorm:"column:handled;type:tinyint;default:0"`
	CreatedAt int64  `json:"addtime" gorm:"column:addtime"`
}

func (DeviceAlarm) TableName() string { return "device_alarms" }

type StringArray []string

func (j StringArray) Value() (driver.Value, error) {
	return strings.Join(j, ","), nil
}

func (j *StringArray) Scan(value interface{}) error {
	switch t := value.(type) {
	case []byte:
		nv := StringArray(strings.Split(string(t), ","))
		*j = nv
	case string:
		nv := StringArray(strings.Split(t, ","))
		*j = nv
	}
	return nil
}

type StringArrayJSON []string

func (j StringArrayJSON) Value() (driver.Value, error) {
	return utils.JSONEncode(&j), nil
}

func (j *StringArrayJSON) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	return utils.JSONDecode(bytes, j)
}

type Int64Array []int64

func (j Int64Array) Value() (driver.Value, error) {
	return strings.Join(its(j), ","), nil
}

func (j *Int64Array) Scan(value interface{}) error {
	switch t := value.(type) {
	case []byte:
		nv := strings.Split(string(t), ",")
		is := Int64Array(sti(nv))
		*j = is
	case string:
		nv := strings.Split(t, ",")
		is := Int64Array(sti(nv))
		*j = is
	}
	return nil
}

type Int64ArrayJSON []int64

func (j Int64ArrayJSON) Value() (driver.Value, error) {
	return utils.JSONEncode(&j), nil
}

func (j *Int64ArrayJSON) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	return utils.JSONDecode(bytes, j)
}
func its(is []int64) []string {
	ss := make([]string, len(is))
	for i, v := range is {
		ss[i] = strconv.FormatInt(v, 10)
	}
	return ss
}
func sti(ss []string) []int64 {
	is := make([]int64, len(ss))
	for i, v := range ss {
		is[i], _ = strconv.ParseInt(v, 10, 64)
	}
	return is
}

func RecordNotFound(e error) bool {
	return errors.Is(e, gorm.ErrRecordNotFound)
}
