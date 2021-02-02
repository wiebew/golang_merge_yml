package main

import (
	"io/ioutil"
	"log"
	"reflect"

	"gopkg.in/yaml.v2"
)

func isMap(typeName string) bool {
	switch typeName {
	case "map[interface {}]interface {}":
		return true
	default:
		return false
	}
}

func mergerecursive(master *map[interface{}]interface{}, merge *map[interface{}]interface{}, level int) {

	for k, v := range *master {
		_, exists := (*merge)[k]
		if exists {
			// key exist in the target yaml
			// if it is a map we need to (recursively) descend into it to check every value in the map
			// this prevents losing values if the master only has one underlying value and the default multiple
			if isMap(reflect.TypeOf(v).String()) {
				masternode := v.(map[interface{}]interface{}) // type assertion (typecast)
				// check if both types are a map types
				if isMap(reflect.TypeOf((*merge)[k]).String()) {
					mergenode := (*merge)[k].(map[interface{}]interface{}) // type assertion
					mergerecursive(&masternode, &mergenode, level+1)
				} else {
					log.Fatal("Key [", k, "] is map/list of values in one yaml and a singular value in the other yaml, can't merge them")
				}
			} else {
				// key is not a map, so we just need to copy the value if they are both non-map types
				if !isMap(reflect.TypeOf((*merge)[k]).String()) {
					(*merge)[k] = v
				} else {
					log.Fatal("Key [", k, "] is map/list of values in one yaml and a singular value in the other yaml, can't merge them")
				}
			}
		} else {
			// key does not exists in the target, just add the whole node/value
			(*merge)[k] = v
		}
	}
}

func merge(masterpath *string, defaultspath *string, merge *map[interface{}]interface{}) {
	// will merge values of both yaml files into merge map
	// values in master will overrule values in defaults

	var master map[interface{}]interface{}
	var defaults map[interface{}]interface{}

	bs, err := ioutil.ReadFile(*masterpath)
	if err != nil {
		panic(err)
	}

	if err := yaml.Unmarshal(bs, &master); err != nil {
		panic(err)
	}

	bs, err = ioutil.ReadFile(*defaultspath)

	if err != nil {
		panic(err)
	}
	if err := yaml.Unmarshal(bs, &defaults); err != nil {
		panic(err)
	}
	*merge = defaults
	mergerecursive(&master, merge, 0)
}

func main() {
	var merged map[interface{}]interface{}
	var manifest string = "./application_manifest.yml"
	var defaults string = "./application_manifest_defaults.yml"

	merge(&manifest, &defaults, &merged)

	bs, err := yaml.Marshal(merged)
	if err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile("merged.yml", bs, 0644); err != nil {
		panic(err)
	}
}
