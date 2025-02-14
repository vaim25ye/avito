package cache

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/vaim25ye/avito/internal/model"
	"github.com/vaim25ye/avito/internal/repository"
)

type Cache struct {
	mu    sync.RWMutex
	store map[int]model.UserInfo
}

func NewCache() *Cache {
	return &Cache{
		store: make(map[int]model.UserInfo),
	}
}

// GetUserInfoByID возвращает UserInfo из кэша (или false, если нет).
func (c *Cache) GetUserInfoByID(userID int) (model.UserInfo, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	info, ok := c.store[userID]
	return info, ok
}

// StartCacheUpdater — запускает горутину, которая каждые interval сек
// вызывает updateCache и обновляет внутренний store.
func StartCacheUpdater(ctx context.Context, repo *repository.Repository, c *Cache, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				log.Println("Cache update started...")
				if err := updateCache(ctx, repo, c); err != nil {
					log.Printf("cache update error: %v\n", err)
				} else {
					log.Println("Cache update finished.")
				}

			case <-ctx.Done():
				log.Println("Cache updater stopped.")
				return
			}
		}
	}()
}

// updateCache — грузит всё из БД, записывает в c.store.
func updateCache(ctx context.Context, repo *repository.Repository, c *Cache) error {
	userInfos, err := repo.LoadAllUserData(ctx)
	if err != nil {
		return err
	}

	newStore := make(map[int]model.UserInfo, len(userInfos))
	for _, ui := range userInfos {
		newStore[ui.User.UserID] = ui
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.store = newStore
	return nil
}
