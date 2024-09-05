package zero

type middleware struct {
	listBeginHandle  []Handler
	listAfterHandle  []Handler
	listBeforeHandle []Handler
	listEndHandle    []Handler
}

var GolbaleMiddleware middleware

func (m *middleware) Use(handler Handler) {
	m.listBeginHandle = append(m.listBeginHandle, handler)
}

func (m *middleware) Before(handler Handler) {
	m.listBeforeHandle = append(m.listBeforeHandle, handler)
}

func (m *middleware) After(handler Handler) {
	m.listAfterHandle = append(m.listAfterHandle, handler)
}

func (m *middleware) End(handler Handler) {
	m.listEndHandle = append(m.listEndHandle, handler)
}

func (m *middleware) HandleBegin(ctx *Ctx) {
	for _, h := range m.listBeginHandle {
		h(ctx)
	}
}

func (m *middleware) HandleBefore(ctx *Ctx) {
	for _, h := range m.listBeforeHandle {
		h(ctx)
	}
}

func (m *middleware) HandleAfter(ctx *Ctx) {
	for _, h := range m.listAfterHandle {
		h(ctx)
	}
}

func (m *middleware) HandleEnd(ctx *Ctx) {
	for _, h := range m.listEndHandle {
		h(ctx)
	}
}
