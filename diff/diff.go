package diff

import (
	"fmt"
	"iter"
	"time"
)

type UpdateableObject[K comparable] interface {
	Updateables() iter.Seq[UpdateableList[K, UpdateableElement]]
	NonUpdateables() iter.Seq[NonUpdateablesList[K, any]]
}

type UpdateableElement interface {
	GetUpdatedAt() time.Time
	IsNil() bool
}

type UpdateableList[K comparable, T UpdateableElement] interface {
	Type() string
	Length() int
	Reset()
	List() iter.Seq2[K, T]
	GetElementByKey(id K) (T, bool)
	SetElementByKey(id K, object T)
}

type NonUpdateablesList[K comparable, T any] interface {
	Type() string
	Reset()
	SetElementByKey(id K, object T)
	List() iter.Seq2[K, T]
}

type Updater[K comparable] struct {
	latestUpdatedAt time.Time
}

func (u *Updater[K]) UpdateCache(new, cache UpdateableObject[K]) (time.Time, bool, error) {
	var (
		countOfUpdatable int
		atLeastOneUpdate bool
		latestUpdatedAt  time.Time
	)

	for n := range new.Updateables() {
		if n.Length() != 0 {
			countOfUpdatable++
		}
		var (
			found, updated bool
			c              UpdateableList[K, UpdateableElement]
		)
		for c = range cache.Updateables() {
			if n.Type() == c.Type() {
				found = true
				break
			}
		}
		if !found {
			return time.Time{}, false, fmt.Errorf(`cannot find updateable list of type %q in cache`, n.Type())
		}

		for k, v := range n.List() {
			if v.IsNil() {
				cachedValue, ok := c.GetElementByKey(k)
				if !ok {
					return time.Time{}, false, fmt.Errorf(`value "%v" in %q was not updated but was not present in cache`, k, n.Type())
				}

				if cachedValue.IsNil() {
					return time.Time{}, false, fmt.Errorf(`value "%v" in %q was not updated but was nil in cache`, k, n.Type())
				}

				n.SetElementByKey(k, cachedValue)

				continue
			}

			updated = true // at least one value in "new" was not null, thus it was updated

			if v.GetUpdatedAt().After(latestUpdatedAt) {
				latestUpdatedAt = v.GetUpdatedAt()
			}
		}

		if updated {
			atLeastOneUpdate = true

			c.Reset()
			// we need to iterate over the cache too because we can't simply do "cache = new", since it's an interface
			// it won't persist after the function ends.
			for k, v := range n.List() {
				c.SetElementByKey(k, v)
			}
		}
	}

	// if there are no updatable lists, that means that the new object is not valid or we got an empty response
	if countOfUpdatable == 0 {
		return time.Time{}, false, fmt.Errorf("no updateable lists found in new object")
	}
	if err := u.replaceNonUpdateables(new, cache); err != nil {
		return time.Time{}, false, err
	}

	// only update updatedAt if we managed to handle the response
	// so that we don't miss any updates in case of an error
	if !latestUpdatedAt.IsZero() {
		// There is a case where all workspaces have not been updated since the last request.
		// In that case updatedAt will be zero.
		u.latestUpdatedAt = latestUpdatedAt
	}

	return u.latestUpdatedAt, atLeastOneUpdate, nil
}

func (u *Updater[K]) replaceNonUpdateables(new, cache UpdateableObject[K]) error {
	for n := range new.NonUpdateables() {
		var (
			found bool
			c     NonUpdateablesList[K, any]
		)
		for c = range cache.NonUpdateables() {
			if n.Type() == c.Type() {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf(`cannot find non updateable list of type %q in cache`, n.Type())
		}

		c.Reset()
		for k, v := range n.List() {
			c.SetElementByKey(k, v)
		}
	}

	return nil
}
