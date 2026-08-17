package generator

import "fmt"

//go:generate go tool enumer -type=syntheticCLIMapstructureTag -linecomment -text
type syntheticCLIMapstructureTag int

const (
	syntheticCLIMapstructureTagSummary     syntheticCLIMapstructureTag = iota // summary
	syntheticCLIMapstructureTagSummaryFile                                    // summary_file
	syntheticCLIMapstructureTagLogger                                         // logger
	syntheticCLIMapstructureTagLoggerFile                                     // logger_file
	syntheticCLIMapstructureTagAnnotation                                     // annotation
	syntheticCLIMapstructureTagOutput                                         // output_format
	syntheticCLIMapstructureTagOutputFile                                     // output_file
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
	runtimeAnnotationModeNone   runtimeAnnotationMode = iota // none
	runtimeAnnotationModeGithub                              // github
)

type resultOutputMode int

const (
	resultOutputModeHuman resultOutputMode = iota
	resultOutputModeJSON
	resultOutputModeYAML
)

func (mode resultOutputMode) String() string {
	switch mode {
	case resultOutputModeHuman:
		return "human"
	case resultOutputModeJSON:
		return "json"
	case resultOutputModeYAML:
		return "yaml"
	default:
		return fmt.Sprintf("resultOutputMode(%d)", mode)
	}
}
