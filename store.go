package tmpstore

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

type Store struct {
	store       map[string]time.Time
	dir         string
	maxFileSize int
	duration    time.Duration
	lifetime    time.Duration
	logger      *log.Logger
}

func New() *Store {
	return &Store{
		store:       make(map[string]time.Time),
		dir:         "",
		maxFileSize: 1 << 20,
		duration:    time.Hour,
		lifetime:    time.Hour,
		logger:      &log.Logger{},
	}
}

func (s *Store) Run() {
	ticker := time.NewTicker(s.duration)
	for {
		<-ticker.C
		for name, t := range s.store {
			if time.Now().Sub(t) > s.lifetime {
				if err := s.Remove(name); err != nil {
					s.logger.Println(err)
				}
			}
		}
	}
}

func (s *Store) SetMaxFileSize(size int) {
	s.maxFileSize = size
}

func (s *Store) SetDuration(duration time.Duration) {
	s.duration = duration
}

func (s *Store) SetLifetime(lifetime time.Duration) {
	s.lifetime = lifetime
}

func (s *Store) SetDirectory(dir string) {
	if !strings.HasSuffix(dir, "/") {
		dir = dir + "/"
	}

	if err := os.MkdirAll(dir, 0660); err != nil {
		log.Println(err)
	}

	s.dir = dir
}

func (s *Store) Dir() string {
	return s.dir
}

func (s *Store) Store(name string, data []byte) error {
	if _, ok := s.store[name]; ok {
		return fmt.Errorf("data already exists %v", name)
	}

	if len(data) > s.maxFileSize {
		return fmt.Errorf("data is too big %v", len(data))
	}

	f, err := os.Create(s.dir + name)
	if err != nil {
		return err
	}

	defer func() {
		if err := f.Close(); err != nil {
			s.logger.Print(err)
		}
	}()

	if _, err := f.Write(data); err != nil {
		return err
	}

	s.store[name] = time.Now()

	return nil
}

func (s *Store) UpdateTime(name string) error {
	if _, ok := s.store[name]; !ok {
		return fmt.Errorf("no data exists %v", name)
	}

	s.store[name] = time.Now()
	return nil
}

func (s *Store) Data(name string) ([]byte, error) {
	if _, ok := s.store[name]; !ok {
		return nil, fmt.Errorf("no data exists %v", name)
	}

	data, err := ioutil.ReadFile(s.dir + name)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *Store) Remove(name string) error {
	if _, ok := s.store[name]; !ok {
		return fmt.Errorf("no data exists %v", name)
	}

	delete(s.store, name)

	if err := os.Remove(s.dir + name); err != nil {
		return err
	}

	return nil
}

func (s *Store) Clear() {
	for name := range s.store {
		if err := s.Remove(name); err != nil {
			s.logger.Println(err)
		}
	}
}
