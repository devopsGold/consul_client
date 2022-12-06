package consul_client

import (
	"encoding/json"
	"fmt"
	consul "github.com/hashicorp/consul/api"
	"reflect"
	"strings"
)

func ConsulClient(keyPath string, data interface{}) ([]byte, error) {

	// data can be either nil or a pointer to the structure to be filled.
	if data != nil && reflect.ValueOf(data).Kind() != reflect.Ptr {
		return nil, fmt.Errorf("invalid pointer. data must be 'nil' or pointer")
	}

	config := consul.DefaultConfig()
	config.Address = "consul.ms:8500"
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
		return nil, fmt.Errorf("value is empty")
	}

	if strings.HasSuffix(keyPath, ".json") {
		if data == nil {
			return nil, fmt.Errorf("no structure passed for json data")
		}

		// try unpack []byte in map[string]interface
		valueData, err := unpackToMap(value.Value)
		if err != nil {
			return nil, err
		}

		// may be data in another format, not map[string]interface
		if len(valueData) == 0 {
			// unpack []byte in []interface
			valueData, err = unpackToSlice(value.Value)
		}

		if err != nil {
			return nil, err
		}

		// data can be either as an object or as a list, otherwise the data is not valid
		if len(valueData) == 0 {
			return nil, fmt.Errorf("failed to unpack data, value is empty")
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

func unpackToMap(value []byte) ([]byte ,error) {

	var object map[string]interface{}
	err := json.Unmarshal(value, &object)
	if err != nil {
		return nil, err
	}

	if len(object) == 0 {
		return nil, nil
	}

	err = mapPrepare(&object)
	if err != nil {
		return nil, err
	}

	valueData, err := json.Marshal(object)
	if err != nil {
		return nil, err
	}

	return valueData, nil
}

func mapPrepare(dataLink *map[string]interface{}) error {
	object := *dataLink

	for key, v := range object {
		// if the value type is a string, and the string ends in .link - use the value as the new path
		if subKeyPath, ok := v.(string); ok && strings.HasSuffix(key, ".link") {
			msgData, err := ConsulClient(subKeyPath, nil)
			if err != nil {
				return fmt.Errorf("no value found for this path: %s", subKeyPath)
			}

			delete(object, key)
			// we cut off the ending .link
			key = key[:len(key)-5]
			object[key] = string(msgData)
		} else if subMap, ok := v.(map[string]interface{}); ok {
			err := mapPrepare(&subMap)
			if err != nil {
				return err
			}
			object[key] = subMap
		} else if subMap, ok := v.([]interface{}); ok {
			err := slicePrepare(&subMap)
			if err != nil {
				return err
			}
			object[key] = subMap
		}
	}

	return nil
}

func slicePrepare(dataLink *[]interface{}) error {
	object := *dataLink

	for idx, value := range object {
		// if the value type is a string, and the string ends in .link - use the value as the new path
		if subMap, ok := value.(map[string]interface{}); ok {
			err := mapPrepare(&subMap)
			if err != nil {
				return err
			}
			object[idx] = subMap
		} else if subMap, ok := value.([]interface{}); ok {
			err := slicePrepare(&subMap)
			if err != nil {
				return err
			}
			object[idx] = subMap
		}
	}

	return nil
}

func unpackToSlice(value []byte) ([]byte ,error) {

	var object []interface{}
	err := json.Unmarshal(value, &object)
	if err != nil {
		return nil, err
	}

	if len(object) == 0 {
		return nil, nil
	}

	valueData, err := json.Marshal(object)
	if err != nil {
		return nil, err
	}

	return valueData, nil
}
