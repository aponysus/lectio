package store

import (
	"context"
	"time"
)

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

type Session struct {
	ID         string
	UserID     int64
	CreatedAt  time.Time
	ExpiresAt  time.Time
	LastSeenAt time.Time
}

type Source struct {
	ID         int64
	Title      string
	Author     string
	Year       *int
	Tradition  string
	Language   string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	EntryCount int64
}

type Tag struct {
	ID        int64
	Slug      string
	Label     string
	CreatedAt time.Time
	Count     int64
}

type Entry struct {
	ID         int64
	SourceID   int64
	Passage    string
	Reflection string
	Mood       string
	Energy     *int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Thread struct {
	ID          int64
	Title       string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Resonance struct {
	ID              int64
	SourceEntryID   int64
	ResonantEntryID int64
	Score           float64
	FactorTag       float64
	FactorTradition float64
	FactorTemporal  float64
	CreatedAt       time.Time
}

type EntryRepository interface {
	Create(context.Context, Entry) (Entry, error)
	GetByID(context.Context, int64) (Entry, error)
	List(context.Context, EntryListFilter) ([]Entry, error)
	Update(context.Context, Entry) (Entry, error)
	Delete(context.Context, int64) error
}

type EntryListFilter struct {
	SourceID *int64
	Tag      string
	From     *time.Time
	To       *time.Time
	Page     int
	PageSize int
}

type SourceRepository interface {
	Create(context.Context, Source) (Source, error)
	GetByID(context.Context, int64) (Source, error)
	List(context.Context, SourceListFilter) ([]Source, error)
	Update(context.Context, Source) (Source, error)
}

type SourceListFilter struct {
	Query    string
	SortBy   string
	Page     int
	PageSize int
}

type TagRepository interface {
	GetBySlug(context.Context, string) (Tag, error)
	List(context.Context, TagListFilter) ([]Tag, error)
	CoOccurrence(context.Context) ([]TagPair, error)
}

type TagListFilter struct {
	Query    string
	Page     int
	PageSize int
}

type TagPair struct {
	LeftSlug  string
	RightSlug string
	Weight    int64
}

type ThreadRepository interface {
	Create(context.Context, Thread) (Thread, error)
	GetByID(context.Context, int64) (Thread, error)
	List(context.Context) ([]Thread, error)
	Update(context.Context, Thread) (Thread, error)
}

type UserRepository interface {
	EnsureBootstrapUser(context.Context, string, string) (User, error)
	GetByEmail(context.Context, string) (User, error)
	GetFirst(context.Context) (User, error)
}

type SessionRepository interface {
	Create(context.Context, Session) error
	GetByID(context.Context, string) (Session, error)
	Touch(context.Context, string, time.Time) error
	Delete(context.Context, string) error
}

type ResonanceRepository interface {
	ReplaceForEntry(context.Context, int64, []Resonance) error
	ListForEntry(context.Context, int64, int) ([]Resonance, error)
}
