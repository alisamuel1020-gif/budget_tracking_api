package main

import (
	"sort"
	"sync"
)

// Summary totals all transactions.
type Summary struct {
	IncomeCents       int64            `json:"incomeCents"`
	ExpenseCents      int64            `json:"expenseCents"`
	BalanceCents      int64            `json:"balanceCents"`
	ExpenseByCategory map[string]int64 `json:"expenseByCategory"`
}

// Store keeps transactions in memory. The mutex makes it safe for concurrent HTTP requests.
type Store struct {
	mu     sync.Mutex
	nextID int
	items  map[int]Transaction
}

func NewStore() *Store {
	return &Store{nextID: 1, items: make(map[int]Transaction)}
}

func (s *Store) Create(in TransactionInput) Transaction {
	s.mu.Lock()
	defer s.mu.Unlock()

	t := Transaction{
		ID:          s.nextID,
		Description: in.Description,
		AmountCents: in.AmountCents,
		Type:        in.Type,
		Category:    in.Category,
		Date:        in.Date,
	}
	s.items[t.ID] = t
	s.nextID++
	return t
}

func (s *Store) Get(id int) (Transaction, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, ok := s.items[id]
	return t, ok
}

func (s *Store) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.items[id]; !ok {
		return false
	}
	delete(s.items, id)
	return true
}

// List returns all transactions, newest date first.
func (s *Store) List() []Transaction {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Transaction, 0, len(s.items))
	for _, t := range s.items {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Date != out[j].Date {
			return out[i].Date > out[j].Date
		}
		return out[i].ID > out[j].ID
	})
	return out
}

func (s *Store) Summary() Summary {
	s.mu.Lock()
	defer s.mu.Unlock()

	sum := Summary{ExpenseByCategory: make(map[string]int64)}
	for _, t := range s.items {
		if t.Type == TypeIncome {
			sum.IncomeCents += t.AmountCents
		} else {
			sum.ExpenseCents += t.AmountCents
			sum.ExpenseByCategory[t.Category] += t.AmountCents
		}
	}
	sum.BalanceCents = sum.IncomeCents - sum.ExpenseCents
	return sum
}
