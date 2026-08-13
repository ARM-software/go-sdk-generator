package generator

//go:generate go tool enumer -type=syntheticCLIMapstructureTag -linecomment -text
type syntheticCLIMapstructureTag int

const (
	syntheticCLIMapstructureTagSummary     syntheticCLIMapstructureTag = iota // summary
	syntheticCLIMapstructureTagSummaryFile                                    // summary_file
	syntheticCLIMapstructureTagLogger                                         // logger
	syntheticCLIMapstructureTagLoggerFile                                     // logger_file
	syntheticCLIMapstructureTagAnnotation                                     // annotation
)

//go:generate go tool enumer -type=runtimeLoggerMode -linecomment -text
type runtimeLoggerMode int

const (
	runtimeLoggerModeConsole runtimeLoggerMode = iota // console
	runtimeLoggerModeRich                             // rich
	runtimeLoggerModeFile                             // file
	runtimeLoggerModeNone                             // none
	runtimeLoggerModeCustom                           // custom
)

//go:generate go tool enumer -type=runtimeAnnotationMode -linecomment -text
type runtimeAnnotationMode int

const (
	runtimeAnnotationModeGithub runtimeAnnotationMode = iota // github
	runtimeAnnotationModeNone                                // none
)
