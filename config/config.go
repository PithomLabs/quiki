package config

import "errors"

// the singleton instance, set in main
var Conf *Config

// configuration, fetch conf values with conf.Get()
type Config struct {
	path string                 // file path
	vars map[string]interface{} // root variable map
	line *uint                  // current line for warnings and errors
}

// new config
func New(path string) *Config {
	return &Config{
		path: path,
		vars: make(map[string]interface{}),
	}
}

// get string value
func (conf *Config) Get(varName string) string {

	// get the map
	where, lastPart := conf.getWhere(varName, false)
	if where == nil {
		conf.warn("could not Get @" + varName)
		return ""
	}

	// get the string value
	iface := where[lastPart]
	switch str := iface.(type) {
	case string:
		return str
	case nil:
		return ""
	default:
		conf.warn("@" + varName + " is not a string")
	}

	return ""
}

// get map
func (conf *Config) GetMap(varName string) map[string]interface{} {

	// get the location
	where, lastPart := conf.getWhere(varName, false)
	if where == nil {
		conf.warn("could not GetMap @" + varName)
		return nil
	}

	// get the map value
	iface := where[lastPart]
	switch aMap := iface.(type) {
	case map[string]interface{}:
		return aMap
	case nil:
		return nil
	default:
		conf.warn("@" + varName + " is not a map")
	}

	return nil
}

// get map with only string values
func (conf *Config) GetStringMap(varName string) map[string]string {
	aMap := conf.GetMap(varName)
	stringMap := make(map[string]string, len(aMap))
	for key, iface := range aMap {
		switch str := iface.(type) {
		case string:
			stringMap[key] = str
		}
	}
	return stringMap
}

// set string value
func (conf *Config) Set(varName string, value string) {

	// get the map
	where, lastPart := conf.getWhere(varName, true)
	if where == nil {
		conf.warn("could not Set @" + varName)
		return
	}

	// set the string value
	where[lastPart] = value
}

// get bool value
func (conf *Config) GetBool(varName string) bool {
	return isTrueString(conf.Get(varName))
}

// same as Get() except it returns an error if the value is missing
func (conf *Config) Require(varName string) (string, error) {
	val := conf.Get(varName)
	if !isTrueString(val) {
		return "", errors.New(conf.getWarn("@" + varName + " is required"))
	}
	return val, nil
}

// return an error if any of the passed variables are missing
func (conf *Config) RequireMany(varNames ...string) error {
	for _, varName := range varNames {
		if _, err := conf.Require(varName); err != nil {
			return err
		}
	}
	return nil
}

// string variable value is true or no?
func isTrueString(str string) bool {
	if str == "" || str == "0" {
		return false
	}
	return true
}
