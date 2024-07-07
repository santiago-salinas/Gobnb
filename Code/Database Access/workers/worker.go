package workers

type Worker interface {
	Use(...WorkerMiddlewareFunc)
	Health() error
	Config(address string) error
	Listen(quantity int, queueName string, processBody func([]byte) error) (Channel, error)
	Close() error
	Send(queue string, messages ...[]byte) error
}

type WorkerMiddlewareFunc func(name string) WorkerMiddleware

type WorkerMiddleware interface {
	Start()
	Stop(error)
}

type Channel interface {
	Close() error
}

func BuildRabbitWorker(address string) (worker Worker, err error) {
	worker = new(rabbitWorker)
	err = worker.Config(address)
	return
}
