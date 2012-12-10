package errors

import (
	"regexp"
)

var (
	MongoDupKeyRegex = regexp.MustCompile("duplicate\\s+key")
)

const (
	ClientNotFound           = "Client not found"
	InvalidIndex             = "Invalid Index"
	InvalidFileTypeForFilter = "Invalid file type for filter"
	MongoDocNotFound         = "Document not found"
	NoAuth                   = "No Authorization"
	WorkUnitQueueEmpty       = "Workunit queue is empty"
	UnAuth                   = "User Unauthorized"
)