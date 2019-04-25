Consul Client
=============

This package allows you to unpack data in a json format into the structure.

There is support for specifying symbolic links.


Usage
=====

```go
package main

import (
	"fmt"
	"github.com/wedoca/consul_client"
)

type MyConfig struct {
	HostName  string   `json:"host_name"`
	FirstName string   `json:"first_name"`
	Data      []string `json:"data"`
}

func main() {
	
	//	raw data in consul KV:
	//	{
	//		"host_name": "localhost",
	//		"first_name.link": "some_path/folder/folder/key",
	//		"data": ["a", "b", "c"]
	//	}
	
	m := &MyConfig{}
	_, err := consul_client.ConsulClient("/first_path/key", m)
	if err != nil {
		fmt.Println(err)
		return
	}
	
	// or
	myStringValue, err := consul_client.ConsulClient("/second_path/key_with_simple_string_value", nil)
	if err != nil {
		fmt.Println(err)
	}
	
	fmt.Println("my string value:", myStringValue)

}
```