package services

import (
	"fmt"
	"pocketbase_go/logger"
	pubsub "pocketbase_go/publish-subscribe"

	"github.com/go-redis/redis/v8"
)

type NotificationService struct {
	Publishers    *pubsub.Publisher
	Channels      map[string]*pubsub.RedisSubscriptionChannel
	CancelMethods map[string]func()
	RedisClient   *redis.Client
}

func NewNotificationService(redisClient *redis.Client) *NotificationService {
	channels := make(map[string]*pubsub.RedisSubscriptionChannel)
	cancelMethods := make(map[string]func())

	return &NotificationService{
		Publishers:    pubsub.NewRedisPublisher(redisClient),
		RedisClient:   redisClient,
		Channels:      channels,
		CancelMethods: cancelMethods,
	}
}

func (n *NotificationService) OpenChannel(channel string) error {
	logger.Info("Service: Opening channel")
	if n.RedisClient == nil {
		logger.Error("Service: Redis client not initialized")
		return fmt.Errorf("Redis client not initialized")
	}

	if _, ok := n.Channels[channel]; ok {
		logger.Error("Service: ", channel, " already exists")
		return fmt.Errorf("Channel already exists")
	}
	newChannel := pubsub.NewRedisSubsciptionChannel(n.RedisClient, channel)
	n.Channels[channel] = newChannel
	return nil
}

func (n *NotificationService) SubscribeToChannel(channel string, subscriber string, handler func(message string)) error {
	if _, ok := n.Channels[channel]; !ok {
		logger.Error("Service: ", channel, " does not exist")
		return fmt.Errorf("Channel does not exist")
	}
	key := subscriberChannelKey(subscriber, channel)
	if _, ok := n.CancelMethods[key]; ok {
		logger.Error("Service: ", subscriber, " is already subscribed to ", channel)
		return fmt.Errorf("Subscriber is already subscribed to channel")
	}
	cancel := n.Channels[channel].Subscribe(handler)

	n.CancelMethods[key] = cancel
	return nil
}

func (n *NotificationService) PublishToChannel(channel string, message string) error {
	if _, ok := n.Channels[channel]; !ok {
		logger.Error("Service: ", channel, " does not exist")
		return fmt.Errorf("Channel does not exist")
	}
	n.Publishers.Publish(channel, message)
	return nil
}

func (n *NotificationService) UnsubscribeFromChannel(subscriber string, channel string) error {
	if _, ok := n.Channels[channel]; !ok {
		logger.Error("Service: ", channel, " does not exist")
		return fmt.Errorf("Channel does not exist")
	}
	key := subscriberChannelKey(subscriber, channel)
	n.CancelMethods[key]()
	n.CancelMethods[key] = nil
	return nil
}

func (n *NotificationService) MailMethod(email string) func(message string) {
	logger.Info("Notifications Service subscribed to email notifications for ", email)
	handler := func(message string) {
		logger.Info("Notifications Service - Mail sent to ", email, ":", message)
	}

	return handler
}

func (n *NotificationService) WhatsAppMethod(number string) func(message string) {
	logger.Info("Notifications Service subscribed to WhatsApp notifications for ", number)
	handler := func(message string) {
		logger.Info("Notifications Service - WhatsApp sent to ", number, ":", message)
	}

	return handler
}

func subscriberChannelKey(subscriber string, channel string) string {
	return subscriber + ":" + channel
}
