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

type nonUpdateablesPair[K comparable] struct {
	new, cache NonUpdateablesList[K, any]
}

// UpdateCache updates the cache with the new object and returns the updatedAt to be used for the next request
// and whether the cache was updated.
// If new is an [IncrementalDefinitionsReceiver] with confirmed incremental definitions, omitted [OmittableList]s
// keep their cached copy and the returned updatedAt is the cursor sent by the control plane.
// Omitted lists stay nil in new, so they must be read from the cache.
func (u *Updater[K]) UpdateCache(new, cache UpdateableObject[K]) (time.Time, bool, error) {
	var (
		countOfUpdatable int
		atLeastOneUpdate bool
		latestUpdatedAt  time.Time
		incremental      bool
		cursor           time.Time
	)
	if r, ok := new.(IncrementalDefinitionsReceiver); ok {
		incremental, cursor = r.IncrementalDefinitions()
	}

	nonUpdateables, err := nonUpdateablesToReplace(new, cache, incremental)
	if err != nil {
		return time.Time{}, false, err
	}

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
					return time.Time{}, false, fmt.Errorf(`value "%v" in %q was not updated but was not present in cache: %w`, k, n.Type(), ErrFullFetchRequired)
				}

				if cachedValue.IsNil() {
					return time.Time{}, false, fmt.Errorf(`value "%v" in %q was not updated but was nil in cache: %w`, k, n.Type(), ErrFullFetchRequired)
				}

				n.SetElementByKey(k, cachedValue)

				continue
			}

			updated = true // at least one value in "new" was not null, thus it was updated

			if v.GetUpdatedAt().After(latestUpdatedAt) {
				latestUpdatedAt = v.GetUpdatedAt()
			}
		}

		// The cache needs to be refreshed both when at least one element was updated and when elements were
		// removed. Removals are only signaled by the absence of the corresponding keys in the new list, so
		// comparing the lengths is enough to detect them: an added key always comes with a non-nil value, i.e.
		// it always sets updated to true as well.
		if updated || n.Length() != c.Length() {
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

	for _, l := range nonUpdateables {
		l.cache.Reset()
		for k, v := range l.new.List() {
			l.cache.SetElementByKey(k, v)
		}
	}

	if incremental {
		// non updateables are only sent when changed
		atLeastOneUpdate = atLeastOneUpdate || len(nonUpdateables) > 0
		// the cursor covers non updateables too, so it might be earlier than the latest updateable element
		u.latestUpdatedAt = cursor
	} else if !latestUpdatedAt.IsZero() {
		// only update updatedAt if we managed to handle the response
		// so that we don't miss any updates in case of an error.
		// There is a case where all workspaces have not been updated since the last request.
		// In that case updatedAt will be zero.
		u.latestUpdatedAt = latestUpdatedAt
	}

	return u.latestUpdatedAt, atLeastOneUpdate, nil
}

// nonUpdateablesToReplace pairs the non updateable lists of new with the ones in cache, leaving out the ones
// omitted from an incremental response. It runs before any change to the cache so that errors leave it untouched.
func nonUpdateablesToReplace[K comparable](new, cache UpdateableObject[K], incremental bool) ([]nonUpdateablesPair[K], error) {
	var pairs []nonUpdateablesPair[K]
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
			return nil, fmt.Errorf(`cannot find non updateable list of type %q in cache`, n.Type())
		}

		if incremental {
			o, ok := n.(OmittableList)
			if !ok {
				return nil, fmt.Errorf(`non updateable list of type %q does not implement OmittableList`, n.Type())
			}
			if o.Omitted() {
				if co, ok := c.(OmittableList); ok && co.Omitted() {
					return nil, fmt.Errorf(`non updateable list of type %q was omitted but was nil in cache: %w`, n.Type(), ErrFullFetchRequired)
				}
				continue
			}
		}

		pairs = append(pairs, nonUpdateablesPair[K]{new: n, cache: c})
	}
	return pairs, nil
}
