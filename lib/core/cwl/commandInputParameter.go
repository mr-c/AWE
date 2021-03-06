package cwl

import (
	"fmt"
	"reflect"

	"github.com/davecgh/go-spew/spew"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/mgo.v2/bson"
)

// http://www.commonwl.org/v1.0/CommandLineTool.html#CommandInputParameter

type CommandInputParameter struct {
	Id             string             `yaml:"id,omitempty" bson:"id,omitempty" json:"id,omitempty" mapstructure:"id,omitempty"`
	SecondaryFiles []string           `yaml:"secondaryFiles,omitempty" bson:"secondaryFiles,omitempty" json:"secondaryFiles,omitempty" mapstructure:"secondaryFiles,omitempty"` // TODO string | Expression | array<string | Expression>
	Format         []string           `yaml:"format,omitempty" bson:"format,omitempty" json:"format,omitempty" mapstructure:"format,omitempty"`
	Streamable     bool               `yaml:"streamable,omitempty" bson:"streamable,omitempty" json:"streamable,omitempty" mapstructure:"streamable,omitempty"`
	Type           []CWLType_Type     `yaml:"type,omitempty" bson:"type,omitempty" json:"type,omitempty" mapstructure:"type,omitempty"` // []CommandInputParameterType  CWLType | CommandInputRecordSchema | CommandInputEnumSchema | CommandInputArraySchema | string | array<CWLType | CommandInputRecordSchema | CommandInputEnumSchema | CommandInputArraySchema | string>
	Label          string             `yaml:"label,omitempty" bson:"label,omitempty" json:"label,omitempty" mapstructure:"label,omitempty"`
	Description    string             `yaml:"description,omitempty" bson:"description,omitempty" json:"description,omitempty" mapstructure:"description,omitempty"`
	InputBinding   CommandLineBinding `yaml:"inputBinding" bson:"inputBinding" json:"inputBinding" mapstructure:"inputBinding"`
	Default        CWLType            `yaml:"default,omitempty" bson:"default,omitempty" json:"default,omitempty" mapstructure:"default,omitempty"`
}

func MakeStringMap(v interface{}) (result interface{}, err error) {

	switch v.(type) {
	case bson.M:

		original_map := v.(bson.M)

		new_map := make(map[string]interface{})

		for key_str, value := range original_map {

			new_map[key_str] = value
		}

		result = new_map
		return
	case map[interface{}]interface{}:

		v_map, ok := v.(map[interface{}]interface{})
		if !ok {
			err = fmt.Errorf("casting problem (b)")
			return
		}
		v_string_map := make(map[string]interface{})

		for key, value := range v_map {
			key_string := key.(string)
			v_string_map[key_string] = value
		}

		result = v_string_map
		return

	}
	result = v
	return
}

func NewCommandInputParameter(v interface{}, schemata []CWLType_Type) (input_parameter *CommandInputParameter, err error) {

	//fmt.Println("NewCommandInputParameter:\n")
	//spew.Dump(v)

	v, err = MakeStringMap(v)
	if err != nil {
		return
	}

	switch v.(type) {

	case map[string]interface{}:
		v_map, ok := v.(map[string]interface{})

		if !ok {
			err = fmt.Errorf("casting problem (a)")
			return
		}

		default_value, ok := v_map["default"]
		if ok {
			//fmt.Println("FOUND default key")
			default_input, xerr := NewCWLType("", default_value) // TODO return Int or similar
			if xerr != nil {
				err = xerr
				return
			}
			//spew.Dump(default_input)
			v_map["default"] = default_input
		}

		type_value, ok := v_map["type"]
		if ok {

			type_value, err = NewCommandInputParameterTypeArray(type_value, schemata)
			if err != nil {
				err = fmt.Errorf("(NewCommandInputParameter) NewCommandInputParameterTypeArray returns: %s", err.Error())
				return
			}

			v_map["type"] = type_value
		}

		format_value, ok := v_map["format"]
		if ok {
			format_str, is_string := format_value.(string)
			if is_string {
				v_map["format"] = []string{format_str}
			}

		}

		input_parameter = &CommandInputParameter{}
		err = mapstructure.Decode(v, input_parameter)
		if err != nil {
			err = fmt.Errorf("(NewCommandInputParameter) decode error: %s", err.Error())
			return
		}

	case string:
		v_string := v.(string)
		err = fmt.Errorf("(NewCommandInputParameter) got string %s", v_string)
		return
	default:

		err = fmt.Errorf("(NewCommandInputParameter) type %s unknown", reflect.TypeOf(v))
		return
	}

	return
}

// keyname will be converted into 'Id'-field
// array<CommandInputParameter> | map<CommandInputParameter.id, CommandInputParameter.type> | map<CommandInputParameter.id, CommandInputParameter>
func CreateCommandInputArray(original interface{}, schemata []CWLType_Type) (err error, new_array []*CommandInputParameter) {

	//fmt.Println("CreateCommandInputArray:\n")
	//spew.Dump(original)

	switch original.(type) {

	case map[interface{}]interface{}:
		for k, v := range original.(map[interface{}]interface{}) {

			//var input_parameter CommandInputParameter
			//mapstructure.Decode(v, &input_parameter)
			input_parameter, xerr := NewCommandInputParameter(v, schemata)
			if xerr != nil {
				err = fmt.Errorf("(CreateCommandInputArray) map[interface{}]interface{} NewCommandInputParameter returned: %s", xerr.Error())
				return
			}

			input_parameter.Id = k.(string)

			//fmt.Printf("C")
			new_array = append(new_array, input_parameter)
			//fmt.Printf("D")

		}
	case []interface{}:
		for _, v := range original.([]interface{}) {

			input_parameter, xerr := NewCommandInputParameter(v, schemata)
			if xerr != nil {

				fmt.Println("CommandInputArray:")
				spew.Dump(original)

				err = fmt.Errorf("(CreateCommandInputArray) []interface{} NewCommandInputParameter returned: %s", xerr.Error())
				return
			}
			//var input_parameter CommandInputParameter
			//mapstructure.Decode(v, &input_parameter)

			//empty, err := NewEmpty(v)
			//if err != nil {
			//	return
			//}

			if input_parameter.Id == "" {
				err = fmt.Errorf("(CreateCommandInputArray) no id found")
				return
			}

			//fmt.Printf("C")
			new_array = append(new_array, input_parameter)
			//fmt.Printf("D")

		}
	default:
		err = fmt.Errorf("(CreateCommandInputArray) type unknown")
	}
	//spew.Dump(new_array)
	return
}
