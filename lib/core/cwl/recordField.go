package cwl

import (
	"fmt"
	"reflect"
)

// http://www.commonwl.org/v1.0/CommandLineTool.html#InputRecordField
type RecordField struct {
	Name         string              `yaml:"name,omitempty" json:"name,omitempty" bson:"name,omitempty"`
	Type         []CWLType_Type      `yaml:"type,omitempty" json:"type,omitempty" bson:"type,omitempty"` // CWLType | InputRecordSchema | InputEnumSchema | InputArraySchema | string | array<CWLType | InputRecordSchema | InputEnumSchema | InputArraySchema | string>
	Doc          string              `yaml:"doc,omitempty" json:"doc,omitempty" bson:"doc,omitempty"`
	InputBinding *CommandLineBinding `yaml:"inputBinding,omitempty" json:"inputBinding,omitempty" bson:"inputBinding,omitempty"`
	Label        string              `yaml:"label,omitempty" json:"label,omitempty" bson:"label,omitempty"`
}

func NewRecordFieldFromInterface(native interface{}, schemata []CWLType_Type) (rf *RecordField, err error) {

	native, err = MakeStringMap(native)
	if err != nil {
		return
	}

	switch native.(type) {
	case map[string]interface{}:
		native_map, ok := native.(map[string]interface{})
		if !ok {
			err = fmt.Errorf("(NewInputRecordFieldFromInterface) type switch error")
			return
		}

		rf = &RecordField{}

		name, has_name := native_map["name"]
		if has_name {
			var ok bool
			rf.Name, ok = name.(string)
			if !ok {
				err = fmt.Errorf("(NewInputRecordFieldFromInterface) type error for name (%s)", reflect.TypeOf(name))
				return
			}
		}

		label, has_label := native_map["label"]
		if has_label {
			var ok bool
			rf.Label, ok = label.(string)
			if !ok {
				err = fmt.Errorf("(NewInputRecordFieldFromInterface) type error for label (%s)", reflect.TypeOf(label))
				return
			}
		}

		doc, has_doc := native_map["doc"]
		if has_doc {
			var ok bool
			rf.Doc, ok = doc.(string)
			if !ok {
				err = fmt.Errorf("(NewInputRecordFieldFromInterface) type error for doc (%s)", reflect.TypeOf(doc))
				return
			}
		}

		the_type, has_type := native_map["type"]
		if has_type {

			rf.Type, err = NewCWLType_TypeArray(the_type, schemata, "Input", false)
			if err != nil {
				err = fmt.Errorf("(NewInputRecordFieldFromInterface) NewCWLTypeArray returned: %s", err.Error())
				return
			}

		}

		inputBinding, has_inputBinding := native_map["inputBinding"]
		if has_inputBinding {

			rf.InputBinding, err = NewCommandLineBinding(inputBinding)
			if err != nil {
				err = fmt.Errorf("(NewInputRecordFieldFromInterface) NewCWLTypeArray returned: %s", err.Error())
				return
			}
		}

		return

	default:
		err = fmt.Errorf("(NewInputRecordFieldFromInterface) unknown type")
	}

	return
}
