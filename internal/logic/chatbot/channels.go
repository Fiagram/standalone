package logic_chatbot

type CreatedWebhookSignal struct {
	OfWebhookId uint64
}

type CreatedWebhookChan chan CreatedWebhookSignal

func NewCreatedWebhookChan() CreatedWebhookChan {
	return make(CreatedWebhookChan, 20)
}
