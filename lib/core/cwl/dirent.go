package cwl

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

//"github.com/davecgh/go-spew/spew"

// Dirent http://www.commonwl.org/v1.0/CommandLineTool.html#Dirent
type Dirent struct {
	CWLType_Impl `yaml:",inline" json:",inline" bson:",inline" mapstructure:",squash"`
	Entry        Expression `yaml:"entry" json:"entry" bson:"entry" mapstructure:"entry"`
	Entryname    Expression `yaml:"entryname" json:"entryname" bson:"entryname" mapstructure:"entryname"`
	Writable     bool       `yaml:"writable" json:"writable" bson:"writable" mapstructure:"writable"`
}

func NewDirentFromInterface(id string, original interface{}) (dirent *Dirent, err error) {

	dirent = &Dirent{}
	err = mapstructure.Decode(original, dirent)
	if err != nil {
		err = fmt.Errorf("(NewDirentFromInterface) Could not convert Dirent object: %s", err.Error())
		return
	}

	return
}
