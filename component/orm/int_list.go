package cherryORM

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type IntList []int

// Value return json value, implement driver.Valuer interface
func (m IntList) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	ba, err := m.MarshalJSON()
	return string(ba), err
}

// Scan scan value into Jsonb, implements sql.Scanner interface
func (m *IntList) Scan(val interface{}) error {
	var ba []byte
	switch v := val.(type) {
	case []byte:
		ba = v
	case string:
		ba = []byte(v)
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", val))
	}

	var t []int
	err := json.Unmarshal(ba, &t)
	*m = t
	return err
}

// MarshalJSON to output non base64 encoded []byte
func (m IntList) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	t := ([]int)(m)
	return json.Marshal(t)
}

// UnmarshalJSON to deserialize []byte
func (m *IntList) UnmarshalJSON(b []byte) error {
	var t []int
	err := json.Unmarshal(b, &t)
	*m = t
	return err
}

// GormDataType gorm common data type
func (m IntList) GormDataType() string {
	return "IntList"
}

// GormDBDataType gorm db data type
func (IntList) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite":
		return "JSON"
	case "mysql":
		return "JSON"
	case "postgres":
		return "JSONB"
	}
	return ""
}
