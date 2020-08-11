package testdata

import "github.com/ClareChu/pilot/schedule"

type DemoInterface interface {
	schedule.PipelineInterface
}

type Demo struct {
	
}

func NewDemo() DemoInterface {
	return &Demo{}
}

func invoke()  {
	
}
