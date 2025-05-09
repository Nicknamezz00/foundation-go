package factory

import "sync"

type Provider func(string) interface{}

type Singleton struct {
	cache    map[string]interface{}
	locker   *sync.Mutex
	provider Provider
}

func NewSingleton(provider Provider) Singleton {
	return Singleton{
		cache:    make(map[string]interface{}),
		locker:   &sync.Mutex{},
		provider: provider,
	}
}

func (s *Singleton) Get(name string) interface{} {
	if target, ok := s.cache[name]; ok {
		return target
	}
	s.locker.Lock()
	defer s.locker.Unlock()
	if t, ok := s.cache[name]; ok {
		return t
	}
	s.cache[name] = s.provider(name)
	return s.cache[name]
}
