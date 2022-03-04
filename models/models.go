package models

//TraceData trace data struct
type TraceData struct {
	File string
	Line int
	Fn   string
}

type Flags struct {
	LogLevelVar, StructureVar, DefaultFolderVar string
	VersionVar                                  bool
}
