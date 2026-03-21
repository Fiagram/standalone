package handler

type CreatedWebhookSignal struct {
	OfWebhookId uint64
}

type CreatedWebhookChan chan CreatedWebhookSignal

func NewCreatedWebhookChan() CreatedWebhookChan {
	return make(CreatedWebhookChan, 100)
}
