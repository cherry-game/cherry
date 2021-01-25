package linq

import (
	"fmt"
	"github.com/ahmetb/go-linq/v3"
	"github.com/mitchellh/mapstructure"
	_ "github.com/mitchellh/mapstructure"
	"log"
	"testing"
)

func TestIntSlice(t *testing.T) {
	var ids []int
	for i := 1; i < 150000; i++ {
		ids = append(ids, i)
	}
	defaultFor(ids)
	linqFor(ids)
}

const (
	size = 1000000
)

func TestQueryCompany(t *testing.T) {

	list1 := GetCompanyByCountry("USA")
	t.Log(fmt.Printf("%x", &list1))
	t.Log(list1)

	list2 := GetCompanyByCountry("USA")
	t.Log(fmt.Printf("%x", &list2))
	t.Log(list2)

	name1 := GetCompanyByName("Microsoft")
	t.Log(fmt.Printf("%x", &name1))
	t.Log(name1)

	name2 := GetCompanyByName("Microsoft")
	t.Log(fmt.Printf("%x", &name2))
	t.Log(name2)
}

func BenchmarkSelectWhereFirst(b *testing.B) {
	for n := 0; n < b.N; n++ {
		linq.Range(1, size).Select(func(i interface{}) interface{} {
			return -i.(int)
		}).Where(func(i interface{}) bool {
			return i.(int) > -100
		}).First()
	}
}

func TestIndexOf(t *testing.T) {
	tests := []struct {
		input     interface{}
		predicate func(interface{}) bool
		expected  int
	}{
		{
			input: [9]int{1, 2, 3, 4, 5, 6, 7, 8, 9},
			predicate: func(i interface{}) bool {
				return i.(int) == 3
			},
			expected: 2,
		},
		{
			input: "sstr",
			predicate: func(i interface{}) bool {
				return i.(rune) == 'r'
			},
			expected: 3,
		},
		{
			input: "gadsgsadgsda",
			predicate: func(i interface{}) bool {
				return i.(rune) == 'z'
			},
			expected: -1,
		},
	}

	for _, test := range tests {
		index := linq.From(test.input).IndexOf(test.predicate)
		if index != test.expected {
			t.Errorf("From(%v).IndexOf() expected %v received %v", test.input, test.expected, index)
		}

		index = linq.From(test.input).IndexOfT(test.predicate)
		if index != test.expected {
			t.Errorf("From(%v).IndexOfT() expected %v received %v", test.input, test.expected, index)
		}
	}
}

type MyQuery linq.Query

func (q MyQuery) GreaterThan(threshold int) linq.Query {
	return linq.Query{
		Iterate: func() linq.Iterator {
			next := q.Iterate()
			return func() (item interface{}, ok bool) {
				for item, ok = next(); ok; item, ok = next() {
					if item.(int) > threshold {
						return
					}
				}
				return
			}
		},
	}
}

func TestMapStructure(t *testing.T) {
	test1()
}

func BenchmarkTest1(b *testing.B) {
	for n := 0; n < b.N; n++ {
		test1()
	}
}

func test1() {
	type Person struct {
		Name   string
		Age    int
		Emails []string
		Extra  map[string]string
	}
	input := map[string]interface{}{
		"name":   "Mitchell",
		"age":    91,
		"emails": []string{"one", "two", "three"},
		"extra": map[string]string{
			"twitter": "mitchellh",
		},
	}
	// map 转struct
	var result Person
	err := mapstructure.Decode(input, &result)
	if err != nil {
		panic(err)
	}

	var users = make([]Person, 2)

	var names []string

	var names2 []struct {
		name string
		age  int
	}

	users = append(users, result)

	// linq 使用
	linq.From(users).Where(func(u interface{}) bool {
		p := u.(Person)
		return p.Age > 10
	}).SelectT(func(p Person) string {
		return p.Name
	}).ToSlice(&names)

	for _, name := range names {
		log.Println(name)
	}

	linq.From(users).Where(func(u interface{}) bool {
		p := u.(Person)
		return p.Age > 10
	}).Select(func(p interface{}) interface{} {
		person := p.(Person)
		return struct {
			name string
			age  int
		}{
			name: person.Name,
			age:  person.Age,
		}
	}).ToSlice(&names2)

	for _, name := range names2 {
		log.Println(name)
	}

	// 测试自定义query
	results := MyQuery(linq.Range(1, 100)).GreaterThan(97).Results()
	for _, result := range results {
		log.Println(result)
	}
}
