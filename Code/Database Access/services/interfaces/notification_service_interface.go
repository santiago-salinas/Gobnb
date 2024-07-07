package interfaces

type INotificationService interface {
	OpenChannel(channel string) error
	SubscribeToChannel(channel string, subscriber string, handler func(message string)) error
	PublishToChannel(channel string, message string) error
	UnsubscribeFromChannel(subscriber string, channel string) error
	MailMethod(email string) func(message string)
	WhatsAppMethod(number string) func(message string)
}
