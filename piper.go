package piper

type Inlet interface {
	In() chan<- any
}

type Outlet interface {
	Out() <-chan any
}

type Pipe interface {
	Inlet
	Outlet
}

type Source interface {
	Outlet
}

type Sink interface {
	Inlet
}
