package cwl

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"reflect"
)

// http://www.commonwl.org/v1.0/CommandLineTool.html#CommandInputArraySchema
type CommandInputArraySchema struct { // Items, Type , Label
	ArraySchema  `yaml:",inline" json:",inline" bson:",inline" mapstructure:",squash"`
	InputBinding *CommandLineBinding `yaml:"inputBinding,omitempty" bson:"inputBinding,omitempty" json:"inputBinding,omitempty"`
}

//func (c *CommandOutputArraySchema) Is_CommandOutputParameterType() {}

func (c *CommandInputArraySchema) Type2String() string { return "CommandInputArraySchema" }

func NewCommandInputArraySchema() (coas *CommandInputArraySchema) {

	coas = &CommandInputArraySchema{}
	coas.Type = "array"

	return
}

func NewCommandInputArraySchemaFromInterface(original interface{}) (coas *CommandInputArraySchema, err error) {

	original, err = MakeStringMap(original)
	if err != nil {
		return
	}

	coas = NewCommandInputArraySchema()

	switch original.(type) {

	case map[string]interface{}:
		original_map, ok := original.(map[string]interface{})
		if !ok {
			err = fmt.Errorf("(NewCommandInputArraySchemaFromInterface) type error b")
			return
		}

		items, ok := original_map["items"]
		if ok {
			var items_type []CWLType_Type
			items_type, err = NewCWLType_TypeArray(items, "CommandInput")
			if err != nil {
				err = fmt.Errorf("(NewCommandInputArraySchemaFromInterface) NewCWLType_TypeArray returns: %s", err.Error())
				return
			}
			original_map["items"] = items_type

		}

		err = mapstructure.Decode(original, coas)
		if err != nil {
			err = fmt.Errorf("(NewCommandInputArraySchemaFromInterface) %s", err.Error())
			return
		}
	default:
		err = fmt.Errorf("NewCommandInputArraySchemaFromInterface, unknown type %s", reflect.TypeOf(original))
	}
	return
}
