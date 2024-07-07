package pipes_and_filters

import "mongo-server/mongo_models"

const retryLimit = 3

type SensorReportFilter func(mongo_models.SensorReport) (mongo_models.SensorReport, error)

type SensorReportPipeline struct {
	filters []SensorReportFilter
}

func (p *SensorReportPipeline) Use(f ...SensorReportFilter) {
	p.filters = append(p.filters, f...)
}

func (p *SensorReportPipeline) Run(input mongo_models.SensorReport) error {
	for _, f := range p.filters {
		var err error
		input, err = f(input)
		if err != nil {
			for i := 0; i < retryLimit; i++ {
				input, err = f(input)
				if err == nil {
					break
				}
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}
