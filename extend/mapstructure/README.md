# mapstructure

mapstructure is a Go library for decoding generic map values to structures
and vice versa, while providing helpful error handling.

This library is most useful when decoding values from some data stream (JSON,
Gob, etc.) where you don't _quite_ know the structure of the underlying data
until you read a part of it. You can therefore read a `map[string]interface{}`
and use this library to decode it into the proper underlying native Go
structure.

## Installation

Standard `go get`:

```
$ go get github.com/goinggo/mapstructure
```

## Usage & Example

For usage and examples see the [Godoc](http://godoc.org/github.com/mitchellh/mapstructure).

The `Decode`, `DecodePath` and `DecodeSlicePath` functions have examples associated with it there.

## But Why?!

Go offers fantastic standard libraries for decoding formats such as JSON.
The standard method is to have a struct pre-created, and populate that struct
from the bytes of the encoded format. This is great, but the problem is if
you have configuration or an encoding that changes slightly depending on
specific fields. For example, consider this JSON:

```json
{
  "type": "person",
  "name": "Mitchell"
}
```

Perhaps we can't populate a specific structure without first reading
the "type" field from the JSON. We could always do two passes over the
decoding of the JSON (reading the "type" first, and the rest later).
However, it is much simpler to just decode this into a `map[string]interface{}`
structure, read the "type" key, then use something like this library
to decode it into the proper structure.

## DecodePath

Sometimes you have a large and complex JSON document where you only need to decode
a small part.

```
{
	"userContext": {
		"conversationCredentials": {
	            "sessionToken": "06142010_1:75bf6a413327dd71ebe8f3f30c5a4210a9b11e93c028d6e11abfca7ff"
	    },
	    "valid": true,
	    "isPasswordExpired": false,
	    "cobrandId": 10000004,
	    "channelId": -1,
	    "locale": "en_US",
	    "tncVersion": 2,
	    "applicationId": "17CBE222A42161A3FF450E47CF4C1A00",
	    "cobrandConversationCredentials": {
	        "sessionToken": "06142010_1:b8d011fefbab8bf1753391b074ffedf9578612d676ed2b7f073b5785b"
	    },
	     "preferenceInfo": {
	         "currencyCode": "USD",
	         "timeZone": "PST",
	         "dateFormat": "MM/dd/yyyy",
	         "currencyNotationType": {
	             "currencyNotationType": "SYMBOL"
	         },
	         "numberFormat": {
	             "decimalSeparator": ".",
	             "groupingSeparator": ",",
	             "groupPattern": "###,##0.##"
	         }
	     }
	 },
	 "lastLoginTime": 1375686841,
	 "loginCount": 299,
	 "passwordRecovered": false,
	 "emailAddress": "johndoe@email.com",
	 "loginName": "sptest1",
	 "userId": 10483860,
	 "userType":
	     {
	     "userTypeId": 1,
	     "userTypeName": "normal_user"
	     }
}
```

It is nice to be able to define and pull the documents and fields you need without
having to map the entire JSON structure.

```
type UserType struct {
	UserTypeId   int
	UserTypeName string
}

type NumberFormat struct {
		DecimalSeparator  string `jpath:"userContext.preferenceInfo.numberFormat.decimalSeparator"`
		GroupingSeparator string `jpath:"userContext.preferenceInfo.numberFormat.groupingSeparator"`
		GroupPattern      string `jpath:"userContext.preferenceInfo.numberFormat.groupPattern"`
	}
	
type User struct {
		Session   string   `jpath:"userContext.cobrandConversationCredentials.sessionToken"`
		CobrandId int      `jpath:"userContext.cobrandId"`
		UserType  UserType `jpath:"userType"`
		LoginName string   `jpath:"loginName"`
		NumberFormat       // This can also be a pointer to the struct (*NumberFormat)
}

docScript := []byte(document)
var docMap map[string]interface{}
json.Unmarshal(docScript, &docMap)

var user User
mapstructure.DecodePath(docMap, &user)
```

## DecodeSlicePath

Sometimes you have a slice of documents that you need to decode into a slice of structures

```
[
	{"name":"bill"},
	{"name":"lisa"}
]
```

Just Unmarshal your document into a slice of maps and decode the slice

```
type NameDoc struct {
	Name string `jpath:"name"`
}

sliceScript := []byte(document)
var sliceMap []map[string]interface{}
json.Unmarshal(sliceScript, &sliceMap)

var myslice []NameDoc
err := DecodeSlicePath(sliceMap, &myslice)

var myslice []*NameDoc
err := DecodeSlicePath(sliceMap, &myslice)
```

## Decode Structs With Embedded Slices

Sometimes you have a document with arrays

```
{
	"cobrandId": 10010352,
	"channelId": -1,
	"locale": "en_US",
	"tncVersion": 2,
	"people": [
		{
			"name": "jack",
			"age": {
			"birth":10,
			"year":2000,
			"animals": [
				{
				"barks":"yes",
				"tail":"yes"
				},
				{
				"barks":"no",
				"tail":"yes"
				}
			]
		}
		},
		{
			"name": "jill",
			"age": {
				"birth":11,
				"year":2001
			}
		}
	]
}
```

You can decode within those arrays

```
type Animal struct {
	Barks string `jpath:"barks"`
}

type People struct {
	Age     int      `jpath:"age.birth"` // jpath is relative to the array
	Animals []Animal `jpath:"age.animals"`
}

type Items struct {
	Categories []string `jpath:"categories"`
	Peoples    []People `jpath:"people"` // Specify the location of the array
}

docScript := []byte(document)
var docMap map[string]interface{}
json.Unmarshal(docScript, &docMap)

var items Items
DecodePath(docMap, &items)
```