package testLinq

import "github.com/ahmetb/go-linq/v3"

type Company struct {
	Name    string
	Country string
	City    string
}

var companies = []Company{
	{Name: "Microsoft", Country: "USA", City: "Redmond"},
	{Name: "Google", Country: "USA", City: "Palo Alto"},
	{Name: "Facebook", Country: "USA", City: "Palo Alto"},
	{Name: "Uber", Country: "USA", City: "San Francisco"},
	{Name: "Tweeter", Country: "USA", City: "San Francisco"},
	{Name: "SoundCloud", Country: "Germany", City: "Berlin"},
}

func GetCompanyByCountry(country string) []Company {
	var c []Company

	linq.From(companies).Where(func(i interface{}) bool {
		return i.(Company).Country == country
	}).ToSlice(&c)

	return c
}

func GetCompanyByName(name string) Company {
	return linq.From(companies).Where(func(i interface{}) bool {
		return i.(Company).Name == name
	}).First().(Company)

}
