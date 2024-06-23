package state

import "github.com/wizzomafizzo/tapto/pkg/tokens"

type TokenQueue struct {
	Tokens chan tokens.Token
}

func (q *TokenQueue) Enqueue(t tokens.Token) {
	q.Tokens <- t
}

func (q *TokenQueue) Dequeue() tokens.Token {
	return <-q.Tokens
}

func (q *TokenQueue) Close() {
	close(q.Tokens)
}

func NewTokenQueue() *TokenQueue {
	return &TokenQueue{
		Tokens: make(chan tokens.Token),
	}
}
