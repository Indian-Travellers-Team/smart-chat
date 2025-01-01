package store

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

type User struct {
	Name           string
	Mobile         string
	OTP            string
	AuthToken      string
	AuthExpireTime time.Time
}

type MessagePair struct {
	UserMessage string
	BotMessage  string
}

type Conversation struct {
	AccessToken           string
	AccessTokenExpireTime time.Time
	UserName              string
	UserMobile            string
	Messages              []MessagePair
	Pushed                bool
	TotalTokens           int
}

const conversationKeysList = "conversation_keys_list"

var mc *memcache.Client

func init() {
	mc = memcache.New("localhost:11211") // Update with actual Memcached server address

	// Pre-populate a test user and conversation for testing
	/*
		testUser := User{
			Name:           "Test User",
			Mobile:         "1234567890",
			OTP:            "1234",
			AuthToken:      "testtoken123",
			AuthExpireTime: time.Now().Add(24 * time.Hour),
		}
		StoreUser(testUser)

		testConversation := Conversation{
			AccessToken:           "testtoken123",
			AccessTokenExpireTime: time.Now().Add(1 * time.Hour),
			Messages: []MessagePair{
				{UserMessage: "Hi", BotMessage: "Hello! How can I help you?"},
			},
			Pushed: false,
			TotalTokens: 0,
		}
		StoreConversation("testtoken123", testConversation)
	*/
}

func StoreUser(user User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return mc.Set(&memcache.Item{Key: "user_" + user.AuthToken, Value: data})
}

func GetUser(authToken string) (User, error) {
	item, err := mc.Get("user_" + authToken)
	if err != nil {
		return User{}, errors.New("user not found")
	}
	var user User
	err = json.Unmarshal(item.Value, &user)
	return user, err
}

func StoreConversation(accessToken string, conversation Conversation) error {
	data, err := json.Marshal(conversation)
	if err != nil {
		return err
	}
	key := "conversation_" + accessToken
	err = mc.Set(&memcache.Item{Key: key, Value: data})
	if err != nil {
		return err
	}

	// Then, add this key to the keys list
	return addKeyToList(key)
}

func addKeyToList(key string) error {
	// Fetch the existing list
	item, err := mc.Get(conversationKeysList)
	var keys []string
	if err == nil && item != nil {
		err = json.Unmarshal(item.Value, &keys)
		if err != nil {
			return err
		}
	}

	// Add the new key to the list
	keys = append(keys, key)
	data, err := json.Marshal(keys)
	if err != nil {
		return err
	}

	// Store the updated list back in Memcached
	return mc.Set(&memcache.Item{Key: conversationKeysList, Value: data})
}

func GetConversation(accessToken string) (Conversation, error) {
	item, err := mc.Get("conversation_" + accessToken)
	if err != nil {
		return Conversation{}, errors.New("conversation not found")
	}
	var conversation Conversation
	err = json.Unmarshal(item.Value, &conversation)
	return conversation, err
}

func AppendToConversation(accessToken string, messagePair MessagePair, usedTokens int) error {
	conversation, err := GetConversation(accessToken)
	if err != nil {
		return err
	}

	conversation.Messages = append(conversation.Messages, messagePair)
	conversation.TotalTokens += usedTokens

	return StoreConversation(accessToken, conversation)
}

func FetchAllConversations() ([]Conversation, error) {
	// Implement logic to fetch all conversation keys
	// For simplicity, let's assume all keys start with "conversation_"
	keys, err := GetAllConversationsKeys()
	if err != nil {
		return nil, err
	}

	var conversations []Conversation
	for _, key := range keys {
		item, err := mc.Get(key)
		if err != nil {
			continue // Skip if there's an error fetching a particular conversation
		}

		var conversation Conversation
		err = json.Unmarshal(item.Value, &conversation)
		if err != nil {
			continue // Skip if unmarshaling fails
		}

		conversations = append(conversations, conversation)
	}
	return conversations, nil
}

func GetAllConversationsKeys() ([]string, error) {
	item, err := mc.Get(conversationKeysList)
	if err != nil {
		return nil, err
	}

	var keys []string
	err = json.Unmarshal(item.Value, &keys)
	if err != nil {
		return nil, err
	}

	var conversationKeys []string
	for _, key := range keys {
		if strings.HasPrefix(key, "conversation_") {
			conversationKeys = append(conversationKeys, key)
		}
	}

	return conversationKeys, nil
}

func MarkConversationAsPushed(accessToken string) error {
	conversation, err := GetConversation(accessToken)
	if err != nil {
		return err
	}

	conversation.Pushed = true
	return StoreConversation(accessToken, conversation)
}

func DeleteConversation(accessToken string) error {
	// Construct the key for the conversation
	conversationKey := "conversation_" + accessToken

	// Remove the conversation from Memcached
	err := mc.Delete(conversationKey)
	if err != nil {
		return err
	}

	// Remove the key from the conversation keys list
	return removeKeyFromList(conversationKey)
}

func removeKeyFromList(key string) error {
	// Fetch the existing list
	item, err := mc.Get(conversationKeysList)
	if err != nil {
		return err
	}

	var keys []string
	err = json.Unmarshal(item.Value, &keys)
	if err != nil {
		return err
	}

	// Find and remove the key
	for i, k := range keys {
		if k == key {
			keys = append(keys[:i], keys[i+1:]...)
			break
		}
	}

	// Store the updated list back in Memcached
	data, err := json.Marshal(keys)
	if err != nil {
		return err
	}

	return mc.Set(&memcache.Item{Key: conversationKeysList, Value: data})
}
