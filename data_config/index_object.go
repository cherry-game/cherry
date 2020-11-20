package cherryDataConfig

import "fmt"

type IndexObject struct {
	TableName string
	IndexName string
	Columns   []string
}

func (i *IndexObject) String() string {
	return fmt.Sprintf("indexName=%s, columns=%s", i.IndexName, i.Columns)
}

func NewIndex(tableName, key string) *IndexObject {
	return Build(tableName, key)
}

func Build(tableName string, indexKey ...string) *IndexObject {
	obj := &IndexObject{}
	obj.TableName = tableName

	str := ""
	for i, s := range indexKey {
		if i != 0 {
			str += "_"
		}
		str += s
	}

	obj.IndexName = str
	obj.Columns = indexKey

	return obj
}
