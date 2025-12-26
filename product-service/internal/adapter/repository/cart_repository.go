package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"product-service/internal/core/domain/entities"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/gommon/log"
)

type ICartRepository interface {
	AddToCart(ctx context.Context, userID string, items []entities.CartItem) error
	GetCart(ctx context.Context, userID string) ([]entities.CartItem, error)
	RemoveFromCart(ctx context.Context, userID int64, productID int64) error
	RemoveAllCart(ctx context.Context, userID int64) error
}

type CartRepository struct {
	Client *redis.Client
}

// RemoveAllCart implements [ICartRepository].
func (c *CartRepository) RemoveAllCart(ctx context.Context, userID int64) error {
	return c.Client.Del(ctx, fmt.Sprintf("cart:cart:%d", userID)).Err()
}

// RemoveFromCart implements [ICartRepository].
func (c *CartRepository) RemoveFromCart(ctx context.Context, userID int64, productID int64) error {
	cart, err := c.GetCart(ctx, fmt.Sprintf("cart:%d", userID))
	if err != nil {
		log.Errorf("[CartRedisRepository-1] RemoveFromCart: %v", err)
		return err
	}

	newCart := []entities.CartItem{}
	for _, item := range cart {
		if item.ProductID != productID {
			newCart = append(newCart, item)
		}
	}

	err = c.Client.Del(ctx, fmt.Sprintf("cart:cart:%d", userID)).Err()
	if err != nil {
		log.Errorf("[CartRedisRepository-2] RemoveFromCart: %v", err)
		return err
	}

	return c.AddToCart(ctx, fmt.Sprintf("cart:%d", userID), newCart)
}

// GetCart implements [ICartRepository].
func (c *CartRepository) GetCart(ctx context.Context, userID string) ([]entities.CartItem, error) {
	val, err := c.Client.Get(ctx, fmt.Sprintf("cart:%s", userID)).Result()

	if err == redis.Nil {
		log.Infof("[CartRedisRepository-1] GetCart: Cart not found")
		return nil, nil
	}

	if err != nil {
		log.Errorf("[CartRedisRepository-2] GetCart: %v", err)
		return nil, err
	}

	var items []entities.CartItem
	err = json.Unmarshal([]byte(val), &items)

	if err != nil {
		log.Errorf("[CartRedisRepository-3] GetCart: %v", err)
		return nil, err
	}

	return items, nil
}

// AddToCart implements [ICartRepository].
func (c *CartRepository) AddToCart(ctx context.Context, userID string, items []entities.CartItem) error {
	data, err := json.Marshal(items)
	if err != nil {
		log.Errorf("[CartRedisRepository-1] AddToCart: %v", err)
		return err
	}
	return c.Client.Set(ctx, fmt.Sprintf("cart:%s", userID), data, 0).Err()
}

func NewCartRepository(client *redis.Client) ICartRepository {
	return &CartRepository{
		Client: client,
	}
}
