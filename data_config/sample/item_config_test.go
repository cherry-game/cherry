package cherryDataConfigSample

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/phantacix/cherry/data_config"
	"reflect"
	"strconv"
	"testing"
)

func init() {
	var items []*ItemConfig
	for i := 10; i >= 1; i-- {
		items = append(items, &ItemConfig{
			Id:   i,
			Name: "a" + strconv.Itoa(i),
		})
	}
	cherryDataConfig.Add("itemConfig", items)
}

func TestSlice(t *testing.T) {
	a := []int{1, 2, 3}
	fmt.Println(a)
	modifySlice(a)
	fmt.Println(a)

}

func modifySlice(data []int) {
	data = nil
}

type Test1 struct {
	A int    `json:"a"`
	B string `json:"b"`
}

func (t *Test1) String() string {
	return fmt.Sprintf("a=%d,b=%s", t.A, t.B)
}

func TestReflectTypeOf(t *testing.T) {
	t1 := &Test1{
		A: 1,
		B: "aaa",
	}

	bytes, err := jsoniter.Marshal(&t1)
	if err != nil {
		fmt.Println(err)
	}

	tn := reflect.New(reflect.TypeOf(t1))

	e := jsoniter.Unmarshal(bytes, tn.Interface())
	fmt.Println(e)
	fmt.Println(tn)
}

func TestItemConfig_EqualOne(t *testing.T) {

	item1 := cherryDataConfig.EqualOne("itemConfig", "Id", 1).(*ItemConfig)
	fmt.Println(item1)
	//fmt.Printf("%v -> %v\n", item1, item1)

	item2 := cherryDataConfig.EqualOne("itemConfig", "Name", "a1").(*ItemConfig)
	fmt.Println(item2)
	//fmt.Printf("%v -> %v\n", item2, item2)

	cfg1 := cherryDataConfig.List("itemConfig").([]*ItemConfig)
	fmt.Println(cfg1)
}

func TestGoCache(t *testing.T) {
	var items []*ItemConfig
	for i := 10; i >= 1; i-- {
		items = append(items, &ItemConfig{
			Id:   i,
			Name: "a" + strconv.Itoa(i),
		})
	}

	maps := make(map[string]interface{})

	maps["itemConfig"] = items

	x1 := maps["itemConfig"]
	x2 := x1.([]*ItemConfig)

	fmt.Println(x1, x2)
}

func BenchmarkItemConfig_EqualOne(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cherryDataConfig.EqualOne("itemConfig", "Id", 1)
	}
}
