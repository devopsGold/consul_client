package consul_client

import (
	"encoding/json"
	"errors"
	consul "github.com/hashicorp/consul/api"
	"reflect"
	"strings"
)

func consulClient(keyPath string, data interface{}) ([]byte, error) {

	// data can be either nil or a pointer to the structure to be filled.
	if data != nil && reflect.ValueOf(data).Kind() != reflect.Ptr {
		return nil, errors.New("invalid pointer. data must be 'nil' or pointer")
	}

	config := consul.DefaultConfig()
	config.Address = "127.0.0.1:8500"
	client, err := consul.NewClient(config)

	if err != nil {
		return nil, err
	}
	kv := client.KV()
	value, _, err := kv.Get(keyPath, nil)

	if err != nil {
		return nil, err
	}

	if value == nil {
		return nil, errors.New("value is empty")
	}

	if strings.HasSuffix(keyPath, ".json") {
		if data == nil {
			return nil, errors.New("no structure passed for json data")
		}

		object := map[string]interface{}{}
		err = json.Unmarshal(value.Value, &object)
		if err != nil {
			return nil, err
		}

		for key, v := range object {
			// if the value type is a string, and the string ends in .link - use the value as the new path
			if subKeyPath, ok := v.(string); ok && strings.HasSuffix(key, ".link") {
				msgData, err := consulClient(subKeyPath, nil)
				if err != nil {
					return nil, err
				}

				// we cut off the ending .link
				key = key[:len(key)-5]
				object[key] = string(msgData)
			}
		}

		valueData, err := json.Marshal(object)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(valueData, data)
		if err != nil {
			return nil, err
		}

	} else if strings.HasSuffix(keyPath, ".link") {
		value, _, err = kv.Get(string(value.Value), nil)

		if err != nil {
			return nil, err
		}
	}

	return value.Value, nil
}
