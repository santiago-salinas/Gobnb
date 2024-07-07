package pipes_and_filters

import "mongo-server/mongo_models"

type AppReportFilter func(mongo_models.AppReport) (mongo_models.AppReport, error)

type AppReportPipeline struct {
	filters []AppReportFilter
}

func (p *AppReportPipeline) Use(f ...AppReportFilter) {
	p.filters = append(p.filters, f...)
}

func (p *AppReportPipeline) Run(input mongo_models.AppReport) (err error) {
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