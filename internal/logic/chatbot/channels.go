package logic_chatbot

type CreatedWebhookSignal struct {
	OfWebhookId uint64
}

type TorchSignal struct {
	OfStrategyID uint64 `json:"of_strategy_id"`
	Symbol       string `json:"symbol"`
	Strategy     string `json:"strategy"`
	Type         string `json:"type"`
}

type CreatedWebhookChan chan CreatedWebhookSignal

func NewCreatedWebhookChan() CreatedWebhookChan {
	return make(CreatedWebhookChan, 20)
}

func NewTorchSignalChan() chan TorchSignal {
	return make(chan TorchSignal, 100)
}
