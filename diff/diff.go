package diff

import (
	"fmt"
	"iter"
	"time"
)

type UpdateableElement interface {
	GetUpdatedAt() time.Time
	IsNil() bool
}

type UpdateableList[K comparable, T UpdateableElement] interface {
	List() iter.Seq2[K, T]
	GetElementByKey(id K) (T, bool)
	SetElementByKey(id K, object T)
	Reset()
}

type Updater[K comparable, T UpdateableElement] struct {
	latestUpdatedAt time.Time
}

func (u *Updater[K, T]) UpdateCache(new, cache UpdateableList[K, T]) (time.Time, error) {
	var latestUpdatedAt time.Time

	for k, v := range new.List() { // this value was not updated, populate it with the previous config
		if v.IsNil() {
			cachedValue, ok := cache.GetElementByKey(k)
			if !ok {
				return time.Time{}, fmt.Errorf(`value "%v" was not updated but was not present in cache`, k)
			}

			if cachedValue.IsNil() {
				return time.Time{}, fmt.Errorf(`value "%v" was not updated but was nil in cache`, k)
			}

			new.SetElementByKey(k, cachedValue)

			continue
		}

		if v.GetUpdatedAt().After(latestUpdatedAt) {
			latestUpdatedAt = v.GetUpdatedAt()
		}
	}

	// only update updatedAt if we managed to handle the response
	// so that we don't miss any updates in case of an error
	if !latestUpdatedAt.IsZero() {
		// There is a case where all workspaces have not been updated since the last request.
		// In that case updatedAt will be zero.
		u.latestUpdatedAt = latestUpdatedAt
	}

	// we need to iterate over the cache too because we can't simply do "cache = new", since it's an interface
	// it won't persist after the function ends.
	cache.Reset()
	for k, v := range new.List() {
		cache.SetElementByKey(k, v)
	}

	return u.latestUpdatedAt, nil
}
