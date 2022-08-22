package main

import (
	"fmt"
	"reflect"
	"testing"
)

type IUser interface {
	Hello(name string)
}

type User struct {
	UserHello
}

type UserHello struct {
	Id   int
	Name string
	Age  int
}

func (u UserHello) Hello(name string) {
	fmt.Printf("Hello , %v .  My Name is %v\n", name, u.Name)
}

func TestIUser(t *testing.T) {
	u := User{} //UserHello{ID: 1, Name: "HaHaHa", Age: 88}
	reflectInterface(u)
}

func reflectInterface(u IUser) {
	v := reflect.ValueOf(u)
	mv := v.MethodByName("Hello")
	args := []reflect.Value{reflect.ValueOf("joe")}
	mv.Call(args)

	t := reflect.TypeOf(u)
	//取出匿名字段
	fmt.Printf("%#v\n", t.Field(0))

	//遍历类型的字段
	for i := 0; i < t.NumField(); i++ {
		//根据索引取得字段
		f := t.Field(i)
		//取出字段对应的值
		val := v.Field(i).Interface()
		fmt.Printf("%5s: %v = %v\n", f.Name, f.Type, val)
	}
}
